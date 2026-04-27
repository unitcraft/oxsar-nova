package httpx

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/identity, oxsar/portal и oxsar/billing. При любом изменении
// синхронизируйте КОПИИ:
//   - projects/game-nova/backend/internal/httpx/trace.go
//   - projects/identity/backend/internal/httpx/trace.go
//   - projects/portal/backend/internal/httpx/trace.go
//   - projects/billing/backend/internal/httpx/trace.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"net/http"

	"oxsar/billing/pkg/ids"
	"oxsar/billing/pkg/trace"
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
