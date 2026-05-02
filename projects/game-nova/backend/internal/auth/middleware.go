package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/economy"
	"oxsar/game-nova/internal/httpx"
	"oxsar/game-nova/pkg/jwtrs"
)

type ctxKey int

const (
	UserIDKey ctxKey = 1
	// План 36 Ф.12: lazy-create middleware читает username/email
	// из RSA-claims, чтобы зеркалить юзера в game-db без HTTP-вызова в identity-service.
	rsaClaimsKey ctxKey = 2
)

// RSAMiddleware верифицирует RSA-256 JWT, выпущенный Auth Service.
// Принимает токен либо в Authorization: Bearer, либо в query ?token=
// (для WebSocket-соединений, которые не могут слать custom headers).
//
// В context кладёт ОБА: userID и полные claims — чтобы EnsureUserMiddleware
// мог lazy-зеркалить юзера в game-db (нужен username/email из claims).
// План 36 Ф.12.
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
			ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
			ctx = context.WithValue(ctx, rsaClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RSAClaims достаёт RSA-claims, положенные RSAMiddleware. Возвращает nil
// и false если middleware не стоял (или был legacy HS256-вариант).
func RSAClaims(ctx context.Context) (*jwtrs.Claims, bool) {
	v, ok := ctx.Value(rsaClaimsKey).(*jwtrs.Claims)
	return v, ok && v != nil
}

// UserID достаёт идентификатор пользователя, положенный Middleware.
// Возвращает пустую строку и false, если middleware не стоял.
func UserID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(UserIDKey).(string)
	return v, ok
}

// LastSeenMiddleware обновляет users.last_seen_at = now() при каждом
// аутентифицированном запросе и начисляет ежедневный бонус кредитов
// (economy.CreditDailyLogin) если прошло ≥24 часа с последнего начисления.
// Выполняется асинхронно (fire-and-forget), чтобы не добавлять задержку.
//
// План 72.1.55.E (effects ipcheck): если у юзера ipcheck=true и
// IP запроса отличается от users.last_seen_ip — INSERT уведомление
// в `messages` (folder=11 system, legacy IP_CHECK_ALERT), затем
// UPDATE last_seen_ip. Не блокирует запрос (legacy тоже только
// уведомляет, не выкидывает сессию).
func LastSeenMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Capture IP до next.ServeHTTP — handler может изменить r.
			ip := clientIP(r)
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
				// План 72.1.55.E (effects ipcheck).
				if ip != "" {
					checkIPChange(ctx, pool, uid, ip)
				}
			}()
		})
	}
}

// clientIP возвращает IP клиента: X-Forwarded-For (за reverse-proxy)
// или RemoteAddr fallback. nginx/caddy в проде ставят X-Forwarded-For.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// XFF может быть chain: "real, proxy1, proxy2"; берём первый.
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}
	if ra := r.RemoteAddr; ra != "" {
		// RemoteAddr — "ip:port"; обрезаем порт.
		for i := len(ra) - 1; i >= 0; i-- {
			if ra[i] == ':' {
				return ra[:i]
			}
		}
		return ra
	}
	return ""
}

// checkIPChange сравнивает текущий IP с users.last_seen_ip; при
// расхождении и ipcheck=true → INSERT уведомление + UPDATE
// last_seen_ip. План 72.1.55.E.
func checkIPChange(ctx context.Context, pool *pgxpool.Pool, uid, ip string) {
	var prev *string
	var ipcheck bool
	if err := pool.QueryRow(ctx,
		`SELECT last_seen_ip, ipcheck FROM users WHERE id = $1`, uid,
	).Scan(&prev, &ipcheck); err != nil {
		return
	}
	if prev != nil && *prev != "" && *prev != ip && ipcheck {
		// INSERT system-message (folder=11 system messages) — простой
		// текст. AutoMsg-сервис не используется чтобы не добавлять
		// зависимость от него в auth-middleware.
		_, _ = pool.Exec(ctx, `
			INSERT INTO messages (id, to_user_id, from_user_id, folder, subject, body, created_at)
			VALUES (gen_random_uuid(), $1, NULL, 11,
			        'Вход с другого IP',
			        'Замечен вход в аккаунт с IP ' || $2 || ' (предыдущий: ' || $3 || '). Если это не вы — смените пароль.',
			        now())
		`, uid, ip, *prev)
	}
	// Всегда обновляем last_seen_ip (включая первый login).
	_, _ = pool.Exec(ctx,
		`UPDATE users SET last_seen_ip = $1 WHERE id = $2`, ip, uid)
}
