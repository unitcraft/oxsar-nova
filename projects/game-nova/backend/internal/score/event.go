package score

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
)

// RecalcAllScheduled — точка входа для scheduler'а (план 32 Ф.4).
// Пересчитывает очки всех активных (umode=false, deleted_at IS NULL)
// игроков через RecalcUser. Не работает в транзакции — большой update
// на 10k+ users мог бы блокировать всё. Ошибки отдельных игроков
// логируются внутри, цикл продолжается.
//
// Существующий RecalcAll(ctx, log) в service.go оставлен для
// admin-handler'ов (требует свой log-интерфейс).
func (s *Service) RecalcAllScheduled(ctx context.Context) error {
	return s.recalcAllWithLog(ctx)
}

// RecalcAllEvent возвращает handler для KindScoreRecalcAll.
// Сохранён для совместимости с legacy wait-events, созданных до
// плана 32. Новые запуски идут через scheduler.RecalcAll напрямую.
// При обработке legacy event'а handler reschedule НЕ делает —
// scheduler сам тикает по cron-расписанию.
func (s *Service) RecalcAllEvent() event.Handler {
	return func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		// Пересчёт не в транзакции event'а — большой update может
		// блокировать всё. Goroutine отрабатывает на отдельном connection.
		// Reschedule не делаем — scheduler тикает по cron независимо.
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

