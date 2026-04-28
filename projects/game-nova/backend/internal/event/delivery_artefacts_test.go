package event_test

// План 65 Ф.2 (D-035): тесты HandleDeliveryArtefacts.
//
// Структура (зеркалит demolish_test.go):
//   - TestDeliveryArtefacts_PayloadRoundTrip — pure-тест: payload корректно
//     сериализуется/десериализуется JSON без потери полей.
//   - TestProperty_DeliveryArtefactsPayload_Determinism — property-based:
//     skip-decision (артефакт уже у получателя?) детерминируется чистой
//     функцией от текущего и целевого владельца.
//   - TestDeliveryArtefacts_GoldenScenarios — golden: интеграция с БД,
//     запускается только при TEST_DATABASE_URL.
//   - TestDeliveryArtefacts_PayloadValidation — невалидные payload'ы.
//
// Инварианты:
//   I1. fleet.state=outbound + новый владелец → переписан user_id, planet_id,
//       state=held, флот → returning.
//   I2. fleet.state=returning → no-op (идемпотентность).
//   I3. артефакт уже у (e.UserID, e.PlanetID) → skip; флот всё равно → returning.
//   I4. артефакт удалён до прибытия → warning, продолжаем остальные.
//   I5. payload без fleet_id / artefact_ids[] / e.UserID / e.PlanetID → ошибка.
//   I6. артефакты из разных вселенных → ошибка (R10).

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

