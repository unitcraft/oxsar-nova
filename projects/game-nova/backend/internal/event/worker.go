package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/metrics"
	"github.com/oxsar/nova/backend/pkg/trace"
)

// Handler обрабатывает событие определённого Kind. Вызывается внутри
// отдельной транзакции на event — ошибка не откатывает соседей по батчу.
// Возврат ErrSkip оставляет событие в wait с переносом fire_at (backoff не инкрементит attempt).
// Любая другая ошибка инкрементит attempt и либо планирует retry, либо помечает 'error'.
type Handler func(ctx context.Context, tx pgx.Tx, e Event) error

// Event — запись из таблицы events в доменном виде.
type Event struct {
	ID        string
	UserID    *string
	PlanetID  *string
	Kind      Kind
	FireAt    time.Time
	Payload   json.RawMessage
	CreatedAt time.Time
	Attempt   int
	TraceID   *string
}

// Worker — ядро event-loop. Не хранит состояние между циклами.
type Worker struct {
	db          repo.Exec
	log         *slog.Logger
	handlers    map[Kind]Handler
	interval    time.Duration
	batch       int
	maxBatch    int
	maxAttempts int
}

// MaxAttempts по умолчанию — 3 попытки (initial + 2 retry).
const DefaultMaxAttempts = 3

// DefaultMaxBatch — верхняя граница в адаптивном режиме за один tick-cycle.
// Если БД переполнена, за один тик воркер может обработать до этого числа
// событий (подряд, без ожидания Ticker), чтобы догнать отставание.
const DefaultMaxBatch = 1000

// Backoff: attempt 1 → 10s, 2 → 60s, 3+ → 300s.
func backoff(attempt int) time.Duration {
	switch {
	case attempt <= 1:
		return 10 * time.Second
	case attempt == 2:
		return 60 * time.Second
	default:
		return 300 * time.Second
	}
}

func NewWorker(db repo.Exec, log *slog.Logger) *Worker {
	return &Worker{
		db:          db,
		log:         log,
		handlers:    map[Kind]Handler{},
		interval:    time.Duration(KindBatchProcessIntervalSecond) * time.Second,
		batch:       100,
		maxBatch:    DefaultMaxBatch,
		maxAttempts: DefaultMaxAttempts,
	}
}

// Config настраивает размеры батчей и интервал. Все поля опциональны:
// если 0 — используется default.
type Config struct {
	Interval    time.Duration
	Batch       int
	MaxBatch    int
	MaxAttempts int
}

// WithConfig применяет настройки к существующему Worker.
func (w *Worker) WithConfig(cfg Config) *Worker {
	if cfg.Interval > 0 {
		w.interval = cfg.Interval
	}
	if cfg.Batch > 0 {
		w.batch = cfg.Batch
	}
	if cfg.MaxBatch > 0 {
		w.maxBatch = cfg.MaxBatch
	}
	if cfg.MaxAttempts > 0 {
		w.maxAttempts = cfg.MaxAttempts
	}
	return w
}

// Register добавляет handler для типа события. Паникует при повторной
// регистрации — это конфигурационная ошибка.
func (w *Worker) Register(kind Kind, h Handler) {
	if _, exists := w.handlers[kind]; exists {
		panic(fmt.Sprintf("event handler for kind %d already registered", kind))
	}
	w.handlers[kind] = h
}

// Run запускает цикл. Возврат — только по отмене контекста.
// Адаптивное поведение: если batch заполнен — не ждать Ticker, а
// тикать подряд до пустого или до maxBatch событий за цикл.
func (w *Worker) Run(ctx context.Context) error {
	return w.RunWithGrace(ctx, 0)
}

// RunWithGrace — как Run, но при отмене signalCtx даёт текущему
// handler'у grace времени завершиться. Внутри используются два контекста:
// signalCtx — сигнал «надо останавливаться» (проверяется между events);
// workCtx — контекст handler'а (не отменяется до grace-таймаута).
//
// Если grace == 0 → поведение как Run: отмена мгновенная.
func (w *Worker) RunWithGrace(signalCtx context.Context, grace time.Duration) error {
	workCtx := signalCtx
	var graceCancel context.CancelFunc
	if grace > 0 {
		workCtx = context.Background()
	}

	t := time.NewTicker(w.interval)
	defer t.Stop()
	for {
		select {
		case <-signalCtx.Done():
			if grace > 0 {
				// Даём текущему tick доработать в пределах grace.
				var cctx context.Context
				cctx, graceCancel = context.WithTimeout(workCtx, grace)
				defer graceCancel()
				w.log.InfoContext(workCtx, "event_worker_graceful_stop",
					slog.Duration("grace", grace))
				// Последний тик — чтобы дренажировать почти-готовые events.
				_ = w.tickLoop(cctx)
			}
			return signalCtx.Err()
		case <-t.C:
			if err := w.tickLoop(workCtx); err != nil {
				w.log.ErrorContext(workCtx, "event_tick_error", slog.String("err", err.Error()))
			}
		}
	}
}

