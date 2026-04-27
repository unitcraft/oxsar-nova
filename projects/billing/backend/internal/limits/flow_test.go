package limits

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestFullHardStopFlow — интеграционный e2e:
//   1. Симулируем серию paid-платежей до достижения hard-stop.
//   2. Reconciler → auto-disable.
//   3. CreateOrder через limits.IsActive() возвращает ErrLimitReached.
//   4. Public /api/billing/limits/status → {active:false, message}.
//   5. Admin override → active=true.
//   6. CreateOrder снова работает (IsActive=true).
//
// Скипается без BILLING_TEST_DB_URL.
func TestFullHardStopFlow(t *testing.T) {
	dbURL := os.Getenv("BILLING_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("BILLING_TEST_DB_URL not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// Чистим всё перед стартом.
	cleanup := func() {
		_, _ = pool.Exec(ctx, `TRUNCATE billing_audit_log, billing_alert_state RESTART IDENTITY`)
		_, _ = pool.Exec(ctx, `UPDATE billing_system_state SET payments_active=true,
			last_changed_by=NULL, last_changed_at=NULL,
			last_change_reason=NULL, auto_disabled_at=NULL WHERE id=1`)
		_, _ = pool.Exec(ctx, `DELETE FROM payment_orders WHERE provider='flow54'`)
	}
	cleanup()
	defer cleanup()

	cfg := DefaultConfig()
	cfg.HardStopKop = 100_000_00 // 100 000 ₽ для теста
	cfg.CacheTTL = 100 * time.Millisecond
	svc := New(pool, cfg)

	// 1. 1000 платежей по 100 ₽ = 100 000 ₽ ровно (= hard-stop).
	now := time.Now()
	insertPaid := func(amount int64) {
		paidAt := now.Add(-time.Duration(now.Nanosecond()))
		_, err := pool.Exec(ctx, `INSERT INTO payment_orders
			(user_id, provider, package_id, amount_kop, credits, status, paid_at)
			VALUES ($1, 'flow54', 'pack_test', $2, 0, 'paid', $3)`,
			uuid.New(), amount, paidAt)
		if err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	for i := 0; i < 1000; i++ {
		insertPaid(10000) // 100 ₽ = 10000 копеек
	}
	revenue, err := svc.GetRevenueYTD(ctx)
	if err != nil {
		t.Fatalf("GetRevenueYTD: %v", err)
	}
	if revenue != 100_000_00 {
		t.Fatalf("expected revenue=100000_00 kop, got %d", revenue)
	}

	// 2. Reconciler runs → auto-disable.
	notifier := &captureNotifier{}
	recon := NewReconciler(svc, notifier, nil, time.Hour)
	recon.runOnce(ctx)
	if !notifier.autoDisabledCalled {
		t.Fatalf("expected NotifyAutoDisabled")
	}
	active, _ := svc.IsActive(ctx)
	if active {
		t.Fatalf("expected active=false after reconciler")
	}

	// 3. Симулируем CreateOrder через LimitsChecker — получает ErrLimitReached.
	// (В прямом виде тест на orders.CreateOrder лежит в пакете billing, но
	// ключевой контракт — IsActive() — мы здесь и проверяем.)
	time.Sleep(150 * time.Millisecond) // ждём истечения cache
	active, _ = svc.IsActive(ctx)
	if active {
		t.Fatalf("expected still inactive")
	}

	// 4. Public /api/billing/limits/status → 503-equivalent (active:false).
	h := NewHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/api/billing/limits/status", nil)
	rr := httptest.NewRecorder()
	h.Status(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status code: %d", rr.Code)
	}
	body := rr.Body.String()
	if !contains(body, `"active":false`) || !contains(body, PublicMessage) {
		t.Fatalf("expected active=false and message, got %s", body)
	}

	// 5. Admin override → enable.
	actor := uuid.New()
	ip := "10.0.0.1"
	if err := svc.SetActive(ctx, true, actor, "manual restore for flow test", &ip, "test"); err != nil {
		t.Fatalf("SetActive enable: %v", err)
	}

	// 6. IsActive снова true.
	time.Sleep(150 * time.Millisecond)
	active, _ = svc.IsActive(ctx)
	if !active {
		t.Fatalf("expected active=true after override")
	}

	// 7. Public status снова active=true.
	rr = httptest.NewRecorder()
	h.Status(rr, req)
	body = rr.Body.String()
	if !contains(body, `"active":true`) {
		t.Fatalf("expected active=true after override, got %s", body)
	}

	// 8. Audit-цепочка: должны быть события auto_disable + enable.
	var autoCnt, enableCnt int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM billing_audit_log WHERE action='limit:auto_disable'`).Scan(&autoCnt)
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM billing_audit_log WHERE action='limit:enable'`).Scan(&enableCnt)
	if autoCnt != 1 || enableCnt != 1 {
		t.Fatalf("expected 1 auto_disable + 1 enable in audit, got %d / %d", autoCnt, enableCnt)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
