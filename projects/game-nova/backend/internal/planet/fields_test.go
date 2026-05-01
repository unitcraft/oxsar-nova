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

func TestMaxFields_SmallMoon_NoMoonBase(t *testing.T) {
	// План 72.1.47: для луны без moon_base формула = round(0×3.5)+1 = 1
	// (legacy `getMaxFields()` без аргумента). Голый fmax = 8 — через
	// MaxFieldsDiameterOnly.
	p := &Planet{Diameter: 2000, IsMoon: true}
	got := MaxFields(p, nil, DefaultFieldConsts)
	if got != 1 {
		t.Errorf("MaxFields(moon 2000, no moon_base) = %d, want 1", got)
	}
	gotDiam := MaxFieldsDiameterOnly(p, nil, DefaultFieldConsts)
	if gotDiam != 8 {
		t.Errorf("MaxFieldsDiameterOnly(moon 2000) = %d, want 8", gotDiam)
	}
}

func TestMaxFields_SmallMoonWithBaseAndLab(t *testing.T) {
	// Луна 2000, moon_base=1, moon_lab=2 →
	//   fields = round(1×5)+1 = 6 (moon_lab>0 → multiplier=5)
	//   fmax = round(4)×2 + 2×5 = 18
	//   max = min(6, 18) = 6
	p := &Planet{Diameter: 2000, IsMoon: true}
	got := MaxFields(p, map[int]int{54: 1, 350: 2}, DefaultFieldConsts)
	if got != 6 {
		t.Errorf("MaxFields(moon 2000 + moon_base=1 + moon_lab=2) = %d, want 6", got)
	}
}

func TestMaxFields_LargeMoon_NoMoonBase(t *testing.T) {
	// Большая луна > 2500, без moon_base: max = 1 (нет полей для застройки).
	// Голый fmax (через DiameterOnly) = round(9) = 9.
	p := &Planet{Diameter: 3000, IsMoon: true}
	got := MaxFields(p, nil, DefaultFieldConsts)
	if got != 1 {
		t.Errorf("MaxFields(moon 3000, no moon_base) = %d, want 1", got)
	}
	gotDiam := MaxFieldsDiameterOnly(p, nil, DefaultFieldConsts)
	if gotDiam != 9 {
		t.Errorf("MaxFieldsDiameterOnly(moon 3000) = %d, want 9", gotDiam)
	}
}

func TestMaxFields_MoonClampsToFmax(t *testing.T) {
	// Луна 2000, moon_base=10, moon_lab=0 →
	//   multiplier=3.5, fields = round(35)+1 = 36
	//   fmax = round(4)×2 + 0 = 8
	//   max = min(36, 8) = 8 (упирается в потолок диаметра)
	p := &Planet{Diameter: 2000, IsMoon: true}
	got := MaxFields(p, map[int]int{54: 10}, DefaultFieldConsts)
	if got != 8 {
		t.Errorf("MaxFields(moon 2000 + moon_base=10) = %d, want 8 (clamp to fmax)", got)
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