// tickLoop вызывает tick подряд, пока батч полный и не достигли maxBatch.
// Безопасно при 1 воркере; при N воркерах с SKIP LOCKED — тоже ок,
// каждый дренирует свою подмножество.
func (w *Worker) tickLoop(ctx context.Context) error {
	processed := 0
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if processed >= w.maxBatch {
			return nil
		}
		n, err := w.tickOnce(ctx)
		if err != nil {
			return err
		}
		processed += n
		// Если батч неполный — очередь пуста, ждём следующего тикера.
		if n < w.batch {
			return nil
		}
	}
}

// tickOnce забирает список event-ID и обрабатывает каждое событие в
// отдельной транзакции. Ошибка одного события больше не откатывает
// остальные. Возвращает количество обработанных (или попытавшихся)
// событий — используется адаптивным tickLoop.
func (w *Worker) tickOnce(ctx context.Context) (int, error) {
	// Шаг 1: взять ID батча в короткой read-only транзакции.
	ids, err := w.fetchBatchIDs(ctx)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, nil
	}
	// Шаг 2: обработать каждое отдельно.
	for _, id := range ids {
		if ctx.Err() != nil {
			return len(ids), ctx.Err()
		}
		if err := w.processOne(ctx, id); err != nil {
			w.log.WarnContext(ctx, "event_tx_failed",
				slog.String("event_id", id), slog.String("err", err.Error()))
		}
	}
	return len(ids), nil
}

func (w *Worker) fetchBatchIDs(ctx context.Context) ([]string, error) {
	rows, err := w.db.Pool().Query(ctx, `
		SELECT id FROM events
		WHERE state = 'wait' AND fire_at <= now()
		ORDER BY fire_at, id
		LIMIT $1
	`, w.batch)
	if err != nil {
		return nil, fmt.Errorf("query event ids: %w", err)
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// processOne выполняет полный цикл обработки одного события в
// отдельной транзакции. FOR UPDATE SKIP LOCKED защищает от дубля между
// воркерами.
func (w *Worker) processOne(ctx context.Context, id string) error {
	return w.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var e Event
		var kind int
		err := tx.QueryRow(ctx, `
			SELECT id, user_id, planet_id, kind, fire_at, payload, created_at, attempt, trace_id
			FROM events
			WHERE id = $1 AND state = 'wait' AND fire_at <= now()
			FOR UPDATE SKIP LOCKED
		`, id).Scan(&e.ID, &e.UserID, &e.PlanetID, &kind, &e.FireAt, &e.Payload, &e.CreatedAt, &e.Attempt, &e.TraceID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// Уже взяли другим воркером / не waiting — это нормально.
				return nil
			}
			return fmt.Errorf("select event: %w", err)
		}
		e.Kind = Kind(kind)

		handler, ok := w.handlers[e.Kind]
		if !ok {
			// Неизвестный kind — сразу в error без retry.
			if _, err := tx.Exec(ctx, `
				UPDATE events SET state='error', processed_at=now(),
				                  last_error=$1
				WHERE id=$2
			`, fmt.Sprintf("no handler for kind %d", e.Kind), e.ID); err != nil {
				return fmt.Errorf("mark error no-handler: %w", err)
			}
			w.log.WarnContext(ctx, "event_no_handler",
				slog.String("event_id", e.ID), slog.Int("kind", int(e.Kind)))
			return nil
		}

		// Прокидываем trace_id в context handler'а и во все slog-записи ниже.
		hCtx := ctx
		if e.TraceID != nil && *e.TraceID != "" {
			hCtx = trace.WithTraceID(ctx, *e.TraceID)
		}

		kindStr := strconv.Itoa(int(e.Kind))
		startTs := time.Now()
		hErr := handler(hCtx, tx, e)
		if metrics.EventHandlerSec != nil {
			metrics.EventHandlerSec.WithLabelValues(kindStr).Observe(time.Since(startTs).Seconds())
		}

		if hErr != nil {
			if errors.Is(hErr, ErrSkip) {
				// Skip — сдвинуть fire_at на backoff(attempt+1), но не
				// инкрементить attempt (это не «сбой», а «зависимость не готова»).
				delay := backoff(e.Attempt + 1)
				if _, err := tx.Exec(ctx, `
					UPDATE events SET fire_at = now() + $1::interval,
					                  last_error = $2
					WHERE id = $3
				`, delay.String(), hErr.Error(), e.ID); err != nil {
					return fmt.Errorf("mark skip: %w", err)
				}
				if metrics.EventsProcessed != nil {
					metrics.EventsProcessed.WithLabelValues(kindStr, "skip").Inc()
				}
				return nil
			}
			nextAttempt := e.Attempt + 1
			if nextAttempt >= w.maxAttempts {
				// Исчерпали попытки.
				if _, err := tx.Exec(ctx, `
					UPDATE events SET state='error', processed_at=now(),
					                  attempt=$1, last_error=$2
					WHERE id=$3
				`, nextAttempt, hErr.Error(), e.ID); err != nil {
					return fmt.Errorf("mark error: %w", err)
				}
				w.log.WarnContext(ctx, "event_error_final",
					slog.String("event_id", e.ID), slog.Int("kind", int(e.Kind)),
					slog.Int("attempt", nextAttempt), slog.String("err", hErr.Error()))
				if metrics.EventsProcessed != nil {
					metrics.EventsProcessed.WithLabelValues(kindStr, "error").Inc()
				}
				return nil
			}
			// Retry.
			delay := backoff(nextAttempt)
			retryAt := time.Now().Add(delay)
			if _, err := tx.Exec(ctx, `
				UPDATE events SET attempt=$1, fire_at=$2, next_retry_at=$2, last_error=$3
				WHERE id=$4
			`, nextAttempt, retryAt, hErr.Error(), e.ID); err != nil {
				return fmt.Errorf("schedule retry: %w", err)
			}
			w.log.InfoContext(ctx, "event_retry",
				slog.String("event_id", e.ID), slog.Int("kind", int(e.Kind)),
				slog.Int("attempt", nextAttempt), slog.Duration("delay", delay),
				slog.String("err", hErr.Error()))
			if metrics.EventsProcessed != nil {
				metrics.EventsProcessed.WithLabelValues(kindStr, "retry").Inc()
			}
			return nil
		}

		// Успех.
		if _, err := tx.Exec(ctx, `
			UPDATE events SET state='ok', processed_at=now()
			WHERE id=$1
		`, e.ID); err != nil {
			return fmt.Errorf("mark ok: %w", err)
		}
		if metrics.EventsProcessed != nil {
			metrics.EventsProcessed.WithLabelValues(kindStr, "ok").Inc()
		}
		return nil
	})
}

