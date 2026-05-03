package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"oxsar/game-nova/pkg/ids"
	"oxsar/game-nova/pkg/trace"
)

// Execer — минимальный интерфейс для event.Insert. Реализуется и pgx.Tx,
// и *pgxpool.Pool, и *pgx.Conn — что позволяет вставлять events как из
// транзакции (предпочтительно), так и напрямую через pool (когда
// контекст вставки не требует tx).
type Execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// InsertOpts — параметры для event.Insert. UserID/PlanetID опциональны.
// Если ID пуст — генерируется через ids.New(); иначе используется
// переданный (нужно для случаев, когда вызывающий код заранее знает
// id события, например return_event_id у транспорта).
type InsertOpts struct {
	ID       string
	UserID   *string
	PlanetID *string
	Kind     Kind
	FireAt   time.Time
	Payload  any // JSON-serializable
}

// Insert создаёт запись в events с state='wait' и trace_id из контекста.
// Возвращает id события (сгенерированный или переданный в opts.ID).
//
// Предпочтительный способ вставки event'ов начиная с миграции 0058:
// гарантирует trace_id из HTTP-запроса / родительского event'а,
// чтобы лог HTTP-handler → worker-handler был соединён.
func Insert(ctx context.Context, db Execer, opts InsertOpts) (string, error) {
	id := opts.ID
	if id == "" {
		id = ids.New()
	}
	payload, err := json.Marshal(opts.Payload)
	if err != nil {
		return "", err
	}
	var tid *string
	if t := trace.FromContext(ctx); t != "" {
		tid = &t
	}
	_, err = db.Exec(ctx, `
		INSERT INTO events (id, user_id, planet_id, kind, state, fire_at, payload, trace_id)
		VALUES ($1, $2, $3, $4, 'wait', $5, $6, $7)
	`, id, opts.UserID, opts.PlanetID, int(opts.Kind), opts.FireAt, payload, tid)
	return id, err
}

// pgx.Tx уже удовлетворяет Execer; явная проверка чтобы не сломать
// case использования.
var _ Execer = (pgx.Tx)(nil)
