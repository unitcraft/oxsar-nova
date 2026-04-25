package alien

import (
	"testing"
	"time"

	"github.com/oxsar/nova/backend/internal/battle"
	"github.com/oxsar/nova/backend/internal/config"
	"github.com/oxsar/nova/backend/pkg/rng"
)

// alienTestCatalog — минимальный каталог с alien-кораблями (id 200–204),
// нужен для scaledAlienFleet. Значения приближены к configs/ships.yml.
func alienTestCatalog() *config.Catalog {
	return &config.Catalog{
		Ships: config.ShipCatalog{Ships: map[string]config.ShipSpec{
			"alien_corvette":      {ID: 200, Attack: 200, Shield: 75, Shell: 5000},
			"alien_screen":        {ID: 201, Attack: 22, Shield: 5000, Shell: 30000},
			"alien_paladin":       {ID: 202, Attack: 75, Shield: 50, Shell: 2000},
			"alien_frigate":       {ID: 203, Attack: 1250, Shield: 150, Shell: 15000},
			"alien_torpedocarrier": {ID: 204, Attack: 350, Shield: 100, Shell: 4000},
		}},
	}
}

// fleetAttackSum — суммарный attack флота (quantity × unit.attack).
func fleetAttackSum(units []battle.Unit) float64 {
	var s float64
	for _, u := range units {
		s += float64(u.Quantity) * u.Attack
	}
	return s
}

func TestAlienDistance_DifferentGalaxies(t *testing.T) {
	t.Parallel()
	d := alienDistance(99, 500, 8, 1, 100, 5)
	want := 20000 * 98
	if d != want {
		t.Errorf("alienDistance cross-galaxy = %d, want %d", d, want)
	}
}

func TestAlienDistance_SameGalaxy(t *testing.T) {
	t.Parallel()
	// Same galaxy, different systems.
	d := alienDistance(1, 100, 5, 1, 110, 5)
	want := 2700 + 95*10
	if d != want {
		t.Errorf("alienDistance same-galaxy = %d, want %d", d, want)
	}
}

func TestAlienDistance_SameSystem(t *testing.T) {
	t.Parallel()
	d := alienDistance(1, 100, 3, 1, 100, 8)
	want := 1000 + 5*5
	if d != want {
		t.Errorf("alienDistance same-system = %d, want %d", d, want)
	}
}

func TestAlienDistance_SamePlanet(t *testing.T) {
	t.Parallel()
	d := alienDistance(2, 50, 7, 2, 50, 7)
	if d != 5 {
		t.Errorf("alienDistance same-planet = %d, want 5", d)
	}
}

func TestAlienDistance_Symmetric(t *testing.T) {
	t.Parallel()
	d1 := alienDistance(99, 500, 8, 3, 200, 10)
	d2 := alienDistance(3, 200, 10, 99, 500, 8)
	if d1 != d2 {
		t.Errorf("alienDistance not symmetric: %d != %d", d1, d2)
	}
}

func TestAlienFlightDuration_MinimumOneMinute(t *testing.T) {
	t.Parallel()
	dur := alienFlightDuration(0)
	if dur < time.Minute {
		t.Errorf("alienFlightDuration(0) = %v, want >= 1min", dur)
	}
}

func TestAlienFlightDuration_IncreasesWithDistance(t *testing.T) {
	t.Parallel()
	d1 := alienFlightDuration(1000)
	d2 := alienFlightDuration(100000)
	if d1 >= d2 {
		t.Errorf("longer distance should take longer: %v >= %v", d1, d2)
	}
}

func TestScoreTier(t *testing.T) {
	t.Parallel()
	cases := []struct {
		score int64
		tier  int
	}{
		{0, 1},
		{999, 1},
		{1000, 2},
		{49999, 2},
		{50000, 3},
		{1000000, 3},
	}
	for _, c := range cases {
		if got := scoreTier(c.score); got != c.tier {
			t.Errorf("scoreTier(%d) = %d, want %d", c.score, got, c.tier)
		}
	}
}