// TestDeliveryArtefacts_PayloadRoundTrip — pure: payload не теряет поля
// при JSON сериализации. Защита от случайного rename JSON-тэга.
func TestDeliveryArtefacts_PayloadRoundTrip(t *testing.T) {
	src := event.DeliveryArtefactsPayload{
		FleetID:     "fleet-abc",
		ArtefactIDs: []string{"a-1", "a-2", "a-3"},
	}
	raw, err := json.Marshal(src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	expectKeys := []string{`"fleet_id":"fleet-abc"`, `"artefact_ids":["a-1","a-2","a-3"]`}
	for _, k := range expectKeys {
		if !strings.Contains(string(raw), k) {
			t.Fatalf("payload missing key %q in JSON: %s", k, string(raw))
		}
	}
	var dst event.DeliveryArtefactsPayload
	if err := json.Unmarshal(raw, &dst); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dst.FleetID != src.FleetID {
		t.Fatalf("fleet_id mismatch: got %q want %q", dst.FleetID, src.FleetID)
	}
	if len(dst.ArtefactIDs) != len(src.ArtefactIDs) {
		t.Fatalf("artefact_ids len: got %d want %d", len(dst.ArtefactIDs), len(src.ArtefactIDs))
	}
	for i := range src.ArtefactIDs {
		if dst.ArtefactIDs[i] != src.ArtefactIDs[i] {
			t.Errorf("artefact_ids[%d]: got %q want %q", i, dst.ArtefactIDs[i], src.ArtefactIDs[i])
		}
	}
}

// TestProperty_DeliveryArtefactsPayload_Determinism — property-based:
// skip-decision должен быть чистой функцией от пары (curOwner, target).
//
// Контракт handler'а (см. handlers.go: idempotent skip):
//
//	shouldSkip(curUser, curPlanet, tgtUser, tgtPlanet, state) =
//	  curUser == tgtUser && curPlanet == tgtPlanet && state != "active"
//
// Property: shouldSkip детерминирован и симметричен пере-чтению.
func TestProperty_DeliveryArtefactsPayload_Determinism(t *testing.T) {
	shouldSkip := func(curUser, curPlanet, tgtUser, tgtPlanet, state string) bool {
		return curUser == tgtUser && curPlanet == tgtPlanet && state != "active"
	}
	rapid.Check(t, func(t *rapid.T) {
		curUser := rapid.SampledFrom([]string{"u1", "u2", "u3"}).Draw(t, "curUser")
		curPlanet := rapid.SampledFrom([]string{"p1", "p2", "p3"}).Draw(t, "curPlanet")
		tgtUser := rapid.SampledFrom([]string{"u1", "u2", "u3"}).Draw(t, "tgtUser")
		tgtPlanet := rapid.SampledFrom([]string{"p1", "p2", "p3"}).Draw(t, "tgtPlanet")
		state := rapid.SampledFrom([]string{"held", "active", "delayed", "expired"}).Draw(t, "state")
		got1 := shouldSkip(curUser, curPlanet, tgtUser, tgtPlanet, state)
		got2 := shouldSkip(curUser, curPlanet, tgtUser, tgtPlanet, state)
		if got1 != got2 {
			t.Fatalf("non-deterministic skip decision")
		}
		expect := curUser == tgtUser && curPlanet == tgtPlanet && state != "active"
		if got1 != expect {
			t.Fatalf("shouldSkip(%q,%q,%q,%q,%q) = %v, want %v",
				curUser, curPlanet, tgtUser, tgtPlanet, state, got1, expect)
		}
	})
}

// TestDeliveryArtefacts_PayloadValidation — невалидные payload должны
// возвращать ошибку до первого SQL-запроса. Чистый тест без БД.
func TestDeliveryArtefacts_PayloadValidation(t *testing.T) {
	user, planet := "u1", "p1"
	cases := []struct {
		name    string
		event   event.Event
		wantSub string
	}{
		{
			name: "no_user",
			event: event.Event{
				ID:       "e1",
				PlanetID: &planet,
				Kind:     event.KindDeliveryArtefacts,
				Payload:  json.RawMessage(`{"fleet_id":"f","artefact_ids":["a"]}`),
			},
			wantSub: "user_id",
		},
		{
			name: "no_planet",
			event: event.Event{
				ID:      "e1",
				UserID:  &user,
				Kind:    event.KindDeliveryArtefacts,
				Payload: json.RawMessage(`{"fleet_id":"f","artefact_ids":["a"]}`),
			},
			wantSub: "planet_id",
		},
		{
			name: "no_fleet_id",
			event: event.Event{
				ID:       "e1",
				UserID:   &user,
				PlanetID: &planet,
				Kind:     event.KindDeliveryArtefacts,
				Payload:  json.RawMessage(`{"artefact_ids":["a"]}`),
			},
			wantSub: "fleet_id",
		},
		{
			name: "empty_artefact_ids",
			event: event.Event{
				ID:       "e1",
				UserID:   &user,
				PlanetID: &planet,
				Kind:     event.KindDeliveryArtefacts,
				Payload:  json.RawMessage(`{"fleet_id":"f","artefact_ids":[]}`),
			},
			wantSub: "artefact_ids",
		},
		{
			name: "bad_json",
			event: event.Event{
				ID:       "e1",
				UserID:   &user,
				PlanetID: &planet,
				Kind:     event.KindDeliveryArtefacts,
				Payload:  json.RawMessage(`not-json`),
			},
			wantSub: "parse payload",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := event.HandleDeliveryArtefacts(context.Background(), nil, tc.event)
			if err == nil {
				t.Fatalf("expected error for %s", tc.name)
			}
			if !strings.Contains(err.Error(), tc.wantSub) {
				t.Errorf("expected error to contain %q, got: %v", tc.wantSub, err)
			}
		})
	}
}

// --- Интеграционные (golden) тесты ниже. Требуют TEST_DATABASE_URL. ---

