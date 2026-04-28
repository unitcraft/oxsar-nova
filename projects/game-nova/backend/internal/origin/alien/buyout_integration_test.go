package alien_test

// План 66 Ф.5: интеграционные тесты Buyout-handler.
//
// Запускаются только при TEST_DATABASE_URL — иначе skip (паттерн от
// handlers_integration_test.go / план 65 demolish_test.go).
//
// billing-сторона мокается через mockBilling — реализует BuyoutBilling.
// Сценарии полностью покрывают коды ответов из openapi.yaml:
//
//   - TestBuyout_HappyPath — 200, mission state='ok', тики удалены,
//     billing.Spend вызван 1 раз с корректным IdempotencyKey/Reason.
//   - TestBuyout_AlreadyClosed — mission уже state='ok' → 409
//     ErrMissionAlreadyClosed, billing НЕ вызван.
//   - TestBuyout_ForeignMission — другой owner → 404 (не раскрываем
//     существование), billing НЕ вызван.
//   - TestBuyout_MissionNotFound — несуществующий ID → 404.
//   - TestBuyout_WrongKind — event есть, но KindAlienAttack — 404
//     (для пользователя «нет такой миссии HOLDING»).
//   - TestBuyout_Insufficient — billing вернул ErrInsufficientOxsar
//     → 402 ErrInsufficientOxsars, mission остаётся 'wait'.
//   - TestBuyout_BillingUnavailable — billing вернул ErrBillingUnavailable
//     → 503 ErrBillingUnavailable, mission остаётся 'wait'.
//   - TestBuyout_IdempotencyConflict — billing вернул
//     ErrIdempotencyConflict → 409 ErrIdempotencyConflict,
//     mission остаётся 'wait'.
//   - TestBuyout_DropsAITicks — после успеха все KindAlienHoldingAI
//     этой миссии удалены (state='wait' тики).

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	billingclient "oxsar/game-nova/internal/billing/client"
	"oxsar/game-nova/internal/event"
	originalien "oxsar/game-nova/internal/origin/alien"
	"oxsar/game-nova/internal/repo"
	"oxsar/game-nova/pkg/ids"
)

// mockBilling — фиксирует вход Spend и возвращает заранее заданный err.
// Реализует originalien.BuyoutBilling.
type mockBilling struct {
	mu       sync.Mutex
	calls    []billingclient.SpendInput
	returnFn func(billingclient.SpendInput) error
}

func (m *mockBilling) Spend(_ context.Context, in billingclient.SpendInput) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, in)
	if m.returnFn != nil {
		return m.returnFn(in)
	}
	return nil
}

func (m *mockBilling) Calls() []billingclient.SpendInput {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]billingclient.SpendInput, len(m.calls))
	copy(out, m.calls)
	return out
}

// seedHoldingForBuyout — создаёт KindAlienHolding event с заданным
// owner и состоянием. Возвращает mission_id.
func seedHoldingForBuyout(ctx context.Context, t *testing.T, db repo.Exec,
	planetID, userID, state string) string {

	t.Helper()
	mid := ids.New()
	pl := map[string]any{
		"planet_id": planetID,
		"user_id":   userID,
		"tier":      1,
	}
	raw, _ := json.Marshal(pl)
	err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO events (id, kind, planet_id, user_id, fire_at, payload, state)
			VALUES ($1::uuid, $2, $3::uuid, $4::uuid, $5, $6, $7)
		`, mid, int(event.KindAlienHolding), planetID, userID,
			time.Now().Add(48*time.Hour), raw, state)
		return err
	})
	if err != nil {
		t.Fatalf("seed holding: %v", err)
	}
	return mid
}

// seedHoldingAITick — создаёт KindAlienHoldingAI с holding_event_id=mid,
// чтобы проверить что Buyout удаляет такие тики.
func seedHoldingAITick(ctx context.Context, t *testing.T, db repo.Exec,
	planetID, userID, mid string) string {

	t.Helper()
	tid := ids.New()
	pl := map[string]any{
		"planet_id":        planetID,
		"user_id":          userID,
		"holding_event_id": mid,
		"start_time":       time.Now().UTC(),
	}
	raw, _ := json.Marshal(pl)
	err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO events (id, kind, planet_id, user_id, fire_at, payload, state)
			VALUES ($1::uuid, $2, $3::uuid, $4::uuid, $5, $6, 'wait')
		`, tid, int(event.KindAlienHoldingAI), planetID, userID,
			time.Now().Add(1*time.Hour), raw)
		return err
	})
	if err != nil {
		t.Fatalf("seed ai tick: %v", err)
	}
	return tid
}

func readEventState(t *testing.T, db repo.Exec, eventID string) string {
	t.Helper()
	var st string
	if err := db.Pool().QueryRow(context.Background(),
		`SELECT state FROM events WHERE id = $1`, eventID).Scan(&st); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "<deleted>"
		}
		t.Fatalf("read event state: %v", err)
	}
	return st
}