func TestAlienFleet_AllTiers(t *testing.T) {
	t.Parallel()
	for _, tier := range []int{1, 2, 3} {
		units := alienFleet(tier)
		if len(units) == 0 {
			t.Errorf("alienFleet(%d) returned empty slice", tier)
		}
		for _, u := range units {
			if u.Quantity <= 0 {
				t.Errorf("tier %d: unit %d has non-positive quantity %d", tier, u.UnitID, u.Quantity)
			}
		}
	}
}

func TestAlienFleet_HigherTierStronger(t *testing.T) {
	t.Parallel()
	totalShell := func(tier int) float64 {
		var s float64
		for _, u := range alienFleet(tier) {
			s += float64(u.Quantity) * u.Shell
		}
		return s
	}
	s1, s2, s3 := totalShell(1), totalShell(2), totalShell(3)
	if !(s1 < s2 && s2 < s3) {
		t.Errorf("expected tier1 < tier2 < tier3 total shell: %.0f, %.0f, %.0f", s1, s2, s3)
	}
}

// TestScaledAlienFleet_BonusScaleIncreasesPower — четверговый bonusScale ≥ 1.5
// должен давать флот с большей суммарной атакой, чем обычный (bonusScale=1.0).
func TestScaledAlienFleet_BonusScaleIncreasesPower(t *testing.T) {
	t.Parallel()
	cat := alienTestCatalog()
	defPower := 100_000.0

	// Одинаковый seed, разный bonusScale → разная мощь.
	f1 := scaledAlienFleet(defPower, rng.New(42), cat, 1.0)
	f2 := scaledAlienFleet(defPower, rng.New(42), cat, 1.75) // средний четверг-boost

	a1 := fleetAttackSum(f1)
	a2 := fleetAttackSum(f2)
	if a2 <= a1 {
		t.Errorf("bonusScale=1.75 should yield stronger fleet: a1=%.0f, a2=%.0f", a1, a2)
	}
	// Должен быть примерно в 1.5–2× раза сильнее (с учётом шага округления при наборе юнитов).
	ratio := a2 / a1
	if ratio < 1.3 || ratio > 2.3 {
		t.Errorf("bonusScale ratio out of range: %.2f (want 1.3..2.3)", ratio)
	}
}

// TestScaledAlienFleet_ZeroBonusScaleFallsBackToOne — 0 и отрицательные значения
// bonusScale должны трактоваться как 1.0 (защита от неправильных caller'ов).
func TestScaledAlienFleet_ZeroBonusScaleFallsBackToOne(t *testing.T) {
	t.Parallel()
	cat := alienTestCatalog()
	defPower := 50_000.0
	f0 := scaledAlienFleet(defPower, rng.New(7), cat, 0)
	f1 := scaledAlienFleet(defPower, rng.New(7), cat, 1.0)
	if fleetAttackSum(f0) != fleetAttackSum(f1) {
		t.Errorf("bonusScale=0 should behave like 1.0: %.0f vs %.0f",
			fleetAttackSum(f0), fleetAttackSum(f1))
	}
}

// TestAlienConsts_Sane — guard-тест на легаси-константы: цифры не должны
// неслышно уехать. Менять — только с ADR.
func TestAlienConsts_Sane(t *testing.T) {
	t.Parallel()
	if AlienAttackInterval != 6*24*time.Hour {
		t.Errorf("AlienAttackInterval = %v, want 6d", AlienAttackInterval)
	}
	if AlienHaltingMinTime != 12*time.Hour {
		t.Errorf("AlienHaltingMinTime = %v, want 12h", AlienHaltingMinTime)
	}
	if AlienHaltingMaxTime != 24*time.Hour {
		t.Errorf("AlienHaltingMaxTime = %v, want 24h", AlienHaltingMaxTime)
	}
	if AlienHaltingMaxRealTime != 15*24*time.Hour {
		t.Errorf("AlienHaltingMaxRealTime = %v, want 15d", AlienHaltingMaxRealTime)
	}
	if ThursdayCandidateMultiplier != 5 {
		t.Errorf("ThursdayCandidateMultiplier = %d, want 5", ThursdayCandidateMultiplier)
	}
	// 2h per 50 credit: 1 кредит = 144 секунд продления.
	if AlienHoldingPaySecondsPerCredit != 144.0 {
		t.Errorf("AlienHoldingPaySecondsPerCredit = %.2f, want 144.0",
			AlienHoldingPaySecondsPerCredit)
	}
}
