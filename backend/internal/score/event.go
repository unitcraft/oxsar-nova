package score

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/pkg/ids"
)

// RecalcAllEvent возвращает handler для KindScoreRecalcAll.
// Handler пересчитывает очки всех активных (umode=false, deleted_at IS NULL)
// игроков через RecalcUser + планирует следующий запуск на +24h.
// При ошибках отдельных игроков — логируем, продолжаем цикл.
func (s *Service) RecalcAllEvent() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		// Reschedule next run первым, чтобы даже при fallthrough
		// следующая итерация состоялась.
		nextAt := time.Now().Add(24 * time.Hour)
		if _, err := tx.Exec(ctx, `
			INSERT INTO events (id, kind, state, fire_at, payload)
			VALUES ($1, 70, 'wait', $2, '{}')
		`, ids.New(), nextAt); err != nil {
			return fmt.Errorf("score.recalc_all_event: schedule next: %w", err)
		}

		// Пересчёт не в транзакции event'а — слишком большой update
		// на 10k+ users может блокировать всё. Вместо этого используем
		// отдельный connection из пула, foreground, без tx event'а.
		go func() {
			bgCtx := context.Background()
			if err := s.recalcAllWithLog(bgCtx); err != nil {
				slog.Default().ErrorContext(bgCtx, "score_recalc_all_event_failed",
					slog.String("err", err.Error()))
			}
		}()
		return nil
	}
}

func (s *Service) recalcAllWithLog(ctx context.Context) error {
	start := time.Now()
	rows, err := s.db.Pool().Query(ctx,
		`SELECT id FROM users WHERE umode=false AND deleted_at IS NULL`)
	if err != nil {
		return err
	}
	var uids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		uids = append(uids, id)
	}
	rows.Close()

	var failed int
	for _, id := range uids {
		if err := s.RecalcUser(ctx, id); err != nil {
			failed++
		}
	}
	slog.Default().InfoContext(ctx, "score_recalc_all_done",
		slog.Int("users", len(uids)),
		slog.Int("failed", failed),
		slog.Duration("took", time.Since(start)))
	return nil
}

// BootstrapRecalcAllEvent обеспечивает, что в events есть хотя бы один
// wait-event типа KindScoreRecalcAll (чтобы цикл запустился).
// Вызывается при старте worker'а. Идемпотентно.
func (s *Service) BootstrapRecalcAllEvent(ctx context.Context) error {
	var exists bool
	err := s.db.Pool().QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM events
			WHERE kind = $1 AND state = 'wait'
		)
	`, int(event.KindScoreRecalcAll)).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	// Ставим первый запуск через 1 минуту после старта worker'а.
	fireAt := time.Now().Add(1 * time.Minute)
	_, err = s.db.Pool().Exec(ctx, `
		INSERT INTO events (id, kind, state, fire_at, payload)
		VALUES ($1, $2, 'wait', $3, '{}')
	`, ids.New(), int(event.KindScoreRecalcAll), fireAt)
	return err
}
