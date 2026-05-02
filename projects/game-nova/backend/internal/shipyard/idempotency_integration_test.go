package shipyard_test

// Регрессионный тест плана 72.1.49 — anti-double-submit для
// `POST /api/planets/{id}/shipyard` через idempotency-middleware.
//
// Без БД: используется in-memory fakeRedis (как в
// `internal/billing/client/integration_test.go`), shipyard-handler
// заменён заглушкой, которая инкрементирует счётчик «как если бы
// в БД появилась запись». Покрытие настоящего chi-роутинга и
// настоящего idempotency middleware.

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"

	"oxsar/game-nova/pkg/idempotency"
)

// fakeRedis — мини-реализация redis-интерфейса, требуемого middleware.
// Копия паттерна из internal/billing/client/integration_test.go
// (отдельный package, поэтому не получится переиспользовать).
type fakeRedis struct {
	mu   sync.Mutex
	data map[string]string
}

func newFakeRedis() *fakeRedis { return &fakeRedis{data: map[string]string{}} }

func (f *fakeRedis) Get(ctx context.Context, key string) *redis.StringCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := redis.NewStringCmd(ctx, "GET", key)
	v, ok := f.data[key]
	if !ok {
		cmd.SetErr(redis.Nil)
		return cmd
	}
	cmd.SetVal(v)
	return cmd
}

func (f *fakeRedis) SetNX(ctx context.Context, key string, value any, _ time.Duration) *redis.BoolCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := redis.NewBoolCmd(ctx, "SETNX", key, value)
	if _, ok := f.data[key]; ok {
		cmd.SetVal(false)
		return cmd
	}
	switch v := value.(type) {
	case string:
		f.data[key] = v
	case []byte:
		f.data[key] = string(v)
	}
	cmd.SetVal(true)
	return cmd
}

func (f *fakeRedis) Set(ctx context.Context, key string, value any, _ time.Duration) *redis.StatusCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := redis.NewStatusCmd(ctx, "SET", key, value)
	switch v := value.(type) {
	case string:
		f.data[key] = v
	case []byte:
		f.data[key] = string(v)
	}
	cmd.SetVal("OK")
	return cmd
}

// fakeEnqueue — заглушка shipyardH.Enqueue. В проде этот handler
// делает INSERT в shipyard_queue. Здесь — просто инкремент счётчика
// и фиксированный 201 ответ, чтобы middleware-кеш мог сработать.
func fakeEnqueue(calls *int32, response string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(calls, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(response))
	}
}

func postShipyard(t *testing.T, url, body, idemKey string) (int, string) {
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

// TestIdempotency_ShipyardEnqueue_Dedupe — повторный POST с тем же
// Idempotency-Key и тем же body не должен вызывать handler второй раз;
// клиент получает закешированный ответ. Регрессия gap'а из плана 72.1.2:
// без middleware двойной клик «Построить» создавал две записи в очереди.
func TestIdempotency_ShipyardEnqueue_Dedupe(t *testing.T) {
	rdb := newFakeRedis()
	mw := idempotency.NewMiddlewareWithCmdForTesting(rdb, time.Hour)

	var calls int32
	r := chi.NewRouter()
	r.With(mw.Wrap).Post("/api/planets/{id}/shipyard", fakeEnqueue(&calls, `{"queue_id":"q-1","unit_id":204,"count":10}`))
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := srv.URL + "/api/planets/p-1/shipyard"
	body := `{"unit_id":204,"count":10}`
	key := "shipyard:p-1:user-1:click-1"

	status1, body1 := postShipyard(t, url, body, key)
	if status1 != http.StatusCreated {
		t.Fatalf("first status = %d, want 201; body=%q", status1, body1)
	}

	status2, body2 := postShipyard(t, url, body, key)
	if status2 != http.StatusCreated {
		t.Fatalf("replay status = %d, want 201 (cached); body=%q", status2, body2)
	}
	if body2 != body1 {
		t.Errorf("replay body = %q, first = %q", body2, body1)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("handler called %d times, want 1 (replay must be served from cache)", got)
	}
}

// TestIdempotency_ShipyardEnqueue_DifferentBody_Conflict — реюз ключа
// с другим body — это баг клиента (RFC Idempotency-Key). Middleware
// должен ответить 409, не выполняя handler.
func TestIdempotency_ShipyardEnqueue_DifferentBody_Conflict(t *testing.T) {
	rdb := newFakeRedis()
	mw := idempotency.NewMiddlewareWithCmdForTesting(rdb, time.Hour)

	var calls int32
	r := chi.NewRouter()
	r.With(mw.Wrap).Post("/api/planets/{id}/shipyard", fakeEnqueue(&calls, `{"queue_id":"q-1"}`))
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := srv.URL + "/api/planets/p-1/shipyard"
	key := "shipyard:p-1:user-1:click-1"

	if status, _ := postShipyard(t, url, `{"unit_id":204,"count":10}`, key); status != http.StatusCreated {
		t.Fatalf("first: status %d", status)
	}
	if status, _ := postShipyard(t, url, `{"unit_id":205,"count":1}`, key); status != http.StatusConflict {
		t.Errorf("different-body status = %d, want 409", status)
	}
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Errorf("handler called %d times, want 1 (409 path must not reach handler)", got)
	}
}

// TestIdempotency_ShipyardEnqueue_NoHeader_NoCache — без Idempotency-Key
// поведение прежнее: каждый POST = новая запись (legacy: фронт может
// не слать ключ, например в e2e-скрипте). Это документация
// того, что middleware не блокирует запросы без header'а.
func TestIdempotency_ShipyardEnqueue_NoHeader_NoCache(t *testing.T) {
	rdb := newFakeRedis()
	mw := idempotency.NewMiddlewareWithCmdForTesting(rdb, time.Hour)

	var calls int32
	r := chi.NewRouter()
	r.With(mw.Wrap).Post("/api/planets/{id}/shipyard", fakeEnqueue(&calls, `{"queue_id":"q"}`))
	srv := httptest.NewServer(r)
	defer srv.Close()

	url := srv.URL + "/api/planets/p-1/shipyard"
	for i := 0; i < 3; i++ {
		if status, _ := postShipyard(t, url, `{"unit_id":204,"count":10}`, ""); status != http.StatusCreated {
			t.Fatalf("iter %d: status %d", i, status)
		}
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("handler called %d times, want 3 (no key = no cache)", got)
	}
}
