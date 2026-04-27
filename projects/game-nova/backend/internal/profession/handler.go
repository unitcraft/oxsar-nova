package profession

import (
	"encoding/json"
	"errors"
	"net/http"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(s *Service) *Handler { return &Handler{svc: s} }

// List GET /api/professions
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"professions": h.svc.List()})
}

// Get GET /api/professions/me
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	info, err := h.svc.Get(r.Context(), uid)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, info)
}

// Change POST /api/professions/me  body: {"profession": "miner"}
func (h *Handler) Change(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var body struct {
		Profession string `json:"profession"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Profession == "" {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "profession required"))
		return
	}
	err := h.svc.Change(r.Context(), uid, body.Profession)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrUnknownProfession):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	case errors.Is(err, ErrNotEnoughCredit), errors.Is(err, ErrChangeTooSoon):
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
	default:
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
	}
}
