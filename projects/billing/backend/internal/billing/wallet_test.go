package billing

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/billing/internal/repo"
)

// TestSpendCredit_E2E прогоняет ключевые сценарии wallet-операций
// против РЕАЛЬНОЙ Postgres БД (требует BILLING_TEST_DB_URL).
//
// Сценарии:
//   - Credit + Balance: после пополнения баланс растёт.
//   - Spend success: баланс уменьшается, транзакция записана.
//   - Spend insufficient: возвращает ErrInsufficient, баланс не меняется.
//   - Concurrent spends: SELECT FOR UPDATE гарантирует, что параллельные
//     spend'ы не уведут баланс в минус.
//   - History: возвращает транзакции в обратном хронологическом порядке.
//
// Если BILLING_TEST_DB_URL не задана — тест пропускается (для CI без Postgres).
func TestSpendCredit_E2E(t *testing.T) {
	dbURL := os.Getenv("BILLING_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("BILLING_TEST_DB_URL not set; skipping integration test")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	svc := New(repo.New(pool))

	t.Run("credit then balance", func(t *testing.T) {
		userID := uuid.NewString()
		_, err := svc.Credit(ctx, CreditInput{
			UserID:      userID,
			Amount:      1000,
			Reason:      "top_up",
			FromAccount: "test:topup",
		})
		if err != nil {
			t.Fatalf("credit: %v", err)
		}
		w, err := svc.Balance(ctx, userID, "")
		if err != nil {
			t.Fatalf("balance: %v", err)
		}
		if w.Balance != 1000 {
			t.Errorf("balance = %d, want 1000", w.Balance)
		}
	})

	t.Run("spend success", func(t *testing.T) {
		userID := uuid.NewString()
		_, _ = svc.Credit(ctx, CreditInput{UserID: userID, Amount: 500, Reason: "top_up", FromAccount: "test"})
		_, err := svc.Spend(ctx, SpendInput{
			UserID: userID, Amount: 200, Reason: "shop_purchase",
			ToAccount: "shop:item:1",
		})
		if err != nil {
			t.Fatalf("spend: %v", err)
		}
		w, _ := svc.Balance(ctx, userID, "")
		if w.Balance != 300 {
			t.Errorf("balance after spend = %d, want 300", w.Balance)
		}
	})

	t.Run("spend insufficient", func(t *testing.T) {
		userID := uuid.NewString()
		_, _ = svc.Credit(ctx, CreditInput{UserID: userID, Amount: 100, Reason: "top_up", FromAccount: "test"})
		_, err := svc.Spend(ctx, SpendInput{
			UserID: userID, Amount: 200, Reason: "shop_purchase",
			ToAccount: "shop:item:1",
		})
		if !errors.Is(err, ErrInsufficient) {
			t.Errorf("err = %v, want ErrInsufficient", err)
		}
		w, _ := svc.Balance(ctx, userID, "")
		if w.Balance != 100 {
			t.Errorf("balance after failed spend = %d, want 100", w.Balance)
		}
	})

	t.Run("concurrent spends do not overdraw", func(t *testing.T) {
		userID := uuid.NewString()
		_, _ = svc.Credit(ctx, CreditInput{UserID: userID, Amount: 1000, Reason: "top_up", FromAccount: "test"})

		// Запускаем 50 параллельных spend по 30 каждый.
		// Должны успешно пройти ровно 33 (33×30 = 990 ≤ 1000),
		// остальные 17 получат ErrInsufficient. Но не больше 1000 списано всего.
		const N = 50
		const amount = int64(30)
		var wg sync.WaitGroup
		var success, fail int
		var mu sync.Mutex
		for i := 0; i < N; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := svc.Spend(ctx, SpendInput{
					UserID: userID, Amount: amount, Reason: "test",
					ToAccount: "test:concurrent",
				})
				mu.Lock()
				defer mu.Unlock()
				if err == nil {
					success++
				} else if errors.Is(err, ErrInsufficient) {
					fail++
				} else {
					t.Errorf("unexpected err: %v", err)
				}
			}()
		}
		wg.Wait()

		w, _ := svc.Balance(ctx, userID, "")
		if w.Balance < 0 {
			t.Errorf("balance went negative: %d", w.Balance)
		}
		if int64(success)*amount > 1000 {
			t.Errorf("success=%d × %d = %d > 1000 (overdrawn)", success, amount, int64(success)*amount)
		}
		t.Logf("concurrent: success=%d fail=%d final_balance=%d", success, fail, w.Balance)
	})

	t.Run("history order", func(t *testing.T) {
		userID := uuid.NewString()
		_, _ = svc.Credit(ctx, CreditInput{UserID: userID, Amount: 100, Reason: "top_up", FromAccount: "test"})
		_, _ = svc.Spend(ctx, SpendInput{UserID: userID, Amount: 30, Reason: "shop", ToAccount: "shop:1"})
		_, _ = svc.Spend(ctx, SpendInput{UserID: userID, Amount: 20, Reason: "shop", ToAccount: "shop:2"})
		txs, err := svc.History(ctx, userID, "", 50, 0)
		if err != nil {
			t.Fatalf("history: %v", err)
		}
		if len(txs) != 3 {
			t.Fatalf("history len = %d, want 3", len(txs))
		}
		// Самая свежая — первая.
		if txs[0].Delta != -20 || txs[1].Delta != -30 || txs[2].Delta != 100 {
			t.Errorf("history order: %v %v %v", txs[0].Delta, txs[1].Delta, txs[2].Delta)
		}
	})

	t.Run("invalid amount", func(t *testing.T) {
		_, err := svc.Spend(ctx, SpendInput{UserID: "x", Amount: 0, Reason: "x", ToAccount: "x"})
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("err = %v, want ErrInvalidAmount", err)
		}
		_, err = svc.Spend(ctx, SpendInput{UserID: "x", Amount: -10, Reason: "x", ToAccount: "x"})
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("err = %v, want ErrInvalidAmount", err)
		}
	})
}
