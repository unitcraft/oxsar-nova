package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/pkg/ids"
	"oxsar/game-nova/pkg/trace"
)

// InsertOpts — параметры для event.Insert. UserID/PlanetID опциональны.
type InsertOpts struct {
	UserID   *string
	PlanetID *string
	Kind     Kind
	FireAt   time.Time
	Payload  any // JSON-serializable
}

// Insert создаёт запись в events с state='wait' и trace_id из контекста.
// Возвращает сгенерированный id события.
//
// Предпочтительный способ вставки event'ов начиная с миграции 0058:
// гарантирует trace_id из HTTP-запроса / родительского event'а,
// чтобы лог HTTP-handler → worker-handler был соединён.
func Insert(ctx context.Context, tx pgx.Tx, opts InsertOpts) (string, error) {
	id := ids.New()
	payload, err := json.Marshal(opts.Payload)
	if err != nil {
		return "", err
	}
	var tid *string
	if t := trace.FromContext(ctx); t != "" {
		tid = &t
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload, trace_id)
		VALUES ($1, $2, $3, $4, 'wait', $5, $6, $7)
	`, id, opts.UserID, opts.PlanetID, int(opts.Kind), opts.FireAt, payload, tid)
	return id, err
}
