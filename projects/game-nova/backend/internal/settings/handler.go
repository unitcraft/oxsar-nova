// Package settings реализует GET/PUT /api/settings — настройки аккаунта.
package settings

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/internal/i18n"
)

var validTimezones = map[string]bool{
	"UTC": true, "Europe/Moscow": true, "Europe/Kiev": true, "Europe/Minsk": true,
	"Asia/Yekaterinburg": true, "Asia/Novosibirsk": true, "Asia/Vladivostok": true,
	"Asia/Almaty": true, "Europe/Berlin": true, "Europe/London": true,
	"America/New_York": true, "America/Los_Angeles": true, "Asia/Tokyo": true,
}

type Handler struct {
	pool    *pgxpool.Pool
	automsg AutoMsgSender
	bundle  *i18n.Bundle
}

func NewHandler(pool *pgxpool.Pool) *Handler { return &Handler{pool: pool} }

func (h *Handler) WithBundle(b *i18n.Bundle) *Handler {
	h.bundle = b
	return h
}

func (h *Handler) tr(group, key string, vars map[string]string) string {
	if h.bundle == nil {
		return "[" + group + "." + key + "]"
	}
	return h.bundle.Tr(i18n.LangRu, group, key, vars)
}

type settingsResponse struct {
	Email        string  `json:"email"`
	Language     string  `json:"language"`
	Timezone     string  `json:"timezone"`
	VacationSince *string `json:"vacation_since"`
}

// Get GET /api/settings — текущие настройки.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	var email, language, timezone string
	var vacationSince *time.Time
	err := h.pool.QueryRow(r.Context(),
		`SELECT email, language, timezone, vacation_since FROM users WHERE id = $1`,
		uid,
	).Scan(&email, &language, &timezone, &vacationSince)
	if err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
		return
	}

	resp := settingsResponse{
		Email:    email,
		Language: language,
		Timezone: timezone,
	}
	if vacationSince != nil {
		s := vacationSince.UTC().Format(time.RFC3339)
		resp.VacationSince = &s
	}
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

type updateRequest struct {
	Email    *string `json:"email"`
	Language *string `json:"language"`
	Timezone *string `json:"timezone"`
}

// Update PUT /api/settings — обновить email, language, timezone.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	uid, ok := auth.UserID(r.Context())
	if !ok {
		httpx.WriteError(w, r, httpx.ErrUnauthorized)
		return
	}

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid json"))
		return
	}

	if req.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*req.Email))
		if !strings.Contains(email, "@") || len(email) < 3 {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "invalid email"))
			return
		}
		if err := h.setEmail(r.Context(), uid, email); err != nil {
			if strings.Contains(err.Error(), "unique") {
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrConflict, "email already taken"))
			} else {
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			}
			return
		}
	}

	if req.Language != nil {
		lang := *req.Language
		if lang != "ru" && lang != "en" {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unsupported language"))
			return
		}
		if _, err := h.pool.Exec(r.Context(),
			`UPDATE users SET language = $1 WHERE id = $2`, lang, uid); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
	}

	if req.Timezone != nil {
		tz := *req.Timezone
		if !validTimezones[tz] {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrBadRequest, "unsupported timezone"))
			return
		}
		if _, err := h.pool.Exec(r.Context(),
			`UPDATE users SET timezone = $1 WHERE id = $2`, tz, uid); err != nil {
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrInternal, err.Error()))
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// План 36 Critical-6: ChangePassword переехал в auth-service (POST /auth/password).
// Хеш пароля живёт в auth-db, в game-db password_hash IS NULL.
// Frontend дёргает auth-service напрямую через vite proxy /auth/password.

func (h *Handler) setEmail(ctx context.Context, uid, email string) error {
	tag, err := h.pool.Exec(ctx,
		`UPDATE users SET email = $1 WHERE id = $2`, email, uid)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}