func readEventExists(t *testing.T, db repo.Exec, eventID string) bool {
	t.Helper()
	var n int
	if err := db.Pool().QueryRow(context.Background(),
		`SELECT count(*) FROM events WHERE id = $1`, eventID).Scan(&n); err != nil {
		t.Fatalf("read event count: %v", err)
	}
	return n > 0
}

// makeCfg — Config с быстрым BuyoutBaseOxsars=100 (default), для тестов.
func makeCfg() originalien.Config {
	return originalien.DefaultConfig()
}

func TestBuyout_HappyPath(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	mid := seedHoldingForBuyout(ctx, t, db, up.planetID, up.userID, "wait")
	tickID := seedHoldingAITick(ctx, t, db, up.planetID, up.userID, mid)

	bm := &mockBilling{} // returnFn=nil → success.
	cfg := makeCfg()

	res, err := originalien.Buyout(ctx, db, bm, cfg, up.userID, mid,
		"jwt-token-for-test", "user1:alien_buyout:"+mid)
	if err != nil {
		t.Fatalf("Buyout error: %v", err)
	}
	if res == nil || res.MissionID != mid {
		t.Fatalf("unexpected result: %+v", res)
	}
	if res.CostOxsars != cfg.BuyoutBaseOxsars {
		t.Fatalf("cost = %d, want %d", res.CostOxsars, cfg.BuyoutBaseOxsars)
	}
	if res.FreedAt.IsZero() {
		t.Fatalf("FreedAt is zero")
	}

	// HOLDING закрыт.
	if got := readEventState(t, db, mid); got != "ok" {
		t.Fatalf("HOLDING state = %q, want 'ok'", got)
	}
	// AI-тик удалён.
	if readEventExists(t, db, tickID) {
		t.Fatalf("AI tick %s still exists, expected DELETE", tickID)
	}
	// billing.Spend вызван 1 раз с корректными параметрами.
	calls := bm.Calls()
	if len(calls) != 1 {
		t.Fatalf("billing.Spend calls = %d, want 1", len(calls))
	}
	c := calls[0]
	if c.Amount != cfg.BuyoutBaseOxsars {
		t.Errorf("Spend.Amount = %d, want %d", c.Amount, cfg.BuyoutBaseOxsars)
	}
	if c.Reason != "alien_buyout" {
		t.Errorf("Spend.Reason = %q, want 'alien_buyout'", c.Reason)
	}
	if c.RefID != mid {
		t.Errorf("Spend.RefID = %q, want %q", c.RefID, mid)
	}
	if c.IdempotencyKey == "" {
		t.Errorf("Spend.IdempotencyKey is empty")
	}
}

func TestBuyout_AlreadyClosed(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	mid := seedHoldingForBuyout(ctx, t, db, up.planetID, up.userID, "ok")

	bm := &mockBilling{}
	_, err := originalien.Buyout(ctx, db, bm, makeCfg(), up.userID, mid,
		"jwt", "user1:alien_buyout:"+mid)
	if !errors.Is(err, originalien.ErrMissionAlreadyClosed) {
		t.Fatalf("error = %v, want ErrMissionAlreadyClosed", err)
	}
	if len(bm.Calls()) != 0 {
		t.Fatalf("billing.Spend called %d times, want 0 (pre-check failed)", len(bm.Calls()))
	}
}

func TestBuyout_ForeignMission(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	owner := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, owner)
	intruder := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, intruder)

	mid := seedHoldingForBuyout(ctx, t, db, owner.planetID, owner.userID, "wait")

	bm := &mockBilling{}
	_, err := originalien.Buyout(ctx, db, bm, makeCfg(), intruder.userID, mid,
		"jwt", "intruder:alien_buyout:"+mid)
	if !errors.Is(err, originalien.ErrMissionNotFound) {
		t.Fatalf("error = %v, want ErrMissionNotFound (foreign mission, single 404)", err)
	}
	if len(bm.Calls()) != 0 {
		t.Fatalf("billing.Spend called for foreign mission")
	}
	// Mission жива (мы её не трогали).
	if got := readEventState(t, db, mid); got != "wait" {
		t.Fatalf("HOLDING state = %q, want 'wait' (untouched)", got)
	}
}

func TestBuyout_MissionNotFound(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	bm := &mockBilling{}
	mid := ids.New()
	_, err := originalien.Buyout(ctx, db, bm, makeCfg(), up.userID, mid,
		"jwt", "u:alien_buyout:"+mid)
	if !errors.Is(err, originalien.ErrMissionNotFound) {
		t.Fatalf("error = %v, want ErrMissionNotFound", err)
	}
	if len(bm.Calls()) != 0 {
		t.Fatalf("billing.Spend called for non-existent mission")
	}
}

