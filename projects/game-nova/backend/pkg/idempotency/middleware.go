// Idempotency-middleware (план 77 Ф.2).
//
// Ортогонален существующему FromRequest/Replay/Record helper'у
// (idempotency.go) — middleware применяется на уровне роута
// (mux.With(mw.Wrap)), helper — внутри handler'а. Оба используют один и
// тот же Redis-namespace ("idem:<key>"), поэтому совместимы и не
// дублируют запись.
//
// Дополнительно к helper'у middleware:
//   - Хеширует body (SHA-256) и хранит хеш вместе с закэшированным ответом.
//   - При повторном Idempotency-Key + другом body отвечает 409 (баг клиента,
//     см. R9 / RFC черновик Idempotency-Key).
//   - Атомарно резервирует ключ через SET NX, чтобы исключить race между
//     параллельными запросами с одним ключом.
package idempotency

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// pendingTTL — короткая блокировка ключа на время выполнения handler'а.
	// Защита от race condition между параллельными запросами с одним ключом.
	pendingTTL = 30 * time.Second

	// pendingMarker — value, который кладётся в Redis при резервировании
	// ключа. Когда handler выполнен, marker заменяется на полный Entry с body.
	pendingMarker = "__pending__"

	// maxBodyBytes — максимальный размер body для хеширования и кеширования.
	// 1 MiB — с большим запасом для любых JSON-payload'ов в game-nova.
	maxBodyBytes = 1 << 20
)

// MiddlewareEntry — расширение Entry с body-hash для middleware-уровня.
//
// Хранится в Redis отдельным префиксом "idem-mw:<key>" чтобы не путаться с
// helper-уровневым "idem:<key>" (тот не пишет body-hash).
type MiddlewareEntry struct {
	BodyHash string `json:"body_hash"`
	Status   int    `json:"status"`
	Body     []byte `json:"body"`
}

// redisCmd — минимальный интерфейс операций redis, нужных middleware.
// Позволяет в unit-тестах подменить *redis.Client на in-memory заглушку
// без поднятия miniredis/redis-контейнера. Production-код вызывает
// NewMiddleware с обычным *redis.Client (он удовлетворяет интерфейсу).
type redisCmd interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
}

// Middleware — chi-style HTTP middleware для дедупа мутирующих запросов
// по заголовку Idempotency-Key.
type Middleware struct {
	rdb redisCmd
	ttl time.Duration
}

// NewMiddleware создаёт middleware. ttl=0 → defaultTTL (24h, как в helper'е).
// rdb=nil → middleware отдаёт запрос напрямую без кеширования (для unit-тестов
// без redis или dev без redis).
func NewMiddleware(rdb *redis.Client, ttl time.Duration) *Middleware {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	if rdb == nil {
		return &Middleware{ttl: ttl}
	}
	return &Middleware{rdb: rdb, ttl: ttl}
}

// Wrap оборачивает handler в idempotency-проверку.
//
// Логика:
//  1. Header нет → handler вызывается без кеша.
//  2. Header есть, ключ ещё не резервировался → SETNX резерв + handler;
//     результат пишется в Redis с TTL.
//  3. Header есть, есть pending-marker → handler вызывается без кеша
//     (защита от потерянных записей; повторный SETNX проиграет race).
//  4. Header есть, кеш есть, body совпадает → возвращается кешированный ответ.
//  5. Header есть, кеш есть, body отличается → 409 Conflict.
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" || m.rdb == nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		body, err := readBody(r)
		if err != nil {
			http.Error(w, "request body too large or unreadable", http.StatusRequestEntityTooLarge)
			return
		}
		// После чтения тела восстанавливаем r.Body для downstream handler'а.
		r.Body = io.NopCloser(bytes.NewReader(body))
		bodyHash := hashBody(body)

		mwKey := mwRedisKey(key)
		raw, err := m.rdb.Get(ctx, mwKey).Bytes()
		switch {
		case errors.Is(err, redis.Nil):
			// Ключ ещё не использовался — пытаемся зарезервировать.
			ok, setErr := m.rdb.SetNX(ctx, mwKey, pendingMarker, pendingTTL).Result()
			if setErr != nil {
				next.ServeHTTP(w, r)
				return
			}
			if !ok {
				// Кто-то параллельно зарезервировал — пропускаем без кеширования.
				next.ServeHTTP(w, r)
				return
			}
			rec := &recordingWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			m.persist(ctx, mwKey, bodyHash, rec)
		case err != nil:
			// Redis-сбой: не блокируем игрока, выполняем handler без кеша.
			next.ServeHTTP(w, r)
			return
		default:
			// Запись есть. Может быть pending (handler в полёте у соседа) или
			// полный Entry.
			if string(raw) == pendingMarker {
				next.ServeHTTP(w, r)
				return
			}
			var e MiddlewareEntry
			if err := json.Unmarshal(raw, &e); err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if e.BodyHash != bodyHash {
				http.Error(w, "Idempotency-Key reuse with different body", http.StatusConflict)
				return
			}
			if len(e.Body) > 0 {
				w.Header().Set("Content-Type", "application/json")
			}
			w.WriteHeader(e.Status)
			if len(e.Body) > 0 {
				_, _ = w.Write(e.Body)
			}
		}
	})
}

// newMiddlewareWithCmd — фабрика для тестов: принимает любой redisCmd.
func newMiddlewareWithCmd(rdb redisCmd, ttl time.Duration) *Middleware {
	if ttl <= 0 {
		ttl = defaultTTL
	}
	return &Middleware{rdb: rdb, ttl: ttl}
}

func (m *Middleware) persist(ctx context.Context, mwKey, bodyHash string, rec *recordingWriter) {
	e := MiddlewareEntry{
		BodyHash: bodyHash,
		Status:   rec.status,
		Body:     rec.buf.Bytes(),
	}
	raw, err := json.Marshal(e)
	if err != nil {
		return
	}
	_ = m.rdb.Set(ctx, mwKey, raw, m.ttl).Err()
}

func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return nil, nil
	}
	return io.ReadAll(io.LimitReader(r.Body, maxBodyBytes))
}

func hashBody(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func mwRedisKey(key string) string { return "idem-mw:" + key }

// recordingWriter — http.ResponseWriter, перехватывающий status и body для
// последующего сохранения в Redis. Body буферизуется полностью; этого
// достаточно для JSON-ответов в game-nova (десятки KB, не мегабайты).
type recordingWriter struct {
	http.ResponseWriter
	status      int
	buf         bytes.Buffer
	wroteHeader bool
}

func (rw *recordingWriter) WriteHeader(status int) {
	rw.status = status
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *recordingWriter) Write(p []byte) (int, error) {
	if !rw.wroteHeader {
		rw.status = http.StatusOK
		rw.wroteHeader = true
	}
	rw.buf.Write(p)
	return rw.ResponseWriter.Write(p)
}
