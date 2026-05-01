package repair

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

type disassembleRequest struct {
	UnitID int   `json:"unit_id"`
	Count  int64 `json:"count"`
}

// Enqueue POST /api/planets/{id}/repair/disassemble — поставить юниты
// в очередь на разбор. Только здоровые (damaged=0).
func (h *Handler) EnqueueDisassemble(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	if planetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing planet id"))
		return
	}
	var req disassembleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	q, err := h.svc.EnqueueDisassemble(r.Context(), uid, planetID, req.UnitID, req.Count)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, q)
	case errors.Is(err, ErrInvalidCount),
		errors.Is(err, ErrUnknownUnit),
		errors.Is(err, ErrNotEnoughRes),
		errors.Is(err, ErrNotEnoughShips),
		errors.Is(err, ErrNoRepairBuilding),
		errors.Is(err, ErrInVacation),
		errors.Is(err, ErrIsObserver),
		errors.Is(err, ErrPlanetUnderAttack),
		errors.Is(err, ErrDockOverflow):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

type repairRequest struct {
	UnitID int `json:"unit_id"`
}

// EnqueueRepair POST /api/planets/{id}/repair/repair — починить
// damaged-юнитов заданного типа (всех разом). Стоимость считается
// по legacy-формуле.
func (h *Handler) EnqueueRepair(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	if planetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing planet id"))
		return
	}
	var req repairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	q, err := h.svc.EnqueueRepair(r.Context(), uid, planetID, req.UnitID)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, q)
	case errors.Is(err, ErrUnknownUnit),
		errors.Is(err, ErrNotEnoughRes),
		errors.Is(err, ErrNoRepairBuilding),
		errors.Is(err, ErrNothingToRepair),
		errors.Is(err, ErrInVacation),
		errors.Is(err, ErrIsObserver),
		errors.Is(err, ErrPlanetUnderAttack):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// StartVIP POST /api/planets/{id}/repair/queue/{queueId}/vip — мгновенный
// старт за credit (legacy `Repair.class.php::startEventVIP`).
func (h *Handler) StartVIP(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	queueID := chi.URLParam(r, "queueId")
	if planetID == "" || queueID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing id"))
		return
	}
	res, err := h.svc.StartVIP(r.Context(), uid, planetID, queueID)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, res)
	case errors.Is(err, ErrQueueItemNotFound):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, err.Error()))
	case errors.Is(err, ErrAlreadyDone),
		errors.Is(err, ErrNotEnoughCredit):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// ListDamaged GET /api/planets/{id}/repair/damaged
func (h *Handler) ListDamaged(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	if planetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing planet id"))
		return
	}
	list, err := h.svc.ListDamaged(r.Context(), planetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"damaged": list})
}

// Cancel DELETE /api/planets/{id}/repair/queue/{queueId}
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	queueID := chi.URLParam(r, "queueId")
	if planetID == "" || queueID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing id"))
		return
	}
	err := h.svc.Cancel(r.Context(), uid, planetID, queueID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrQueueItemNotFound):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrNotFound, err.Error()))
	case errors.Is(err, ErrAlreadyDone):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// List GET /api/planets/{id}/repair/queue
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	if planetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing planet id"))
		return
	}
	resp, err := h.svc.List(r.Context(), planetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}