func TestBuyout_WrongKind(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	// Инсертим event KindAlienAttack — не HOLDING.
	mid := ids.New()
	err := db.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			INSERT INTO events (id, kind, planet_id, user_id, fire_at, payload, state)
			VALUES ($1::uuid, $2, $3::uuid, $4::uuid, $5, '{}', 'wait')
		`, mid, int(event.KindAlienAttack), up.planetID, up.userID, time.Now())
		return err
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	bm := &mockBilling{}
	_, perr := originalien.Buyout(ctx, db, bm, makeCfg(), up.userID, mid,
		"jwt", "u:alien_buyout:"+mid)
	if !errors.Is(perr, originalien.ErrMissionNotFound) {
		t.Fatalf("error = %v, want ErrMissionNotFound (wrong kind)", perr)
	}
	if len(bm.Calls()) != 0 {
		t.Fatalf("billing.Spend called for wrong-kind event")
	}
}

func TestBuyout_Insufficient(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	mid := seedHoldingForBuyout(ctx, t, db, up.planetID, up.userID, "wait")
	tickID := seedHoldingAITick(ctx, t, db, up.planetID, up.userID, mid)

	bm := &mockBilling{
		returnFn: func(_ billingclient.SpendInput) error {
			return billingclient.ErrInsufficientOxsar
		},
	}
	_, err := originalien.Buyout(ctx, db, bm, makeCfg(), up.userID, mid,
		"jwt", "u:alien_buyout:"+mid)
	if !errors.Is(err, originalien.ErrInsufficientOxsars) {
		t.Fatalf("error = %v, want ErrInsufficientOxsars", err)
	}
	// Mission осталась 'wait', тики не удалены.
	if got := readEventState(t, db, mid); got != "wait" {
		t.Fatalf("HOLDING state = %q, want 'wait' (insufficient → untouched)", got)
	}
	if !readEventExists(t, db, tickID) {
		t.Fatalf("AI tick deleted on insufficient — should remain")
	}
}

func TestBuyout_BillingUnavailable(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	mid := seedHoldingForBuyout(ctx, t, db, up.planetID, up.userID, "wait")

	bm := &mockBilling{
		returnFn: func(_ billingclient.SpendInput) error {
			return billingclient.ErrBillingUnavailable
		},
	}
	_, err := originalien.Buyout(ctx, db, bm, makeCfg(), up.userID, mid,
		"jwt", "u:alien_buyout:"+mid)
	if !errors.Is(err, originalien.ErrBillingUnavailable) {
		t.Fatalf("error = %v, want ErrBillingUnavailable", err)
	}
	if got := readEventState(t, db, mid); got != "wait" {
		t.Fatalf("HOLDING state = %q, want 'wait' (billing unavailable → untouched)", got)
	}
}

func TestBuyout_IdempotencyConflict(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	mid := seedHoldingForBuyout(ctx, t, db, up.planetID, up.userID, "wait")

	bm := &mockBilling{
		returnFn: func(_ billingclient.SpendInput) error {
			return billingclient.ErrIdempotencyConflict
		},
	}
	_, err := originalien.Buyout(ctx, db, bm, makeCfg(), up.userID, mid,
		"jwt", "u:alien_buyout:"+mid)
	if !errors.Is(err, originalien.ErrIdempotencyConflict) {
		t.Fatalf("error = %v, want ErrIdempotencyConflict", err)
	}
	if got := readEventState(t, db, mid); got != "wait" {
		t.Fatalf("HOLDING state = %q, want 'wait' (idempotency conflict → untouched)", got)
	}
}

func TestBuyout_DropsAITicks(t *testing.T) {
	db, closeFn := openTestDB(t)
	defer closeFn()
	ctx := context.Background()

	up := seedUserPlanet(ctx, t, db, 0, nil)
	defer cleanupUserPlanet(ctx, db, up)

	mid := seedHoldingForBuyout(ctx, t, db, up.planetID, up.userID, "wait")
	// Несколько тиков подряд (как обычно бывает после генерации
	// HOLDING_AI каскада).
	tick1 := seedHoldingAITick(ctx, t, db, up.planetID, up.userID, mid)
	tick2 := seedHoldingAITick(ctx, t, db, up.planetID, up.userID, mid)
	tick3 := seedHoldingAITick(ctx, t, db, up.planetID, up.userID, mid)

	// Чужой тик другой миссии — не должен быть удалён.
	otherMid := seedHoldingForBuyout(ctx, t, db, up.planetID, up.userID, "wait")
	otherTick := seedHoldingAITick(ctx, t, db, up.planetID, up.userID, otherMid)

	bm := &mockBilling{}
	if _, err := originalien.Buyout(ctx, db, bm, makeCfg(), up.userID, mid,
		"jwt", "u:alien_buyout:"+mid); err != nil {
		t.Fatalf("Buyout: %v", err)
	}

	// Свои тики удалены.
	for _, id := range []string{tick1, tick2, tick3} {
		if readEventExists(t, db, id) {
			t.Errorf("tick %s still exists, expected DELETE", id)
		}
	}
	// Чужой тик жив.
	if !readEventExists(t, db, otherTick) {
		t.Errorf("foreign tick %s deleted, expected to remain", otherTick)
	}
}
