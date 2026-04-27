package httpx

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/auth и oxsar/portal. При любом изменении синхронизируйте КОПИИ:
//   - projects/game-nova/backend/internal/httpx/router.go
//   - projects/auth/backend/internal/httpx/router.go
//   - projects/portal/backend/internal/httpx/router.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// RouterDeps агрегирует зависимости, которые router прокидывает в handler-ы
// через chi-middleware (контекст-значения).
type RouterDeps struct {
	Log *slog.Logger
}

// NewRouter возвращает chi.Router с базовыми middleware.
// Сами маршруты навешиваются в cmd/server/routes.go, а не здесь —
// чтобы не плодить cross-import.
func NewRouter(deps RouterDeps) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(TraceIDMiddleware)
	r.Use(Logger(deps.Log))
	r.Use(Recoverer(deps.Log))
	r.Use(middleware.Timeout(15 * 1e9)) // 15s
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, r, http.StatusOK, map[string]string{"status": "ok"})
	})
	return r
}
