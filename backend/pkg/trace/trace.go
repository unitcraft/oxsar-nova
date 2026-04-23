// Package trace — trace_id для cross-cutting observability.
//
// trace_id uuid кладётся в context на входе HTTP-запроса, читается
// при вставке event'ов и попадает в slog логи обработчика (в том числе
// асинхронного worker-handler'а, см. event.Event.TraceID).
package trace

import (
	"context"

	"github.com/oxsar/nova/backend/pkg/ids"
)

type ctxKey struct{}

var traceIDKey = ctxKey{}

// WithTraceID возвращает контекст с данным trace_id. Если id пустой —
// генерируется новый uuid.
func WithTraceID(ctx context.Context, id string) context.Context {
	if id == "" {
		id = ids.New()
	}
	return context.WithValue(ctx, traceIDKey, id)
}

// FromContext возвращает trace_id из контекста; пустая строка если не задан.
func FromContext(ctx context.Context) string {
	v, _ := ctx.Value(traceIDKey).(string)
	return v
}
