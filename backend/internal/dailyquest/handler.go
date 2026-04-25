package dailyquest

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

// Handler — HTTP-адаптер.
//
//	GET  /api/daily-quests           — список quest на сегодня (lazy-gen при первом GET)
//	POST /api/daily-quests/{id}/claim — забрать награду
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List GET /api/daily-quests
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	qs, err := h.svc.List(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if qs == nil {
		qs = []Quest{}
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"quests": qs})
}

// Claim POST /api/daily-quests/{id}/claim
func (h *Handler) Claim(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	defIDStr := chi.URLParam(r, "id")
	defID, err := strconv.Atoi(defIDStr)
	if err != nil || defID <= 0 {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid quest id"))
		return
	}
	credits, m, si, hy, err := h.svc.Claim(r.Context(), uid, defID)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
			"reward_credits":  credits,
			"reward_metal":    m,
			"reward_silicon":  si,
			"reward_hydrogen": hy,
		})
	case errors.Is(err, ErrQuestNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotCompleted),
		errors.Is(err, ErrAlreadyClaimed):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}
