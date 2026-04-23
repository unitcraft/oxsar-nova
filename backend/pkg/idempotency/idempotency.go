// Package idempotency реализует дедупликацию POST-запросов по заголовку
// Idempotency-Key (RFC черновик, широко принятый стандарт).
//
// Использование в handler'е:
//
//	idem := idempotency.FromRequest(r, rdb)
//	if idem.Replay(w) { return }
//	// ... обычная логика ...
//	idem.Record(status, body)
package idempotency

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

const defaultTTL = 24 * time.Hour

// Entry — закэшированный результат операции.
type Entry struct {
	Status int    `json:"status"`
	Body   []byte `json:"body"`
}

// Key — контекст одного idempotency-ключа для конкретного запроса.
type Key struct {
	key string
	rdb *redis.Client
	ctx context.Context
}

// FromRequest возвращает Key для текущего запроса.
// Если rdb == nil или заголовок Idempotency-Key отсутствует — возвращает
// no-op Key (Replay всегда false, Record ничего не делает).
func FromRequest(r *http.Request, rdb *redis.Client) *Key {
	if rdb == nil {
		return &Key{}
	}
	k := r.Header.Get("Idempotency-Key")
	if k == "" {
		return &Key{}
	}
	return &Key{key: k, rdb: rdb, ctx: r.Context()}
}

// Replay проверяет кэш: если есть — пишет сохранённый ответ и возвращает true.
// Handler должен сразу вернуться.
func (k *Key) Replay(w http.ResponseWriter) bool {
	if k.rdb == nil || k.key == "" {
		return false
	}
	raw, err := k.rdb.Get(k.ctx, redisKey(k.key)).Bytes()
	if err != nil {
		return false
	}
	var e Entry
	if err := json.Unmarshal(raw, &e); err != nil {
		return false
	}
	if len(e.Body) > 0 {
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(e.Status)
	if len(e.Body) > 0 {
		_, _ = w.Write(e.Body)
	}
	return true
}

// Record сохраняет результат операции в Redis.
func (k *Key) Record(status int, body []byte) {
	if k.rdb == nil || k.key == "" {
		return
	}
	e := Entry{Status: status, Body: body}
	raw, err := json.Marshal(e)
	if err != nil {
		return
	}
	_ = k.rdb.Set(k.ctx, redisKey(k.key), raw, defaultTTL).Err()
}

func redisKey(key string) string { return "idem:" + key }
