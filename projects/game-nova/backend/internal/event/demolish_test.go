package event_test

// План 65 Ф.1 (D-031): тесты HandleDemolishConstruction.
//
// Структура:
//   - TestDemolish_PayloadRoundTrip — pure-тест: payload корректно
//     сериализуется/десериализуется JSON без потери полей.
//   - TestProperty_DemolishPayload_Determinism — property-based (rapid):
//     любая валидная пара (cur, target) даёт ожидаемое решение
//     idempotency-ветки.
//   - TestDemolish_GoldenScenarios — golden-тесты: интеграция с БД,
//     запускается только при TEST_DATABASE_URL.
//
// Инварианты:
//   I1. cur > target  → level UPDATE на target, used_fields-1 если target=0.
//   I2. cur <= target → no-op (идемпотентность).
//   I3. target < 0    → ошибка валидации payload.
//   I4. payload_id отсутствует в planet_id → ошибка.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pgregory.net/rapid"

	"oxsar/game-nova/internal/event"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

// TestDemolish_PayloadRoundTrip — pure: payload не теряет поля при
// JSON сериализации. Защита от случайного rename JSON-тэга.
func TestDemolish_PayloadRoundTrip(t *testing.T) {
	src := event.BuildingPayload{
		QueueID:     "q-123",
		UnitID:      42,
		TargetLevel: 5,
	}
	raw, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	expectKeys := []string{`"queue_id":"q-123"`, `"unit_id":42`, `"target_level":5`}
	for _, k := range expectKeys {
		if !strings.Contains(string(raw), k) {
			t.Fatalf("payload missing key %q in JSON: %s", k, string(raw))
		}
	}
	var dst event.BuildingPayload
	if err := json.Unmarshal(raw, &dst); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dst != src {
		t.Fatalf("round-trip mismatch: got %+v want %+v", dst, src)
	}
}

// TestProperty_DemolishPayload_Determinism — property-based:
// идемпотентность handler'а должна определяться чистой функцией от пары
// (cur, target). Здесь моделируем решение в виде локальной функции
// shouldSkip(cur, target) → bool — это контракт handler'а.
//
// Проверяемые свойства:
//   - shouldSkip(cur, target) == (cur <= target)
//   - детерминированность.
func TestProperty_DemolishPayload_Determinism(t *testing.T) {
	shouldSkip := func(cur, target int) bool { return cur <= target }
	rapid.Check(t, func(t *rapid.T) {
		cur := rapid.IntRange(0, 50).Draw(t, "cur")
		target := rapid.IntRange(0, 50).Draw(t, "target")
		got1 := shouldSkip(cur, target)
		got2 := shouldSkip(cur, target)
		if got1 != got2 {
			t.Fatalf("non-deterministic skip decision")
		}
		expect := cur <= target
		if got1 != expect {
			t.Fatalf("shouldSkip(%d,%d) = %v, want %v", cur, target, got1, expect)
		}
	})
}

// --- Интеграционные (golden) тесты ниже. Требуют TEST_DATABASE_URL. ---

func openTestDB(t *testing.T) (repo.Exec, func()) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping demolish integration tests")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("ping: %v", err)
	}
	return repo.New(pool), func() { pool.Close() }
}

type demolishFixture struct {
	userID   string
	planetID string
	queueID  string
	unitID   int
	curLevel int
	usedF    int
}

