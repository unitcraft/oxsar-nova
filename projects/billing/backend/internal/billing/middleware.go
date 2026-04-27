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
	ctxUserID      ctxKey = 1
	ctxUsername    ctxKey = 2
	ctxRoles       ctxKey = 3
	ctxPermissions ctxKey = 4
)

// AuthMiddleware верифицирует RSA-256 JWT, выпущенный identity-service
// (план 51: rename auth → identity).
//
// План 52: дополнительно кладёт permissions из claims в context для
// permission-based проверок через RequirePermission.
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
			ctx = context.WithValue(ctx, ctxPermissions, claims.Permissions)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission — фабрика middleware, возвращает middleware,
// пропускающий запрос только если у юзера в JWT есть указанный permission.
//
// Применение для admin-endpoints billing-сервиса (план 54):
//
//	r.With(billing.AuthMiddleware(ver)).
//	  With(billing.RequirePermission("billing:read")).
//	  Get("/api/admin/billing/payments", h.ListPayments)
//
// Если permission отсутствует — 403. Использовать после AuthMiddleware,
// иначе context будет пустой и permission-check всегда возвратит false.
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			perms, _ := r.Context().Value(ctxPermissions).([]string)
			for _, p := range perms {
				if p == permission {
					next.ServeHTTP(w, r)
					return
				}
			}
			httpx.WriteError(w, r, httpx.ErrForbidden)
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
//
// Deprecated: используйте RequirePermission для гранулярных проверок.
// Оставлено как backward-compat для существующих handler-ов.
func HasRole(r *http.Request, role string) bool {
	roles, _ := r.Context().Value(ctxRoles).([]string)
	for _, x := range roles {
		if x == role {
			return true
		}
	}
	return false
}

// PermissionsFromCtx достаёт permissions из контекста (для inline-проверок
// в handler-ах, не через middleware).
func PermissionsFromCtx(r *http.Request) []string {
	v, _ := r.Context().Value(ctxPermissions).([]string)
	return v
}
