package aiadvisor_test

// Integration-тесты автопилота. Требуют реальный PostgreSQL по
// TEST_DATABASE_URL (см. event/worker_test.go testPool). При отсутствии
// переменной все тесты пропускаются.
//
// Тесты создают изолированный test-user в users и убирают за собой
// записи через test_marker в полях. Не используют миграции вне
// уже применённых к схеме (миграция 0100_autopilot должна быть
// накатана перед запуском).

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/aiadvisor"
	"oxsar/game-nova/internal/config"
	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/planet"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/internal/score"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping autopilot integration tests")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping: %v", err)
	}
	return pool
}

// createTestUser вставляет минимальную запись в users с заданным
// балансом кредитов и возвращает id. Helper bricks при ошибке.
func createTestUser(t *testing.T, pool *pgxpool.Pool, credit float64) string {
	t.Helper()
	ctx := context.Background()
	id := uuid.NewString()
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, username, email, password_hash, credit)
		VALUES ($1, $2, $3, '', $4)
	`, id, "autopilot_test_"+id[:8], "autopilot_test_"+id[:8]+"@test.local", credit)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	return id
}

func cleanupTestUser(t *testing.T, pool *pgxpool.Pool, userID string) {
	ctx := context.Background()
	_, _ = pool.Exec(ctx, `DELETE FROM events WHERE user_id = $1`, userID)
	_, _ = pool.Exec(ctx, `DELETE FROM ai_advisor_log WHERE user_id = $1`, userID)
	_, _ = pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
}

// newServiceForTest собирает AutopilotService с минимальным набором
// зависимостей для интеграционных тестов: пустой каталог,
// нулевые тюнинги, реальные planet/score сервисы.
//
// Catalog/Buildings/Research пусты → scoreCandidates вернёт [] →
// Compute запишет ready+empty recs. Для тестирования pipeline
// (Enqueue/Compute/Result/Execute) этого достаточно.
func newServiceForTest(pool *pgxpool.Pool) *aiadvisor.AutopilotService {
	db := repo.New(pool)
	cat := &config.Catalog{}
	planetRepo := planet.NewRepository(pool)
	planetSvc := planet.NewService(db, planetRepo, cat)
	scoreSvc := score.NewServiceWithCoeffs(db, cat, config.PointsCoefficients{})
	cfg := config.AIAdvisorConfig{
		MaxPerDay:            5,
		MaxTokens:            1024,
		AutopilotCostCredits: 2,
	}
	return aiadvisor.NewAutopilotService(
		db, cfg,
		aiadvisor.GameTuning{},
		cat, planetSvc, scoreSvc,
		nil, nil, // building/research не нужны при пустом каталоге
	)
}

func TestAutopilot_Enqueue_NotEnoughCredit(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	uid := createTestUser(t, pool, 1) // меньше cost=2
	defer cleanupTestUser(t, pool, uid)

	svc := newServiceForTest(pool)
	_, err := svc.Enqueue(context.Background(), uid, aiadvisor.StrategyEconomy)
	if err != aiadvisor.ErrNotEnoughCredit {
		t.Fatalf("Enqueue with credit=1: got %v, want ErrNotEnoughCredit", err)
	}
}

func TestAutopilot_Enqueue_RateLimitReached(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	uid := createTestUser(t, pool, 100)
	defer cleanupTestUser(t, pool, uid)

	ctx := context.Background()
	// Заполняем лимит (MaxPerDay=5).
	for i := 0; i < 5; i++ {
		_, err := pool.Exec(ctx, `
			INSERT INTO ai_advisor_log (id, user_id, model, question, answer, credits, created_at, source, status)
			VALUES ($1, $2, 'autopilot', '', '', 0, now(), 'autopilot', 'ready')
		`, uuid.NewString(), uid)
		if err != nil {
			t.Fatalf("seed log: %v", err)
		}
	}

	svc := newServiceForTest(pool)
	_, err := svc.Enqueue(ctx, uid, aiadvisor.StrategyEconomy)
	if err != aiadvisor.ErrRateLimitReached {
		t.Fatalf("Enqueue at limit: got %v, want ErrRateLimitReached", err)
	}
}

func TestAutopilot_Enqueue_InvalidStrategy(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	uid := createTestUser(t, pool, 100)
	defer cleanupTestUser(t, pool, uid)

	svc := newServiceForTest(pool)
	_, err := svc.Enqueue(context.Background(), uid, aiadvisor.Strategy("garbage"))
	if err != aiadvisor.ErrInvalidStrategy {
		t.Fatalf("Enqueue garbage: got %v, want ErrInvalidStrategy", err)
	}
}

func TestAutopilot_Enqueue_Success(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := context.Background()
	uid := createTestUser(t, pool, 10)
	defer cleanupTestUser(t, pool, uid)

	svc := newServiceForTest(pool)
	res, err := svc.Enqueue(ctx, uid, aiadvisor.StrategyEconomy)
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}
	if res.JobID == "" {
		t.Fatal("expected non-empty job_id")
	}

	// Кредиты списались (10 - 2 = 8).
	var balance float64
	if err := pool.QueryRow(ctx, `SELECT credit FROM users WHERE id=$1`, uid).Scan(&balance); err != nil {
		t.Fatalf("read balance: %v", err)
	}
	if balance != 8 {
		t.Errorf("credit after Enqueue = %v, want 8", balance)
	}

	// Лог создан со status=pending.
	var status, source, strategy string
	if err := pool.QueryRow(ctx, `
		SELECT status, source, strategy FROM ai_advisor_log WHERE id=$1
	`, res.JobID).Scan(&status, &source, &strategy); err != nil {
		t.Fatalf("read log: %v", err)
	}
	if status != aiadvisor.AutopilotStatusPending {
		t.Errorf("status = %q, want pending", status)
	}
	if source != aiadvisor.AutopilotSource {
		t.Errorf("source = %q, want autopilot", source)
	}
	if strategy != string(aiadvisor.StrategyEconomy) {
		t.Errorf("strategy = %q, want economy", strategy)
	}

	// Event KindAutopilotAdvise создан.
	var eventCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM events WHERE user_id=$1 AND kind=$2 AND state='wait'
	`, uid, int(event.KindAutopilotAdvise)).Scan(&eventCount); err != nil {
		t.Fatalf("count events: %v", err)
	}
	if eventCount != 1 {
		t.Errorf("event count = %d, want 1", eventCount)
	}
}