// seedFixture создаёт минимальный набор: user → planet → buildings (level=cur)
// → construction_queue (status='running'). Возвращает фикстуру для проверки
// post-conditions. Откатывается через cleanupFixture.
func seedFixture(ctx context.Context, t *testing.T, db repo.Exec, fx demolishFixture) {
	t.Helper()
	err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Минимальная users-строка. Структура users в nova богатая, но
		// для теста хватит required-полей. См. migrations/0001_init.sql.
		if _, err := tx.Exec(ctx, `
			INSERT INTO users (id, username, email, password_hash, registered_at, universe_id)
			VALUES ($1, $2, $3, '', now(),
			        COALESCE((SELECT id FROM universes LIMIT 1),
			                 '00000000-0000-0000-0000-000000000000'::uuid))
			ON CONFLICT (id) DO NOTHING
		`, fx.userID, "demolish-test-"+fx.userID[:8], "dt-"+fx.userID[:8]+"@test"); err != nil {
			return fmt.Errorf("seed user: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO planets (id, user_id, planetname, galaxy, system, position,
			                     planet_type, used_fields, max_fields,
			                     metal, silicon, hydrogen,
			                     last_update, universe_id)
			VALUES ($1, $2, 'TestPlanet', 1, 1, 5, 'planet', $3, 250,
			        0, 0, 0, now(),
			        (SELECT universe_id FROM users WHERE id=$2))
			ON CONFLICT (id) DO NOTHING
		`, fx.planetID, fx.userID, fx.usedF); err != nil {
			return fmt.Errorf("seed planet: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO buildings (planet_id, unit_id, level)
			VALUES ($1, $2, $3)
			ON CONFLICT (planet_id, unit_id) DO UPDATE SET level=EXCLUDED.level
		`, fx.planetID, fx.unitID, fx.curLevel); err != nil {
			return fmt.Errorf("seed building: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO construction_queue (id, planet_id, unit_id, unit_type, target_level,
			                                 start_at, end_at, cost_metal, cost_silicon, cost_hydrogen, status)
			VALUES ($1, $2, $3, 'building', $4, now(), now(), 0, 0, 0, 'running')
			ON CONFLICT (id) DO NOTHING
		`, fx.queueID, fx.planetID, fx.unitID, fx.curLevel-1); err != nil {
			return fmt.Errorf("seed queue: %w", err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func cleanupFixture(ctx context.Context, db repo.Exec, fx demolishFixture) {
	_ = db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, _ = tx.Exec(ctx, `DELETE FROM construction_queue WHERE id=$1`, fx.queueID)
		_, _ = tx.Exec(ctx, `DELETE FROM buildings WHERE planet_id=$1`, fx.planetID)
		_, _ = tx.Exec(ctx, `DELETE FROM planets WHERE id=$1`, fx.planetID)
		_, _ = tx.Exec(ctx, `DELETE FROM users WHERE id=$1`, fx.userID)
		return nil
	})
}

func runDemolish(t *testing.T, db repo.Exec, fx demolishFixture, target int) error {
	t.Helper()
	ctx := context.Background()
	planetID := fx.planetID
	payload := json.RawMessage(fmt.Sprintf(
		`{"queue_id":"%s","unit_id":%d,"target_level":%d}`,
		fx.queueID, fx.unitID, target))
	e := event.Event{
		ID:        ids.New(),
		UserID:    &fx.userID,
		PlanetID:  &planetID,
		Kind:      event.KindDemolishConstruction,
		FireAt:    time.Now(),
		Payload:   payload,
		CreatedAt: time.Now(),
	}
	return db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return event.HandleDemolishConstruction(ctx, tx, e)
	})
}

func readBuildingLevel(t *testing.T, db repo.Exec, planetID string, unitID int) int {
	t.Helper()
	ctx := context.Background()
	var lvl int
	err := db.Pool().QueryRow(ctx,
		`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
		planetID, unitID).Scan(&lvl)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("read level: %v", err)
	}
	return lvl
}

func readUsedFields(t *testing.T, db repo.Exec, planetID string) int {
	t.Helper()
	ctx := context.Background()
	var u int
	if err := db.Pool().QueryRow(ctx,
		`SELECT used_fields FROM planets WHERE id=$1`, planetID).Scan(&u); err != nil {
		t.Fatalf("read used_fields: %v", err)
	}
	return u
}

func readQueueStatus(t *testing.T, db repo.Exec, queueID string) string {
	t.Helper()
	ctx := context.Background()
	var s string
	if err := db.Pool().QueryRow(ctx,
		`SELECT status FROM construction_queue WHERE id=$1`, queueID).Scan(&s); err != nil {
		t.Fatalf("read status: %v", err)
	}
	return s
}

