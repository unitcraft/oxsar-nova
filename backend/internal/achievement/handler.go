package achievement

import (
	"net/http"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// List GET /api/achievements
//
// Перед возвратом прогоняем CheckAll — идемпотентно открывает всё,
// что игрок заслужил с последнего визита. Это lazy-trigger подход:
// не влияет на hot-path (building/attack handler'ы) и гарантирует,
// что UI показывает актуальное состояние без событий.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	if err := h.svc.CheckAll(r.Context(), uid); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	list, err := h.svc.List(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"achievements": list})
}
