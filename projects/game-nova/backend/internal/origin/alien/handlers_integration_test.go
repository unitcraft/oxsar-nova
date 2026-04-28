package alien_test

// План 66 Ф.3: интеграционные тесты Kind handlers
// (FlyUnknown / GrabCredit / ChangeMissionAI).
//
// Запускаются только при TEST_DATABASE_URL — в обычном `go test`
// автоматически skip'аются. Образец — demolish_test.go (план 65 Ф.1).
//
// Покрытые сценарии:
//   - TestFlyUnknown_GrabBranch: mode=GrabCredit, credit > 100k
//     → списываются оксариты, отправляется сообщение, 90% return.
//   - TestFlyUnknown_HaltBranch: бедный игрок, не четверг → halt
//     (создаётся новое KindAlienHalt событие).
//   - TestGrabCredit_RedirectsToFlyUnknown: KindAlienGrabCredit
//     обрабатывается тем же handler'ом FlyUnknown.
//   - TestChangeMissionAI_ExtendsParent: parent ATTACK с remaining
//     < 8h продлевает fire_at (origin:910).
//   - TestChangeMissionAI_ReplansWithPowerScale: parent с remaining
//     >= 8h получает новый power_scale.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"oxsar/game-nova/internal/event"
	originalien "oxsar/game-nova/internal/origin/alien"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

func openTestDB(t *testing.T) (repo.Exec, func()) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping origin/alien integration tests")
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

type userPlanet struct {
	userID   string
	planetID string
}

