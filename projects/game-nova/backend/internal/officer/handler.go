package officer

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

// List GET /api/officers
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	list, err := h.svc.List(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"officers": list})
}

// Activate POST /api/officers/{key}/activate
func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	key := chi.URLParam(r, "key")
	if key == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "missing key"))
		return
	}
	var body struct {
		AutoRenew bool `json:"auto_renew"`
	}
	// body опционален — игнорируем ошибки декодирования (пустое тело → auto_renew=false)
	_ = json.NewDecoder(r.Body).Decode(&body)
	e, err := h.svc.Activate(r.Context(), uid, key, body.AutoRenew)
	switch {
	case err == nil:
		httpx.WriteJSON(w, r, http.StatusCreated, e)
	case errors.Is(err, ErrOfficerNotFound):
		httpx.WriteError(w, r, httpx.ErrNotFound)
	case errors.Is(err, ErrAlreadyActive),
		errors.Is(err, ErrGroupActive),
		errors.Is(err, ErrNotEnoughCredit):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}
