// Package storage инкапсулирует подключения к БД и кешу.
package storage

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/identity, oxsar/portal и oxsar/billing. При любом изменении
// синхронизируйте КОПИИ:
//   - projects/game-nova/backend/internal/storage/postgres.go
//   - projects/identity/backend/internal/storage/postgres.go
//   - projects/portal/backend/internal/storage/postgres.go
//   - projects/billing/backend/internal/storage/postgres.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// startupRetries — сколько попыток ping'a делать при старте. Полезно,
// когда Docker embedded DNS ещё не warmed up или pg стартует медленно
// (холодный том, большой buffer pool). Полная выдержка:
// 0.5+1+2+4+8+16 ≈ 31s.
const startupRetries = 6

// OpenPostgres возвращает pgxpool, настроенный на контекст приложения.
// Вызывающий обязан вызвать pool.Close() при остановке. При первом
// ping'e применяется ретрай с экспоненциальной задержкой — чтобы не
// падать от ленивого DNS/пустой pg buffer pool.
func OpenPostgres(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parse pg url: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("new pg pool: %w", err)
	}

	var lastErr error
	delay := 500 * time.Millisecond
	for attempt := 1; attempt <= startupRetries; attempt++ {
		if err := pool.Ping(ctx); err == nil {
			return pool, nil
		} else {
			lastErr = err
			slog.Warn("pg ping failed, retrying",
				slog.Int("attempt", attempt),
				slog.String("err", err.Error()))
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("ping pg aborted: %w", ctx.Err())
		case <-time.After(delay):
			delay *= 2
		}
	}
	pool.Close()
	return nil, fmt.Errorf("ping pg after %d attempts: %w", startupRetries, lastErr)
}

// OpenRedis возвращает redis-клиент, ping'нутый на доступность.
// Такой же ретрай, как у Postgres.
func OpenRedis(ctx context.Context, url string) (*redis.Client, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	c := redis.NewClient(opt)

	var lastErr error
	delay := 500 * time.Millisecond
	for attempt := 1; attempt <= startupRetries; attempt++ {
		if err := c.Ping(ctx).Err(); err == nil {
			return c, nil
		} else {
			lastErr = err
			slog.Warn("redis ping failed, retrying",
				slog.Int("attempt", attempt),
				slog.String("err", err.Error()))
		}
		select {
		case <-ctx.Done():
			_ = c.Close()
			return nil, fmt.Errorf("ping redis aborted: %w", ctx.Err())
		case <-time.After(delay):
			delay *= 2
		}
	}
	_ = c.Close()
	return nil, fmt.Errorf("ping redis after %d attempts: %w", startupRetries, lastErr)
}
