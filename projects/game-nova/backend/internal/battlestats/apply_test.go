package battlestats

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/battle"
	"oxsar/game-nova/pkg/ids"
)

// План 72.1.1: integration-тесты ApplyBattleResult (TEST_DATABASE_URL).
// Без БД skipped (как destroy_building_test.go).

func openTestPool(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping battlestats integration tests")
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
	return pool, func() { pool.Close() }
}

func seedUser(t *testing.T, ctx context.Context, tx pgx.Tx, userID, suffix string) {
	t.Helper()
	if _, err := tx.Exec(ctx, `
		INSERT INTO users (id, username, email, password_hash, registered_at, universe_id)
		VALUES ($1, $2, $3, '', now(),
		        COALESCE((SELECT id FROM universes LIMIT 1),
		                 '00000000-0000-0000-0000-000000000000'::uuid))
		ON CONFLICT (id) DO NOTHING
	`, userID, "bstats-"+suffix, "bstats-"+suffix+"@test"); err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func cleanupUser(ctx context.Context, tx pgx.Tx, userID string) {
	_, _ = tx.Exec(ctx, `DELETE FROM user_experience WHERE user_id=$1`, userID)
	_, _ = tx.Exec(ctx, `DELETE FROM users WHERE id=$1`, userID)
}

func seedBattleReport(t *testing.T, ctx context.Context, tx pgx.Tx, battleID, attID, defID string) {
	t.Helper()
	if _, err := tx.Exec(ctx, `
		INSERT INTO battle_reports (id, attacker_user_id, defender_user_id, planet_id,
		                            seed, winner, rounds,
		                            debris_metal, debris_silicon,
		                            loot_metal, loot_silicon, loot_hydrogen, report)
		VALUES ($1, $2, $3, NULL, 0, 'attackers', 1, 0, 0, 0, 0, 0, '{}'::jsonb)
	`, battleID, attID, defID); err != nil {
		t.Fatalf("seed battle_reports: %v", err)
	}
}

func readUserStats(t *testing.T, ctx context.Context, tx pgx.Tx, userID string) (ePoints, bePoints, points, uPoints float64, battles, uCount int) {
	t.Helper()
	if err := tx.QueryRow(ctx, `
		SELECT e_points, be_points, points, u_points, battles, u_count
		FROM users WHERE id=$1
	`, userID).Scan(&ePoints, &bePoints, &points, &uPoints, &battles, &uCount); err != nil {
		t.Fatalf("read user stats: %v", err)
	}
	return
}

// TestApplyBattleResult_HappyPath — простой бой 1 атак vs 1 защ:
// атакующий получает +AtterExp в e_points/be_points, +1 battles,
// защитник — то же по своим Defender-полям.
func TestApplyBattleResult_HappyPath(t *testing.T) {
	pool, closeFn := openTestPool(t)
	defer closeFn()
	ctx := context.Background()

	atkID := ids.New()
	defID := ids.New()
	battleID := ids.New()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() {
		cleanupUser(ctx, tx, atkID)
		cleanupUser(ctx, tx, defID)
		_, _ = tx.Exec(ctx, `DELETE FROM battle_reports WHERE id=$1`, battleID)
		_ = tx.Rollback(ctx)
	}()

	seedUser(t, ctx, tx, atkID, "att")
	seedUser(t, ctx, tx, defID, "def")
	// Дадим обоим стартовый points/u_count чтобы списания списались.
	if _, err := tx.Exec(ctx, `
		UPDATE users SET points=10000, u_points=10000, u_count=100 WHERE id=ANY($1::uuid[])
	`, []string{atkID, defID}); err != nil {
		t.Fatalf("init points: %v", err)
	}
	seedBattleReport(t, ctx, tx, battleID, atkID, defID)

	report := battle.Report{
		Winner:      "attackers",
		Rounds:      3,
		AttackerExp: 7,
		DefenderExp: 3,
		Attackers: []battle.SideResult{{
			UserID: atkID, LostPoints: 100, LostUnits: 5,
		}},
		Defenders: []battle.SideResult{{
			UserID: defID, LostPoints: 800, LostUnits: 50,
		}},
	}

	if err := ApplyBattleResult(ctx, tx, report, battleID); err != nil {
		t.Fatalf("apply: %v", err)
	}

	atkE, atkBE, atkP, atkUP, atkB, atkUC := readUserStats(t, ctx, tx, atkID)
	if atkE != 7 || atkBE != 7 || atkB != 1 {
		t.Fatalf("attacker exp/battles: got e=%v be=%v batt=%d, want 7/7/1", atkE, atkBE, atkB)
	}
	if atkP != 9900 || atkUP != 9900 || atkUC != 95 {
		t.Fatalf("attacker losses: got p=%v up=%v uc=%d, want 9900/9900/95", atkP, atkUP, atkUC)
	}

	defE, defBE, defP, defUP, defB, defUC := readUserStats(t, ctx, tx, defID)
	if defE != 3 || defBE != 3 || defB != 1 {
		t.Fatalf("defender exp/battles: got e=%v be=%v batt=%d, want 3/3/1", defE, defBE, defB)
	}
	if defP != 9200 || defUP != 9200 || defUC != 50 {
		t.Fatalf("defender losses: got p=%v up=%v uc=%d, want 9200/9200/50", defP, defUP, defUC)
	}

	// user_experience: 2 строки.
	var n int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM user_experience WHERE battle_id=$1`,
		battleID).Scan(&n); err != nil {
		t.Fatalf("count user_experience: %v", err)
	}
	if n != 2 {
		t.Fatalf("user_experience rows: got %d, want 2", n)
	}
}

// TestApplyBattleResult_Idempotent — повторный вызов с тем же battleID
// возвращает ErrAlreadyApplied и не меняет данные.
func TestApplyBattleResult_Idempotent(t *testing.T) {
	pool, closeFn := openTestPool(t)
	defer closeFn()
	ctx := context.Background()

	atkID := ids.New()
	defID := ids.New()
	battleID := ids.New()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() {
		cleanupUser(ctx, tx, atkID)
		cleanupUser(ctx, tx, defID)
		_, _ = tx.Exec(ctx, `DELETE FROM battle_reports WHERE id=$1`, battleID)
		_ = tx.Rollback(ctx)
	}()

	seedUser(t, ctx, tx, atkID, "att2")
	seedUser(t, ctx, tx, defID, "def2")
	seedBattleReport(t, ctx, tx, battleID, atkID, defID)

	report := battle.Report{
		Winner: "draw", Rounds: 2, AttackerExp: 4, DefenderExp: 4,
		Attackers: []battle.SideResult{{UserID: atkID}},
		Defenders: []battle.SideResult{{UserID: defID}},
	}

	if err := ApplyBattleResult(ctx, tx, report, battleID); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	atkE1, _, _, _, atkB1, _ := readUserStats(t, ctx, tx, atkID)

	// Второй вызов — должен вернуть ErrAlreadyApplied.
	err2 := ApplyBattleResult(ctx, tx, report, battleID)
	if !errors.Is(err2, ErrAlreadyApplied) {
		t.Fatalf("second apply: want ErrAlreadyApplied, got %v", err2)
	}

	atkE2, _, _, _, atkB2, _ := readUserStats(t, ctx, tx, atkID)
	if atkE1 != atkE2 || atkB1 != atkB2 {
		t.Fatalf("second apply changed state: e %v→%v, batt %d→%d",
			atkE1, atkE2, atkB1, atkB2)
	}
}

// TestApplyBattleResult_AliensSkipped — IsAliens-сторона skip'ается:
// для unit-теста только NPC vs реальный юзер (defender), у которого
// должны зачислиться очки.
func TestApplyBattleResult_AliensSkipped(t *testing.T) {
	pool, closeFn := openTestPool(t)
	defer closeFn()
	ctx := context.Background()

	defID := ids.New()
	battleID := ids.New()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	defer func() {
		cleanupUser(ctx, tx, defID)
		_, _ = tx.Exec(ctx, `DELETE FROM battle_reports WHERE id=$1`, battleID)
		_ = tx.Rollback(ctx)
	}()

	seedUser(t, ctx, tx, defID, "alien-def")
	if _, err := tx.Exec(ctx, `INSERT INTO battle_reports (id, attacker_user_id, defender_user_id, planet_id, seed, winner, rounds, debris_metal, debris_silicon, loot_metal, loot_silicon, loot_hydrogen, report) VALUES ($1, NULL, $2, NULL, 0, 'defenders', 1, 0, 0, 0, 0, 0, '{}'::jsonb)`,
		battleID, defID); err != nil {
		t.Fatalf("seed report: %v", err)
	}

	report := battle.Report{
		Winner: "defenders", Rounds: 2, AttackerExp: 5, DefenderExp: 9,
		Attackers: []battle.SideResult{{UserID: "aliens", IsAliens: true}},
		Defenders: []battle.SideResult{{UserID: defID, LostUnits: 2, LostPoints: 50}},
	}

	if err := ApplyBattleResult(ctx, tx, report, battleID); err != nil {
		t.Fatalf("apply: %v", err)
	}

	// Defender получил опыт.
	defE, defBE, _, _, defB, _ := readUserStats(t, ctx, tx, defID)
	if defE != 9 || defBE != 9 || defB != 1 {
		t.Fatalf("defender: got e=%v be=%v batt=%d, want 9/9/1", defE, defBE, defB)
	}

	// В user_experience нет записи для "aliens" UserID (он не uuid и
	// должен skip'нуться).
	var aliensRows int
	_ = tx.QueryRow(ctx, `SELECT count(*) FROM user_experience WHERE user_id::text = 'aliens'`).
		Scan(&aliensRows)
	if aliensRows != 0 {
		t.Fatalf("aliens row leaked into user_experience: %d", aliensRows)
	}
}

// Sanity-проверка что fmt используется (фиксирует unused import при
// будущих рефакторах test-файла).
var _ = fmt.Sprintf
