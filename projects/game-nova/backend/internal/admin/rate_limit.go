package admin

import (
	"net/http"
	"sync"
	"time"

	"oxsar/game-nova/internal/auth"
	"oxsar/game-nova/internal/httpx"
)

// План 14 Ф.8.2 — rate-limit write-действий админа.
//
// Защита от human-error: «разом забанил 1000 аккаунтов». Лимит
// по умолчанию: 100 write-действий в час на одного админа. Счётчик
// in-memory — перезапуск сервера обнуляет. Для продакшна этого
// достаточно, так как админов мало и лимит — soft-safeguard, а не
// anti-DDoS.
//
// Write-методы: POST, PUT, PATCH, DELETE. GET не ограничивается.
//
// При превышении: 429 Too Many Requests + заголовок Retry-After.

const (
	// AdminRateLimitWrites — максимум write-запросов за окно.
	AdminRateLimitWrites = 100
	// AdminRateLimitWindow — окно.
	AdminRateLimitWindow = time.Hour
)

type rateCounter struct {
	count    int
	windowAt time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	perAdmin map[string]*rateCounter
}

var adminLimiter = &rateLimiter{perAdmin: map[string]*rateCounter{}}

// RateLimitMiddleware — один инстанс на процесс (shared). Закрывает
// write-действия. GET проходит без учёта.
func RateLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isWriteMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			uid, ok := auth.UserID(r.Context())
			if !ok {
				// Без авторизации — пусть дальше разбирается auth-слой.
				next.ServeHTTP(w, r)
				return
			}
			if !adminLimiter.allow(uid) {
				retry := int(AdminRateLimitWindow.Seconds())
				w.Header().Set("Retry-After", itoa(retry))
				httpx.WriteError(w, r, httpx.Wrap(httpx.ErrRateLimit,
					"admin rate limit exceeded (100 writes/hour); try later"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// allow — inc counter if under limit. Сбрасывает окно автоматически.
func (rl *rateLimiter) allow(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	c, ok := rl.perAdmin[userID]
	if !ok || now.Sub(c.windowAt) > AdminRateLimitWindow {
		rl.perAdmin[userID] = &rateCounter{count: 1, windowAt: now}
		return true
	}
	if c.count >= AdminRateLimitWrites {
		return false
	}
	c.count++
	return true
}

// --- helpers ---

func isWriteMethod(m string) bool {
	switch m {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	}
	return false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
