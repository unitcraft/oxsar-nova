package alliance

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// List GET /api/alliances
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context(), 50)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"alliances": list})
}

// Get GET /api/alliances/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	al, members, err := h.svc.Get(r.Context(), id)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
			"alliance": al,
			"members":  members,
		})
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// My GET /api/alliances/me
func (h *Handler) My(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	al, members, err := h.svc.MyAlliance(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if al == nil {
		httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"alliance": nil, "members": nil})
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"alliance": al, "members": members})
}

// Create POST /api/alliances
// Body: {"tag":"TAG","name":"Full Name","description":"..."}
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req struct {
		Tag         string `json:"tag"`
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	al, err := h.svc.Create(r.Context(), uid, req.Tag, req.Name, req.Description)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, map[string]any{"alliance": al})
	case errors.Is(err, ErrAlreadyMember):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "already in an alliance"))
	case errors.Is(err, ErrTagTaken):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "tag already taken"))
	case errors.Is(err, ErrNameTaken):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "name already taken"))
	case errors.Is(err, ErrInvalidTag):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "tag must be 3–5 latin letters/digits"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Join POST /api/alliances/{id}/join
func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	err := h.svc.Join(r.Context(), uid, id)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrAlreadyMember):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "already in an alliance"))
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Leave POST /api/alliances/leave
func (h *Handler) Leave(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	err := h.svc.Leave(r.Context(), uid)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotMember):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "not in an alliance"))
	case errors.Is(err, ErrCannotLeaveOwn):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "owner must disband the alliance first"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Disband DELETE /api/alliances/{id}
func (h *Handler) Disband(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	err := h.svc.Disband(r.Context(), uid, id)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrNotOwner):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}
