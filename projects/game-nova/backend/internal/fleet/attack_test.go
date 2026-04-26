package fleet

import (
	"testing"

	"github.com/oxsar/nova/backend/internal/battle"
	"github.com/oxsar/nova/backend/internal/config"
)

// minCatalog — минимальный каталог для тестов grabLoot/calcDebris.
// ships — map[name]ShipSpec, как в YAML.
func minCatalog(ships map[string]config.ShipSpec) *config.Catalog {
	return &config.Catalog{
		Ships: config.ShipCatalog{Ships: ships},
	}
}

// TestGrabLoot_UnlimitedCargo — если cargo > 50% ресурсов, берём ровно 50%.
func TestGrabLoot_UnlimitedCargo(t *testing.T) {
	t.Parallel()
	cat := minCatalog(map[string]config.ShipSpec{
		"large_transport": {ID: 203, Cost: config.ResCost{Metal: 6000, Silicon: 6000}, Cargo: 25000},
	})
	survivors := []unitStack{{UnitID: 203, Count: 10}} // cargo = 250_000
	got := grabLoot(10000, 8000, 4000, survivors, cat, 0, 0, 0)
	if got.Metal != 5000 {
		t.Errorf("metal: want 5000, got %d", got.Metal)
	}
	if got.Silicon != 4000 {
		t.Errorf("silicon: want 4000, got %d", got.Silicon)
	}
	if got.Hydrogen != 2000 {
		t.Errorf("hydrogen: want 2000, got %d", got.Hydrogen)
	}
}

// TestGrabLoot_CargoConstraint — cargo ограничивает лут пропорционально.
func TestGrabLoot_CargoConstraint(t *testing.T) {
	t.Parallel()
	cat := minCatalog(map[string]config.ShipSpec{
		"small_transport": {ID: 202, Cost: config.ResCost{Metal: 2000, Silicon: 2000}, Cargo: 5000},
	})
	survivors := []unitStack{{UnitID: 202, Count: 1}} // cargo = 5000
	// 50% of planet: 5000m + 4000si + 2000h = 11000 want, but cargo = 5000
	got := grabLoot(10000, 8000, 4000, survivors, cat, 0, 0, 0)
	total := got.Metal + got.Silicon + got.Hydrogen
	if total > 5000 {
		t.Errorf("total loot %d exceeds cargo 5000", total)
	}
}

// TestGrabLoot_NoSurvivors — без выживших кораблей лут = 0.
func TestGrabLoot_NoSurvivors(t *testing.T) {
	t.Parallel()
	cat := minCatalog(nil)
	got := grabLoot(10000, 8000, 4000, nil, cat, 0, 0, 0)
	if got.Metal != 0 || got.Silicon != 0 || got.Hydrogen != 0 {
		t.Errorf("expected zero loot, got %+v", got)
	}
}

// TestGrabLoot_CarryReducesFree — уже загруженный carry уменьшает свободный cargo.
func TestGrabLoot_CarryReducesFree(t *testing.T) {
	t.Parallel()
	cat := minCatalog(map[string]config.ShipSpec{
		"large_transport": {ID: 203, Cost: config.ResCost{Metal: 6000, Silicon: 6000}, Cargo: 25000},
	})
	survivors := []unitStack{{UnitID: 203, Count: 1}} // cargo = 25000
	// carry уже 24000 → free = 1000
	got := grabLoot(10000, 8000, 4000, survivors, cat, 8000, 8000, 8000)
	total := got.Metal + got.Silicon + got.Hydrogen
	if total > 1000 {
		t.Errorf("total loot %d exceeds free cargo 1000", total)
	}
}

// TestCalcDebris_Basic — 30% metal + 30% silicon от lost ships.
func TestCalcDebris_Basic(t *testing.T) {
	t.Parallel()
	cat := minCatalog(map[string]config.ShipSpec{
		"fighter": {ID: 204, Cost: config.ResCost{Metal: 3000, Silicon: 1000}, Cargo: 50},
	})
	rep := battle.Report{
		Attackers: []battle.SideResult{{
			Units: []battle.UnitResult{
				{UnitID: 204, QuantityStart: 10, QuantityEnd: 4}, // 6 lost
			},
		}},
		Defenders: nil,
	}
	// 6 lost × (3000m + 1000si) × 30% = 5400m + 1800si
	dm, ds := calcDebris(rep, nil, cat)
	if dm != 5400 {
		t.Errorf("debris metal: want 5400, got %d", dm)
	}
	if ds != 1800 {
		t.Errorf("debris silicon: want 1800, got %d", ds)
	}
}

// TestCalcDebris_DefenseExcluded — defense не идёт в debris.
func TestCalcDebris_DefenseExcluded(t *testing.T) {
	t.Parallel()
	cat := minCatalog(nil)
	rep := battle.Report{
		Defenders: []battle.SideResult{{
			Units: []battle.UnitResult{
				{UnitID: 401, QuantityStart: 5, QuantityEnd: 0},
			},
		}},
	}
	defIDs := map[int]bool{401: true}
	dm, ds := calcDebris(rep, defIDs, cat)
	if dm != 0 || ds != 0 {
		t.Errorf("defense should produce zero debris, got %d %d", dm, ds)
	}
}

// TestDeriveSeed_Deterministic — одинаковый fleetID → одинаковый seed.
func TestDeriveSeed_Deterministic(t *testing.T) {
	t.Parallel()
	id := "550e8400-e29b-41d4-a716-446655440000"
	s1 := deriveSeed(id)
	s2 := deriveSeed(id)
	if s1 != s2 {
		t.Errorf("deriveSeed not deterministic: %d != %d", s1, s2)
	}
}

// TestDeriveSeed_Different — разные ID дают разные seed (high probability).
func TestDeriveSeed_Different(t *testing.T) {
	t.Parallel()
	s1 := deriveSeed("aaaa")
	s2 := deriveSeed("bbbb")
	if s1 == s2 {
		t.Error("different IDs produced same seed")
	}
}
