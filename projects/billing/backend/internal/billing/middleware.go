package billing

import (
	"context"
	"net/http"
	"strings"

	"oxsar/billing/internal/httpx"
	"oxsar/billing/pkg/jwtrs"
)

type ctxKey int

const (
	ctxUserID    ctxKey = 1
	ctxUsername  ctxKey = 2
	ctxRoles     ctxKey = 3
)

// AuthMiddleware верифицирует RSA-256 JWT, выпущенный auth-service.
// Принимает токен в Authorization: Bearer <jwt>.
func AuthMiddleware(ver *jwtrs.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
				token = strings.TrimPrefix(h, "Bearer ")
			}
			if token == "" {
				httpx.WriteError(w, r, httpx.ErrUnauthorized)
				return
			}
			claims, err := ver.Parse(token, "access")
			if err != nil {
				httpx.WriteError(w, r, httpx.ErrUnauthorized)
				return
			}
			ctx := r.Context()
			ctx = context.WithValue(ctx, ctxUserID, claims.Subject)
			ctx = context.WithValue(ctx, ctxUsername, claims.Username)
			ctx = context.WithValue(ctx, ctxRoles, claims.Roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserIDFromCtx достаёт user_id из контекста (для idempotency middleware
// и handlers).
func UserIDFromCtx(r *http.Request) (string, bool) {
	v, ok := r.Context().Value(ctxUserID).(string)
	return v, ok && v != ""
}

// HasRole проверяет наличие роли (admin и т.п.).
func HasRole(r *http.Request, role string) bool {
	roles, _ := r.Context().Value(ctxRoles).([]string)
	for _, x := range roles {
		if x == role {
			return true
		}
	}
	return false
}
