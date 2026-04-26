package economy

import (
	"testing"
	"time"
)

func TestBuildDuration_Decreases_WithRobotics(t *testing.T) {
	t.Parallel()
	cost := Cost{Metal: 100_000, Silicon: 50_000}
	t0 := BuildDuration(60, cost, 0, 0, 1)
	t1 := BuildDuration(60, cost, 5, 0, 1)
	if t1 >= t0 {
		t.Fatalf("roboLevel=5 should be faster than 0: %v >= %v", t1, t0)
	}
}

func TestBuildDuration_Decreases_WithNano(t *testing.T) {
	t.Parallel()
	cost := Cost{Metal: 100_000, Silicon: 50_000}
	t0 := BuildDuration(60, cost, 0, 0, 1)
	t1 := BuildDuration(60, cost, 0, 1, 1)
	if t1 >= t0 {
		t.Fatalf("nanoLevel=1 should halve duration: %v >= %v", t1, t0)
	}
}

func TestBuildDuration_GameSpeed(t *testing.T) {
	t.Parallel()
	cost := Cost{Metal: 100_000, Silicon: 50_000}
	t1 := BuildDuration(60, cost, 0, 0, 1)
	t2 := BuildDuration(60, cost, 0, 0, 2)
	if t2 >= t1 {
		t.Fatalf("gameSpeed=2 should be faster: %v >= %v", t2, t1)
	}
}

func TestBuildDuration_MinimumOneSecond(t *testing.T) {
	t.Parallel()
	// Очень маленькая стоимость → не должна быть < 1s.
	cost := Cost{Metal: 1, Silicon: 1}
	dur := BuildDuration(60, cost, 30, 10, 100)
	if dur < time.Second {
		t.Fatalf("build duration must be >= 1s, got %v", dur)
	}
}

func TestBuildDuration_ZeroBaseSeconds(t *testing.T) {
	t.Parallel()
	cost := Cost{Metal: 10_000, Silicon: 5_000}
	// baseSeconds=0 должен подставить 60.
	dur := BuildDuration(0, cost, 0, 0, 1)
	if dur < time.Second {
		t.Fatalf("got %v, want >= 1s", dur)
	}
}

func TestEnergyDemand_ZeroLevel(t *testing.T) {
	t.Parallel()
	if d := EnergyDemand(10, 0); d != 0 {
		t.Fatalf("EnergyDemand(10,0) = %v, want 0", d)
	}
}

func TestEnergyDemand_MonotonicallyIncreases(t *testing.T) {
	t.Parallel()
	prev := 0.0
	for lvl := 1; lvl <= 15; lvl++ {
		d := EnergyDemand(10, lvl)
		if d <= prev {
			t.Fatalf("EnergyDemand not monotone at level %d: %v <= %v", lvl, d, prev)
		}
		prev = d
	}
}

func TestEnergyOutput_ZeroLevel(t *testing.T) {
	t.Parallel()
	if o := EnergyOutput(20, 0); o != 0 {
		t.Fatalf("EnergyOutput(20,0) = %v, want 0", o)
	}
}

func TestEnergyOutput_MonotonicallyIncreases(t *testing.T) {
	t.Parallel()
	prev := 0.0
	for lvl := 1; lvl <= 15; lvl++ {
		o := EnergyOutput(20, lvl)
		if o <= prev {
			t.Fatalf("EnergyOutput not monotone at level %d: %v <= %v", lvl, o, prev)
		}
		prev = o
	}
}

func TestEnergyOutput_SameFormulaAsDemand(t *testing.T) {
	t.Parallel()
	// Обе функции используют одну формулу — при одинаковых параметрах результат тот же.
	for lvl := 1; lvl <= 10; lvl++ {
		if EnergyOutput(15, lvl) != EnergyDemand(15, lvl) {
			t.Fatalf("output != demand at level %d with same params", lvl)
		}
	}
}

func TestProductionPerHour_Increases(t *testing.T) {
	t.Parallel()
	prev := 0.0
	for lvl := 1; lvl <= 20; lvl++ {
		p := ProductionPerHour(30, lvl, 1.0)
		if p <= prev {
			t.Fatalf("production not monotone at level %d: %v <= %v", lvl, p, prev)
		}
		prev = p
	}
}

func TestProductionPerHour_FactorScales(t *testing.T) {
	t.Parallel()
	p1 := ProductionPerHour(30, 5, 1.0)
	p2 := ProductionPerHour(30, 5, 2.0)
	if p2 != p1*2 {
		t.Fatalf("factor 2x should double production: %v != %v", p2, p1*2)
	}
}
