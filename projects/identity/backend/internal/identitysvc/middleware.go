package identitysvc

import (
	"context"
	"net/http"
	"strings"

	"oxsar/identity/internal/httpx"
	"oxsar/identity/pkg/jwtrs"
)

type ctxKey int

const ctxUserID ctxKey = 1

// Middleware проверяет RSA-256 JWT и кладёт userID в контекст.
func Middleware(ver *jwtrs.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := ""
			if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
				token = strings.TrimPrefix(h, "Bearer ")
			} else if q := r.URL.Query().Get("token"); q != "" {
				token = q
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
			ctx := context.WithValue(r.Context(), ctxUserID, claims.Subject)
			// План 52 Ф.2: permissions из JWT в context — для permission-checks
			// в admin-handler-ах (RBAC). Используем typed key из rbac_handler.go.
			if len(claims.Permissions) > 0 {
				ctx = context.WithValue(ctx, ctxKeyPermissions{}, claims.Permissions)
			}
			// actor uuid — для audit-log в RBAC handler-ах.
			ctx = context.WithValue(ctx, ctxKeyUserID{}, claims.Subject)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func userIDFromCtx(r *http.Request) (string, bool) {
	v, ok := r.Context().Value(ctxUserID).(string)
	return v, ok && v != ""
}
