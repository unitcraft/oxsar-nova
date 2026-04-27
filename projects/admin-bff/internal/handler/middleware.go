package handler

import (
	"context"
	"crypto/subtle"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"oxsar/admin-bff/internal/httpx"
	"oxsar/admin-bff/internal/identityclient"
	"oxsar/admin-bff/internal/session"
)

type ctxKeySession struct{}

// SessionLookup — middleware: читает session ID из cookie, достаёт сессию
// из Redis, кладёт в context. При отсутствии — 401 (для protected routes).
//
// Refresh access-token выполняется тут же, если до exp осталось <
// refreshLeadTime. Это прозрачно для frontend.
func SessionLookup(
	store *session.Store,
	identity *identityclient.Client,
	refreshLeadTime time.Duration,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := session.SessionIDFromRequest(r)
			if id == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "no_session", "")
				return
			}
			sess, err := store.Get(r.Context(), id)
			if err != nil {
				if errors.Is(err, session.ErrNotFound) {
					httpx.WriteError(w, http.StatusUnauthorized, "session_expired", "")
					return
				}
				slog.ErrorContext(r.Context(), "session lookup failed", slog.String("err", err.Error()))
				httpx.WriteError(w, http.StatusInternalServerError, "session_lookup_failed", "")
				return
			}

			// Lazy refresh: если access близок к истечению, обновляем атомарно.
			if !sess.AccessTokenExp.IsZero() && time.Until(sess.AccessTokenExp) < refreshLeadTime {
				ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
				newTokens, err := identity.Refresh(ctx, sess.RefreshToken)
				cancel()
				if err != nil {
					if errors.Is(err, identityclient.ErrInvalidToken) {
						_ = store.Delete(r.Context(), id)
						httpx.WriteError(w, http.StatusUnauthorized, "refresh_invalid", "")
						return
					}
					slog.WarnContext(r.Context(), "token refresh failed",
						slog.String("err", err.Error()))
					// Продолжаем со старым токеном — backend сам вернёт 401 если истёк.
				} else {
					sess.AccessToken = newTokens.AccessToken
					sess.RefreshToken = newTokens.RefreshToken
					sess.AccessTokenExp = newTokens.AccessTokenExp
					sess.Claims = session.Claims{
						Subject:     newTokens.Subject,
						Username:    newTokens.Username,
						Roles:       newTokens.Roles,
						Permissions: newTokens.Permissions,
					}
				}
			}

			// Touch — обновляет LastSeenAt + продлевает sliding TTL.
			if err := store.Touch(r.Context(), sess); err != nil {
				slog.WarnContext(r.Context(), "session touch failed",
					slog.String("err", err.Error()))
			}

			ctx := context.WithValue(r.Context(), ctxKeySession{}, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CSRF — double-submit middleware для state-changing запросов.
// Frontend читает admin_csrf cookie (он не HttpOnly) и шлёт
// header X-CSRF-Token. Они должны совпадать.
//
// Применяется только к POST/PUT/PATCH/DELETE — GET/HEAD/OPTIONS пропускаются.
func CSRF() func(http.Handler) http.Handler {
	safeMethods := map[string]bool{
		http.MethodGet:     true,
		http.MethodHead:    true,
		http.MethodOptions: true,
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if safeMethods[r.Method] {
				next.ServeHTTP(w, r)
				return
			}
			sess, ok := SessionFromContext(r.Context())
			if !ok {
				httpx.WriteError(w, http.StatusUnauthorized, "no_session", "")
				return
			}
			header := r.Header.Get("X-CSRF-Token")
			if header == "" || subtle.ConstantTimeCompare([]byte(header), []byte(sess.CSRFToken)) != 1 {
				httpx.WriteError(w, http.StatusForbidden, "csrf_mismatch", "")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SessionFromContext — извлекает текущую сессию из request context.
func SessionFromContext(ctx context.Context) (*session.Session, bool) {
	sess, ok := ctx.Value(ctxKeySession{}).(*session.Session)
	return sess, ok
}

// ContextWithSession — кладёт сессию в context. Используется SessionLookup
// и тестами, которым нужен запрос с готовой сессией без поднятия Redis.
func ContextWithSession(ctx context.Context, sess *session.Session) context.Context {
	return context.WithValue(ctx, ctxKeySession{}, sess)
}
