package rocket

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/galaxy"
	"github.com/oxsar/nova/backend/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

type launchRequest struct {
	Dst          galaxy.Coords `json:"dst"`
	Count        int64         `json:"count"`
	TargetUnitID int           `json:"target_unit_id"` // 0 = без приоритета
}

// Launch POST /api/planets/{id}/rockets/launch
func (h *Handler) Launch(w http.ResponseWriter, r *http.Request) {
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
	var req launchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	res, err := h.svc.Launch(r.Context(), uid, planetID, req.Dst, req.Count, req.TargetUnitID)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, res)
	case errors.Is(err, ErrInvalidInput),
		errors.Is(err, ErrNoRockets),
		errors.Is(err, ErrTargetNotFound),
		errors.Is(err, ErrSiloLimit):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrPlanetOwnership):
		httpx.WriteError(w, r, httpx.ErrForbidden)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Stock GET /api/planets/{id}/rockets
func (h *Handler) Stock(w http.ResponseWriter, r *http.Request) {
	if _, ok := auth.UserID(r.Context()); !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	planetID := chi.URLParam(r, "id")
	if planetID == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing planet id"))
		return
	}
	n, err := h.svc.Stock(r.Context(), planetID)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]int64{"count": n})
}
