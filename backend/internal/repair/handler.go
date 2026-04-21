package repair

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
		errors.Is(err, ErrNoRepairBuilding):
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
	list, err := h.svc.List(r.Context(), planetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"queue": list})
}
