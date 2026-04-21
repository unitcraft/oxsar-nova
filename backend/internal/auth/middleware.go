package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/oxsar/nova/backend/internal/httpx"
)

type ctxKey int

const userIDKey ctxKey = 1

// Middleware проверяет Authorization: Bearer <access> и кладёт userID
// в контекст. При отсутствии токена возвращает 401.
func Middleware(j *JWTIssuer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if !strings.HasPrefix(h, "Bearer ") {
				httpx.WriteError(w, r, httpx.ErrUnauthorized)
				return
			}
			uid, err := j.Parse(strings.TrimPrefix(h, "Bearer "), "access")
			if err != nil {
				httpx.WriteError(w, r, httpx.ErrUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserID достаёт идентификатор пользователя, положенный Middleware.
// Возвращает пустую строку и false, если middleware не стоял.
func UserID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}
