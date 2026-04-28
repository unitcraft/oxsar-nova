package client_test

// Integration-тест end-to-end (план 77 Ф.4).
//
// Сценарий: HTTP-роут с idempotency-middleware вызывает billing-client.
// fake-billing считает оксары на «кошельке». Проверяем:
//  1. Первый запрос списывает оксары и кеширует ответ.
//  2. Повторный запрос с тем же Idempotency-Key возвращает тот же ответ
//     БЕЗ повторного вызова billing — оксары не списываются дважды.
//  3. Запрос с тем же ключом, но другим body → 409.
//
// Тест runs без Redis: использует in-memory fake (см. middleware_test.go),
// поэтому не требует docker / miniredis.

import (
	"context"
	"encoding/json"
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

	billingclient "oxsar/game-nova/internal/billing/client"
	"oxsar/game-nova/pkg/idempotency"
)

// fakeWallet — мини-биллинг с одним кошельком в памяти.
type fakeWallet struct {
	mu       sync.Mutex
	balance  int64
	calls    int32
	seenKeys map[string]int64 // idempotency-key -> сумма списания (для дедупа на side billing)
}

func (fw *fakeWallet) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&fw.calls, 1)
		var req struct {
			Amount    int64  `json:"amount"`
			Reason    string `json:"reason"`
			ToAccount string `json:"to_account"`
		}
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &req)
		idemKey := r.Header.Get("Idempotency-Key")

		fw.mu.Lock()
		defer fw.mu.Unlock()
		if idemKey != "" {
			if _, ok := fw.seenKeys[idemKey]; ok {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"replayed":true}`))
				return
			}
		}
		if fw.balance < req.Amount {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		fw.balance -= req.Amount
		if idemKey != "" {
			fw.seenKeys[idemKey] = req.Amount
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})
}

// fakeRedis — копия из middleware_test.go (нужна и здесь — отдельный package).
type fakeRedis struct {
	mu   sync.Mutex
	data map[string]string
}

func newFakeRedis() *fakeRedis {
	return &fakeRedis{data: map[string]string{}}
}

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

func (f *fakeRedis) SetNX(ctx context.Context, key string, value any, expiration time.Duration) *redis.BoolCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := redis.NewBoolCmd(ctx, "SETNX", key, value)
	if _, ok := f.data[key]; ok {
		cmd.SetVal(false)
		return cmd
	}
	if s, ok := value.(string); ok {
		f.data[key] = s
	} else if b, ok := value.([]byte); ok {
		f.data[key] = string(b)
	}
	cmd.SetVal(true)
	return cmd
}

func (f *fakeRedis) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	f.mu.Lock()
	defer f.mu.Unlock()
	cmd := redis.NewStatusCmd(ctx, "SET", key, value)
	if s, ok := value.(string); ok {
		f.data[key] = s
	} else if b, ok := value.([]byte); ok {
		f.data[key] = string(b)
	}
	cmd.SetVal("OK")
	return cmd
}

// teleportHandler — гипотетический handler, который списывает оксары и
// возвращает «ОК / 402». Имитирует будущий план 65 Ф.6.
func teleportHandler(c *billingclient.Client, billingCalls *int32) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(billingCalls, 1)
		err := c.Spend(r.Context(), billingclient.SpendInput{
			UserToken:      "test.jwt",
			Amount:         100,
			Reason:         "teleport_planet",
			RefID:          "planet-1",
			ToAccount:      "system:teleport",
			IdempotencyKey: r.Header.Get("Idempotency-Key"),
		})
		switch {
		case err == nil:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"teleported":true}`))
		case errors.Is(err, billingclient.ErrInsufficientOxsar):
			w.WriteHeader(http.StatusPaymentRequired)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func TestE2E_SpendAndIdempotentReplay(t *testing.T) {
	wallet := &fakeWallet{balance: 1000, seenKeys: map[string]int64{}}
	billingSrv := httptest.NewServer(wallet.handler())
	defer billingSrv.Close()

	c := billingclient.New(billingSrv.URL)

	rdb := newFakeRedis()
	mw := idempotency.NewMiddlewareWithCmdForTesting(rdb, time.Hour)

	var billingCalls int32
	gameSrv := httptest.NewServer(mw.Wrap(teleportHandler(c, &billingCalls)))
	defer gameSrv.Close()

	// Первый запрос: списание 100, баланс 1000 → 900.
	status1, body1 := postJSON(t, gameSrv.URL, `{"planet_id":1}`, "user-1:teleport:planet-1")
	if status1 != 200 {
		t.Fatalf("first status %d, body=%q", status1, body1)
	}
	if !strings.Contains(body1, `"teleported":true`) {
		t.Errorf("first body = %q", body1)
	}
	if wallet.balance != 900 {
		t.Errorf("balance after first = %d, want 900", wallet.balance)
	}

	// Повторный запрос: middleware должен вернуть кеш, handler не вызовется,
	// billing — тоже.
	status2, body2 := postJSON(t, gameSrv.URL, `{"planet_id":1}`, "user-1:teleport:planet-1")
	if status2 != 200 {
		t.Fatalf("replay status %d", status2)
	}
	if body2 != body1 {
		t.Errorf("replay body = %q, first = %q", body2, body1)
	}
	if wallet.balance != 900 {
		t.Errorf("balance after replay = %d, want still 900 (no double-spend)", wallet.balance)
	}
	if got := atomic.LoadInt32(&billingCalls); got != 1 {
		t.Errorf("handler invoked %d times, want 1 (replay should not reach handler)", got)
	}
	if got := atomic.LoadInt32(&wallet.calls); got != 1 {
		t.Errorf("billing reached %d times, want 1", got)
	}

	// Запрос с тем же ключом, но другим body → 409.
	status3, _ := postJSON(t, gameSrv.URL, `{"planet_id":2}`, "user-1:teleport:planet-1")
	if status3 != http.StatusConflict {
		t.Errorf("conflict status = %d, want 409", status3)
	}
	if wallet.balance != 900 {
		t.Errorf("balance after 409 = %d, want still 900", wallet.balance)
	}
}

func TestE2E_InsufficientOxsar_PropagatedAs402(t *testing.T) {
	wallet := &fakeWallet{balance: 50, seenKeys: map[string]int64{}}
	billingSrv := httptest.NewServer(wallet.handler())
	defer billingSrv.Close()

	c := billingclient.New(billingSrv.URL)
	rdb := newFakeRedis()
	mw := idempotency.NewMiddlewareWithCmdForTesting(rdb, time.Hour)

	var billingCalls int32
	gameSrv := httptest.NewServer(mw.Wrap(teleportHandler(c, &billingCalls)))
	defer gameSrv.Close()

	status, _ := postJSON(t, gameSrv.URL, `{"planet_id":1}`, "user-2:teleport:planet-1")
	if status != http.StatusPaymentRequired {
		t.Fatalf("status = %d, want 402", status)
	}
	if wallet.balance != 50 {
		t.Errorf("balance changed despite 402: %d", wallet.balance)
	}
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
