package exchange

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

// PgRepo — pgx-реализация Repo.
type PgRepo struct {
	db repo.Exec
}

func NewPgRepo(db repo.Exec) *PgRepo { return &PgRepo{db: db} }

// encodeCursor / decodeCursor — opaque cursor для list-pagination.
// Формат: base64("<created_at_unix_nanos>:<id>"). Сортировка лотов:
// ORDER BY created_at DESC, id DESC.
func encodeCursor(t time.Time, id string) string {
	raw := strconv.FormatInt(t.UnixNano(), 10) + ":" + id
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(c string) (time.Time, string, error) {
	if c == "" {
		return time.Time{}, "", nil
	}
	b, err := base64.RawURLEncoding.DecodeString(c)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("decode cursor: %w", err)
	}
	parts := strings.SplitN(string(b), ":", 2)
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid cursor format")
	}
	ns, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor timestamp: %w", err)
	}
	return time.Unix(0, ns).UTC(), parts[1], nil
}

func (r *PgRepo) ListLots(ctx context.Context, f ListFilters) ([]Lot, string, error) {
	if f.Limit <= 0 || f.Limit > 100 {
		f.Limit = 50
	}
	status := "active"
	if f.Status != nil && *f.Status != "" {
		status = *f.Status
	}
	args := []any{status}
	conds := []string{"l.status = $1"}
	idx := 2
	if f.ArtifactUnitID != nil {
		conds = append(conds, fmt.Sprintf("l.artifact_unit_id = $%d", idx))
		args = append(args, *f.ArtifactUnitID)
		idx++
	}
	if f.MinPrice != nil {
		conds = append(conds, fmt.Sprintf("l.price_oxsarit >= $%d", idx))
		args = append(args, *f.MinPrice)
		idx++
	}
	if f.MaxPrice != nil {
		conds = append(conds, fmt.Sprintf("l.price_oxsarit <= $%d", idx))
		args = append(args, *f.MaxPrice)
		idx++
	}
	if f.SellerID != nil && *f.SellerID != "" {
		conds = append(conds, fmt.Sprintf("l.seller_user_id = $%d", idx))
		args = append(args, *f.SellerID)
		idx++
	}
	if f.Cursor != "" {
		ct, cid, err := decodeCursor(f.Cursor)
		if err != nil {
			return nil, "", err
		}
		conds = append(conds, fmt.Sprintf("(l.created_at, l.id) < ($%d, $%d)", idx, idx+1))
		args = append(args, ct, cid)
		idx += 2
	}
	args = append(args, f.Limit+1) // +1 для определения next_cursor
	q := fmt.Sprintf(`
		SELECT l.id, l.seller_user_id, COALESCE(u.username, ''),
		       l.artifact_unit_id, l.quantity, l.price_oxsarit,
		       l.status, l.created_at, l.expires_at,
		       l.buyer_user_id, l.sold_at, l.expire_event_id
		FROM exchange_lots l
		LEFT JOIN users u ON u.id = l.seller_user_id
		WHERE %s
		ORDER BY l.created_at DESC, l.id DESC
		LIMIT $%d
	`, strings.Join(conds, " AND "), idx)
	rows, err := r.db.Pool().Query(ctx, q, args...)
	if err != nil {
		return nil, "", fmt.Errorf("list lots query: %w", err)
	}
	defer rows.Close()
	lots := make([]Lot, 0, f.Limit)
	for rows.Next() {
		var l Lot
		if err := rows.Scan(&l.ID, &l.SellerUserID, &l.SellerUsername,
			&l.ArtifactUnitID, &l.Quantity, &l.PriceOxsarit,
			&l.Status, &l.CreatedAt, &l.ExpiresAt,
			&l.BuyerUserID, &l.SoldAt, &l.ExpireEventID); err != nil {
			return nil, "", err
		}
		l.UnitPriceOxsarit = l.PriceOxsarit / int64(l.Quantity)
		lots = append(lots, l)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	var nextCursor string
	if len(lots) > f.Limit {
		nextCursor = encodeCursor(lots[f.Limit-1].CreatedAt, lots[f.Limit-1].ID)
		lots = lots[:f.Limit]
	}
	return lots, nextCursor, nil
}

func (r *PgRepo) GetLot(ctx context.Context, id string) (Lot, error) {
	var l Lot
	err := r.db.Pool().QueryRow(ctx, `
		SELECT l.id, l.seller_user_id, COALESCE(u.username, ''),
		       l.artifact_unit_id, l.quantity, l.price_oxsarit,
		       l.status, l.created_at, l.expires_at,
		       l.buyer_user_id, l.sold_at, l.expire_event_id
		FROM exchange_lots l
		LEFT JOIN users u ON u.id = l.seller_user_id
		WHERE l.id = $1
	`, id).Scan(&l.ID, &l.SellerUserID, &l.SellerUsername,
		&l.ArtifactUnitID, &l.Quantity, &l.PriceOxsarit,
		&l.Status, &l.CreatedAt, &l.ExpiresAt,
		&l.BuyerUserID, &l.SoldAt, &l.ExpireEventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Lot{}, ErrLotNotFound
		}
		return Lot{}, fmt.Errorf("get lot: %w", err)
	}
	l.UnitPriceOxsarit = l.PriceOxsarit / int64(l.Quantity)
	return l, nil
}

