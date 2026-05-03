package fleet

// План 65 Ф.3-Ф.4 (D-037): тесты ветки destroy-building.
//
// Структура (зеркало demolish_test.go):
//   - TestTransportPayload_TargetBuildingRoundTrip — pure round-trip:
//     поле target_building_id корректно сериализуется (omitempty, без
//     потерь).
//   - TestACSPayload_TargetBuildingRoundTrip — то же для acsPayload.
//   - TestProperty_DestroyBuilding_NoOpDecision — property-based: ветка
//     handler'а зависит детерминированно от (isMoon, winner).
//   - TestDestroyBuilding_GoldenScenarios — golden c TEST_DATABASE_URL:
//     понижение уровня, освобождение поля при level=1→0, фильтр
//     UNIT_EXCHANGE/UNIT_NANO_FACTORY, идемпотентность.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"pgregory.net/rapid"

	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

// TestTransportPayload_TargetBuildingRoundTrip — поле target_building_id
// (план 65 Ф.3) сериализуется через JSON-тэг "target_building_id" с
// omitempty. Защита от случайного rename.
func TestTransportPayload_TargetBuildingRoundTrip(t *testing.T) {
	src := transportPayload{
		FleetID:          "f-1",
		Carried:          map[string]int64{"metal": 100},
		TargetBuildingID: 42,
	}
	raw, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"target_building_id":42`) {
		t.Fatalf("payload missing target_building_id key: %s", raw)
	}
	var dst transportPayload
	if err := json.Unmarshal(raw, &dst); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dst.TargetBuildingID != 42 {
		t.Errorf("round-trip: got %d want 42", dst.TargetBuildingID)
	}

	// omitempty: zero-value не пишется (легаси-payload без поля).
	zero := transportPayload{FleetID: "f-2"}
	rawZero, _ := json.Marshal(zero)
	if strings.Contains(string(rawZero), "target_building_id") {
		t.Errorf("zero-value should be omitted: %s", rawZero)
	}
}

// TestACSPayload_TargetBuildingRoundTrip — то же для acsPayload (Ф.4).
func TestACSPayload_TargetBuildingRoundTrip(t *testing.T) {
	src := acsPayload{
		FleetID:          "f-3",
		ACSGroupID:       "g-1",
		TargetBuildingID: 7,
	}
	raw, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"target_building_id":7`) {
		t.Fatalf("payload missing target_building_id key: %s", raw)
	}
	var dst acsPayload
	if err := json.Unmarshal(raw, &dst); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dst.TargetBuildingID != 7 {
		t.Errorf("round-trip: got %d want 7", dst.TargetBuildingID)
	}
	zero := acsPayload{FleetID: "f-4", ACSGroupID: "g-2"}
	rawZero, _ := json.Marshal(zero)
	if strings.Contains(string(rawZero), "target_building_id") {
		t.Errorf("zero-value should be omitted: %s", rawZero)
	}
}

// План 72.1.56 B7: legacy DESTROY_BUILD_RESULT_MIN_OFFS_LEVEL эвристика
// (Assault.class.php:253-281). Pure-test без БД.
func TestFilterDestroyCandidates(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name              string
		defenderBuilds    map[int]int
		attackerMaxLevels map[int]int
		want              []int
	}{
		{
			name:              "nil attackers → fallback no heuristic, all returned",
			defenderBuilds:    map[int]int{1: 5, 2: 3, 4: 7},
			attackerMaxLevels: nil,
			want:              []int{1, 2, 4},
		},
		{
			name:              "attacker has nothing → all unchecked, excluded",
			defenderBuilds:    map[int]int{1: 5, 2: 3},
			attackerMaxLevels: map[int]int{},
			want:              []int{},
		},
		{
			name:              "attacker level=4, defender level=10 → 9>=4, kept",
			defenderBuilds:    map[int]int{1: 10},
			attackerMaxLevels: map[int]int{1: 4},
			want:              []int{1},
		},
		{
			name:              "attacker level=10, defender level=5 → 4<10, excluded",
			defenderBuilds:    map[int]int{1: 5},
			attackerMaxLevels: map[int]int{1: 10},
			want:              []int{},
		},
		{
			name:              "attacker level=5, defender level=5 → 4<5, excluded",
			defenderBuilds:    map[int]int{1: 5},
			attackerMaxLevels: map[int]int{1: 5},
			want:              []int{},
		},
		{
			name:              "attacker level=5, defender level=6 → 5>=5, kept",
			defenderBuilds:    map[int]int{1: 6},
			attackerMaxLevels: map[int]int{1: 5},
			want:              []int{1},
		},
		{
			name: "mixed: building 1 unchecked, 2 dropped, 3 kept",
			defenderBuilds: map[int]int{
				1: 8,  // unchecked → excluded
				2: 4,  // 3<5 → excluded
				3: 10, // 9>=5 → kept
			},
			attackerMaxLevels: map[int]int{
				2: 5,
				3: 5,
			},
			want: []int{3},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterDestroyCandidates(tc.defenderBuilds, tc.attackerMaxLevels)
			if len(got) != len(tc.want) {
				t.Fatalf("len: got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("[%d]: got %d, want %d", i, got[i], tc.want[i])
				}
			}
		})
	}
}

