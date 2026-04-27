package billing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/billing/internal/repo"
)

// TestReconcile_Mismatch проверяет что Reconciler находит и замораживает
// кошельки, у которых wallet.balance не совпадает с SUM(transactions.delta).
//
// Сценарий:
//   1. Создаём юзера, пополняем 1000.
//   2. Вручную через UPDATE подкручиваем balance до 1500 (имитация бага).
//   3. Запускаем runOnce.
//   4. Кошелёк должен быть frozen, в frozen_reason — описание расхождения.
func TestReconcile_Mismatch(t *testing.T) {
	dbURL := os.Getenv("BILLING_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("BILLING_TEST_DB_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	svc := New(repo.New(pool))
	rec := NewReconciler(svc, time.Hour)

	userID := uuid.NewString()

	// 1. Pre-fill: credit 1000.
	if _, err := svc.Credit(ctx, CreditInput{
		UserID: userID, Amount: 1000, Reason: "top_up", FromAccount: "test",
	}); err != nil {
		t.Fatalf("credit: %v", err)
	}
	w, _ := svc.Balance(ctx, userID, "")
	if w.Balance != 1000 || w.Frozen {
		t.Fatalf("setup: balance=%d frozen=%v want 1000/false", w.Balance, w.Frozen)
	}

	// 2. Корвет: подкрутить balance вручную (имитация рассинхронизации).
	if _, err := pool.Exec(ctx,
		`UPDATE wallets SET balance = 1500 WHERE id = $1`, w.ID); err != nil {
		t.Fatalf("update wallet: %v", err)
	}

	// 3. Reconcile.
	res := rec.runOnce(ctx)
	if res.Frozen < 1 {
		t.Errorf("frozen=%d, want >=1", res.Frozen)
	}
	if res.Mismatched < 1 {
		t.Errorf("mismatched=%d, want >=1", res.Mismatched)
	}

	// 4. Проверяем что наш кошелёк теперь frozen.
	w2, _ := svc.Balance(ctx, userID, "")
	if !w2.Frozen {
		t.Errorf("wallet not frozen after reconcile")
	}

	// 5. После freeze попытка spend → ErrFrozen.
	_, err = svc.Spend(ctx, SpendInput{
		UserID: userID, Amount: 10, Reason: "test", ToAccount: "test:after-freeze",
	})
	if err != ErrFrozen {
		t.Errorf("spend after freeze: err=%v, want ErrFrozen", err)
	}
}

// TestReconcile_NoMismatch проверяет что чистый Reconcile не замораживает
// валидные кошельки (balance == SUM(delta)).
func TestReconcile_NoMismatch(t *testing.T) {
	dbURL := os.Getenv("BILLING_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("BILLING_TEST_DB_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	svc := New(repo.New(pool))
	rec := NewReconciler(svc, time.Hour)

	userID := uuid.NewString()
	if _, err := svc.Credit(ctx, CreditInput{
		UserID: userID, Amount: 500, Reason: "top_up", FromAccount: "test",
	}); err != nil {
		t.Fatalf("credit: %v", err)
	}
	if _, err := svc.Spend(ctx, SpendInput{
		UserID: userID, Amount: 100, Reason: "test", ToAccount: "test:1",
	}); err != nil {
		t.Fatalf("spend: %v", err)
	}

	// runOnce пробежит по ВСЕМ кошелькам (не только нашему). Любой mismatch
	// от других тестов в том же DB упадёт сюда. Поэтому проверяем только
	// СВОЙ кошелёк: после reconcile он не frozen.
	rec.runOnce(ctx)
	w, _ := svc.Balance(ctx, userID, "")
	if w.Frozen {
		t.Errorf("clean wallet frozen by reconcile")
	}
	if w.Balance != 400 {
		t.Errorf("balance = %d, want 400", w.Balance)
	}
}
