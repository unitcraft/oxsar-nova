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
	svc     *Service
	db      *pgxpool.Pool
	vacation *VacationService
}

func NewHandler(svc *Service, db *pgxpool.Pool) *Handler {
	return &Handler{svc: svc, db: db, vacation: NewVacationService(db)}
}

type registerRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	User   User   `json:"user"`
	Tokens Tokens `json:"tokens"`
}

// Register POST /api/auth/register?ref=<userID>
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}
	in := RegisterInput{
		Username:   req.Username,
		Email:      req.Email,
		Password:   req.Password,
		ReferredBy: r.URL.Query().Get("ref"),
	}
	u, t, err := h.svc.Register(r.Context(), in)
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

// Me GET /api/me — возвращает user_id, username, role, credit, profession текущего пользователя.
// Требует Middleware (Bearer / ?token=).
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	var username, role, profession string
	var credit float64
	if err := h.db.QueryRow(context.Background(),
		`SELECT username, COALESCE(role::text, ''), credit, COALESCE(profession, 'none') FROM users WHERE id=$1`,
		uid,
	).Scan(&username, &role, &credit, &profession); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{
		"user_id":    uid,
		"username":   username,
		"role":       role,
		"credit":     credit,
		"profession": profession,
	})
}

// SetVacation POST /api/me/vacation — включить режим отпуска.
func (h *Handler) SetVacation(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	if err := h.vacation.SetVacation(r.Context(), uid); err != nil {
		switch {
		case errors.Is(err, ErrVacationAlreadyActive):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "vacation already active"))
		case errors.Is(err, ErrVacationIntervalNotMet):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "vacation interval not met (20 days)"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UnsetVacation DELETE /api/me/vacation — выйти из режима отпуска.
func (h *Handler) UnsetVacation(w http.ResponseWriter, r *http.Request) {
	uid, ok := UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}
	if err := h.vacation.UnsetVacation(r.Context(), uid); err != nil {
		switch {
		case errors.Is(err, ErrVacationNotActive):
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "vacation not active"))
		default:
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
