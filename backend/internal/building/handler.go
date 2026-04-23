package building

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/internal/requirements"
)

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type enqueueRequest struct {
	UnitID int `json:"unit_id"`
}

// Enqueue POST /api/planets/{id}/buildings
func (h *Handler) Enqueue(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")

	var req enqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}

	item, err := h.svc.Enqueue(r.Context(), uid, planetID, req.UnitID)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, item)
	case errors.Is(err, ErrQueueBusy):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "queue busy"))
	case errors.Is(err, ErrNotEnoughRes):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "not enough resources"))
	case errors.Is(err, ErrUnknownUnit):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unknown unit"))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrMoonOnly):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "moon-only building"))
	case errors.Is(err, ErrPlanetOnly):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "not available on moon"))
	case errors.Is(err, ErrMaxLevelReached):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "max level reached"))
	case requirements.IsNotMet(err):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Levels GET /api/planets/{id}/buildings
func (h *Handler) Levels(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	levels, err := h.svc.Levels(r.Context(), planetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	buildSecs, err := h.svc.BuildSecondsMap(r.Context(), planetID, levels)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	unmet, err := h.svc.RequirementsUnmet(r.Context(), uid, planetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"levels":               levels,
		"build_seconds":        buildSecs,
		"requirements_unmet":   unmet,
	})
}

// List GET /api/planets/{id}/buildings/queue
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	items, err := h.svc.List(r.Context(), planetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"queue": items})
}

// Cancel DELETE /api/planets/{id}/buildings/queue/{taskId}
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	taskID := chi.URLParam(r, "taskId")
	if err := h.svc.Cancel(r.Context(), uid, taskID); err != nil {
		if errors.Is(err, ErrQueueItemNotFound) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