func (r *PgRepo) GetLotItems(ctx context.Context, lotID string) ([]string, error) {
	rows, err := r.db.Pool().Query(ctx,
		`SELECT artefact_id FROM exchange_lot_items WHERE lot_id = $1 ORDER BY artefact_id`,
		lotID)
	if err != nil {
		return nil, fmt.Errorf("get lot items: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *PgRepo) CountActiveLotsBySeller(ctx context.Context, tx pgx.Tx, sellerID string) (int, error) {
	var n int
	err := tx.QueryRow(ctx,
		`SELECT COUNT(*) FROM exchange_lots
		 WHERE seller_user_id = $1 AND status = 'active'`,
		sellerID).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count active lots: %w", err)
	}
	return n, nil
}

func (r *PgRepo) AvgUnitPrice(ctx context.Context, tx pgx.Tx,
	artifactUnitID int, window time.Duration) (*int64, error) {
	// Берём AVG(price/quantity) только по successfully bought-историям
	// за окно. Использует индекс ix_exchange_history_pricing.
	var avg *float64
	err := tx.QueryRow(ctx, `
		SELECT AVG(l.price_oxsarit::float / l.quantity)
		FROM exchange_history h
		JOIN exchange_lots l ON l.id = h.lot_id
		WHERE h.event_kind = 'bought'
		  AND l.artifact_unit_id = $1
		  AND h.created_at > now() - $2::interval
	`, artifactUnitID, window.String()).Scan(&avg)
	if err != nil {
		return nil, fmt.Errorf("avg unit price: %w", err)
	}
	if avg == nil {
		return nil, nil
	}
	v := int64(*avg)
	return &v, nil
}

func (r *PgRepo) SelectAvailableArtefacts(ctx context.Context, tx pgx.Tx,
	sellerID string, artifactUnitID int, n int) ([]string, error) {
	// Артефакт доступен для escrow если:
	//  - принадлежит seller'у;
	//  - state='held' (не activated/listed/expired/consumed);
	//  - НЕ находится в active-лоте exchange_lot_items (защита от повторного
	//    listing'а если кто-то изменил state вручную);
	//  - НЕ выставлен в artefact_offers (другая фича — single-item market).
	rows, err := tx.Query(ctx, `
		SELECT au.id
		FROM artefacts_user au
		WHERE au.user_id = $1
		  AND au.unit_id = $2
		  AND au.state = 'held'
		  AND NOT EXISTS (
		      SELECT 1 FROM exchange_lot_items eli
		      JOIN exchange_lots el ON el.id = eli.lot_id
		      WHERE eli.artefact_id = au.id AND el.status = 'active'
		  )
		  AND NOT EXISTS (
		      SELECT 1 FROM artefact_offers ao WHERE ao.artefact_id = au.id
		  )
		ORDER BY au.id
		LIMIT $3
		FOR UPDATE OF au
	`, sellerID, artifactUnitID, n)
	if err != nil {
		return nil, fmt.Errorf("select available artefacts: %w", err)
	}
	defer rows.Close()
	out := make([]string, 0, n)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (r *PgRepo) MarkArtefactsListed(ctx context.Context, tx pgx.Tx, artefactIDs []string) error {
	if len(artefactIDs) == 0 {
		return nil
	}
	_, err := tx.Exec(ctx,
		`UPDATE artefacts_user SET state = 'listed' WHERE id = ANY($1)`,
		artefactIDs)
	if err != nil {
		return fmt.Errorf("mark artefacts listed: %w", err)
	}
	return nil
}

func (r *PgRepo) MarkArtefactsHeld(ctx context.Context, tx pgx.Tx,
	artefactIDs []string, newOwnerID, newPlanetID string) error {
	if len(artefactIDs) == 0 {
		return nil
	}
	if newOwnerID == "" {
		// cancel/expire — owner не меняется, только state.
		_, err := tx.Exec(ctx,
			`UPDATE artefacts_user SET state = 'held' WHERE id = ANY($1)`,
			artefactIDs)
		if err != nil {
			return fmt.Errorf("mark artefacts held: %w", err)
		}
		return nil
	}
	_, err := tx.Exec(ctx, `
		UPDATE artefacts_user
		SET state = 'held',
		    user_id = $2,
		    planet_id = $3
		WHERE id = ANY($1)
	`, artefactIDs, newOwnerID, newPlanetID)
	if err != nil {
		return fmt.Errorf("transfer artefacts to buyer: %w", err)
	}
	return nil
}

func (r *PgRepo) InsertLot(ctx context.Context, tx pgx.Tx, l Lot) (Lot, error) {
	if l.ID == "" {
		l.ID = ids.New()
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now().UTC()
	}
	l.Status = "active"
	_, err := tx.Exec(ctx, `
		INSERT INTO exchange_lots
			(id, seller_user_id, artifact_unit_id, quantity, price_oxsarit,
			 created_at, expires_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'active')
	`, l.ID, l.SellerUserID, l.ArtifactUnitID, l.Quantity, l.PriceOxsarit,
		l.CreatedAt, l.ExpiresAt)
	if err != nil {
		return Lot{}, fmt.Errorf("insert lot: %w", err)
	}
	l.UnitPriceOxsarit = l.PriceOxsarit / int64(l.Quantity)
	return l, nil
}

func (r *PgRepo) InsertLotItems(ctx context.Context, tx pgx.Tx, lotID string, artefactIDs []string) error {
	if len(artefactIDs) == 0 {
		return errors.New("insert lot items: empty artefactIDs")
	}
	rows := make([][]any, 0, len(artefactIDs))
	for _, id := range artefactIDs {
		rows = append(rows, []any{lotID, id})
	}
	_, err := tx.CopyFrom(ctx,
		pgx.Identifier{"exchange_lot_items"},
		[]string{"lot_id", "artefact_id"},
		pgx.CopyFromRows(rows))
	if err != nil {
		return fmt.Errorf("copy lot items: %w", err)
	}
	return nil
}

func (r *PgRepo) SetLotExpireEvent(ctx context.Context, tx pgx.Tx, lotID, eventID string) error {
	_, err := tx.Exec(ctx,
		`UPDATE exchange_lots SET expire_event_id = $1 WHERE id = $2`,
		eventID, lotID)
	if err != nil {
		return fmt.Errorf("set lot expire event: %w", err)
	}
	return nil
}

func (r *PgRepo) LockLotForUpdate(ctx context.Context, tx pgx.Tx, id string) (Lot, error) {
	var l Lot
	err := tx.QueryRow(ctx, `
		SELECT id, seller_user_id, artifact_unit_id, quantity, price_oxsarit,
		       status, created_at, expires_at,
		       buyer_user_id, sold_at, expire_event_id
		FROM exchange_lots WHERE id = $1 FOR UPDATE
	`, id).Scan(&l.ID, &l.SellerUserID, &l.ArtifactUnitID, &l.Quantity,
		&l.PriceOxsarit, &l.Status, &l.CreatedAt, &l.ExpiresAt,
		&l.BuyerUserID, &l.SoldAt, &l.ExpireEventID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Lot{}, ErrLotNotFound
		}
		return Lot{}, fmt.Errorf("lock lot: %w", err)
	}
	l.UnitPriceOxsarit = l.PriceOxsarit / int64(l.Quantity)
	return l, nil
}

func (r *PgRepo) MarkLotSold(ctx context.Context, tx pgx.Tx, lotID, buyerID string, soldAt time.Time) error {
	_, err := tx.Exec(ctx, `
		UPDATE exchange_lots
		SET status = 'sold', buyer_user_id = $2, sold_at = $3
		WHERE id = $1 AND status = 'active'
	`, lotID, buyerID, soldAt)
	if err != nil {
		return fmt.Errorf("mark lot sold: %w", err)
	}
	return nil
}

func (r *PgRepo) MarkLotCancelled(ctx context.Context, tx pgx.Tx, lotID string) error {
	_, err := tx.Exec(ctx,
		`UPDATE exchange_lots SET status = 'cancelled' WHERE id = $1 AND status = 'active'`,
		lotID)
	if err != nil {
		return fmt.Errorf("mark lot cancelled: %w", err)
	}
	return nil
}

func (r *PgRepo) MarkLotExpired(ctx context.Context, tx pgx.Tx, lotID string) error {
	_, err := tx.Exec(ctx,
		`UPDATE exchange_lots SET status = 'expired' WHERE id = $1 AND status = 'active'`,
		lotID)
	if err != nil {
		return fmt.Errorf("mark lot expired: %w", err)
	}
	return nil
}

func (r *PgRepo) CancelExpireEvent(ctx context.Context, tx pgx.Tx, eventID, reason string) error {
	if eventID == "" {
		return nil // лот мог быть создан до expire-event INSERT'а — no-op.
	}
	// Помечаем 'ok' с last_error=reason, чтобы worker не подобрал заново.
	// state enum ограничен wait/start/ok/error (см. 0001_init.sql); 'ok'
	// + last_error=reason — стандартный паттерн «отменили без обработки».
	_, err := tx.Exec(ctx, `
		UPDATE events
		SET state = 'ok', processed_at = now(), last_error = $2
		WHERE id = $1 AND state = 'wait'
	`, eventID, reason)
	if err != nil {
		return fmt.Errorf("cancel expire event: %w", err)
	}
	return nil
}

func (r *PgRepo) InsertHistory(ctx context.Context, tx pgx.Tx,
	lotID, eventKind string, actorUserID *string, payload []byte) error {
	if len(payload) == 0 {
		payload = []byte("{}")
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO exchange_history (id, lot_id, event_kind, actor_user_id, payload)
		VALUES ($1, $2, $3, $4, $5)
	`, ids.New(), lotID, eventKind, actorUserID, payload)
	if err != nil {
		return fmt.Errorf("insert history: %w", err)
	}
	return nil
}

func (r *PgRepo) SelectHomePlanet(ctx context.Context, tx pgx.Tx, userID string) (string, error) {
	var planetID string
	err := tx.QueryRow(ctx, `
		SELECT id FROM planets
		WHERE user_id = $1 AND destroyed_at IS NULL AND is_moon = false
		ORDER BY created_at ASC
		LIMIT 1
	`, userID).Scan(&planetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserHasNoPlanet
		}
		return "", fmt.Errorf("select home planet: %w", err)
	}
	return planetID, nil
}

func (r *PgRepo) SpendOxsarits(ctx context.Context, tx pgx.Tx, userID string, amount int64) error {
	tag, err := tx.Exec(ctx, `
		UPDATE users SET credit = credit - $2
		WHERE id = $1 AND credit >= $2
	`, userID, amount)
	if err != nil {
		return fmt.Errorf("spend oxsarits: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrInsufficientOxsarits
	}
	return nil
}

func (r *PgRepo) AddOxsarits(ctx context.Context, tx pgx.Tx, userID string, amount int64) error {
	_, err := tx.Exec(ctx,
		`UPDATE users SET credit = credit + $2 WHERE id = $1`,
		userID, amount)
	if err != nil {
		return fmt.Errorf("add oxsarits: %w", err)
	}
	return nil
}

func (r *PgRepo) SelectActiveLotsBySeller(ctx context.Context, tx pgx.Tx, sellerID string) ([]Lot, error) {
	rows, err := tx.Query(ctx, `
		SELECT id, seller_user_id, artifact_unit_id, quantity, price_oxsarit,
		       status, created_at, expires_at,
		       buyer_user_id, sold_at, expire_event_id
		FROM exchange_lots
		WHERE seller_user_id = $1 AND status = 'active'
		ORDER BY id
		FOR UPDATE
	`, sellerID)
	if err != nil {
		return nil, fmt.Errorf("select active lots by seller: %w", err)
	}
	defer rows.Close()
	var out []Lot
	for rows.Next() {
		var l Lot
		if err := rows.Scan(&l.ID, &l.SellerUserID, &l.ArtifactUnitID, &l.Quantity,
			&l.PriceOxsarit, &l.Status, &l.CreatedAt, &l.ExpiresAt,
			&l.BuyerUserID, &l.SoldAt, &l.ExpireEventID); err != nil {
			return nil, err
		}
		l.UnitPriceOxsarit = l.PriceOxsarit / int64(l.Quantity)
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r *PgRepo) Stats(ctx context.Context, window time.Duration) ([]StatsRow, error) {
	// Объединяем active-counts с pricing-aggregates через FULL OUTER JOIN
	// по unit_id: чтобы вернуть и unit'ы без active-лотов (но с историей)
	// и unit'ы с active-лотами без bought-истории.
	rows, err := r.db.Pool().Query(ctx, `
		WITH active AS (
			SELECT artifact_unit_id AS uid, COUNT(*) AS n
			FROM exchange_lots WHERE status = 'active'
			GROUP BY artifact_unit_id
		),
		bought AS (
			SELECT l.artifact_unit_id AS uid,
			       AVG(l.price_oxsarit::float / l.quantity) AS avgp,
			       SUM(l.price_oxsarit) AS vol
			FROM exchange_history h
			JOIN exchange_lots l ON l.id = h.lot_id
			WHERE h.event_kind = 'bought' AND h.created_at > now() - $1::interval
			GROUP BY l.artifact_unit_id
		)
		SELECT COALESCE(a.uid, b.uid),
		       COALESCE(a.n, 0),
		       b.avgp,
		       COALESCE(b.vol, 0)
		FROM active a
		FULL OUTER JOIN bought b ON a.uid = b.uid
		ORDER BY 1
	`, window.String())
	if err != nil {
		return nil, fmt.Errorf("stats query: %w", err)
	}
	defer rows.Close()
	var out []StatsRow
	for rows.Next() {
		var s StatsRow
		var avgp *float64
		if err := rows.Scan(&s.ArtifactUnitID, &s.ActiveLots, &avgp, &s.Last30dVolume); err != nil {
			return nil, err
		}
		if avgp != nil {
			v := int64(*avgp)
			s.AvgUnitPrice = &v
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