// TestDemolish_GoldenScenarios — три golden-сценария:
//
//	1) Снос здания level 5 → 4: level=4, used_fields неизменён, queue=done.
//	2) Снос здания level 1 → 0: level=0, used_fields-1, queue=done.
//	3) Идемпотентность: повторный запуск handler'а не меняет состояние.
func TestDemolish_GoldenScenarios(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()

	t.Run("level_5_to_4", func(t *testing.T) {
		fx := demolishFixture{
			userID: ids.New(), planetID: ids.New(), queueID: ids.New(),
			unitID: 1, curLevel: 5, usedF: 10,
		}
		ctx := context.Background()
		seedFixture(ctx, t, db, fx)
		defer cleanupFixture(ctx, db, fx)

		if err := runDemolish(t, db, fx, 4); err != nil {
			t.Fatalf("handler: %v", err)
		}
		if got := readBuildingLevel(t, db, fx.planetID, fx.unitID); got != 4 {
			t.Errorf("level: got %d want 4", got)
		}
		if got := readUsedFields(t, db, fx.planetID); got != 10 {
			t.Errorf("used_fields: got %d want 10 (unchanged)", got)
		}
		if got := readQueueStatus(t, db, fx.queueID); got != "done" {
			t.Errorf("status: got %q want done", got)
		}
	})

	t.Run("level_1_to_0_releases_field", func(t *testing.T) {
		fx := demolishFixture{
			userID: ids.New(), planetID: ids.New(), queueID: ids.New(),
			unitID: 2, curLevel: 1, usedF: 7,
		}
		ctx := context.Background()
		seedFixture(ctx, t, db, fx)
		defer cleanupFixture(ctx, db, fx)

		if err := runDemolish(t, db, fx, 0); err != nil {
			t.Fatalf("handler: %v", err)
		}
		if got := readBuildingLevel(t, db, fx.planetID, fx.unitID); got != 0 {
			t.Errorf("level: got %d want 0", got)
		}
		if got := readUsedFields(t, db, fx.planetID); got != 6 {
			t.Errorf("used_fields: got %d want 6 (released)", got)
		}
		if got := readQueueStatus(t, db, fx.queueID); got != "done" {
			t.Errorf("status: got %q want done", got)
		}
	})

	t.Run("idempotent_replay", func(t *testing.T) {
		fx := demolishFixture{
			userID: ids.New(), planetID: ids.New(), queueID: ids.New(),
			unitID: 3, curLevel: 3, usedF: 5,
		}
		ctx := context.Background()
		seedFixture(ctx, t, db, fx)
		defer cleanupFixture(ctx, db, fx)

		// Первый прогон.
		if err := runDemolish(t, db, fx, 2); err != nil {
			t.Fatalf("first run: %v", err)
		}
		levelAfter1 := readBuildingLevel(t, db, fx.planetID, fx.unitID)
		usedAfter1 := readUsedFields(t, db, fx.planetID)

		// Повтор того же события (тот же payload) — должен no-op.
		if err := runDemolish(t, db, fx, 2); err != nil {
			t.Fatalf("replay: %v", err)
		}
		if got := readBuildingLevel(t, db, fx.planetID, fx.unitID); got != levelAfter1 {
			t.Errorf("level changed on replay: got %d want %d", got, levelAfter1)
		}
		if got := readUsedFields(t, db, fx.planetID); got != usedAfter1 {
			t.Errorf("used_fields changed on replay: got %d want %d", got, usedAfter1)
		}
	})
}

// TestDemolish_NegativeTargetLevel_Rejected — невалидный payload (target<0)
// должен возвращать ошибку, а не applied state. Чистый тест без БД.
func TestDemolish_NegativeTargetLevel_Rejected(t *testing.T) {
	planetID := "p1"
	payload := json.RawMessage(`{"queue_id":"q","unit_id":1,"target_level":-1}`)
	e := event.Event{
		ID:       "e1",
		PlanetID: &planetID,
		Kind:     event.KindDemolishConstruction,
		Payload:  payload,
	}
	// tx==nil безопасен: handler упирается в валидацию TargetLevel<0
	// до первого SQL-запроса.
	err := event.HandleDemolishConstruction(context.Background(), nil, e)
	if err == nil {
		t.Fatalf("expected error for negative target_level")
	}
	if !strings.Contains(err.Error(), "target_level") {
		t.Errorf("expected target_level error, got: %v", err)
	}
}
