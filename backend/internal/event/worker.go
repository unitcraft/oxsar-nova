package event

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/oxsar/nova/backend/internal/repo"
)

// Handler обрабатывает событие определённого Kind. Вызывается внутри
// транзакции воркера; если вернёт ошибку — событие помечается error
// и не повторяется автоматически (защита от шторма).
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
}

// Worker — ядро event-loop. Не хранит состояние между циклами.
type Worker struct {
	db       repo.Exec
	log      *slog.Logger
	handlers map[Kind]Handler
	interval time.Duration
	batch    int
}

func NewWorker(db repo.Exec, log *slog.Logger) *Worker {
	return &Worker{
		db:       db,
		log:      log,
		handlers: map[Kind]Handler{},
		interval: time.Duration(KindBatchProcessIntervalSecond) * time.Second,
		batch:    100,
	}
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
func (w *Worker) Run(ctx context.Context) error {
	t := time.NewTicker(w.interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			if err := w.tick(ctx); err != nil {
				w.log.ErrorContext(ctx, "event_tick_error", slog.String("err", err.Error()))
			}
		}
	}
}

// tick забирает и обрабатывает пачку готовых событий.
func (w *Worker) tick(ctx context.Context) error {
	return w.db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT id, user_id, planet_id, kind, fire_at, payload, created_at
			FROM events
			WHERE state = 'wait' AND fire_at <= now()
			ORDER BY fire_at
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		`, w.batch)
		if err != nil {
			return fmt.Errorf("query events: %w", err)
		}
		var batch []Event
		for rows.Next() {
			var e Event
			var kind int
			if err := rows.Scan(&e.ID, &e.UserID, &e.PlanetID, &kind, &e.FireAt, &e.Payload, &e.CreatedAt); err != nil {
				rows.Close()
				return fmt.Errorf("scan event: %w", err)
			}
			e.Kind = Kind(kind)
			batch = append(batch, e)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("rows err: %w", err)
		}

		for _, e := range batch {
			if err := w.process(ctx, tx, e); err != nil {
				w.log.WarnContext(ctx, "event_failed",
					slog.String("event_id", e.ID), slog.Int("kind", int(e.Kind)), slog.String("err", err.Error()))
				if _, uErr := tx.Exec(ctx, `UPDATE events SET state='error', processed_at=now() WHERE id=$1`, e.ID); uErr != nil {
					return fmt.Errorf("mark error: %w", uErr)
				}
				continue
			}
			if _, err := tx.Exec(ctx, `UPDATE events SET state='ok', processed_at=now() WHERE id=$1`, e.ID); err != nil {
				return fmt.Errorf("mark ok: %w", err)
			}
		}
		return nil
	})
}

func (w *Worker) process(ctx context.Context, tx pgx.Tx, e Event) error {
	h, ok := w.handlers[e.Kind]
	if !ok {
		return fmt.Errorf("no handler for kind %d", e.Kind)
	}
	return h(ctx, tx, e)
}

// ErrSkip — хэндлер возвращает её, когда событие нужно оставить в wait
// (например, зависимость ещё не готова). Не используется по умолчанию,
// но зарезервировано на будущее.
var ErrSkip = errors.New("event: skip for now")
