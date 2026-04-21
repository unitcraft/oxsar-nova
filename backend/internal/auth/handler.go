package auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/oxsar/nova/backend/internal/httpx"
)

// Handler — HTTP-адаптер к auth.Service.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User   User   `json:"user"`
	Tokens Tokens `json:"tokens"`
}

// Register POST /api/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	u, t, err := h.svc.Register(r.Context(), RegisterInput(req))
	if err != nil {
		switch {
		case errors.Is(err, ErrUserExists):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "user exists"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, err.Error()))
		}
		return
	}
	httpx.WriteJSON(w, r, http.StatusCreated, authResponse{User: u, Tokens: t})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login POST /api/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	u, t, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, authResponse{User: u, Tokens: t})
}

type refreshRequest struct {
	Refresh string `json:"refresh"`
}

// Refresh POST /api/auth/refresh
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	t, err := h.svc.Refresh(req.Refresh)
	if err != nil {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]Tokens{"tokens": t})
}
