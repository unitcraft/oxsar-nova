package artefact

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// List GET /api/artefacts — инвентарь текущего пользователя.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	items, err := h.svc.ListUser(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"artefacts": items})
}

// Activate POST /api/artefacts/{id}/activate
func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	rec, err := h.svc.Activate(r.Context(), uid, id)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, rec)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrAlreadyActive):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "already active"))
	case errors.Is(err, ErrNonStackable):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "non-stackable already active"))
	case errors.Is(err, ErrPlanetRequired):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "artefact requires planet"))
	case errors.Is(err, ErrUnknownArtefact):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unknown artefact"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Deactivate POST /api/artefacts/{id}/deactivate
func (h *Handler) Deactivate(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	if err := h.svc.Deactivate(r.Context(), uid, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
