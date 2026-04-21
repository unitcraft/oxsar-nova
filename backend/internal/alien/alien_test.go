package alien

import (
	"testing"
	"time"
)

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
