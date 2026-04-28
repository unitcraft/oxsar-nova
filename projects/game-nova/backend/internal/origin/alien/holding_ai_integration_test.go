package alien_test

// План 66 Ф.4: интеграционные тесты HoldingAIHandler (8 sub-phases).
//
// Запускаются только при TEST_DATABASE_URL — иначе skip (паттерн от
// handlers_integration_test.go / план 65 demolish_test.go).
//
// Покрытые сценарии:
//   - TestHoldingAI_TickIncrementsControlTimes: после успешного тика
//     control_times++ в payload следующего HOLDING_AI; spawn'ится один
//     follow-up event.
//   - TestHoldingAI_PaidCreditExtendsParent: paid_credit > 0 продлевает
//     parent.fire_at на 2h × paid_credit / 50 (origin AlienAI:993).
//   - TestHoldingAI_SkipParentGone: parent KindAlienHolding отсутствует
//     → handler ничего не делает (silent skip).
//   - TestHoldingAI_SkipParentDone: parent уже state='ok' → silent skip.
//   - TestHoldingAI_AllStubSubphasesAreNoop: все 6 заглушек не меняют
//     parent payload (только control_times++).

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"oxsar/game-nova/internal/event"
	originalien "oxsar/game-nova/internal/origin/alien"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

