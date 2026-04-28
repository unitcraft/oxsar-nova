package event

// План 68 Ф.4: event-handler'ы биржи артефактов.
//
// KindExchangeExpire — фоновое истечение лота. Создаётся вместе с лотом
// (см. exchange.Service.CreateLot, fire_at=expires_at). При срабатывании
// возвращает escrow-артефакты seller'у (state='listed' → 'held') и
// помечает лот expired. Идемпотентен через WHERE status='active' —
// повторный запуск handler'а после buy/cancel будет no-op.
//
// KindExchangeBan — служебный handler. Принимает payload {seller_user_id,
// reason}. SELECT все active-лоты seller'а FOR UPDATE → return escrow +
// status='cancelled' + history(event_kind='banned'). Используется
// автомодерацией / админ-tool'ами при бане игрока.
//
// R8: метрики через metrics.ExchangeEventTotal{kind, status}.
// R10: nova однобазная (universe=отдельная БД), universe_id не нужен.
// R13: payload — типизированные структуры (см. ExchangeExpirePayload,
// ExchangeBanPayload).

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/pkg/metrics"
)

// ExchangeExpirePayload — payload события KindExchangeExpire.
//
// LotID — id лота, который должен истечь. Достаточно: handler сам
// восстанавливает остальное (seller_user_id, артефакты в lot_items).
type ExchangeExpirePayload struct {
	LotID string `json:"lot_id"`
}

// ExchangeBanPayload — payload события KindExchangeBan.
//
// SellerUserID и Reason — для audit. e.UserID может быть nil (системный ban)
// или модератором. Reason — символическое имя ('banned_by_moderator',
// 'fraud_detected' и т.п.).
type ExchangeBanPayload struct {
	SellerUserID string `json:"seller_user_id"`
	Reason       string `json:"reason"`
}

// HandleExchangeExpire обрабатывает истечение лота.
//
// Алгоритм:
//   1. SELECT lot FOR UPDATE WHERE status='active'. Если не active или
//      не существует — no-op (лот купили/отменили до срабатывания event'а).
//   2. SELECT artefact_id FROM exchange_lot_items WHERE lot_id.
//   3. UPDATE artefacts_user SET state='held' WHERE id=ANY($items)
//      AND state='listed' (защита от рассинхрона: если артефакт уже
//      в другом state — не трогаем, но это не ошибка).
//   4. UPDATE lot SET status='expired'.
//   5. INSERT exchange_history (event_kind='expired', actor=NULL).
//
// Идемпотентность: повтор после успешного завершения не находит
// active-лот (status='expired') и завершается no-op.
func HandleExchangeExpire(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl ExchangeExpirePayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("parse exchange expire payload: %w", err)
	}
	if pl.LotID == "" {
		return fmt.Errorf("exchange expire payload missing lot_id")
	}

	// 1. Lock lot. Берём минимальный set полей.
	var (
		sellerUserID string
		status       string
	)
	err := tx.QueryRow(ctx, `
		SELECT seller_user_id, status FROM exchange_lots
		WHERE id = $1 FOR UPDATE
	`, pl.LotID).Scan(&sellerUserID, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.InfoContext(ctx, "exchange_expire_skip_lot_missing",
				slog.String("event_id", e.ID),
				slog.String("lot_id", pl.LotID))
			recordExchangeEvent("expire", "noop")
			return nil
		}
		return fmt.Errorf("lock lot: %w", err)
	}
	if status != "active" {
		slog.InfoContext(ctx, "exchange_expire_skip_not_active",
			slog.String("event_id", e.ID),
			slog.String("lot_id", pl.LotID),
			slog.String("status", status))
		recordExchangeEvent("expire", "noop")
		return nil
	}

	// 2. Items.
	rows, err := tx.Query(ctx,
		`SELECT artefact_id FROM exchange_lot_items WHERE lot_id = $1`,
		pl.LotID)
	if err != nil {
		return fmt.Errorf("select lot items: %w", err)
	}
	var items []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		items = append(items, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	// 3. Return artefacts (state='listed' → 'held').
	if len(items) > 0 {
		if _, err := tx.Exec(ctx, `
			UPDATE artefacts_user SET state = 'held'
			WHERE id = ANY($1) AND state = 'listed'
		`, items); err != nil {
			return fmt.Errorf("return artefacts: %w", err)
		}
	}

	// 4. Lot → expired.
	if _, err := tx.Exec(ctx, `
		UPDATE exchange_lots SET status = 'expired'
		WHERE id = $1 AND status = 'active'
	`, pl.LotID); err != nil {
		return fmt.Errorf("mark expired: %w", err)
	}

	// 5. History.
	historyPayload, _ := json.Marshal(map[string]string{"event_id": e.ID})
	if _, err := tx.Exec(ctx, `
		INSERT INTO exchange_history (id, lot_id, event_kind, actor_user_id, payload)
		VALUES (gen_random_uuid(), $1, 'expired', NULL, $2)
	`, pl.LotID, historyPayload); err != nil {
		return fmt.Errorf("insert history: %w", err)
	}

	slog.InfoContext(ctx, "exchange_lot_expired",
		slog.String("event_id", e.ID),
		slog.String("lot_id", pl.LotID),
		slog.String("seller_user_id", sellerUserID),
		slog.Int("items_returned", len(items)))
	recordExchangeEvent("expire", "ok")
	return nil
}

