package planet

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/oxsar/nova/backend/internal/auth"
	"github.com/oxsar/nova/backend/internal/httpx"
)

// Handler — HTTP-адаптер к planet.Service. Протектед auth-middleware'ом.
type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

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
