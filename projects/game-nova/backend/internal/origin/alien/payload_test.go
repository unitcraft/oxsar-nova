package alien

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestMissionPayload_RoundTrip — pure: payload не теряет поля при
// JSON-сериализации. Защита от случайного rename JSON-тэга.
func TestMissionPayload_RoundTrip(t *testing.T) {
	src := MissionPayload{
		Mode:         int(ModeFlyUnknown),
		Tier:         2,
		UserID:       "u-1",
		PlanetID:     "p-1",
		Galaxy:       3, System: 100, Position: 7,
		Ships: Fleet{
			{UnitID: 200, Quantity: 50, ShellPercent: 100},
			{UnitID: 203, Quantity: 5, ShellPercent: 100},
		},
		Metal:        1_000_000,
		Silicon:      500_000,
		Hydrogen:     200_000,
		ControlTimes: 1,
		PowerScale:   1.5,
		AlienActor:   true,
		AddTech:      map[int]int{15: 7, 16: 8},
	}
	raw, err := json.Marshal(&src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	expectKeys := []string{
		`"mode":33`,
		`"user_id":"u-1"`,
		`"planet_id":"p-1"`,
		`"control_times":1`,
		`"power_scale":1.5`,
		`"alien_actor":true`,
		`"add_tech":{"15":7,"16":8}`,
		`"ships":[`,
	}
	for _, k := range expectKeys {
		if !strings.Contains(string(raw), k) {
			t.Errorf("payload missing %q in JSON: %s", k, string(raw))
		}
	}
	var dst MissionPayload
	if err := json.Unmarshal(raw, &dst); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dst.Mode != src.Mode || dst.UserID != src.UserID ||
		dst.PlanetID != src.PlanetID || dst.ControlTimes != src.ControlTimes ||
		dst.PowerScale != src.PowerScale || dst.AlienActor != src.AlienActor {
		t.Errorf("round-trip mismatch: got %+v want %+v", dst, src)
	}
	if len(dst.Ships) != len(src.Ships) {
		t.Errorf("ships len: got %d want %d", len(dst.Ships), len(src.Ships))
	}
	if dst.AddTech[15] != 7 || dst.AddTech[16] != 8 {
		t.Errorf("add_tech: %v", dst.AddTech)
	}
}

func TestChangeMissionPayload_RoundTrip(t *testing.T) {
	src := ChangeMissionPayload{
		ParentEventID: "11111111-2222-3333-4444-555555555555",
		UserID:        "u-1",
		PlanetID:      "p-1",
		ControlTimes:  3,
		AlienActor:    true,
	}
	raw, err := json.Marshal(&src)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, k := range []string{
		`"parent_event_id":"11111111-2222-3333-4444-555555555555"`,
		`"control_times":3`,
		`"alien_actor":true`,
	} {
		if !strings.Contains(string(raw), k) {
			t.Errorf("missing %q in JSON: %s", k, string(raw))
		}
	}
	var dst ChangeMissionPayload
	if err := json.Unmarshal(raw, &dst); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dst != src {
		t.Errorf("round-trip mismatch: got %+v want %+v", dst, src)
	}
}
