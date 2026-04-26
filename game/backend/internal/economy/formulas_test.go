package economy

import (
	"math"
	"testing"
)

// --- Prod/Cons static functions (plan 16) ---

func TestMetalmineProdMetal(t *testing.T) {
	t.Parallel()
	// floor(30 * 10 * pow(1.1 + 5*0.0006, 10))
	want := math.Floor(30 * 10 * math.Pow(1.1+5*0.0006, 10))
	got := MetalmineProdMetal(10, 5)
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("MetalmineProdMetal(10,5): got %v, want %v", got, want)
	}
	if MetalmineProdMetal(0, 5) != 0 {
		t.Fatal("level 0 must return 0")
	}
}

func TestSiliconLabProdSilicon(t *testing.T) {
	t.Parallel()
	// floor(20 * 8 * pow(1.1 + 3*0.0007, 8))
	want := math.Floor(20 * 8 * math.Pow(1.1+3*0.0007, 8))
	got := SiliconLabProdSilicon(8, 3)
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("SiliconLabProdSilicon(8,3): got %v, want %v", got, want)
	}
}

func TestHydrogenLabProdHydrogen(t *testing.T) {
	t.Parallel()
	// floor(10 * 5 * pow(1.1+2*0.0008, 5) * (-0.002*40 + 1.28))
	want := math.Floor(10 * 5 * math.Pow(1.1+2*0.0008, 5) * (-0.002*40 + 1.28))
	got := HydrogenLabProdHydrogen(5, 2, 40)
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("HydrogenLabProdHydrogen(5,2,40): got %v, want %v", got, want)
	}
}

func TestMoonHydrogenLabProdHydrogen(t *testing.T) {
	t.Parallel()
	// floor(100 * 3 * pow(1.1+0*0.0008, 3) * (-0.002*50 + 1.28))
	want := math.Floor(100 * 3 * math.Pow(1.1+0*0.0008, 3) * (-0.002*50 + 1.28))
	got := MoonHydrogenLabProdHydrogen(3, 0, 50)
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("MoonHydrogenLabProdHydrogen(3,0,50): got %v, want %v", got, want)
	}
}

func TestSolarPlantProdEnergy(t *testing.T) {
	t.Parallel()
	// floor(20 * 10 * pow(1.1+3*0.0005, 10))
	want := math.Floor(20 * 10 * math.Pow(1.1+3*0.0005, 10))
	got := SolarPlantProdEnergy(10, 3)
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("SolarPlantProdEnergy(10,3): got %v, want %v", got, want)
	}
}

func TestHydrogenPlantProdEnergy(t *testing.T) {
	t.Parallel()
	// floor(50 * 6 * pow(1.15+2*0.005, 6))
	want := math.Floor(50 * 6 * math.Pow(1.15+2*0.005, 6))
	got := HydrogenPlantProdEnergy(6, 2)
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("HydrogenPlantProdEnergy(6,2): got %v, want %v", got, want)
	}
}

func TestSolarSatelliteProdEnergy(t *testing.T) {
	t.Parallel()
	// tempC=40: perUnit = min(floor(40/4+20), 50) = min(30, 50) = 30
	// * pow(1.05, 0) = 30
	got := SolarSatelliteProdEnergy(40, 0)
	if math.Abs(got-30) > 1e-6 {
		t.Fatalf("SolarSatelliteProdEnergy(40,0): got %v, want 30", got)
	}
	// tempC=200: perUnit = min(floor(200/4+20), 50) = min(70, 50) = 50
	got = SolarSatelliteProdEnergy(200, 0)
	if math.Abs(got-50) > 1e-6 {
		t.Fatalf("SolarSatelliteProdEnergy(200,0): got %v, want 50", got)
	}
}

func TestGraviProdEnergy(t *testing.T) {
	t.Parallel()
	// 300000 * pow(3, 0) = 300000
	got := GraviProdEnergy(1, 300000)
	if math.Abs(got-300000) > 1e-6 {
		t.Fatalf("GraviProdEnergy(1): got %v, want 300000", got)
	}
	// 300000 * pow(3, 2) = 2700000
	got = GraviProdEnergy(3, 300000)
	if math.Abs(got-2700000) > 1e-6 {
		t.Fatalf("GraviProdEnergy(3): got %v, want 2700000", got)
	}
}

func TestMineConsEnergy(t *testing.T) {
	t.Parallel()
	// metalmine base=10: floor(10 * 10 * pow(1.1-3*0.0005, 10))
	want := math.Floor(10 * 10 * math.Pow(1.1-3*0.0005, 10))
	got := MineConsEnergy(10, 10, 3)
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("MineConsEnergy(10,10,3): got %v, want %v", got, want)
	}
	if MineConsEnergy(10, 0, 3) != 0 {
		t.Fatal("level 0 must return 0")
	}
}

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
