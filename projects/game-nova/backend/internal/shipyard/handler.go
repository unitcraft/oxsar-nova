package shipyard

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
	UnitID int   `json:"unit_id"`
	Count  int64 `json:"count"`
}

// Enqueue POST /api/planets/{id}/shipyard
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
	item, err := h.svc.Enqueue(r.Context(), uid, planetID, req.UnitID, req.Count)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, item)
	case errors.Is(err, ErrInvalidCount):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "count must be positive"))
	case errors.Is(err, ErrUnknownUnit):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unknown unit"))
	case errors.Is(err, ErrNotEnoughRes):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "not enough resources"))
	case errors.Is(err, ErrNoShipyard):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "shipyard required"))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	case errors.Is(err, ErrUmodeBlocked), errors.Is(err, ErrObserverBlocked):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	case errors.Is(err, ErrQueueFull):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, err.Error()))
	case errors.Is(err, ErrCapacityExceeded):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "shield/rocket capacity exceeded"))
	case requirements.IsNotMet(err):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// List GET /api/planets/{id}/shipyard/queue
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

// Cancel DELETE /api/planets/{id}/shipyard/{queueId}
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	queueID := chi.URLParam(r, "queueId")
	err := h.svc.Cancel(r.Context(), uid, planetID, queueID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrQueueItemNotFound):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, "queue item not found"))
	case errors.Is(err, ErrAlreadyDone):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "already completed"))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// StartVIP POST /api/planets/{id}/shipyard/{queueId}/vip (план 72.1.44).
//
// Legacy `EventHandler::startConstructionEventVIP` для UNIT_TYPE_FLEET/
// DEFENSE: мгновенный старт shipyard-задачи за credits.
func (h *Handler) StartVIP(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	queueID := chi.URLParam(r, "queueId")
	item, err := h.svc.StartVIP(r.Context(), uid, planetID, queueID)
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

// Capacity GET /api/planets/{id}/shipyard/capacity (план 72.1.41).
// Legacy `Shipyard` показывает freeShieldFields/freeRocketFields.
func (h *Handler) Capacity(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	cap, err := h.svc.Capacity(r.Context(), uid, planetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, cap)
}

// Inventory GET /api/planets/{id}/shipyard/inventory
// Возвращает inventory + costs (per-unit) + seconds (per-unit с учётом
// shipyard/nano уровней). Pixel-perfect клон legacy (план 72.1 ч.20.3).
func (h *Handler) Inventory(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	ships, defense, err := h.svc.Inventory(r.Context(), planetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	shipCosts, defCosts := h.svc.CostsMap()
	shipSecs, defSecs, err := h.svc.SecondsMap(r.Context(), planetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	// План 72.1.45 §5: damaged/shell_percent для UI inventory (legacy shipyard.tpl).
	shipsDmg, defenseDmg, err := h.svc.DamagedInventory(r.Context(), planetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"ships":         ships,
		"defense":       defense,
		"ship_costs":    shipCosts,
		"defense_costs": defCosts,
		"ship_seconds":  shipSecs,
		"defense_seconds": defSecs,
		"ships_damaged":   shipsDmg,
		"defense_damaged": defenseDmg,
	})
}
