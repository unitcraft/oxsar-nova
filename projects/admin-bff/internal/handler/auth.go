// Package handler — HTTP-handlers для admin-bff.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"oxsar/admin-bff/internal/httpx"
	"oxsar/admin-bff/internal/identityclient"
	"oxsar/admin-bff/internal/session"
)

// Auth — handler для login/logout/me.
type Auth struct {
	identity     *identityclient.Client
	sessions     *session.Store
	cookieDomain string
	cookieSecure bool
	idleTimeout  time.Duration
}

func NewAuth(
	identity *identityclient.Client,
	sessions *session.Store,
	cookieDomain string,
	cookieSecure bool,
	idleTimeout time.Duration,
) *Auth {
	return &Auth{
		identity:     identity,
		sessions:     sessions,
		cookieDomain: cookieDomain,
		cookieSecure: cookieSecure,
		idleTimeout:  idleTimeout,
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type meResponse struct {
	Username    string   `json:"username"`
	Subject     string   `json:"sub"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	CSRFToken   string   `json:"csrf_token"`
}

// Login — принимает username+password, обращается к identity, создаёт
// сессию в Redis, ставит admin_session + admin_csrf cookies.
func (a *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_body", "malformed JSON")
		return
	}
	if req.Username == "" || req.Password == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing_credentials", "username and password required")
		return
	}

	tokens, err := a.identity.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, identityclient.ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "")
			return
		}
		slog.ErrorContext(r.Context(), "identity login failed", slog.String("err", err.Error()))
		httpx.WriteError(w, http.StatusBadGateway, "identity_unavailable", "")
		return
	}

	sess := session.Session{
		AccessToken:    tokens.AccessToken,
		RefreshToken:   tokens.RefreshToken,
		AccessTokenExp: tokens.AccessTokenExp,
		Claims: session.Claims{
			Subject:     tokens.Subject,
			Username:    tokens.Username,
			Roles:       tokens.Roles,
			Permissions: tokens.Permissions,
		},
		IP:        httpx.RemoteIP(r),
		UserAgent: r.UserAgent(),
	}
	id, csrf, err := a.sessions.Create(r.Context(), sess)
	if err != nil {
		slog.ErrorContext(r.Context(), "session create failed", slog.String("err", err.Error()))
		httpx.WriteError(w, http.StatusInternalServerError, "session_create_failed", "")
		return
	}

	session.SetCookies(w, id, csrf, a.cookieDomain, a.cookieSecure, a.idleTimeout)
	httpx.WriteJSON(w, http.StatusOK, meResponse{
		Username:    tokens.Username,
		Subject:     tokens.Subject,
		Roles:       tokens.Roles,
		Permissions: tokens.Permissions,
		CSRFToken:   csrf,
	})
}

// Logout — удаляет сессию из Redis, отзывает refresh-token в identity,
// чистит cookies.
func (a *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	id := session.SessionIDFromRequest(r)
	if id != "" {
		if sess, err := a.sessions.Get(r.Context(), id); err == nil {
			// Best-effort revoke в identity. Не блокируем при ошибке —
			// local-session всё равно удаляем.
			ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
			defer cancel()
			if err := a.identity.Logout(ctx, sess.RefreshToken); err != nil {
				slog.WarnContext(r.Context(), "identity logout failed",
					slog.String("err", err.Error()))
			}
		}
		_ = a.sessions.Delete(r.Context(), id)
	}
	session.ClearCookies(w, a.cookieDomain, a.cookieSecure)
	w.WriteHeader(http.StatusNoContent)
}

// Me — возвращает claims summary текущей сессии (для UI guards).
// 401 если нет валидной сессии.
func (a *Auth) Me(w http.ResponseWriter, r *http.Request) {
	sess, ok := SessionFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "no_session", "")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, meResponse{
		Username:    sess.Claims.Username,
		Subject:     sess.Claims.Subject,
		Roles:       sess.Claims.Roles,
		Permissions: sess.Claims.Permissions,
		CSRFToken:   sess.CSRFToken,
	})
}
