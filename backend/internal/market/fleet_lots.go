package market

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// FleetLot — лот с пакетом кораблей. sell_fleet: map[unit_id]count.
type FleetLot struct {
	ID          string           `json:"id"`
	SellerID    string           `json:"seller_id"`
	SellerName  string           `json:"seller_name"`
	PlanetID    string           `json:"planet_id"`
	SellFleet   map[string]int64 `json:"sell_fleet"` // ключ — unit_id (string для JSON)
	BuyResource string           `json:"buy_resource"`
	BuyAmount   int64            `json:"buy_amount"`
	State       string           `json:"state"`
	CreatedAt   string           `json:"created_at"`
}

// ListFleetLots возвращает открытые лоты кораблей.
func (s *Service) ListFleetLots(ctx context.Context, limit int) ([]FleetLot, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := s.db.Pool().Query(ctx, `
		SELECT ml.id, ml.seller_id, COALESCE(u.username, ''),
		       ml.planet_id, ml.sell_fleet, ml.buy_resource, ml.buy_amount,
		       ml.state, ml.created_at
		FROM market_lots ml
		LEFT JOIN users u ON u.id = ml.seller_id
		WHERE ml.state = 'open' AND ml.kind = 'fleet'
		ORDER BY ml.created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("market: list fleet lots: %w", err)
	}
	defer rows.Close()

	var out []FleetLot
	for rows.Next() {
		var lot FleetLot
		var fleetJSON []byte
		var createdAt interface{ Format(string) string }
		_ = createdAt
		if err := rows.Scan(&lot.ID, &lot.SellerID, &lot.SellerName,
			&lot.PlanetID, &fleetJSON, &lot.BuyResource, &lot.BuyAmount,
			&lot.State, &lot.CreatedAt); err != nil {
			return nil, fmt.Errorf("market: scan fleet lot: %w", err)
		}
		if err := json.Unmarshal(fleetJSON, &lot.SellFleet); err != nil {
			return nil, fmt.Errorf("market: parse fleet: %w", err)
		}
		out = append(out, lot)
	}
	return out, rows.Err()
}

// CreateFleetLot создаёт лот с кораблями. fleet — map[unit_id]count.
// Атомарно: списываем корабли с планеты продавца.
func (s *Service) CreateFleetLot(ctx context.Context, userID, planetID string,
	fleet map[int]int64, buyResource string, buyAmount int64) (FleetLot, error) {
	if len(fleet) == 0 {
		return FleetLot{}, ErrInvalidAmount
	}
	for _, cnt := range fleet {
		if cnt <= 0 {
			return FleetLot{}, ErrInvalidAmount
		}
	}
	if !validResource(buyResource) {
		return FleetLot{}, ErrInvalidResource
	}
	if buyAmount <= 0 {
		return FleetLot{}, ErrInvalidAmount
	}

	// Преобразуем для JSON.
	fleetStr := make(map[string]int64, len(fleet))
	for id, cnt := range fleet {
		fleetStr[fmt.Sprintf("%d", id)] = cnt
	}
	fleetJSON, err := json.Marshal(fleetStr)
	if err != nil {
		return FleetLot{}, fmt.Errorf("market: marshal fleet: %w", err)
	}

	var out FleetLot
	err = s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Проверка владельца планеты.
		var ownerID string
		if err := tx.QueryRow(ctx,
			`SELECT user_id FROM planets WHERE id = $1 FOR UPDATE`,
			planetID).Scan(&ownerID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPlanetNotFound
			}
			return fmt.Errorf("market: read planet: %w", err)
		}
		if ownerID != userID {
			return ErrPlanetOwnership
		}
		// Списать корабли с планеты (FOR UPDATE поштучно).
		for id, cnt := range fleet {
			var have int64
			err := tx.QueryRow(ctx,
				`SELECT count FROM ships WHERE planet_id = $1 AND unit_id = $2 FOR UPDATE`,
				planetID, id).Scan(&have)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return ErrNotEnough
				}
				return fmt.Errorf("market: read ships: %w", err)
			}
			if have < cnt {
				return ErrNotEnough
			}
			if _, err := tx.Exec(ctx,
				`UPDATE ships SET count = count - $1 WHERE planet_id = $2 AND unit_id = $3`,
				cnt, planetID, id); err != nil {
				return fmt.Errorf("market: subtract ship: %w", err)
			}
		}
		// Создать лот.
		return tx.QueryRow(ctx, `
			INSERT INTO market_lots (seller_id, planet_id, kind, sell_fleet, buy_resource, buy_amount, state)
			VALUES ($1, $2, 'fleet', $3, $4, $5, 'open')
			RETURNING id, seller_id, planet_id, buy_resource, buy_amount, state, created_at
		`, userID, planetID, fleetJSON, buyResource, buyAmount,
		).Scan(&out.ID, &out.SellerID, &out.PlanetID, &out.BuyResource, &out.BuyAmount, &out.State, &out.CreatedAt)
	})
	if err != nil {
		return FleetLot{}, err
	}
	out.SellFleet = fleetStr
	return out, nil
}

// AcceptFleetLot покупает лот: списываем ресурсы у покупателя, зачисляем продавцу,
// переносим ships на планету покупателя.
func (s *Service) AcceptFleetLot(ctx context.Context, buyerID, buyerPlanetID, lotID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Lot.
		var sellerID, buyResource, state string
		var buyAmount int64
		var fleetJSON []byte
		err := tx.QueryRow(ctx, `
			SELECT seller_id, buy_resource, buy_amount, state, sell_fleet
			FROM market_lots WHERE id = $1 AND kind = 'fleet' FOR UPDATE
		`, lotID).Scan(&sellerID, &buyResource, &buyAmount, &state, &fleetJSON)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrLotNotFound
			}
			return fmt.Errorf("market: read fleet lot: %w", err)
		}
		if state != "open" {
			return ErrLotNotOpen
		}
		if sellerID == buyerID {
			return ErrOwnLot
		}
		// Buyer planet.
		var buyerPlanetOwner string
		var buyerPlanetBalance int64
		if err := tx.QueryRow(ctx,
			`SELECT user_id, `+buyResource+`::bigint FROM planets WHERE id = $1 FOR UPDATE`,
			buyerPlanetID).Scan(&buyerPlanetOwner, &buyerPlanetBalance); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPlanetNotFound
			}
			return fmt.Errorf("market: read buyer planet: %w", err)
		}
		if buyerPlanetOwner != buyerID {
			return ErrPlanetOwnership
		}
		if buyerPlanetBalance < buyAmount {
			return ErrNotEnough
		}
		// Find seller planet (первая не-луна).
		var sellerPlanetID string
		if err := tx.QueryRow(ctx,
			`SELECT id FROM planets WHERE user_id = $1 AND destroyed_at IS NULL AND is_moon = false
			 ORDER BY sort_order, created_at LIMIT 1`,
			sellerID).Scan(&sellerPlanetID); err != nil {
			return fmt.Errorf("market: seller planet: %w", err)
		}
		// Списываем у покупателя, зачисляем продавцу.
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET `+buyResource+` = `+buyResource+` - $1 WHERE id = $2`,
			buyAmount, buyerPlanetID); err != nil {
			return fmt.Errorf("market: debit buyer: %w", err)
		}
		if _, err := tx.Exec(ctx,
			`UPDATE planets SET `+buyResource+` = `+buyResource+` + $1 WHERE id = $2`,
			buyAmount, sellerPlanetID); err != nil {
			return fmt.Errorf("market: credit seller: %w", err)
		}
		// Зачислить корабли покупателю.
		var fleet map[string]int64
		if err := json.Unmarshal(fleetJSON, &fleet); err != nil {
			return fmt.Errorf("market: parse fleet: %w", err)
		}
		for idStr, cnt := range fleet {
			var id int
			if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
				return fmt.Errorf("market: bad unit id %q: %w", idStr, err)
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO ships (planet_id, unit_id, count)
				VALUES ($1, $2, $3)
				ON CONFLICT (planet_id, unit_id) DO UPDATE SET count = ships.count + EXCLUDED.count
			`, buyerPlanetID, id, cnt); err != nil {
				return fmt.Errorf("market: credit ship: %w", err)
			}
		}
		// Закрыть лот.
		if _, err := tx.Exec(ctx,
			`UPDATE market_lots SET state = 'accepted', buyer_id = $1, updated_at = now() WHERE id = $2`,
			buyerID, lotID); err != nil {
			return fmt.Errorf("market: close fleet lot: %w", err)
		}
		return nil
	})
}

// CancelFleetLot отменяет лот и возвращает корабли продавцу.
func (s *Service) CancelFleetLot(ctx context.Context, userID, lotID string) error {
	return s.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var sellerID, planetID, state string
		var fleetJSON []byte
		err := tx.QueryRow(ctx, `
			SELECT seller_id, planet_id, state, sell_fleet
			FROM market_lots WHERE id = $1 AND kind = 'fleet' FOR UPDATE
		`, lotID).Scan(&sellerID, &planetID, &state, &fleetJSON)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrLotNotFound
			}
			return err
		}
		if sellerID != userID {
			return ErrPlanetOwnership
		}
		if state != "open" {
			return ErrLotNotOpen
		}
		var fleet map[string]int64
		if err := json.Unmarshal(fleetJSON, &fleet); err != nil {
			return err
		}
		for idStr, cnt := range fleet {
			var id int
			if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
				INSERT INTO ships (planet_id, unit_id, count)
				VALUES ($1, $2, $3)
				ON CONFLICT (planet_id, unit_id) DO UPDATE SET count = ships.count + EXCLUDED.count
			`, planetID, id, cnt); err != nil {
				return err
			}
		}
		_, err = tx.Exec(ctx,
			`UPDATE market_lots SET state = 'cancelled', updated_at = now() WHERE id = $1`, lotID)
		return err
	})
}
