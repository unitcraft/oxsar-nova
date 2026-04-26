package planet

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

// Handler — HTTP-адаптер к planet.Service. Протектед auth-middleware'ом.
type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// Reorder PATCH /api/planets/order — обновить sort_order по списку planet_ids.
func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var req struct {
		PlanetIDs []string `json:"planet_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	if err := h.svc.Reorder(r.Context(), uid, req.PlanetIDs); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.WriteError(w, r, httpx.ErrForbidden)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// List GET /api/planets — все планеты текущего пользователя.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	ps, err := h.svc.ListByUser(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"planets": ps})
}

// Get GET /api/planets/{id}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	p, err := h.svc.Get(r.Context(), id)
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	if p.UserID != uid {
		httpx.WriteError(w, r, httpx.ErrForbidden)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, p)
}

// Rename PATCH /api/planets/{id} — переименовать планету.
func (h *Handler) Rename(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}

	err := h.svc.Rename(r.Context(), uid, id, body.Name)
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		if errors.Is(err, ErrInvalidInput) {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "renamed"})
}

// SetHome POST /api/planets/{id}/set-home — установить как домашнюю планету.
func (h *Handler) SetHome(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")

	err := h.svc.SetHome(r.Context(), uid, id)
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		if errors.Is(err, ErrMoonRestricted) {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "home set"})
}

// Abandon DELETE /api/planets/{id} — покинуть (удалить) планету.
func (h *Handler) Abandon(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")

	err := h.svc.Abandon(r.Context(), uid, id)
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		if errors.Is(err, ErrMoonRestricted) || errors.Is(err, ErrOnlyPlanet) || errors.Is(err, ErrCannotAbandonHome) {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "abandoned"})
}

// ResourceReport GET /api/planets/{id}/resource-report — отчёт о производстве ресурсов.
func (h *Handler) ResourceReport(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")

	report, err := h.svc.ResourceReport(r.Context(), uid, id)
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, report)
}

// Forecast GET /api/planets/{id}/forecast?hours=N
// План 17 G1. Прогноз ресурсов через N часов с учётом storage cap.
func (h *Handler) Forecast(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "id required"))
		return
	}
	hours := 4
	if v := r.URL.Query().Get("hours"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			hours = n
		}
	}
	res, err := h.svc.Forecast(r.Context(), id, hours)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, res)
}

// ResourceUpdate POST /api/planets/{id}/resource-update — обновить факторы производства.
func (h *Handler) ResourceUpdate(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")

	var body struct {
		Factors map[string]int `json:"factors"` // unit_id: factor %
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		return
	}

	err := h.svc.UpdateResourceFactors(r.Context(), uid, id, body.Factors)
	if err != nil {
		if err == ErrNotFound {
			httpx.WriteError(w, r, httpx.ErrNotFound)
			return
		}
		if errors.Is(err, ErrInvalidInput) {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
			return
		}
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "updated"})
}
