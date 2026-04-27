package identitysvc

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// JTIBlacklist хранит ревокированные jti в Redis. План 36 Critical-4.
//
// Ключ — `revoked:<jti>`, значение — пустая строка, TTL = оставшееся время
// жизни токена. После истечения TTL запись удаляется автоматически (Redis EXPIRE).
type JTIBlacklist struct {
	rdb *redis.Client
}

// NewJTIBlacklist создаёт blacklist. Если rdb=nil, методы становятся no-op
// (revoke ничего не делает, IsRevoked всегда false). Это нужно для dev,
// когда Redis недоступен.
func NewJTIBlacklist(rdb *redis.Client) *JTIBlacklist {
	return &JTIBlacklist{rdb: rdb}
}

// Revoke помечает jti отозванным до момента expiresAt.
func (b *JTIBlacklist) Revoke(ctx context.Context, jti string, expiresAt time.Time) error {
	if b.rdb == nil {
		return nil
	}
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Токен уже истёк — добавлять в blacklist бессмысленно.
		return nil
	}
	return b.rdb.Set(ctx, "revoked:"+jti, "", ttl).Err()
}

// IsRevoked проверяет, есть ли jti в blacklist. При ошибке Redis возвращает
// (false, err) — fail-open: если Redis сломан, верифицируем как обычно.
// Production-режим может предпочесть fail-closed (возвращать true при ошибке).
func (b *JTIBlacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	if b.rdb == nil {
		return false, nil
	}
	n, err := b.rdb.Exists(ctx, "revoked:"+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
