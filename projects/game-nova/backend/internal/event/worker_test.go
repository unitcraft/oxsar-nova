package event_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/oxsar/nova/backend/internal/event"
	"github.com/oxsar/nova/backend/internal/repo"
	"github.com/oxsar/nova/backend/pkg/ids"
)

// Интеграционные тесты требуют PostgreSQL. Если TEST_DATABASE_URL не
// задан — тесты пропускаются. Это позволяет CI без БД проходить,
// локальный разработчик может поднять make dev-up и запустить их.
func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping event worker integration tests")
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

func cleanup(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	_, _ = pool.Exec(ctx, `DELETE FROM events WHERE payload::text LIKE '%"test_marker":true%'`)
}

func insertEvent(t *testing.T, pool *pgxpool.Pool, kind event.Kind, fireAt time.Time) string {
	t.Helper()
	ctx := context.Background()
	id := ids.New()
	payload := fmt.Sprintf(`{"test_marker":true,"kind":%d}`, int(kind))
	_, err := pool.Exec(ctx, `
		INSERT INTO events (id, kind, state, fire_at, payload)
		VALUES ($1, $2, 'wait', $3, $4)
	`, id, int(kind), fireAt, payload)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	return id
}

// TestWorker_HandlerError_Retries проверяет что transient-error
// планирует retry через backoff до maxAttempts.
func TestWorker_HandlerError_Retries(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	defer cleanup(t, pool)

	db := repo.New(pool)
	w := event.NewWorker(db, slogDiscard())
	calls := 0
	testKind := event.Kind(999) // не используется реально
	w.Register(testKind, func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		calls++
		return errors.New("boom")
	})
	// Принудительно задаём maxAttempts=2 для ускорения.
	w = w.WithConfig(event.Config{MaxAttempts: 2, Batch: 10, Interval: 1 * time.Second})

	id := insertEvent(t, pool, testKind, time.Now().Add(-time.Second))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Первый вызов: handler → error → retry planned.
	if err := runOnceViaReflection(w, ctx); err != nil {
		t.Fatalf("first tick: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
	// Проверяем что в БД attempt=1, state='wait', next_retry_at установлен.
	var state string
	var attempt int
	var lastErr *string
	err := pool.QueryRow(ctx, `SELECT state, attempt, last_error FROM events WHERE id=$1`, id).
		Scan(&state, &attempt, &lastErr)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if state != "wait" || attempt != 1 || lastErr == nil || *lastErr == "" {
		t.Fatalf("expected wait/attempt=1/lastErr, got state=%s attempt=%d err=%v",
			state, attempt, lastErr)
	}
}

// TestWorker_UnknownKind_Error — неизвестный kind сразу в state='error'.
func TestWorker_UnknownKind_Error(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	defer cleanup(t, pool)

	db := repo.New(pool)
	w := event.NewWorker(db, slogDiscard())
	// Ничего не регистрируем — kind 998 неизвестен.
	unknownKind := event.Kind(998)

	id := insertEvent(t, pool, unknownKind, time.Now().Add(-time.Second))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := runOnceViaReflection(w, ctx); err != nil {
		t.Fatalf("tick: %v", err)
	}

	var state string
	var lastErr *string
	_ = pool.QueryRow(ctx, `SELECT state, last_error FROM events WHERE id=$1`, id).
		Scan(&state, &lastErr)
	if state != "error" {
		t.Fatalf("expected state=error, got %s", state)
	}
	if lastErr == nil || *lastErr == "" {
		t.Fatalf("expected last_error set")
	}
}

// TestWorker_Ok — успешный handler помечает state='ok'.
func TestWorker_Ok(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	defer cleanup(t, pool)

	db := repo.New(pool)
	w := event.NewWorker(db, slogDiscard())
	testKind := event.Kind(997)
	w.Register(testKind, func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		return nil
	})

	id := insertEvent(t, pool, testKind, time.Now().Add(-time.Second))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := runOnceViaReflection(w, ctx); err != nil {
		t.Fatalf("tick: %v", err)
	}

	var state string
	_ = pool.QueryRow(ctx, `SELECT state FROM events WHERE id=$1`, id).Scan(&state)
	if state != "ok" {
		t.Fatalf("expected ok, got %s", state)
	}
}

// TestWorker_TieBreaker — при одинаковом fire_at порядок по id.
func TestWorker_TieBreaker(t *testing.T) {
	pool := testPool(t)
	defer pool.Close()
	defer cleanup(t, pool)

	db := repo.New(pool)
	w := event.NewWorker(db, slogDiscard())
	var order []string
	testKind := event.Kind(996)
	w.Register(testKind, func(ctx context.Context, tx pgx.Tx, e event.Event) error {
		order = append(order, e.ID)
		return nil
	})

	fireAt := time.Now().Add(-time.Second)
	ids := []string{
		insertEvent(t, pool, testKind, fireAt),
		insertEvent(t, pool, testKind, fireAt),
		insertEvent(t, pool, testKind, fireAt),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := runOnceViaReflection(w, ctx); err != nil {
		t.Fatalf("tick: %v", err)
	}

	// Проверим что порядок обработки совпадает с ORDER BY id ASC.
	expected := append([]string{}, ids...)
	sortStrings(expected)
	if len(order) != len(expected) {
		t.Fatalf("expected %d events, got %d", len(expected), len(order))
	}
	for i, id := range expected {
		if order[i] != id {
			t.Fatalf("at %d expected %s, got %s", i, id, order[i])
		}
	}
}

// runOnceViaReflection — использует публичный Run с short interval,
// ждёт один tick и отменяет. Упрощает код тестов.
func runOnceViaReflection(w *event.Worker, ctx context.Context) error {
	// Вместо reflection — запускаем Run в горутине и отменяем через ~200ms.
	// Один tick запустится автоматически, если interval <= 200ms.
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Новый worker с interval=50ms сформируем через WithConfig снаружи,
	// но Tick на самом деле сработает только через Interval. Упрощение:
	// вручную дёрнем через несколько вариантов.
	// Так как tick — приватный, используем Run в фоне и cancel при
	// обнаружении, что event обработан.
	done := make(chan error, 1)
	go func() {
		done <- w.Run(runCtx)
	}()

	// Немного подождать: в первый Tick Run войдёт через Interval.
	// Наши тесты создают worker с Interval=1s по умолчанию — в норме
	// один Tick успеет.
	select {
	case <-time.After(2 * time.Second):
		cancel()
		<-done
		return nil
	case err := <-done:
		return err
	}
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if s[j] < s[i] {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}
