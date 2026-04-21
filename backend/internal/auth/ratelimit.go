package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/oxsar/nova/backend/internal/httpx"
)

// RateLimiter — sliding-window rate limiter на Redis.
// Если Redis недоступен — пропускаем запрос (fail-open).
type RateLimiter struct {
	rdb     *redis.Client
	limit   int           // max requests
	window  time.Duration // per window
	keyFunc func(r *http.Request) string
}

// NewIPRateLimiter создаёт limiter по IP-адресу клиента.
// limit=10, window=1m — разумный дефолт для /api/auth/login.
func NewIPRateLimiter(rdb *redis.Client, limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rdb:    rdb,
		limit:  limit,
		window: window,
		keyFunc: func(r *http.Request) string {
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = r.RemoteAddr
			}
			return "ratelimit:" + ip
		},
	}
}

// Middleware возвращает chi-совместимый middleware.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if rl.rdb == nil {
			next.ServeHTTP(w, r)
			return
		}
		key := rl.keyFunc(r)
		exceeded, err := rl.check(r.Context(), key)
		if err != nil {
			// Redis ошибка — fail-open, пропускаем.
			next.ServeHTTP(w, r)
			return
		}
		if exceeded {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", rl.window.Seconds()))
			httpx.WriteError(w, r, httpx.Wrap(httpx.ErrRateLimit, "too many requests, try later"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// check инкрементирует счётчик и возвращает true если лимит превышен.
func (rl *RateLimiter) check(ctx context.Context, key string) (bool, error) {
	pipe := rl.rdb.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, rl.window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	return incr.Val() > int64(rl.limit), nil
}
