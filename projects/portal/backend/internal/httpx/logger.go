// Package httpx содержит HTTP-обвязки: логгер-middleware, error-writer,
// request-id, recoverer. Все handler-ы должны писать ответы через httpx,
// чтобы формат ошибок был единым.
package httpx

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/auth, oxsar/portal и oxsar/billing. При любом изменении
// синхронизируйте КОПИИ:
//   - projects/game-nova/backend/internal/httpx/logger.go
//   - projects/auth/backend/internal/httpx/logger.go
//   - projects/portal/backend/internal/httpx/logger.go
//   - projects/billing/backend/internal/httpx/logger.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"log/slog"
	"net/http"
	"time"
)

// Logger — middleware, логирующий каждый запрос с длительностью и статусом.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			log.InfoContext(r.Context(), "http_request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", sw.status),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (s *statusWriter) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}
