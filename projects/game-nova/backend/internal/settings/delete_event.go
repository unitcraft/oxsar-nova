// Event-handler для KindAccountDelete (план 72.1.30).
//
// Воркер вызывает Handler.AccountDeleteEventHandler() для каждого
// fire_at <= now() event'а с kind=90. Handler выполняет физический
// soft-delete: UPDATE users.deleted_at, anonymize username/email,
// alliance_id=NULL, cancel market_lots.
//
// Если юзер отменил удаление через CancelDeletion (delete_at=NULL),
// event помечен state='cancelled' воркером перед нашим хэндлером —
// до нас не дойдёт. Дополнительно проверим delete_at внутри tx.

package settings

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
)

// AccountDeleteEventHandler — фабрика event.Handler для KindAccountDelete=90.
//
// Использование (server/main.go или worker/main.go):
//
//	dispatcher.Register(event.KindAccountDelete, settingsH.AccountDeleteEventHandler())
func (h *Handler) AccountDeleteEventHandler() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		uid := ""
		if e.UserID != nil {
			uid = *e.UserID
		}
		if uid == "" {
			return fmt.Errorf("account-delete: no user_id in event")
		}

		// Idempotent: проверяем что delete_at всё ещё установлен и
		// прошёл (не сбросили CancelDeletion).
		var deletePending bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM users
				WHERE id = $1
				  AND deleted_at IS NULL
				  AND delete_at IS NOT NULL
				  AND delete_at <= now()
			)
		`, uid).Scan(&deletePending); err != nil {
			return fmt.Errorf("account-delete: check pending: %w", err)
		}
		if !deletePending {
			// Уже отменено или уже soft-deleted — выходим без ошибки.
			return nil
		}

		// Soft-delete + анонимизация (legacy `Preferences::updateDeletion`
		// + cron финализирует тем же UPDATE).
		if _, err := tx.Exec(ctx, `
			UPDATE users SET
				deleted_at = now(),
				username = '[deleted_' || substr(id::text, 1, 8) || ']',
				email = '[deleted_' || substr(id::text, 1, 8) || ']',
				alliance_id = NULL,
				delete_at = NULL
			WHERE id = $1
		`, uid); err != nil {
			return fmt.Errorf("account-delete: soft-delete user: %w", err)
		}
		// Закрыть открытые лоты игрока.
		if _, err := tx.Exec(ctx, `
			UPDATE market_lots SET state = 'cancelled'
			WHERE seller_id = $1 AND state = 'open'
		`, uid); err != nil {
			// market_lots может отсутствовать в окружении — не критично.
			_ = err
		}
		return nil
	}
}