func TestAutopilot_Result_NotFound(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := context.Background()
	uid := createTestUser(t, pool, 10)
	defer cleanupTestUser(t, pool, uid)

	svc := newServiceForTest(pool)
	_, err := svc.Result(ctx, uid, "00000000-0000-0000-0000-000000000000")
	if err != aiadvisor.ErrJobNotFound {
		t.Fatalf("Result missing job: got %v, want ErrJobNotFound", err)
	}
}

func TestAutopilot_Result_Pending(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := context.Background()
	uid := createTestUser(t, pool, 10)
	defer cleanupTestUser(t, pool, uid)

	svc := newServiceForTest(pool)
	enq, err := svc.Enqueue(ctx, uid, aiadvisor.StrategyMilitary)
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	res, err := svc.Result(ctx, uid, enq.JobID)
	if err != nil {
		t.Fatalf("Result: %v", err)
	}
	if res.Status != aiadvisor.AutopilotStatusPending {
		t.Errorf("Result status = %q, want pending", res.Status)
	}
	if res.Strategy != aiadvisor.StrategyMilitary {
		t.Errorf("Result strategy = %q, want military", res.Strategy)
	}
	if len(res.Recs) != 0 {
		t.Errorf("Result recs = %d, want 0 (pending)", len(res.Recs))
	}
}

func TestAutopilot_Result_OtherUser_NotFound(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := context.Background()
	uid := createTestUser(t, pool, 10)
	defer cleanupTestUser(t, pool, uid)
	otherUID := createTestUser(t, pool, 10)
	defer cleanupTestUser(t, pool, otherUID)

	svc := newServiceForTest(pool)
	enq, err := svc.Enqueue(ctx, uid, aiadvisor.StrategyEconomy)
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	// Другой user не должен видеть чужой job.
	if _, err := svc.Result(ctx, otherUID, enq.JobID); err != aiadvisor.ErrJobNotFound {
		t.Errorf("Result for other user: got %v, want ErrJobNotFound", err)
	}
}

func TestAutopilot_Compute_Idempotent(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	ctx := context.Background()
	uid := createTestUser(t, pool, 10)
	defer cleanupTestUser(t, pool, uid)

	svc := newServiceForTest(pool)
	enq, err := svc.Enqueue(ctx, uid, aiadvisor.StrategyEconomy)
	if err != nil {
		t.Fatalf("Enqueue: %v", err)
	}

	// Достанем event через прямое чтение и эмулируем воркер:
	// worker.processOne открывает pgx-транзакцию и вызывает handler.
	ev := readAdviseEvent(t, pool, uid)

	// Запускаем Compute через прямую pgx-транзакцию.
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := svc.Compute(ctx, tx, ev); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("Compute first: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Проверим status=ready.
	res, err := svc.Result(ctx, uid, enq.JobID)
	if err != nil {
		t.Fatalf("Result after first Compute: %v", err)
	}
	if res.Status != aiadvisor.AutopilotStatusReady {
		t.Fatalf("status after first Compute = %q, want ready", res.Status)
	}

	// Повторный Compute — no-op (status уже ready).
	tx2, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin2: %v", err)
	}
	if err := svc.Compute(ctx, tx2, ev); err != nil {
		_ = tx2.Rollback(ctx)
		t.Fatalf("Compute second: %v", err)
	}
	if err := tx2.Commit(ctx); err != nil {
		t.Fatalf("commit2: %v", err)
	}

	res2, err := svc.Result(ctx, uid, enq.JobID)
	if err != nil {
		t.Fatalf("Result after second Compute: %v", err)
	}
	if res2.Status != aiadvisor.AutopilotStatusReady {
		t.Errorf("status after idempotent Compute = %q, want ready", res2.Status)
	}
}

// readAdviseEvent достаёт event KindAutopilotAdvise из БД для user'а.
// Используется для эмуляции worker'а в Compute-тесте.
func readAdviseEvent(t *testing.T, pool *pgxpool.Pool, userID string) event.Event {
	t.Helper()
	ctx := context.Background()
	var (
		id      string
		kindInt int
		fireAt  time.Time
		payload []byte
	)
	err := pool.QueryRow(ctx, `
		SELECT id, kind, fire_at, payload
		FROM events
		WHERE user_id = $1 AND kind = $2 AND state = 'wait'
		LIMIT 1
	`, userID, int(event.KindAutopilotAdvise)).Scan(&id, &kindInt, &fireAt, &payload)
	if err != nil {
		t.Fatalf("read event: %v", err)
	}
	uid := userID
	return event.Event{
		ID:      id,
		UserID:  &uid,
		Kind:    event.Kind(kindInt),
		FireAt:  fireAt,
		Payload: json.RawMessage(payload),
	}
}

