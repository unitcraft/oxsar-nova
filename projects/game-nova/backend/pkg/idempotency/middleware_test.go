package idempotency

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// fakeRedis — потокобезопасная in-memory заглушка минимального
// redis-интерфейса (Get/Set/SetNX). Достаточно для unit-тестов middleware
// без поднятия miniredis или контейнера.
type fakeRedis struct {
	mu   sync.Mutex
	data map[string]fakeEntry
	now  func() time.Time
}

type fakeEntry struct {
	value     string
	expiresAt time.Time
}

func newFakeRedis() *fakeRedis {
	return &fakeRedis{
		data: map[string]fakeEntry{},
		now:  time.Now,
	}
}

func (f *fakeRedis) get(key string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.data[key]
	if !ok {
		return "", false
	}
	if !e.expiresAt.IsZero() && f.now().After(e.expiresAt) {
		delete(f.data, key)
		return "", false
	}
	return e.value, true
}

func (f *fakeRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	v, ok := f.get(key)
	cmd := redis.NewStringCmd(ctx, "GET", key)
	if !ok {
		cmd.SetErr(redis.Nil)
		return cmd
	}
	cmd.SetVal(v)
	return cmd
}

func (f *fakeRedis) SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(ctx, "SETNX", key, value)
	f.mu.Lock()
	defer f.mu.Unlock()
	if e, ok := f.data[key]; ok {
		if e.expiresAt.IsZero() || f.now().Before(e.expiresAt) {
			cmd.SetVal(false)
			return cmd
		}
	}
	f.data[key] = fakeEntry{
		value:     toString(value),
		expiresAt: f.now().Add(expiration),
	}
	cmd.SetVal(true)
	return cmd
}

func (f *fakeRedis) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx, "SET", key, value)
	f.mu.Lock()
	defer f.mu.Unlock()
	exp := time.Time{}
	if expiration > 0 {
		exp = f.now().Add(expiration)
	}
	f.data[key] = fakeEntry{value: toString(value), expiresAt: exp}
	cmd.SetVal("OK")
	return cmd
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	default:
		return ""
	}
}

func newServer(mw *Middleware, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(mw.Wrap(handler))
}

func postJSON(t *testing.T, url, body, idemKey string) (int, string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if idemKey != "" {
		req.Header.Set("Idempotency-Key", idemKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post: %v", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(raw)
}

func TestMiddleware_NoHeader_PassesThrough(t *testing.T) {
	rdb := newFakeRedis()
	mw := newMiddlewareWithCmd(rdb, time.Hour)
	var calls int32
	srv := newServer(mw, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	defer srv.Close()

	for i := 0; i < 3; i++ {
		status, body := postJSON(t, srv.URL, `{"x":1}`, "")
		if status != 200 || body != `{"ok":true}` {
			t.Fatalf("got %d %q", status, body)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("calls = %d, want 3 (no caching without header)", got)
	}
}

func TestMiddleware_FirstRequest_StoresAndReplays(t *testing.T) {
	rdb := newFakeRedis()
	mw := newMiddlewareWithCmd(rdb, time.Hour)
	var calls int32
	srv := newServer(mw, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":42}`))
	})
	defer srv.Close()

	status1, body1 := postJSON(t, srv.URL, `{"x":1}`, "k-1")
	if status1 != 201 || body1 != `{"id":42}` {
		t.Fatalf("first: %d %q", status1, body1)
	}
	// Повторный запрос с тем же ключом и body — должен вернуть кеш без
	// повторного вызова handler'а.
	status2, body2 := postJSON(t, srv.URL, `{"x":1}`, "k-1")
	if status2 != 201 || body2 != `{"id":42}` {
		t.Fatalf("replay: %d %q", status2, body2)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1 (replay should use cache)", got)
	}
}

func TestMiddleware_DifferentBody_Returns409(t *testing.T) {
	rdb := newFakeRedis()
	mw := newMiddlewareWithCmd(rdb, time.Hour)
	srv := newServer(mw, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
	defer srv.Close()

	status1, _ := postJSON(t, srv.URL, `{"x":1}`, "k-2")
	if status1 != 200 {
		t.Fatalf("first status %d", status1)
	}
	status2, _ := postJSON(t, srv.URL, `{"x":2}`, "k-2")
	if status2 != http.StatusConflict {
		t.Errorf("second status = %d, want 409", status2)
	}
}

func TestMiddleware_TTL_Expires(t *testing.T) {
	rdb := newFakeRedis()
	// Имитируем «текущее время» — сдвинем после первой записи.
	clock := time.Now()
	rdb.now = func() time.Time { return clock }
	mw := newMiddlewareWithCmd(rdb, 100*time.Millisecond)
	var calls int32
	srv := newServer(mw, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	postJSON(t, srv.URL, `{}`, "k-3")
	// Сдвигаем «время» на 200ms вперёд — запись истекла.
	clock = clock.Add(200 * time.Millisecond)
	postJSON(t, srv.URL, `{}`, "k-3")
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2 (TTL expired between requests)", got)
	}
}

func TestMiddleware_NilRedis_PassesThrough(t *testing.T) {
	mw := NewMiddleware(nil, time.Hour)
	var calls int32
	srv := newServer(mw, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	postJSON(t, srv.URL, `{}`, "k-4")
	postJSON(t, srv.URL, `{}`, "k-4")
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("calls = %d, want 2 (no redis = no cache)", got)
	}
}

// errRedis — заглушка с принудительной ошибкой Get, для проверки fallback'а.
type errRedis struct{}

func (errRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	cmd := redis.NewStringCmd(ctx, "GET", key)
	cmd.SetErr(errors.New("redis down"))
	return cmd
}
func (errRedis) SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd {
	cmd := redis.NewBoolCmd(ctx, "SETNX", key, value)
	cmd.SetErr(errors.New("redis down"))
	return cmd
}
func (errRedis) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	cmd := redis.NewStatusCmd(ctx, "SET", key, value)
	cmd.SetErr(errors.New("redis down"))
	return cmd
}

func TestMiddleware_RedisFailure_FallsThrough(t *testing.T) {
	mw := newMiddlewareWithCmd(errRedis{}, time.Hour)
	var calls int32
	srv := newServer(mw, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	status, _ := postJSON(t, srv.URL, `{}`, "k-5")
	if status != 200 {
		t.Errorf("status = %d, want 200 (fall-through on redis error)", status)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("calls = %d, want 1", got)
	}
}
