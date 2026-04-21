package economy

import (
	"math"
	"testing"
)

func TestCostForLevel_Level1IsBase(t *testing.T) {
	t.Parallel()
	got := CostForLevel(Cost{Metal: 60, Silicon: 15}, 1.5, 1)
	if got.Metal != 60 || got.Silicon != 15 {
		t.Fatalf("expected base cost at level 1, got %+v", got)
	}
}

func TestCostForLevel_Monotone(t *testing.T) {
	t.Parallel()
	prev := int64(0)
	for lvl := 1; lvl <= 10; lvl++ {
		c := CostForLevel(Cost{Metal: 60, Silicon: 15}, 1.5, lvl)
		if c.Metal <= prev && lvl > 1 {
			t.Fatalf("cost not monotonic at level %d: %d <= %d", lvl, c.Metal, prev)
		}
		prev = c.Metal
	}
}

func TestProductionPerHour_ZeroLevel(t *testing.T) {
	t.Parallel()
	if ProductionPerHour(30, 0, 1) != 0 {
		t.Fatalf("production at level 0 must be 0")
	}
}

func TestEnergyRatio(t *testing.T) {
	t.Parallel()
	if r := EnergyRatio(100, 50); r != 1 {
		t.Fatalf("surplus energy must cap at 1, got %v", r)
	}
	if r := EnergyRatio(50, 100); math.Abs(r-0.5) > 1e-9 {
		t.Fatalf("expected 0.5 ratio, got %v", r)
	}
	if r := EnergyRatio(0, 0); r != 1 {
		t.Fatalf("no demand => ratio 1, got %v", r)
	}
}
