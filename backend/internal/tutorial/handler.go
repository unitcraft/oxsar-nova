package tutorial

import (
	"net/http"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Status GET /api/tutorial
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	uid, _ := auth.UserID(r.Context())
	steps, state, err := h.svc.Status(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"steps":    steps,
		"state":    state,
		"complete": state >= 6,
	})
}