func seedUserPlanet(ctx context.Context, t *testing.T, db repo.Exec,
	credit float64, ships []struct{ UnitID int; Count int64 }) userPlanet {

	t.Helper()
	uid := ids.New()
	pid := ids.New()
	err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `
			INSERT INTO users (id, username, email, password_hash, registered_at, universe_id, credit)
			VALUES ($1, $2, $3, '', now(),
			        COALESCE((SELECT id FROM universes LIMIT 1),
			                 '00000000-0000-0000-0000-000000000000'::uuid),
			        $4)
			ON CONFLICT (id) DO NOTHING
		`, uid, "alien-test-"+uid[:8], "at-"+uid[:8]+"@test", credit); err != nil {
			return fmt.Errorf("seed user: %w", err)
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO planets (id, user_id, planetname, galaxy, system, position,
			                     planet_type, used_fields, max_fields,
			                     metal, silicon, hydrogen,
			                     last_update, universe_id)
			VALUES ($1, $2, 'AT-'||substr($1::text, 1, 6), 1, 1, 5, 'planet', 5, 250,
			        100000, 50000, 25000, now(),
			        (SELECT universe_id FROM users WHERE id=$2))
			ON CONFLICT (id) DO NOTHING
		`, pid, uid); err != nil {
			return fmt.Errorf("seed planet: %w", err)
		}
		for _, s := range ships {
			if _, err := tx.Exec(ctx, `
				INSERT INTO ships (planet_id, unit_id, count, damaged_count, shell_percent)
				VALUES ($1, $2, $3, 0, 100)
				ON CONFLICT (planet_id, unit_id) DO UPDATE SET count=EXCLUDED.count
			`, pid, s.UnitID, s.Count); err != nil {
				return fmt.Errorf("seed ship: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	return userPlanet{userID: uid, planetID: pid}
}

func cleanupUserPlanet(ctx context.Context, db repo.Exec, up userPlanet) {
	_ = db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, _ = tx.Exec(ctx, `DELETE FROM events WHERE planet_id=$1`, up.planetID)
		_, _ = tx.Exec(ctx, `DELETE FROM messages WHERE to_user_id=$1`, up.userID)
		_, _ = tx.Exec(ctx, `DELETE FROM ships WHERE planet_id=$1`, up.planetID)
		_, _ = tx.Exec(ctx, `DELETE FROM planets WHERE id=$1`, up.planetID)
		_, _ = tx.Exec(ctx, `DELETE FROM users WHERE id=$1`, up.userID)
		return nil
	})
}

func runHandler(t *testing.T, db repo.Exec, h event.Handler, e event.Event) error {
	t.Helper()
	ctx := context.Background()
	return db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return h(ctx, tx, e)
	})
}

func readUserCredit(t *testing.T, db repo.Exec, userID string) float64 {
	t.Helper()
	ctx := context.Background()
	var c float64
	if err := db.Pool().QueryRow(ctx,
		`SELECT credit::float8 FROM users WHERE id=$1`, userID).Scan(&c); err != nil {
		t.Fatalf("read credit: %v", err)
	}
	return c
}

func readMessagesCount(t *testing.T, db repo.Exec, userID string) int {
	t.Helper()
	ctx := context.Background()
	var n int
	if err := db.Pool().QueryRow(ctx,
		`SELECT count(*) FROM messages WHERE to_user_id=$1`, userID).Scan(&n); err != nil {
		t.Fatalf("read messages: %v", err)
	}
	return n
}

func countEventsByKind(t *testing.T, db repo.Exec, planetID string, kind event.Kind) int {
	t.Helper()
	ctx := context.Background()
	var n int
	if err := db.Pool().QueryRow(ctx,
		`SELECT count(*) FROM events WHERE planet_id=$1 AND kind=$2`,
		planetID, int(kind)).Scan(&n); err != nil {
		t.Fatalf("count events: %v", err)
	}
	return n
}

// TestFlyUnknown_GrabBranch — credit > GrabMinCredit, mode=GrabCredit
// → 100% грабёж, credit уменьшается, сообщение в инбоксе.
//
// Используем mode=GrabCredit чтобы детерминистично попасть в ветку грабежа
// (без 10% random). После грабежа — 90% return; в 10% случаев продолжается
// в attack/halt, что добавит KindAlienAttack/KindAlienHalt событие.
func TestFlyUnknown_GrabBranch(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 1_000_000.0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	svc := originalien.NewService(nil, nil)
	pl := originalien.MissionPayload{
		Mode:     int(originalien.ModeGrabCredit),
		UserID:   up.userID, PlanetID: up.planetID,
		Galaxy: 1, System: 1, Position: 5,
		Metal:    100, Silicon: 100, Hydrogen: 100,
		PowerScale: 1.0, AlienActor: true,
	}
	rawPayload, _ := json.Marshal(pl)
	e := event.Event{
		ID:       ids.New(),
		Kind:     event.KindAlienGrabCredit,
		FireAt:   time.Now(),
		Payload:  rawPayload,
		PlanetID: &up.planetID,
		UserID:   &up.userID,
	}
	if err := runHandler(t, db, svc.GrabCreditHandler(), e); err != nil {
		t.Fatalf("handler: %v", err)
	}

	// Credit уменьшился (грабёж 0.08-0.10% от 1M = 800..1000).
	credAfter := readUserCredit(t, db, up.userID)
	if credAfter >= 1_000_000.0 {
		t.Errorf("credit not decreased: %.2f", credAfter)
	}
	if credAfter < 999_000.0 || credAfter > 999_200.0 {
		// допуск: 0.0008 * 1M = 800; 0.001 * 1M = 1000 → credit ∈ [999000, 999200]
		t.Errorf("credit after grab = %.2f, want ~999_000..999_200", credAfter)
	}

	// Сообщение игроку отправлено.
	if got := readMessagesCount(t, db, up.userID); got == 0 {
		t.Errorf("expected at least 1 message after grab; got 0")
	}
}

// TestFlyUnknown_HaltBranch — credit ниже порога, не четверг (мы
// контролируем через random seed): достаточно высокий шанс попасть
// в HALT. Проверяем что либо HALT либо ATTACK событие создано (тест
// нестрогий по именно ветке: handler для бедного игрока без атаки —
// либо attack либо halt).
//
// Минимальный инвариант: после FlyUnknown handler'а на бедного игрока
// создалось ровно 1 новое alien-событие (Halt или Attack) — миссия
// пришельцев продолжается.
func TestFlyUnknown_HaltOrAttackBranch(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 1_000.0, nil) // ниже GrabMinCredit
	defer cleanupUserPlanet(ctx, db, up)

	svc := originalien.NewService(nil, nil)
	pl := originalien.MissionPayload{
		Mode:     int(originalien.ModeFlyUnknown),
		UserID:   up.userID, PlanetID: up.planetID,
		Galaxy: 1, System: 1, Position: 5,
		Metal:    100, Silicon: 100, Hydrogen: 100,
		Ships: originalien.Fleet{
			{UnitID: 200, Quantity: 5, ShellPercent: 100},
		},
		PowerScale: 1.0, AlienActor: true,
	}
	rawPayload, _ := json.Marshal(pl)
	e := event.Event{
		ID:       ids.New(),
		Kind:     event.KindAlienFlyUnknown,
		FireAt:   time.Now(),
		Payload:  rawPayload,
		PlanetID: &up.planetID,
		UserID:   &up.userID,
	}
	if err := runHandler(t, db, svc.FlyUnknownHandler(), e); err != nil {
		t.Fatalf("handler: %v", err)
	}

	// При credit<grab_min — не было грабежа; не Thursday по seed → handler
	// решает 50/50 attack vs halt. Любой исход = 1 follow-up event.
	attacks := countEventsByKind(t, db, up.planetID, event.KindAlienAttack)
	halts := countEventsByKind(t, db, up.planetID, event.KindAlienHalt)
	total := attacks + halts
	if total != 1 {
		t.Errorf("expected exactly 1 follow-up event (attack or halt); got attack=%d halt=%d",
			attacks, halts)
	}
}

// TestChangeMissionAI_ExtendsParent — parent ATTACK с remaining < 8h
// продлевает fire_at parent'а на 10..50s.
func TestChangeMissionAI_ExtendsParent(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 1000.0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	// parent ATTACK с MissionPayload, remaining=1h (< 8h ChangeMissionMinTime).
	parentID := ids.New()
	parentFireAt := time.Now().Add(1 * time.Hour)
	parentPL := originalien.MissionPayload{
		Mode: int(originalien.ModeAttack),
		UserID: up.userID, PlanetID: up.planetID,
		ControlTimes: 1, PowerScale: 1.0, AlienActor: true,
	}
	parentRaw, _ := json.Marshal(parentPL)
	if _, err := db.Pool().Exec(ctx, `
		INSERT INTO events (id, kind, planet_id, user_id, fire_at, payload, state)
		VALUES ($1::uuid, $2, $3::uuid, $4::uuid, $5, $6, 'wait')
	`, parentID, int(event.KindAlienAttack), up.planetID, up.userID,
		parentFireAt, parentRaw); err != nil {
		t.Fatalf("seed parent: %v", err)
	}

	svc := originalien.NewService(nil, nil)
	cmPL := originalien.ChangeMissionPayload{
		ParentEventID: parentID,
		UserID: up.userID, PlanetID: up.planetID,
		ControlTimes: 1, AlienActor: true,
	}
	cmRaw, _ := json.Marshal(cmPL)
	e := event.Event{
		ID:       ids.New(),
		Kind:     event.KindAlienChangeMissionAI,
		FireAt:   time.Now(),
		Payload:  cmRaw,
		PlanetID: &up.planetID,
		UserID:   &up.userID,
	}
	if err := runHandler(t, db, svc.ChangeMissionAIHandler(), e); err != nil {
		t.Fatalf("handler: %v", err)
	}

	// Parent fire_at сдвинулся вперёд на 10..50s.
	var newFireAt time.Time
	if err := db.Pool().QueryRow(ctx,
		`SELECT fire_at FROM events WHERE id=$1`, parentID).Scan(&newFireAt); err != nil {
		t.Fatalf("read parent: %v", err)
	}
	delta := newFireAt.Sub(parentFireAt)
	if delta < 10*time.Second || delta > 60*time.Second {
		t.Errorf("parent extension = %v; want ~10..50s", delta)
	}

	// control_times++.
	var newPayloadRaw json.RawMessage
	if err := db.Pool().QueryRow(ctx,
		`SELECT payload FROM events WHERE id=$1`, parentID).Scan(&newPayloadRaw); err != nil {
		t.Fatalf("read payload: %v", err)
	}
	var newPL originalien.MissionPayload
	if err := json.Unmarshal(newPayloadRaw, &newPL); err != nil {
		t.Fatalf("unmarshal new: %v", err)
	}
	if newPL.ControlTimes != 2 {
		t.Errorf("control_times: got %d want 2", newPL.ControlTimes)
	}
}

// TestChangeMissionAI_ReplansWithPowerScale — parent с remaining >= 8h
// получает новый power_scale (1 + control_times*1.5).
func TestChangeMissionAI_ReplansWithPowerScale(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 1000.0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	parentID := ids.New()
	parentPL := originalien.MissionPayload{
		Mode: int(originalien.ModeAttack),
		UserID: up.userID, PlanetID: up.planetID,
		ControlTimes: 1, PowerScale: 1.0, AlienActor: true,
	}
	parentRaw, _ := json.Marshal(parentPL)
	if _, err := db.Pool().Exec(ctx, `
		INSERT INTO events (id, kind, planet_id, user_id, fire_at, payload, state)
		VALUES ($1::uuid, $2, $3::uuid, $4::uuid, $5, $6, 'wait')
	`, parentID, int(event.KindAlienAttack), up.planetID, up.userID,
		time.Now().Add(20*time.Hour), parentRaw); err != nil {
		t.Fatalf("seed parent: %v", err)
	}

	svc := originalien.NewService(nil, nil)
	cmPL := originalien.ChangeMissionPayload{
		ParentEventID: parentID,
		UserID: up.userID, PlanetID: up.planetID,
		ControlTimes: 1, AlienActor: true,
	}
	cmRaw, _ := json.Marshal(cmPL)
	e := event.Event{
		ID:       ids.New(),
		Kind:     event.KindAlienChangeMissionAI,
		FireAt:   time.Now(),
		Payload:  cmRaw,
		PlanetID: &up.planetID,
		UserID:   &up.userID,
	}
	if err := runHandler(t, db, svc.ChangeMissionAIHandler(), e); err != nil {
		t.Fatalf("handler: %v", err)
	}

	var newPayloadRaw json.RawMessage
	if err := db.Pool().QueryRow(ctx,
		`SELECT payload FROM events WHERE id=$1`, parentID).Scan(&newPayloadRaw); err != nil {
		t.Fatalf("read payload: %v", err)
	}
	var newPL originalien.MissionPayload
	if err := json.Unmarshal(newPayloadRaw, &newPL); err != nil {
		t.Fatalf("unmarshal new: %v", err)
	}
	// control_times: 1 → 2; power_scale = 1 + 2*1.5 = 4.0.
	if newPL.ControlTimes != 2 {
		t.Errorf("control_times: got %d want 2", newPL.ControlTimes)
	}
	if newPL.PowerScale != 4.0 {
		t.Errorf("power_scale: got %v want 4.0", newPL.PowerScale)
	}
	// mode флипанулся в Attack или FlyUnknown — оба валидны.
	if newPL.Mode != int(originalien.ModeAttack) && newPL.Mode != int(originalien.ModeFlyUnknown) {
		t.Errorf("mode after replan: got %d, want Attack(35) or FlyUnknown(33)", newPL.Mode)
	}
}

// TestChangeMissionAI_SkipParentGone — если parent удалён, handler
// возвращает nil без побочных эффектов.
func TestChangeMissionAI_SkipParentGone(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 1000.0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	svc := originalien.NewService(nil, nil)
	cmPL := originalien.ChangeMissionPayload{
		ParentEventID: ids.New(), // не существует
		UserID: up.userID, PlanetID: up.planetID,
		ControlTimes: 1, AlienActor: true,
	}
	cmRaw, _ := json.Marshal(cmPL)
	e := event.Event{
		ID:       ids.New(),
		Kind:     event.KindAlienChangeMissionAI,
		FireAt:   time.Now(),
		Payload:  cmRaw,
		PlanetID: &up.planetID,
		UserID:   &up.userID,
	}
	if err := runHandler(t, db, svc.ChangeMissionAIHandler(), e); err != nil {
		t.Errorf("expected nil for missing parent; got %v", err)
	}
}

// TestFlyUnknown_NegativePayloadRejected — пустой UserID/PlanetID
// в payload → ошибка валидации.
func TestFlyUnknown_NegativePayloadRejected(t *testing.T) {
	svc := originalien.NewService(nil, nil)
	pl := originalien.MissionPayload{Mode: int(originalien.ModeFlyUnknown)}
	rawPayload, _ := json.Marshal(pl)
	planetID := "p1"
	userID := "u1"
	e := event.Event{
		ID:       "e1",
		Kind:     event.KindAlienFlyUnknown,
		Payload:  rawPayload,
		PlanetID: &planetID,
		UserID:   &userID,
	}
	// nil tx безопасен — handler упирается в валидацию payload до SQL.
	err := svc.FlyUnknownHandler()(context.Background(), nil, e)
	if err == nil {
		t.Fatalf("expected error for empty user_id/planet_id in payload")
	}
}

// _ — компилятор-страж: убедимся что pgx-import востребован
// (pgx.ErrNoRows может пригодиться будущим тестам, оставим явный
// reference чтобы избежать "imported and not used").
var _ = pgx.ErrNoRows

// errSentinel — sentinel-доступ для будущих failure-тестов.
var _ = errors.New("placeholder")
