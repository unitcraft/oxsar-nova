package httpx

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recoverer ловит panic'и в handler'ах и превращает их в 500.
// Пригодится, пока в проде всё же попадут баги с nil-разыменованием.
func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.ErrorContext(r.Context(), "http_panic",
						slog.Any("panic", rec),
						slog.String("stack", string(debug.Stack())),
					)
					WriteError(w, r, ErrInternal)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