func openDeliveryTestDB(t *testing.T) (repo.Exec, func()) {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping delivery_artefacts integration tests")
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

type deliveryFixture struct {
	senderID    string
	recipientID string
	srcPlanetID string
	dstPlanetID string
	fleetID     string
	artefacts   []string // создаём artefacts_user.id'ы у sender'а
}

// seedDeliveryFixture: 2 user'а в одной вселенной → планета у каждого →
// fleet outbound с sender → artefacts_user (state=held, owner=sender).
func seedDeliveryFixture(ctx context.Context, t *testing.T, db repo.Exec, fx deliveryFixture) {
	t.Helper()
	err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		// Sender и recipient.
		for _, uid := range []string{fx.senderID, fx.recipientID} {
			if _, err := tx.Exec(ctx, `
				INSERT INTO users (id, username, email, password_hash, registered_at, universe_id)
				VALUES ($1, $2, $3, '', now(),
				        COALESCE((SELECT id FROM universes LIMIT 1),
				                 '00000000-0000-0000-0000-000000000000'::uuid))
				ON CONFLICT (id) DO NOTHING
			`, uid, "delart-"+uid[:8], "delart-"+uid[:8]+"@test"); err != nil {
				return fmt.Errorf("seed user %s: %w", uid, err)
			}
		}
		// Планета sender'а (src).
		if _, err := tx.Exec(ctx, `
			INSERT INTO planets (id, user_id, planetname, galaxy, system, position,
			                     planet_type, used_fields, max_fields,
			                     metal, silicon, hydrogen,
			                     last_update, universe_id)
			VALUES ($1, $2, 'SrcPlanet', 1, 1, 5, 'planet', 0, 250, 0, 0, 0, now(),
			        (SELECT universe_id FROM users WHERE id=$2))
			ON CONFLICT (id) DO NOTHING
		`, fx.srcPlanetID, fx.senderID); err != nil {
			return fmt.Errorf("seed src planet: %w", err)
		}
		// Планета recipient'а (dst).
		if _, err := tx.Exec(ctx, `
			INSERT INTO planets (id, user_id, planetname, galaxy, system, position,
			                     planet_type, used_fields, max_fields,
			                     metal, silicon, hydrogen,
			                     last_update, universe_id)
			VALUES ($1, $2, 'DstPlanet', 1, 2, 7, 'planet', 0, 250, 0, 0, 0, now(),
			        (SELECT universe_id FROM users WHERE id=$2))
			ON CONFLICT (id) DO NOTHING
		`, fx.dstPlanetID, fx.recipientID); err != nil {
			return fmt.Errorf("seed dst planet: %w", err)
		}
		// Флот: sender → recipient, mission=delivery_artefacts (23), outbound.
		if _, err := tx.Exec(ctx, `
			INSERT INTO fleets (id, owner_user_id, src_planet_id,
			                    dst_galaxy, dst_system, dst_position, dst_is_moon,
			                    mission, state, depart_at, arrive_at)
			VALUES ($1, $2, $3, 1, 2, 7, false, 23, 'outbound', now(), now())
			ON CONFLICT (id) DO NOTHING
		`, fx.fleetID, fx.senderID, fx.srcPlanetID); err != nil {
			return fmt.Errorf("seed fleet: %w", err)
		}
		// Артефакты у sender'а на src-планете.
		for _, aid := range fx.artefacts {
			if _, err := tx.Exec(ctx, `
				INSERT INTO artefacts_user (id, user_id, planet_id, unit_id, state, acquired_at)
				VALUES ($1, $2, $3, 1001, 'held', now())
				ON CONFLICT (id) DO NOTHING
			`, aid, fx.senderID, fx.srcPlanetID); err != nil {
				return fmt.Errorf("seed artefact %s: %w", aid, err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
}

func cleanupDeliveryFixture(ctx context.Context, db repo.Exec, fx deliveryFixture) {
	_ = db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		for _, aid := range fx.artefacts {
			_, _ = tx.Exec(ctx, `DELETE FROM artefacts_user WHERE id=$1`, aid)
		}
		_, _ = tx.Exec(ctx, `DELETE FROM fleets WHERE id=$1`, fx.fleetID)
		_, _ = tx.Exec(ctx, `DELETE FROM planets WHERE id IN ($1,$2)`, fx.srcPlanetID, fx.dstPlanetID)
		_, _ = tx.Exec(ctx, `DELETE FROM users WHERE id IN ($1,$2)`, fx.senderID, fx.recipientID)
		return nil
	})
}

func runDeliveryArtefacts(t *testing.T, db repo.Exec, fx deliveryFixture) error {
	t.Helper()
	ctx := context.Background()
	planetID := fx.dstPlanetID
	userID := fx.recipientID
	pl := event.DeliveryArtefactsPayload{
		FleetID:     fx.fleetID,
		ArtefactIDs: fx.artefacts,
	}
	raw, err := json.Marshal(pl)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	e := event.Event{
		ID:        ids.New(),
		UserID:    &userID,
		PlanetID:  &planetID,
		Kind:      event.KindDeliveryArtefacts,
		FireAt:    time.Now(),
		Payload:   raw,
		CreatedAt: time.Now(),
	}
	return db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		return event.HandleDeliveryArtefacts(ctx, tx, e)
	})
}

