package limits

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestHighestPassed — pure-функция, не требует БД.
func TestHighestPassed(t *testing.T) {
	cases := []struct {
		name        string
		revenue     int64
		hardStop    int64
		expected    Threshold
	}{
		{"zero revenue", 0, 230_000_000, ThresholdNone},
		{"50% — none", 115_000_000, 230_000_000, ThresholdNone},
		{"79% — none", 181_700_000, 230_000_000, ThresholdNone},
		{"80% exact — Threshold80", 184_000_000, 230_000_000, Threshold80},
		{"85% — Threshold80", 195_500_000, 230_000_000, Threshold80},
		{"90% exact — Threshold90", 207_000_000, 230_000_000, Threshold90},
		{"95% exact — Threshold95", 218_500_000, 230_000_000, Threshold95},
		{"99% — Threshold95", 227_700_000, 230_000_000, Threshold95},
		{"100% exact — ThresholdHard", 230_000_000, 230_000_000, ThresholdHard},
		{"over — ThresholdHard", 240_000_000, 230_000_000, ThresholdHard},
		{"zero hard stop — none", 1, 0, ThresholdNone},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HighestPassed(tc.revenue, tc.hardStop)
			if got != tc.expected {
				t.Errorf("revenue=%d hard=%d → got %d, want %d",
					tc.revenue, tc.hardStop, got, tc.expected)
			}
		})
	}
}

func TestStartOfYear(t *testing.T) {
	moscow, _ := time.LoadLocation("Europe/Moscow")
	now := time.Date(2026, 7, 15, 12, 30, 0, 0, moscow)
	soy := startOfYear(now, moscow)
	want := time.Date(2026, 1, 1, 0, 0, 0, 0, moscow)
	if !soy.Equal(want) {
		t.Errorf("got %v, want %v", soy, want)
	}
}

func TestAlertColumn(t *testing.T) {
	cases := map[Threshold]string{
		Threshold80:   "threshold_80_sent",
		Threshold90:   "threshold_90_sent",
		Threshold95:   "threshold_95_sent",
		ThresholdHard: "threshold_hard_sent",
		ThresholdNone: "",
	}
	for th, want := range cases {
		got := alertColumn(th)
		if got != want {
			t.Errorf("threshold=%d: got %q, want %q", th, got, want)
		}
	}
}