// RunMetricsUpdater периодически обновляет queue-depth gauge'ы.
// Запускается рядом с Run в goroutine.
func (w *Worker) RunMetricsUpdater(ctx context.Context) error {
	t := time.NewTicker(15 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			w.updateMetrics(ctx)
		}
	}
}

func (w *Worker) updateMetrics(ctx context.Context) {
	if metrics.EventsQueue == nil {
		return
	}
	rows, err := w.db.Pool().Query(ctx,
		`SELECT state, COUNT(*) FROM events GROUP BY state`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var state string
		var count int64
		if err := rows.Scan(&state, &count); err == nil {
			metrics.EventsQueue.WithLabelValues(state).Set(float64(count))
		}
	}

	var lag *float64
	_ = w.db.Pool().QueryRow(ctx, `
		SELECT EXTRACT(EPOCH FROM (now() - MIN(fire_at)))
		FROM events WHERE state='wait' AND fire_at <= now()
	`).Scan(&lag)
	if lag != nil {
		metrics.EventsLagSec.Set(*lag)
	} else {
		metrics.EventsLagSec.Set(0)
	}
}

// ErrSkip — хэндлер возвращает её, когда событие нужно оставить в wait
// (например, зависимость ещё не готова). Сдвигает fire_at на backoff,
// но не инкрементирует attempt.
var ErrSkip = errors.New("event: skip for now")

// DeadLetterThreshold — после скольки дней error-event переносится
// в events_dead. Ограничивает рост основной таблицы events.
const DeadLetterThreshold = 7 * 24 * time.Hour

// PruneErrors перемещает error-события старше threshold'а в events_dead.
// Возвращает количество перенесённых записей.
func (w *Worker) PruneErrors(ctx context.Context) (int64, error) {
	var moved int64
	err := w.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			WITH moved AS (
				DELETE FROM events
				WHERE state = 'error'
				  AND processed_at IS NOT NULL
				  AND processed_at < now() - $1::interval
				RETURNING id, user_id, planet_id, kind, fire_at, payload,
				          created_at, processed_at, attempt, last_error
			)
			INSERT INTO events_dead (
				id, user_id, planet_id, kind, fire_at, payload,
				created_at, processed_at, attempt, last_error
			)
			SELECT id, user_id, planet_id, kind, fire_at, payload,
			       created_at, processed_at, attempt, last_error
			FROM moved
		`, DeadLetterThreshold.String())
		if err != nil {
			return fmt.Errorf("prune errors: %w", err)
		}
		moved = tag.RowsAffected()
		return nil
	})
	return moved, err
}

// RunPruner запускает демон, раз в сутки перемещающий устаревшие error-events.
// Отдельная goroutine от Run (чтобы не задерживать tick).
func (w *Worker) RunPruner(ctx context.Context) error {
	t := time.NewTicker(24 * time.Hour)
	defer t.Stop()
	// Первый прогон через 1 минуту после старта (прогрев).
	firstRun := time.NewTimer(time.Minute)
	defer firstRun.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-firstRun.C:
			w.runPruneOnce(ctx)
		case <-t.C:
			w.runPruneOnce(ctx)
		}
	}
}

func (w *Worker) runPruneOnce(ctx context.Context) {
	moved, err := w.PruneErrors(ctx)
	if err != nil {
		w.log.ErrorContext(ctx, "event_prune_failed", slog.String("err", err.Error()))
		return
	}
	if moved > 0 {
		w.log.InfoContext(ctx, "event_pruned",
			slog.Int64("moved", moved),
			slog.String("threshold", DeadLetterThreshold.String()))
	}
}