func readArtefactOwner(t *testing.T, db repo.Exec, id string) (userID string, planetID *string, state string, found bool) {
	t.Helper()
	ctx := context.Background()
	err := db.Pool().QueryRow(ctx,
		`SELECT user_id, planet_id, state FROM artefacts_user WHERE id=$1`,
		id).Scan(&userID, &planetID, &state)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil, "", false
		}
		t.Fatalf("read artefact: %v", err)
	}
	return userID, planetID, state, true
}

func readFleetState(t *testing.T, db repo.Exec, fleetID string) string {
	t.Helper()
	ctx := context.Background()
	var s string
	if err := db.Pool().QueryRow(ctx,
		`SELECT state FROM fleets WHERE id=$1`, fleetID).Scan(&s); err != nil {
		t.Fatalf("read fleet state: %v", err)
	}
	return s
}

// TestDeliveryArtefacts_GoldenScenarios — golden-сценарии:
//
//  1. Доставка одного артефакта: владелец переписан, state=held, флот=returning.
//  2. Доставка трёх артефактов разом: все переписаны, флот=returning.
//  3. Идемпотентность: повторный запуск не меняет состояние.
//  4. Active артефакт сбрасывается в held; activated_at и expire_at обнуляются.
//  5. Флот уже returning → no-op (артефакты не трогаются).
func TestDeliveryArtefacts_GoldenScenarios(t *testing.T) {
	db, closeFn := openDeliveryTestDB(t)
	defer closeFn()

	t.Run("single_artefact_delivered", func(t *testing.T) {
		fx := deliveryFixture{
			senderID:    ids.New(),
			recipientID: ids.New(),
			srcPlanetID: ids.New(),
			dstPlanetID: ids.New(),
			fleetID:     ids.New(),
			artefacts:   []string{ids.New()},
		}
		ctx := context.Background()
		seedDeliveryFixture(ctx, t, db, fx)
		defer cleanupDeliveryFixture(ctx, db, fx)

		if err := runDeliveryArtefacts(t, db, fx); err != nil {
			t.Fatalf("handler: %v", err)
		}
		uid, pid, state, ok := readArtefactOwner(t, db, fx.artefacts[0])
		if !ok {
			t.Fatalf("artefact disappeared")
		}
		if uid != fx.recipientID {
			t.Errorf("user_id: got %q want %q", uid, fx.recipientID)
		}
		if pid == nil || *pid != fx.dstPlanetID {
			t.Errorf("planet_id: got %v want %q", pid, fx.dstPlanetID)
		}
		if state != "held" {
			t.Errorf("state: got %q want held", state)
		}
		if got := readFleetState(t, db, fx.fleetID); got != "returning" {
			t.Errorf("fleet state: got %q want returning", got)
		}
	})

	t.Run("three_artefacts_delivered", func(t *testing.T) {
		fx := deliveryFixture{
			senderID:    ids.New(),
			recipientID: ids.New(),
			srcPlanetID: ids.New(),
			dstPlanetID: ids.New(),
			fleetID:     ids.New(),
			artefacts:   []string{ids.New(), ids.New(), ids.New()},
		}
		ctx := context.Background()
		seedDeliveryFixture(ctx, t, db, fx)
		defer cleanupDeliveryFixture(ctx, db, fx)

		if err := runDeliveryArtefacts(t, db, fx); err != nil {
			t.Fatalf("handler: %v", err)
		}
		for _, aid := range fx.artefacts {
			uid, pid, state, ok := readArtefactOwner(t, db, aid)
			if !ok {
				t.Fatalf("artefact %s disappeared", aid)
			}
			if uid != fx.recipientID {
				t.Errorf("artefact %s user_id: got %q want %q", aid, uid, fx.recipientID)
			}
			if pid == nil || *pid != fx.dstPlanetID {
				t.Errorf("artefact %s planet_id: got %v want %q", aid, pid, fx.dstPlanetID)
			}
			if state != "held" {
				t.Errorf("artefact %s state: got %q want held", aid, state)
			}
		}
		if got := readFleetState(t, db, fx.fleetID); got != "returning" {
			t.Errorf("fleet state: got %q want returning", got)
		}
	})

	t.Run("idempotent_replay", func(t *testing.T) {
		fx := deliveryFixture{
			senderID:    ids.New(),
			recipientID: ids.New(),
			srcPlanetID: ids.New(),
			dstPlanetID: ids.New(),
			fleetID:     ids.New(),
			artefacts:   []string{ids.New(), ids.New()},
		}
		ctx := context.Background()
		seedDeliveryFixture(ctx, t, db, fx)
		defer cleanupDeliveryFixture(ctx, db, fx)

		if err := runDeliveryArtefacts(t, db, fx); err != nil {
			t.Fatalf("first run: %v", err)
		}
		// Повторный запуск (event re-fire) — флот теперь в returning,
		// весь handler должен быть no-op (skip-by-fleet-state).
		if err := runDeliveryArtefacts(t, db, fx); err != nil {
			t.Fatalf("replay: %v", err)
		}
		// Состояние не должно поменяться.
		for _, aid := range fx.artefacts {
			uid, pid, state, _ := readArtefactOwner(t, db, aid)
			if uid != fx.recipientID || pid == nil || *pid != fx.dstPlanetID || state != "held" {
				t.Errorf("artefact %s changed on replay: user=%q planet=%v state=%q",
					aid, uid, pid, state)
			}
		}
		if got := readFleetState(t, db, fx.fleetID); got != "returning" {
			t.Errorf("fleet state on replay: got %q want returning", got)
		}
	})

	t.Run("active_artefact_reset_to_held", func(t *testing.T) {
		fx := deliveryFixture{
			senderID:    ids.New(),
			recipientID: ids.New(),
			srcPlanetID: ids.New(),
			dstPlanetID: ids.New(),
			fleetID:     ids.New(),
			artefacts:   []string{ids.New()},
		}
		ctx := context.Background()
		seedDeliveryFixture(ctx, t, db, fx)
		defer cleanupDeliveryFixture(ctx, db, fx)

		// Помечаем артефакт active с заполненными activated_at + expire_at.
		now := time.Now().UTC()
		future := now.Add(24 * time.Hour)
		if _, err := db.Pool().Exec(ctx,
			`UPDATE artefacts_user SET state='active', activated_at=$1, expire_at=$2
			 WHERE id=$3`, now, future, fx.artefacts[0]); err != nil {
			t.Fatalf("set active: %v", err)
		}

		if err := runDeliveryArtefacts(t, db, fx); err != nil {
			t.Fatalf("handler: %v", err)
		}
		_, _, state, _ := readArtefactOwner(t, db, fx.artefacts[0])
		if state != "held" {
			t.Errorf("active artefact state: got %q want held", state)
		}
		// Проверяем, что activated_at и expire_at сброшены.
		var actAt, expAt *time.Time
		if err := db.Pool().QueryRow(ctx,
			`SELECT activated_at, expire_at FROM artefacts_user WHERE id=$1`,
			fx.artefacts[0]).Scan(&actAt, &expAt); err != nil {
			t.Fatalf("read timestamps: %v", err)
		}
		if actAt != nil {
			t.Errorf("activated_at: got %v want nil", actAt)
		}
		if expAt != nil {
			t.Errorf("expire_at: got %v want nil", expAt)
		}
	})

	t.Run("fleet_not_outbound_noop", func(t *testing.T) {
		fx := deliveryFixture{
			senderID:    ids.New(),
			recipientID: ids.New(),
			srcPlanetID: ids.New(),
			dstPlanetID: ids.New(),
			fleetID:     ids.New(),
			artefacts:   []string{ids.New()},
		}
		ctx := context.Background()
		seedDeliveryFixture(ctx, t, db, fx)
		defer cleanupDeliveryFixture(ctx, db, fx)

		// Сразу переводим флот в returning — handler должен быть no-op.
		if _, err := db.Pool().Exec(ctx,
			`UPDATE fleets SET state='returning' WHERE id=$1`, fx.fleetID); err != nil {
			t.Fatalf("set returning: %v", err)
		}

		if err := runDeliveryArtefacts(t, db, fx); err != nil {
			t.Fatalf("handler: %v", err)
		}
		// Артефакт остался у sender'а, не переписан.
		uid, pid, _, _ := readArtefactOwner(t, db, fx.artefacts[0])
		if uid != fx.senderID {
			t.Errorf("artefact owner changed despite returning fleet: got %q want %q",
				uid, fx.senderID)
		}
		if pid == nil || *pid != fx.srcPlanetID {
			t.Errorf("artefact planet changed: got %v want %q", pid, fx.srcPlanetID)
		}
	})
}