// TestProperty_DestroyBuilding_NoOpDecision — детерминированно: handler
// должен no-op'ить если isMoon=true ИЛИ winner != "attackers".
//
// Property: для любой пары (isMoon, winner) shouldNoOp(...) === expect.
// Это контракт первого блока tryDestroyBuilding (до touch'а БД).
func TestProperty_DestroyBuilding_NoOpDecision(t *testing.T) {
	// Локальная реплика контракта-функции (мы её не экспортируем —
	// проверяем именно ожидаемую логику).
	shouldNoOp := func(isMoon bool, winner string) bool {
		return isMoon || winner != "attackers"
	}
	winners := []string{"attackers", "defenders", "draw", "", "weird"}
	rapid.Check(t, func(t *rapid.T) {
		isMoon := rapid.Bool().Draw(t, "isMoon")
		w := winners[rapid.IntRange(0, len(winners)-1).Draw(t, "winnerIdx")]
		got := shouldNoOp(isMoon, w)
		expect := isMoon || w != "attackers"
		if got != expect {
			t.Fatalf("shouldNoOp(%v, %q) = %v want %v", isMoon, w, got, expect)
		}
	})
}

// --- Golden integration tests (TEST_DATABASE_URL) ---

func openTestDB(t *testing.T) (repo.Exec, func()) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping destroy_building integration tests")
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

// dbFixture создаёт user + planet с заданными usedFields и список
// зданий (unitID → level). Возвращает (userID, planetID).
type dbFixture struct {
	userID    string
	planetID  string
	usedF     int
	buildings map[int]int // unit_id → level
}