// TestService_E2E — integration-тесты, требуют BILLING_TEST_DB_URL.
//
// Покрывают:
//   - IsActive: возвращает true/false из billing_system_state.
//   - SetActive: меняет флаг + пишет audit, кеш сбрасывается.
//   - AutoDisable: idempotent (повторный вызов не дублирует audit).
//   - GetRevenueYTD: считает только status='paid' и в текущем году.
//   - MarkAlerted: первый раз true, повторно false.
//   - Hard-stop flow: при revenue >= hard_stop reconciler.runOnce ставит
//     payments_active=false и AlertedAt[hard]!=nil.
//   - Admin override после auto-disable восстанавливает active=true.
//
// Skipped if BILLING_TEST_DB_URL not set (CI без Postgres).
func TestService_E2E(t *testing.T) {
	dbURL := os.Getenv("BILLING_TEST_DB_URL")
	if dbURL == "" {
		t.Skip("BILLING_TEST_DB_URL not set — skipping limits integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	// Чистим релевантные таблицы. Тесты предполагают что БД создана с
	// миграциями 0001-0002 (test-БД должна быть свежая для каждого прогона).
	cleanup := func() {
		_, _ = pool.Exec(ctx, `TRUNCATE billing_audit_log, billing_alert_state RESTART IDENTITY`)
		_, _ = pool.Exec(ctx, `UPDATE billing_system_state SET payments_active=true,
			last_changed_by=NULL, last_changed_at=NULL,
			last_change_reason=NULL, auto_disabled_at=NULL WHERE id=1`)
		_, _ = pool.Exec(ctx, `DELETE FROM payment_orders WHERE provider='test54'`)
	}
	cleanup()
	defer cleanup()

	cfg := DefaultConfig()
	cfg.HardStopKop = 1000_00 // 1000 ₽ для удобства тестов
	cfg.CacheTTL = 0          // отключаем cache
	svc := New(pool, cfg)

	t.Run("IsActive default true", func(t *testing.T) {
		active, err := svc.IsActive(ctx)
		if err != nil || !active {
			t.Fatalf("expected active=true, got %v err=%v", active, err)
		}
	})

	t.Run("SetActive false → IsActive false + audit", func(t *testing.T) {
		actor := uuid.New()
		ip := "127.0.0.1"
		err := svc.SetActive(ctx, false, actor, "test reason", &ip, "go-test")
		if err != nil {
			t.Fatalf("SetActive: %v", err)
		}
		active, err := svc.IsActive(ctx)
		if err != nil || active {
			t.Fatalf("expected active=false, got %v err=%v", active, err)
		}
		// Audit-запись.
		var cnt int
		err = pool.QueryRow(ctx, `SELECT count(*) FROM billing_audit_log
			WHERE actor_id=$1 AND action='limit:disable'`, actor).Scan(&cnt)
		if err != nil || cnt != 1 {
			t.Fatalf("expected 1 audit row, got %d err=%v", cnt, err)
		}
		// Восстановим.
		_ = svc.SetActive(ctx, true, actor, "restore", &ip, "go-test")
	})

	t.Run("SetActive empty reason rejected", func(t *testing.T) {
		err := svc.SetActive(ctx, false, uuid.New(), "", nil, "")
		if err == nil {
			t.Fatal("expected error on empty reason")
		}
	})

	t.Run("AutoDisable idempotent", func(t *testing.T) {
		cleanup()
		err := svc.AutoDisable(ctx, "auto-test")
		if err != nil {
			t.Fatalf("AutoDisable: %v", err)
		}
		// Повторный вызов — не должен бросить и не дублировать audit.
		err = svc.AutoDisable(ctx, "auto-test")
		if err != nil {
			t.Fatalf("AutoDisable 2: %v", err)
		}
		var cnt int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM billing_audit_log
			WHERE action='limit:auto_disable'`).Scan(&cnt)
		if cnt != 1 {
			t.Errorf("expected 1 audit row, got %d", cnt)
		}
	})

	t.Run("MarkAlerted first true, second false", func(t *testing.T) {
		cleanup()
		year := time.Now().Year()
		fresh, err := svc.MarkAlerted(ctx, year, Threshold80)
		if err != nil || !fresh {
			t.Fatalf("first mark: fresh=%v err=%v", fresh, err)
		}
		fresh, err = svc.MarkAlerted(ctx, year, Threshold80)
		if err != nil || fresh {
			t.Fatalf("second mark: fresh=%v err=%v", fresh, err)
		}
		at, err := svc.AlertedAt(ctx, year, Threshold80)
		if err != nil || at == nil {
			t.Fatalf("AlertedAt: at=%v err=%v", at, err)
		}
	})

	t.Run("GetRevenueYTD sums paid only in current year", func(t *testing.T) {
		cleanup()
		now := time.Now()
		startOfThisYear := time.Date(now.Year(), 1, 15, 12, 0, 0, 0, time.UTC)
		// Добавим 3 платежа: 2 paid в текущем году, 1 pending, 1 в прошлом году.
		insert := func(amount int64, status string, paidAt *time.Time) {
			_, _ = pool.Exec(ctx, `INSERT INTO payment_orders
				(user_id, provider, package_id, amount_kop, credits, status, paid_at)
				VALUES ($1, 'test54', 'pack_test', $2, 0, $3, $4)`,
				uuid.New(), amount, status, paidAt)
		}
		paid1 := startOfThisYear.Add(time.Hour)
		paid2 := startOfThisYear.Add(48 * time.Hour)
		insert(50000, "paid", &paid1)  // 500 ₽
		insert(70000, "paid", &paid2)  // 700 ₽
		insert(99999, "pending", nil)  // не учитывается
		lastYear := time.Date(now.Year()-1, 6, 1, 0, 0, 0, 0, time.UTC)
		insert(123456, "paid", &lastYear) // прошлый год — не учитывается

		got, err := svc.GetRevenueYTD(ctx)
		if err != nil {
			t.Fatalf("GetRevenueYTD: %v", err)
		}
		want := int64(50000 + 70000)
		if got != want {
			t.Errorf("got %d, want %d", got, want)
		}
	})

	t.Run("Hard-stop flow: reconciler auto-disables, override re-enables", func(t *testing.T) {
		cleanup()
		now := time.Now()
		// Платёж >= hard_stop: 1500 ₽ при пороге 1000 ₽.
		paidAt := time.Date(now.Year(), 1, 15, 0, 0, 0, 0, time.UTC)
		_, err := pool.Exec(ctx, `INSERT INTO payment_orders
			(user_id, provider, package_id, amount_kop, credits, status, paid_at)
			VALUES ($1, 'test54', 'pack_big', 150000, 0, 'paid', $2)`,
			uuid.New(), paidAt)
		if err != nil {
			t.Fatalf("insert: %v", err)
		}

		notifier := &captureNotifier{}
		recon := NewReconciler(svc, notifier, nil, time.Hour)
		recon.runOnce(ctx)

		// payments_active должен стать false.
		active, err := svc.IsActive(ctx)
		if err != nil || active {
			t.Fatalf("expected active=false after reconciler, got %v err=%v", active, err)
		}
		// AutoDisabledAt заполнен.
		st, err := svc.GetState(ctx)
		if err != nil {
			t.Fatalf("GetState: %v", err)
		}
		if st.AutoDisabledAt == nil {
			t.Errorf("expected AutoDisabledAt non-nil")
		}
		// Notifier получил уведомление.
		if !notifier.autoDisabledCalled {
			t.Errorf("expected NotifyAutoDisabled called")
		}

		// Admin override → active=true.
		actor := uuid.New()
		ip := "10.0.0.1"
		if err := svc.SetActive(ctx, true, actor, "manual restore", &ip, "ui"); err != nil {
			t.Fatalf("SetActive enable: %v", err)
		}
		active2, _ := svc.IsActive(ctx)
		if !active2 {
			t.Errorf("expected active=true after override")
		}
		// auto_disabled_at должен быть очищен (по логике SetActive).
		st2, _ := svc.GetState(ctx)
		if st2.AutoDisabledAt != nil {
			t.Errorf("expected AutoDisabledAt cleared after enable, got %v", st2.AutoDisabledAt)
		}
	})

	t.Run("Soft-warning at 80%: notifier called once per year", func(t *testing.T) {
		cleanup()
		// Платёж = 80% от hard_stop = 800 ₽.
		now := time.Now()
		paidAt := time.Date(now.Year(), 2, 1, 0, 0, 0, 0, time.UTC)
		_, _ = pool.Exec(ctx, `INSERT INTO payment_orders
			(user_id, provider, package_id, amount_kop, credits, status, paid_at)
			VALUES ($1, 'test54', 'pack_80', 80000, 0, 'paid', $2)`,
			uuid.New(), paidAt)

		notifier := &captureNotifier{}
		recon := NewReconciler(svc, notifier, nil, time.Hour)
		recon.runOnce(ctx)

		if notifier.thresholdCount != 1 {
			t.Errorf("expected 1 threshold notification, got %d", notifier.thresholdCount)
		}

		// Повторный run в том же году — не шлёт снова.
		recon.runOnce(ctx)
		if notifier.thresholdCount != 1 {
			t.Errorf("expected still 1 after second run, got %d", notifier.thresholdCount)
		}
	})
}

// captureNotifier — test-дублёр Notifier, считает вызовы.
type captureNotifier struct {
	thresholdCount     int
	lastThreshold      Threshold
	autoDisabledCalled bool
}

func (n *captureNotifier) NotifyThresholdReached(_ context.Context, t Threshold, _, _ int64) {
	n.thresholdCount++
	n.lastThreshold = t
}
func (n *captureNotifier) NotifyAutoDisabled(_ context.Context, _, _ int64) {
	n.autoDisabledCalled = true
}
