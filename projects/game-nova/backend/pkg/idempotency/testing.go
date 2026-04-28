package idempotency

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCmdForTesting — экспорт интерфейса redisCmd для cross-package тестов.
// Не предназначен для использования в production-коде; нужен только чтобы
// integration-тесты в других пакетах могли подсунуть in-memory заглушку
// без поднятия miniredis или контейнера.
type RedisCmdForTesting interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
}

// NewMiddlewareWithCmdForTesting — конструктор middleware с произвольной
// реализацией redis-интерфейса. См. RedisCmdForTesting.
func NewMiddlewareWithCmdForTesting(rdb RedisCmdForTesting, ttl time.Duration) *Middleware {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &Middleware{rdb: rdb, ttl: ttl}
}
