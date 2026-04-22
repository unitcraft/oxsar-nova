package research

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

// Enqueue POST /api/planets/{id}/research
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
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "research in progress"))
	case errors.Is(err, ErrNotEnoughRes):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "not enough resources"))
	case errors.Is(err, ErrUnknownUnit):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unknown research"))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrNoResearchLab):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "research lab required"))
	case requirements.IsNotMet(err):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// List GET /api/research — все текущие исследования игрока + уровни.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	queue, err := h.svc.List(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	levels, err := h.svc.Levels(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	resSecs, err := h.svc.ResearchSecondsMap(r.Context(), uid, levels)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"queue":            queue,
		"levels":           levels,
		"research_seconds": resSecs,
	})
}
