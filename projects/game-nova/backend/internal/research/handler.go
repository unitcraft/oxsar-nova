package research

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/requirements"
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
	case errors.Is(err, ErrUmodeBlocked):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	case errors.Is(err, ErrObserverBlocked):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	case errors.Is(err, ErrMaxLevelReached):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "max research level reached"))
	case requirements.IsNotMet(err):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// StartVIP POST /api/research/{queueId}/vip — VIP-instant старт
// исследования за credits (план 72.1.44 cross-cut).
func (h *Handler) StartVIP(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	queueID := chi.URLParam(r, "queueId")
	item, err := h.svc.StartVIP(r.Context(), uid, queueID)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, item)
	case errors.Is(err, ErrQueueItemNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrNotEnoughCredit):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "not enough credit"))
	case errors.Is(err, ErrVIPAlreadyStarted):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "task already started"))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Cancel DELETE /api/research/{queueId} — отмена исследования с
// возвратом ресурсов. План 72.1.39 / правило 1:1 (legacy
// Research::abort).
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	queueID := chi.URLParam(r, "queueId")
	if err := h.svc.Cancel(r.Context(), uid, queueID); err != nil {
		switch {
		case errors.Is(err, ErrQueueItemNotFound):
			httpx.WriteError(w, r, httpx.ErrNotFound)
		case errors.Is(err, ErrPlanetOwnership):
			httpx.WriteError(w, r, httpx.ErrForbidden)
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
	resCosts := h.svc.ResearchCostsMap(levels)
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"queue":            queue,
		"levels":           levels,
		"research_seconds": resSecs,
		"research_costs":   resCosts,
	})
}
