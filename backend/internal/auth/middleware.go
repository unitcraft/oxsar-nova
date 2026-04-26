package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oxsar/nova/backend/internal/economy"
	"github.com/oxsar/nova/backend/internal/httpx"
	"github.com/oxsar/nova/backend/pkg/jwtrs"
)

type ctxKey int

const userIDKey ctxKey = 1

// Middleware проверяет Authorization: Bearer <access> и кладёт userID
// в контекст. При отсутствии токена возвращает 401.
// Для WebSocket-соединений (которые не могут слать custom headers) принимает
// токен через query-param ?token=<access>.
func Middleware(j *JWTIssuer) func(http.Handler) http.Handler {
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
			uid, err := j.Parse(token, "access")
			if err != nil {
				httpx.WriteError(w, r, httpx.ErrUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RSAMiddleware — аналог Middleware, но верифицирует RSA-256 токены
// выданные Auth Service. Используется когда AUTH_JWKS_URL задан.
func RSAMiddleware(ver *jwtrs.Verifier) func(http.Handler) http.Handler {
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
			ctx := context.WithValue(r.Context(), userIDKey, claims.Subject)
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

// LastSeenMiddleware обновляет users.last_seen_at = now() при каждом
// аутентифицированном запросе и начисляет ежедневный бонус кредитов
// (economy.CreditDailyLogin) если прошло ≥24 часа с последнего начисления.
// Выполняется асинхронно (fire-and-forget), чтобы не добавлять задержку.
func LastSeenMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			uid, ok := UserID(r.Context())
			if !ok || uid == "" {
				return
			}
			go func() {
				ctx := context.Background()
				_, _ = pool.Exec(ctx,
					`UPDATE users SET last_seen_at = now() WHERE id = $1`, uid)
				// Ежедневный бонус: начислять если last_daily_credit_at IS NULL
				// или прошло ≥24 часа.
				_, _ = pool.Exec(ctx, `
					UPDATE users
					SET credit = credit + $1,
					    last_daily_credit_at = now()
					WHERE id = $2
					  AND (last_daily_credit_at IS NULL
					       OR last_daily_credit_at < now() - interval '24 hours')
				`, economy.CreditDailyLogin, uid)
			}()
		})
	}
}
