package planet

import "testing"

func TestMaxFields_StartedPlanet(t *testing.T) {
	// Homeworld в тestseed: diameter=18800.
	// round((18800/1000)^2) = round(353.44) = 353
	// + PLANET_FIELD_ADDITION (10) = 363
	p := &Planet{Diameter: 18800, IsMoon: false}
	got := MaxFields(p, nil, DefaultFieldConsts)
	if got != 363 {
		t.Errorf("MaxFields(18800) = %d, want 363", got)
	}
}

func TestMaxFields_TerraFormerBonus(t *testing.T) {
	p := &Planet{Diameter: 18800, IsMoon: false}
	// terra_former level 3 → +15
	got := MaxFields(p, map[int]int{58: 3}, DefaultFieldConsts)
	if got != 378 {
		t.Errorf("MaxFields with terra_former=3 = %d, want 378", got)
	}
}

func TestMaxFields_SmallMoon(t *testing.T) {
	// Маленькая луна ≤ 2500 → base × 2.
	// diameter=2000: round(4) = 4; × 2 = 8; no moon_lab → 8.
	p := &Planet{Diameter: 2000, IsMoon: true}
	got := MaxFields(p, nil, DefaultFieldConsts)
	if got != 8 {
		t.Errorf("MaxFields(moon 2000) = %d, want 8", got)
	}
}

func TestMaxFields_SmallMoonWithLab(t *testing.T) {
	// Луна 2000, moon_lab level 2 → 8 + 2×5 = 18.
	p := &Planet{Diameter: 2000, IsMoon: true}
	got := MaxFields(p, map[int]int{350: 2}, DefaultFieldConsts)
	if got != 18 {
		t.Errorf("MaxFields(moon 2000 + moon_lab=2) = %d, want 18", got)
	}
}

func TestMaxFields_LargeMoon(t *testing.T) {
	// Большая луна > 2500: base = round((3000/1000)^2) = 9, без ×2.
	p := &Planet{Diameter: 3000, IsMoon: true}
	got := MaxFields(p, nil, DefaultFieldConsts)
	if got != 9 {
		t.Errorf("MaxFields(moon 3000) = %d, want 9", got)
	}
}

func TestMaxFields_NilBuildings(t *testing.T) {
	// nil-карта — читаем lookups как 0, не паникуем.
	p := &Planet{Diameter: 12000}
	got := MaxFields(p, nil, DefaultFieldConsts)
	// round(144) + 0 + 10 = 154
	if got != 154 {
		t.Errorf("MaxFields with nil map = %d, want 154", got)
	}
}
