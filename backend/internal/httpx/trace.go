package httpx

import (
	"net/http"

	"github.com/oxsar/nova/backend/pkg/ids"
	"github.com/oxsar/nova/backend/pkg/trace"
)

const traceHeader = "X-Trace-Id"

// TraceIDMiddleware читает X-Trace-Id из входящего запроса (если есть)
// или генерирует новый uuid-v4. Кладёт в context через trace.WithTraceID
// и возвращает в response-header для клиента.
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(traceHeader)
		if id == "" {
			id = ids.New()
		}
		w.Header().Set(traceHeader, id)
		ctx := trace.WithTraceID(r.Context(), id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