// HandleExchangeBan отзывает все active-лоты seller'а.
//
// Алгоритм:
//   1. SELECT lots FOR UPDATE WHERE seller_user_id AND status='active'.
//   2. Для каждого лота:
//      - SELECT items, UPDATE artefacts_user SET state='held'
//        WHERE state='listed' (так же защита от рассинхрона).
//      - UPDATE lot SET status='cancelled'.
//      - INSERT history (event_kind='banned', actor=e.UserID — модератор).
//      - UPDATE связанный expire-event SET state='ok' (если был).
//   3. Если active-лотов нет — no-op (idempotent).
//
// Использование: создание этого event'а — ответственность админ/модерации
// (внутри game-nova: будущий admin-handler ставит event с payload и
// fire_at=now()).
func HandleExchangeBan(ctx context.Context, tx pgx.Tx, e Event) error {
	var pl ExchangeBanPayload
	if err := json.Unmarshal(e.Payload, &pl); err != nil {
		return fmt.Errorf("parse exchange ban payload: %w", err)
	}
	if pl.SellerUserID == "" {
		return fmt.Errorf("exchange ban payload missing seller_user_id")
	}

	// 1. Lock all active lots.
	rows, err := tx.Query(ctx, `
		SELECT id, expire_event_id FROM exchange_lots
		WHERE seller_user_id = $1 AND status = 'active'
		ORDER BY id
		FOR UPDATE
	`, pl.SellerUserID)
	if err != nil {
		return fmt.Errorf("select seller lots: %w", err)
	}
	type lotRef struct {
		id            string
		expireEventID *string
	}
	var lots []lotRef
	for rows.Next() {
		var l lotRef
		if err := rows.Scan(&l.id, &l.expireEventID); err != nil {
			rows.Close()
			return err
		}
		lots = append(lots, l)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	if len(lots) == 0 {
		slog.InfoContext(ctx, "exchange_ban_no_active_lots",
			slog.String("event_id", e.ID),
			slog.String("seller_user_id", pl.SellerUserID))
		recordExchangeEvent("ban", "noop")
		return nil
	}

	// 2. Process each lot.
	historyPayload, _ := json.Marshal(map[string]string{
		"seller_user_id": pl.SellerUserID,
		"reason":         pl.Reason,
		"event_id":       e.ID,
	})
	for _, lr := range lots {
		// Items.
		ir, err := tx.Query(ctx,
			`SELECT artefact_id FROM exchange_lot_items WHERE lot_id = $1`, lr.id)
		if err != nil {
			return fmt.Errorf("select items for ban: %w", err)
		}
		var items []string
		for ir.Next() {
			var id string
			if err := ir.Scan(&id); err != nil {
				ir.Close()
				return err
			}
			items = append(items, id)
		}
		ir.Close()
		if err := ir.Err(); err != nil {
			return err
		}

		if len(items) > 0 {
			if _, err := tx.Exec(ctx, `
				UPDATE artefacts_user SET state = 'held'
				WHERE id = ANY($1) AND state = 'listed'
			`, items); err != nil {
				return fmt.Errorf("return artefacts on ban: %w", err)
			}
		}
		if _, err := tx.Exec(ctx, `
			UPDATE exchange_lots SET status = 'cancelled'
			WHERE id = $1 AND status = 'active'
		`, lr.id); err != nil {
			return fmt.Errorf("cancel lot on ban: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO exchange_history (id, lot_id, event_kind, actor_user_id, payload)
			VALUES (gen_random_uuid(), $1, 'banned', $2, $3)
		`, lr.id, e.UserID, historyPayload); err != nil {
			return fmt.Errorf("insert ban history: %w", err)
		}
		// Cancel expire-event (если был).
		if lr.expireEventID != nil && *lr.expireEventID != "" {
			if _, err := tx.Exec(ctx, `
				UPDATE events SET state='ok', processed_at=now(),
				                  last_error='cancelled by exchange_ban'
				WHERE id = $1 AND state = 'wait'
			`, *lr.expireEventID); err != nil {
				return fmt.Errorf("cancel expire event on ban: %w", err)
			}
		}
	}

	slog.InfoContext(ctx, "exchange_ban_applied",
		slog.String("event_id", e.ID),
		slog.String("seller_user_id", pl.SellerUserID),
		slog.String("reason", pl.Reason),
		slog.Int("lots_cancelled", len(lots)))
	recordExchangeEvent("ban", "ok")
	return nil
}

// recordExchangeEvent — обновление ExchangeEventTotal{kind,status} (R8).
func recordExchangeEvent(kind, status string) {
	if metrics.ExchangeEventTotal != nil {
		metrics.ExchangeEventTotal.WithLabelValues(kind, status).Inc()
	}
}
