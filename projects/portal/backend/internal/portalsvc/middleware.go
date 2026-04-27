package portalsvc

import (
	"context"
	"net/http"
	"strings"

	"oxsar/portal/internal/httpx"
	"oxsar/portal/pkg/jwtrs"
)

type ctxKey int

const (
	ctxUserID     ctxKey = 1
	ctxAuthorName ctxKey = 2
	ctxRoles      ctxKey = 3
)

// Middleware верифицирует RSA-256 JWT и кладёт userID + username в контекст.
func Middleware(ver *jwtrs.Verifier) func(http.Handler) http.Handler {
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
			ctx = context.WithValue(ctx, ctxAuthorName, claims.Username)
			ctx = context.WithValue(ctx, ctxRoles, claims.Roles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminMiddleware разрешает доступ только пользователям с ролью "admin".
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roles, _ := r.Context().Value(ctxRoles).([]string)
		for _, role := range roles {
			if role == "admin" {
				next.ServeHTTP(w, r)
				return
			}
		}
		httpx.WriteError(w, r, httpx.ErrForbidden)
	})
}

func userIDFromCtx(r *http.Request) (string, bool) {
	v, ok := r.Context().Value(ctxUserID).(string)
	return v, ok && v != ""
}

func authorNameFromCtx(r *http.Request) (string, bool) {
	v, ok := r.Context().Value(ctxAuthorName).(string)
	return v, ok && v != ""
}

// UserIDFromContext возвращает UUID-юзера, помещённый Middleware'ом в
// context. Экспортируется для пакетов вне portalsvc (план 56 — package
// report использует тот же auth-paypload).
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxUserID).(string)
	return v, ok && v != ""
}

// RolesFromContext возвращает роли пользователя из JWT (по умолчанию nil).
func RolesFromContext(ctx context.Context) []string {
	v, _ := ctx.Value(ctxRoles).([]string)
	return v
}
