package auth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oxsar/nova/backend/internal/httpx"
)

// Handler — HTTP-адаптер к auth.Service.
type Handler struct {
	svc *Service
	db  *pgxpool.Pool
}

func NewHandler(svc *Service, db *pgxpool.Pool) *Handler { return &Handler{svc: svc, db: db} }

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

// Me GET /api/me — возвращает user_id и username текущего пользователя.
// Требует Middleware (Bearer / ?token=).
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var username, role string
	var credit float64
	if err := h.db.QueryRow(context.Background(),
		`SELECT username, COALESCE(role::text, ''), credit FROM users WHERE id=$1`, uid).Scan(&username, &role, &credit); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"user_id":  uid,
		"username": username,
		"role":     role,
		"credit":   credit,
	})
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