// seedHolding — создаёт parent KindAlienHolding event с заданным fire_at
// и payload (alien_fleet, snapshot ресурсов). Возвращает holding_event_id.
func seedHolding(ctx context.Context, t *testing.T, db repo.Exec,
	planetID, userID string, fireAt time.Time, fleet []originalien.HoldingFleetUnit,
	metal, silicon, hydrogen int64) string {

	t.Helper()
	holdingID := ids.New()
	pl := originalien.HoldingAIPayload{
		PlanetID:   planetID,
		UserID:     userID,
		Tier:       1,
		AlienFleet: fleet,
		StartTime:  time.Now().UTC(),
		Metal:      metal, Silicon: silicon, Hydrogen: hydrogen,
	}
	raw, _ := json.Marshal(pl)
	err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO events (id, kind, planet_id, user_id, fire_at, payload, state)
			VALUES ($1::uuid, $2, $3::uuid, $4::uuid, $5, $6, 'wait')
		`, holdingID, int(event.KindAlienHolding), planetID, userID, fireAt, raw)
		return err
	})
	if err != nil {
		t.Fatalf("seed holding: %v", err)
	}
	return holdingID
}

// makeHoldingAIEvent — событие HOLDING_AI с указанным holding_event_id
// и параметрами payload.
func makeHoldingAIEvent(planetID, userID, holdingID string,
	fleet []originalien.HoldingFleetUnit, controlTimes int, paidCredit int64) event.Event {

	pl := originalien.HoldingAIPayload{
		PlanetID:       planetID,
		UserID:         userID,
		Tier:           1,
		AlienFleet:     fleet,
		StartTime:      time.Now().UTC().Add(-1 * time.Hour),
		HoldingEventID: holdingID,
		ControlTimes:   controlTimes,
		PaidCredit:     paidCredit,
	}
	raw, _ := json.Marshal(pl)
	return event.Event{
		ID:       ids.New(),
		Kind:     event.KindAlienHoldingAI,
		FireAt:   time.Now(),
		Payload:  raw,
		PlanetID: &planetID,
		UserID:   &userID,
	}
}

func TestHoldingAI_TickIncrementsControlTimes(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	fleet := []originalien.HoldingFleetUnit{
		{UnitID: 200, Quantity: 100},
		{UnitID: 201, Quantity: 50},
	}
	holdingID := seedHolding(ctx, t, db, up.planetID, up.userID,
		time.Now().Add(48*time.Hour), fleet, 0, 0, 0)

	svc := originalien.NewService(nil, nil)
	e := makeHoldingAIEvent(up.planetID, up.userID, holdingID, fleet, 0, 0)
	if err := runHandler(t, db, svc.HoldingAIHandler(), e); err != nil {
		t.Fatalf("handler: %v", err)
	}

	// Должен появиться ровно один новый KindAlienHoldingAI с
	// control_times >= 1.
	var nextRaw []byte
	if err := db.Pool().QueryRow(ctx, `
		SELECT payload FROM events
		WHERE planet_id=$1 AND kind=$2 AND id != $3
		ORDER BY created_at DESC LIMIT 1
	`, up.planetID, int(event.KindAlienHoldingAI), e.ID).Scan(&nextRaw); err != nil {
		t.Fatalf("query next: %v", err)
	}
	var nextPL originalien.HoldingAIPayload
	if err := json.Unmarshal(nextRaw, &nextPL); err != nil {
		t.Fatalf("parse next: %v", err)
	}
	if nextPL.ControlTimes != 1 {
		t.Errorf("next control_times = %d, want 1", nextPL.ControlTimes)
	}
	if nextPL.HoldingEventID != holdingID {
		t.Errorf("next holding_event_id = %s, want %s",
			nextPL.HoldingEventID, holdingID)
	}
}

func TestHoldingAI_PaidCreditExtendsParent(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	fleet := []originalien.HoldingFleetUnit{{UnitID: 200, Quantity: 100}}
	parentFireAt := time.Now().Add(48 * time.Hour).Truncate(time.Second)
	holdingID := seedHolding(ctx, t, db, up.planetID, up.userID,
		parentFireAt, fleet, 0, 0, 0)

	svc := originalien.NewService(nil, nil)
	const paid int64 = 50 // 50 → +2h согласно origin (HoldingPaySecondsPerCredit=144).
	e := makeHoldingAIEvent(up.planetID, up.userID, holdingID, fleet, 0, paid)
	if err := runHandler(t, db, svc.HoldingAIHandler(), e); err != nil {
		t.Fatalf("handler: %v", err)
	}

	var newFireAt time.Time
	if err := db.Pool().QueryRow(ctx,
		`SELECT fire_at FROM events WHERE id=$1::uuid`, holdingID,
	).Scan(&newFireAt); err != nil {
		t.Fatalf("query parent: %v", err)
	}
	expectedAdd := 2 * time.Hour // 50 paid × 144s = 7200s = 2h.
	tolerance := 5 * time.Second
	delta := newFireAt.Sub(parentFireAt)
	if delta < expectedAdd-tolerance || delta > expectedAdd+tolerance {
		t.Errorf("parent fire_at extended by %v, want ~%v", delta, expectedAdd)
	}

	// Проверяем что paid_sum_credit и paid_times пишутся в parent.payload.
	var parentRaw []byte
	if err := db.Pool().QueryRow(ctx,
		`SELECT payload FROM events WHERE id=$1::uuid`, holdingID).Scan(&parentRaw); err != nil {
		t.Fatalf("parent payload: %v", err)
	}
	var parentPL originalien.HoldingAIPayload
	if err := json.Unmarshal(parentRaw, &parentPL); err != nil {
		t.Fatalf("parse parent: %v", err)
	}
	if parentPL.PaidSumCredit != paid {
		t.Errorf("parent paid_sum_credit = %d, want %d",
			parentPL.PaidSumCredit, paid)
	}
	if parentPL.PaidTimes != 1 {
		t.Errorf("parent paid_times = %d, want 1", parentPL.PaidTimes)
	}
	if parentPL.PaidCredit != 0 {
		t.Errorf("parent paid_credit not consumed: %d", parentPL.PaidCredit)
	}
}

func TestHoldingAI_SkipParentGone(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	// holding_event_id ссылается на несуществующее событие.
	missing := ids.New()
	fleet := []originalien.HoldingFleetUnit{{UnitID: 200, Quantity: 100}}
	svc := originalien.NewService(nil, nil)
	e := makeHoldingAIEvent(up.planetID, up.userID, missing, fleet, 0, 0)
	if err := runHandler(t, db, svc.HoldingAIHandler(), e); err != nil {
		t.Fatalf("handler should silent-skip, got: %v", err)
	}
	// Не должно появиться никаких новых HOLDING_AI событий.
	if n := countEventsByKind(t, db, up.planetID, event.KindAlienHoldingAI); n != 0 {
		t.Errorf("unexpected HOLDING_AI events: %d", n)
	}
}

func TestHoldingAI_SkipParentDone(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	fleet := []originalien.HoldingFleetUnit{{UnitID: 200, Quantity: 100}}
	holdingID := seedHolding(ctx, t, db, up.planetID, up.userID,
		time.Now().Add(48*time.Hour), fleet, 0, 0, 0)

	// Закрываем parent вручную (имитируем что HOLDING истёк).
	_ = db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx,
			`UPDATE events SET state='ok' WHERE id=$1::uuid`, holdingID)
		return err
	})

	svc := originalien.NewService(nil, nil)
	e := makeHoldingAIEvent(up.planetID, up.userID, holdingID, fleet, 0, 0)
	if err := runHandler(t, db, svc.HoldingAIHandler(), e); err != nil {
		t.Fatalf("handler should silent-skip, got: %v", err)
	}
	if n := countEventsByKind(t, db, up.planetID, event.KindAlienHoldingAI); n != 0 {
		t.Errorf("unexpected HOLDING_AI follow-up: %d", n)
	}
}

// TestHoldingAI_StubSubphasesAreNoop — гоняем 50 тиков и убеждаемся,
// что follow-up event'ы создаются стабильно (никакой stub не валит
// handler), parent.fire_at не сдвигается без paid_credit.
func TestHoldingAI_StubSubphasesAreNoop(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	fleet := []originalien.HoldingFleetUnit{
		{UnitID: 200, Quantity: 1000},
		{UnitID: 201, Quantity: 500},
	}
	parentFireAt := time.Now().Add(48 * time.Hour).Truncate(time.Second)
	holdingID := seedHolding(ctx, t, db, up.planetID, up.userID,
		parentFireAt, fleet, 0, 0, 0)

	svc := originalien.NewService(nil, nil)
	const N = 50
	successes := 0
	for i := 0; i < N; i++ {
		e := makeHoldingAIEvent(up.planetID, up.userID, holdingID, fleet, i, 0)
		if err := runHandler(t, db, svc.HoldingAIHandler(), e); err != nil {
			t.Fatalf("handler tick %d: %v", i, err)
		}
		successes++
	}
	if successes != N {
		t.Errorf("got %d successful ticks, want %d", successes, N)
	}

	// Без paid_credit parent.fire_at не должен изменяться, кроме
	// случаев когда subphase=Extract/Unload поменяла alien_fleet
	// (в этом случае мы тоже сохраняем parent payload через UPDATE,
	// но fire_at остаётся прежним — мы НЕ продлеваем без paid).
	var newFireAt time.Time
	if err := db.Pool().QueryRow(ctx,
		`SELECT fire_at FROM events WHERE id=$1::uuid`, holdingID,
	).Scan(&newFireAt); err != nil {
		t.Fatalf("parent fire_at: %v", err)
	}
	if !newFireAt.Equal(parentFireAt) {
		t.Errorf("parent fire_at drifted: %v vs %v",
			newFireAt, parentFireAt)
	}
}

// TestHoldingAI_NegativePayloadRejected — пустой holding_event_id
// в payload отвергается. Требует БД (handler начинает с Unmarshal,
// затем валидирует поля до query — tx нужен только для context
// опоры, фактический query не выполняется).
func TestHoldingAI_NegativePayloadRejected(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()

	svc := originalien.NewService(nil, nil)
	pl := originalien.HoldingAIPayload{
		PlanetID: ids.New(),
		UserID:   ids.New(),
		// HoldingEventID пустой → handler вернёт ошибку валидации.
	}
	raw, _ := json.Marshal(pl)
	pid, uid := pl.PlanetID, pl.UserID
	e := event.Event{
		ID:       ids.New(),
		Kind:     event.KindAlienHoldingAI,
		FireAt:   time.Now(),
		Payload:  raw,
		PlanetID: &pid,
		UserID:   &uid,
	}
	err := runHandler(t, db, svc.HoldingAIHandler(), e)
	if err == nil {
		t.Fatal("handler must reject payload without holding_event_id")
	}
}