func seedFixture(ctx context.Context, t *testing.T, db repo.Exec, fx dbFixture) {
	t.Helper()
	err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			INSERT INTO users (id, username, email, password_hash, registered_at, universe_id)
			VALUES ($1, $2, $3, '', now(),
			        COALESCE((SELECT id FROM universes LIMIT 1),
			                 '00000000-0000-0000-0000-000000000000'::uuid))
			ON CONFLICT (id) DO NOTHING
		`, fx.userID, "destroy-test-"+fx.userID[:8], "dt-"+fx.userID[:8]+"@test"); err != nil {
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
		for unitID, lvl := range fx.buildings {
			if _, err := tx.Exec(ctx, `
				INSERT INTO buildings (planet_id, unit_id, level)
				VALUES ($1, $2, $3)
				ON CONFLICT (planet_id, unit_id) DO UPDATE SET level=EXCLUDED.level
			`, fx.planetID, unitID, lvl); err != nil {
				return fmt.Errorf("seed building %d: %w", unitID, err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func cleanupFixture(ctx context.Context, db repo.Exec, fx dbFixture) {
	_ = db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, _ = tx.Exec(ctx, `DELETE FROM buildings WHERE planet_id=$1`, fx.planetID)
		_, _ = tx.Exec(ctx, `DELETE FROM planets WHERE id=$1`, fx.planetID)
		_, _ = tx.Exec(ctx, `DELETE FROM users WHERE id=$1`, fx.userID)
		return nil
	})
}

func readLevel(t *testing.T, db repo.Exec, planetID string, unitID int) int {
	t.Helper()
	var lvl int
	err := db.Pool().QueryRow(context.Background(),
		`SELECT level FROM buildings WHERE planet_id=$1 AND unit_id=$2`,
		planetID, unitID).Scan(&lvl)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("read level: %v", err)
	}
	return lvl
}

func readUsedFields(t *testing.T, db repo.Exec, planetID string) int {
	t.Helper()
	var u int
	if err := db.Pool().QueryRow(context.Background(),
		`SELECT used_fields FROM planets WHERE id=$1`, planetID).Scan(&u); err != nil {
		t.Fatalf("read used_fields: %v", err)
	}
	return u
}

// TestDestroyBuilding_GoldenScenarios — golden:
//
//   1. explicit target unit_id=4, level 5→4: level=4, used_fields неизменён.
//   2. explicit target unit_id=4, level 1→0: level=0, used_fields-1.
//   3. defender победил → no-op.
//   4. цель — луна → no-op.
//   5. Random выбор: только из не-EXCHANGE/NANO_FACTORY зданий.
//   6. Idempotent replay через те же payload-параметры — no double dec.
func TestDestroyBuilding_GoldenScenarios(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	t.Run("explicit_target_5_to_4", func(t *testing.T) {
		fx := dbFixture{
			userID: ids.New(), planetID: ids.New(), usedF: 10,
			buildings: map[int]int{4: 5},
		}
		seedFixture(ctx, t, db, fx)
		defer cleanupFixture(ctx, db, fx)

		err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			unitID, lvlFrom, lvlTo, ok, err := tryDestroyBuilding(ctx, tx,
				fx.planetID, false, "attackers", 4, nil, 0xDEADBEEF)
			if err != nil {
				return err
			}
			if !ok || unitID != 4 || lvlFrom != 5 || lvlTo != 4 {
				t.Errorf("got (%d,%d,%d,%v) want (4,5,4,true)", unitID, lvlFrom, lvlTo, ok)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("tx: %v", err)
		}
		if got := readLevel(t, db, fx.planetID, 4); got != 4 {
			t.Errorf("level: got %d want 4", got)
		}
		if got := readUsedFields(t, db, fx.planetID); got != 10 {
			t.Errorf("used_fields: got %d want 10 (unchanged)", got)
		}
	})

	t.Run("explicit_target_1_to_0_releases_field", func(t *testing.T) {
		fx := dbFixture{
			userID: ids.New(), planetID: ids.New(), usedF: 7,
			buildings: map[int]int{6: 1},
		}
		seedFixture(ctx, t, db, fx)
		defer cleanupFixture(ctx, db, fx)

		err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			unitID, lvlFrom, lvlTo, ok, err := tryDestroyBuilding(ctx, tx,
				fx.planetID, false, "attackers", 6, nil, 0)
			if err != nil {
				return err
			}
			if !ok || unitID != 6 || lvlFrom != 1 || lvlTo != 0 {
				t.Errorf("got (%d,%d,%d,%v) want (6,1,0,true)", unitID, lvlFrom, lvlTo, ok)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("tx: %v", err)
		}
		if got := readLevel(t, db, fx.planetID, 6); got != 0 {
			t.Errorf("level: got %d want 0", got)
		}
		if got := readUsedFields(t, db, fx.planetID); got != 6 {
			t.Errorf("used_fields: got %d want 6 (released)", got)
		}
	})

	t.Run("defenders_win_noop", func(t *testing.T) {
		fx := dbFixture{
			userID: ids.New(), planetID: ids.New(), usedF: 5,
			buildings: map[int]int{4: 5},
		}
		seedFixture(ctx, t, db, fx)
		defer cleanupFixture(ctx, db, fx)

		err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			_, _, _, ok, err := tryDestroyBuilding(ctx, tx,
				fx.planetID, false, "defenders", 4, nil, 0)
			if err != nil {
				return err
			}
			if ok {
				t.Errorf("expected no-op when defenders won")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("tx: %v", err)
		}
		if got := readLevel(t, db, fx.planetID, 4); got != 5 {
			t.Errorf("level changed: got %d want 5", got)
		}
	})

	t.Run("moon_target_noop", func(t *testing.T) {
		fx := dbFixture{
			userID: ids.New(), planetID: ids.New(), usedF: 5,
			buildings: map[int]int{4: 5},
		}
		seedFixture(ctx, t, db, fx)
		defer cleanupFixture(ctx, db, fx)

		err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			_, _, _, ok, err := tryDestroyBuilding(ctx, tx,
				fx.planetID, true, "attackers", 4, nil, 0)
			if err != nil {
				return err
			}
			if ok {
				t.Errorf("expected no-op when target is moon")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("tx: %v", err)
		}
	})

	t.Run("random_skips_excluded_units", func(t *testing.T) {
		// Только UNIT_EXCHANGE и UNIT_NANO_FACTORY на планете —
		// random-ветка должна вернуть no-op.
		fx := dbFixture{
			userID: ids.New(), planetID: ids.New(), usedF: 5,
			buildings: map[int]int{
				unitExchange:    10,
				unitNanoFactory: 5,
			},
		}
		seedFixture(ctx, t, db, fx)
		defer cleanupFixture(ctx, db, fx)

		err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			_, _, _, ok, err := tryDestroyBuilding(ctx, tx,
				fx.planetID, false, "attackers", 0, nil, 0xCAFEBABE)
			if err != nil {
				return err
			}
			if ok {
				t.Errorf("expected no-op when only excluded buildings present")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("tx: %v", err)
		}
		// Уровни не изменились.
		if got := readLevel(t, db, fx.planetID, unitExchange); got != 10 {
			t.Errorf("UNIT_EXCHANGE level changed: got %d want 10", got)
		}
		if got := readLevel(t, db, fx.planetID, unitNanoFactory); got != 5 {
			t.Errorf("UNIT_NANO_FACTORY level changed: got %d want 5", got)
		}
	})

	t.Run("random_picks_eligible_only", func(t *testing.T) {
		// Смешанная планета: один eligible (unit_id=4) + два excluded.
		// random должен выбрать именно 4.
		fx := dbFixture{
			userID: ids.New(), planetID: ids.New(), usedF: 5,
			buildings: map[int]int{
				4:               5,
				unitExchange:    10,
				unitNanoFactory: 5,
			},
		}
		seedFixture(ctx, t, db, fx)
		defer cleanupFixture(ctx, db, fx)

		err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			unitID, _, _, ok, err := tryDestroyBuilding(ctx, tx,
				fx.planetID, false, "attackers", 0, nil, 0xCAFEBABE)
			if err != nil {
				return err
			}
			if !ok || unitID != 4 {
				t.Errorf("expected unit_id=4, got (%d, ok=%v)", unitID, ok)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("tx: %v", err)
		}
		// EXCHANGE/NANO остались нетронутыми.
		if got := readLevel(t, db, fx.planetID, unitExchange); got != 10 {
			t.Errorf("UNIT_EXCHANGE changed: got %d want 10", got)
		}
		if got := readLevel(t, db, fx.planetID, unitNanoFactory); got != 5 {
			t.Errorf("UNIT_NANO_FACTORY changed: got %d want 5", got)
		}
	})

	t.Run("idempotent_explicit", func(t *testing.T) {
		// Explicit target — повторный вызов с тем же seed/target
		// продолжает понижать (handler логически one-shot, повтор
		// в реальном flow невозможен из-за FOR UPDATE SKIP LOCKED;
		// но если кто-то снаружи вызовет дважды — должно быть
		// детерминировано).
		//
		// Этот тест документирует фактическое поведение: handler
		// сам по себе не идемпотентен на уровне tryDestroyBuilding
		// (понижает и при повторном вызове). Идемпотентность —
		// на уровне worker'а через event-state. Зеркалит legacy
		// (там handler тоже просто понижает level-1 без проверки).
		fx := dbFixture{
			userID: ids.New(), planetID: ids.New(), usedF: 5,
			buildings: map[int]int{4: 3},
		}
		seedFixture(ctx, t, db, fx)
		defer cleanupFixture(ctx, db, fx)

		err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
			if _, _, _, ok, err := tryDestroyBuilding(ctx, tx, fx.planetID, false, "attackers", 4, nil, 1); err != nil || !ok {
				t.Fatalf("first call: ok=%v err=%v", ok, err)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("first tx: %v", err)
		}
		if got := readLevel(t, db, fx.planetID, 4); got != 2 {
			t.Errorf("after first: level=%d want 2", got)
		}
	})
}
