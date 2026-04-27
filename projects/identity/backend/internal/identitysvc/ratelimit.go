package identitysvc

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"oxsar/identity/internal/httpx"
)

// RateLimiter — IP-based rate-limiter поверх Redis (sliding-window через INCR + EXPIRE).
// Используется на /auth/login и /auth/register для защиты от brute-force.
//
// Алгоритм: для каждой пары (key, IP) ведём счётчик в Redis с TTL = window.
// Если на момент запроса счётчик >= limit — 429. Иначе INCR и пропускаем.
//
// Если Redis недоступен (rdb == nil или ошибка) — fail-open: пропускаем.
// Это сознательный trade-off: лучше пустить всех, чем заблокировать sign-in
// при инциденте Redis. Для production можно поменять на fail-closed.
type RateLimiter struct {
	rdb    *redis.Client
	prefix string
	limit  int
	window time.Duration
}

// NewRateLimiter создаёт IP-based лимитер.
//   - prefix — namespace ключей в Redis (`rl:login`, `rl:register`).
//   - limit — максимум запросов в окне.
//   - window — длительность окна.
func NewRateLimiter(rdb *redis.Client, prefix string, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{rdb: rdb, prefix: prefix, limit: limit, window: window}
}

// Middleware возвращает chi-совместимую middleware-функцию.
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if rl.rdb == nil {
				next.ServeHTTP(w, r)
				return
			}
			ip := clientIP(r)
			key := rl.prefix + ":" + ip
			ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
			defer cancel()

			// INCR + EXPIRE атомарно через pipeline.
			pipe := rl.rdb.TxPipeline()
			incr := pipe.Incr(ctx, key)
			pipe.Expire(ctx, key, rl.window)
			if _, err := pipe.Exec(ctx); err != nil {
				// fail-open
				next.ServeHTTP(w, r)
				return
			}
			count := incr.Val()
			if count > int64(rl.limit) {
				retryAfter := int(rl.window.Seconds())
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				httpx.WriteError(w, r, httpx.ErrRateLimit)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// clientIP извлекает IP из X-Forwarded-For (если за nginx/proxy), иначе RemoteAddr.
// Берёт первый IP из X-Forwarded-For (он самый «настоящий», далее цепочка proxy).
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return strings.TrimSpace(xr)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
