package alien

import (
	"encoding/json"
	"testing"
	"time"

	"oxsar/game-nova/internal/battle"
)

// TestSurvivorsToStacks — компактное представление флота пришельцев
// после боя: оставляем только юниты с QuantityEnd > 0.
func TestSurvivorsToStacks(t *testing.T) {
	t.Parallel()
	in := []battle.UnitResult{
		{UnitID: 200, QuantityStart: 10, QuantityEnd: 7},
		{UnitID: 201, QuantityStart: 5, QuantityEnd: 0}, // погиб целиком — отфильтровать
		{UnitID: 202, QuantityStart: 3, QuantityEnd: 3},
	}
	got := survivorsToStacks(in)
	if len(got) != 2 {
		t.Fatalf("want 2 stacks, got %d", len(got))
	}
	if got[0].UnitID != 200 || got[0].Quantity != 7 {
		t.Errorf("stack[0] = %+v, want {200, 7}", got[0])
	}
	if got[1].UnitID != 202 || got[1].Quantity != 3 {
		t.Errorf("stack[1] = %+v, want {202, 3}", got[1])
	}
}

// TestSurvivorsToStacks_EmptyInput — на пустой вход отдаём пустой (не nil) slice.
func TestSurvivorsToStacks_EmptyInput(t *testing.T) {
	t.Parallel()
	got := survivorsToStacks(nil)
	if got == nil {
		t.Error("want empty slice, got nil (may break json.Marshal consistency)")
	}
	if len(got) != 0 {
		t.Errorf("want len=0, got %d", len(got))
	}
}

// TestSurvivorsToStacks_AllDead — все погибли → пустой slice, не nil.
func TestSurvivorsToStacks_AllDead(t *testing.T) {
	t.Parallel()
	in := []battle.UnitResult{
		{UnitID: 200, QuantityEnd: 0},
		{UnitID: 201, QuantityEnd: 0},
	}
	got := survivorsToStacks(in)
	if len(got) != 0 {
		t.Errorf("want 0 survivors, got %d", len(got))
	}
}

// TestHoldingPayload_JSONRoundtrip — payload корректно сериализуется
// и десериализуется. Критично для events.payload jsonb.
func TestHoldingPayload_JSONRoundtrip(t *testing.T) {
	t.Parallel()
	orig := holdingPayload{
		PlanetID: "p-123",
		UserID:   "u-456",
		Tier:     2,
		AlienFleet: []fleetStack{
			{UnitID: 200, Quantity: 20},
			{UnitID: 202, Quantity: 5},
		},
		StartTime:      time.Date(2026, 4, 24, 12, 0, 0, 0, time.UTC),
		HoldingEventID: "h-789",
		PaidCredit:     500,
		PaidTimes:      3,
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got holdingPayload
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.PlanetID != orig.PlanetID || got.UserID != orig.UserID ||
		got.Tier != orig.Tier || got.HoldingEventID != orig.HoldingEventID ||
		got.PaidCredit != orig.PaidCredit || got.PaidTimes != orig.PaidTimes {
		t.Errorf("scalar fields mismatch: %+v vs %+v", got, orig)
	}
	if !got.StartTime.Equal(orig.StartTime) {
		t.Errorf("StartTime mismatch: %v vs %v", got.StartTime, orig.StartTime)
	}
	if len(got.AlienFleet) != 2 ||
		got.AlienFleet[0].UnitID != 200 || got.AlienFleet[0].Quantity != 20 ||
		got.AlienFleet[1].UnitID != 202 || got.AlienFleet[1].Quantity != 5 {
		t.Errorf("AlienFleet mismatch: %+v", got.AlienFleet)
	}
}

// TestHoldingPayload_OmitsEmptyPaid — когда PaidCredit/PaidTimes нулевые,
// они не сериализуются (omitempty), чтобы payload был компактнее.
func TestHoldingPayload_OmitsEmptyPaid(t *testing.T) {
	t.Parallel()
	hp := holdingPayload{
		PlanetID: "p-1", UserID: "u-1", Tier: 1,
		StartTime: time.Now().UTC(),
	}
	data, err := json.Marshal(hp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	if containsField(s, `"paid_credit"`) {
		t.Errorf("paid_credit should be omitted when zero: %s", s)
	}
	if containsField(s, `"paid_times"`) {
		t.Errorf("paid_times should be omitted when zero: %s", s)
	}
	if containsField(s, `"holding_event_id"`) {
		t.Errorf("holding_event_id should be omitted when empty: %s", s)
	}
}

func containsField(s, f string) bool {
	for i := 0; i+len(f) <= len(s); i++ {
		if s[i:i+len(f)] == f {
			return true
		}
	}
	return false
}
