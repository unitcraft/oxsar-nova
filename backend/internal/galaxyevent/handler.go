package galaxyevent

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/httpx"
)

// Handler — HTTP-адаптер.
//
// Public:
//   GET  /api/galaxy-event           — текущее событие или 204 No Content
//
// Admin (план 14 + 17 F):
//   POST   /api/admin/galaxy-events  body: {kind, duration_hours, params}
//   DELETE /api/admin/galaxy-events/{id}
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// Active GET /api/galaxy-event
func (h *Handler) Active(w http.ResponseWriter, r *http.Request) {
	e, err := h.svc.Active(r.Context())
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusOK, e)
	case errors.Is(err, ErrNoActive):
		w.WriteHeader(http.StatusNoContent)
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}

// Create POST /api/admin/galaxy-events
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Kind          string         `json:"kind"`
		DurationHours int            `json:"duration_hours"`
		Params        map[string]any `json:"params"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	e, err := h.svc.Create(r.Context(), body.Kind, body.DurationHours, body.Params)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusCreated, e)
}

// Cancel DELETE /api/admin/galaxy-events/{id}
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid id"))
		return
	}
	if err := h.svc.Cancel(r.Context(), id); err != nil {
		if errors.Is(err, ErrNoActive) {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
