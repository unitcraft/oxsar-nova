package galaxy

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

// Handler — HTTP-адаптер чтения галактики. На текущем этапе только
// GET /api/galaxy/{g}/{s}.
type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }

// System GET /api/galaxy/{g}/{s}
func (h *Handler) System(w http.ResponseWriter, r *http.Request) {
	g, err1 := strconv.Atoi(chi.URLParam(r, "g"))
	s, err2 := strconv.Atoi(chi.URLParam(r, "s"))
	if err1 != nil || err2 != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid coords"))
		return
	}
	if err := (Coords{Galaxy: g, System: s, Position: 1}).Validate(); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	uid, _ := auth.UserID(r.Context())
	view, err := h.repo.ReadSystem(r.Context(), g, s, uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, view)
}
