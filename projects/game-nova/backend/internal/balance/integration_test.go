package balance

import (
	"testing"
)

// TestLoadFor_OriginAppliesRealOverride — integration-тест: реальный
// файл configs/balance/origin.yaml (сгенерированный
// cmd/tools/import-legacy-balance) применяется поверх дефолта и
// перекрывает числа Lancer / Shadow / зданий.
//
// Это критерий приёма Ф.3 (план 64): для existing universeID="origin"
// LoadFor возвращает bundle с origin-числами, отличающимися от
// modern-defaults.
func TestLoadFor_OriginAppliesRealOverride(t *testing.T) {
	t.Parallel()
	l := NewLoader(configsRoot)
	def, err := l.LoadDefaults()
	if err != nil {
		t.Fatal(err)
	}
	origin, err := l.LoadFor("origin")
	if err != nil {
		t.Fatalf("LoadFor origin: %v", err)
	}
	if !origin.HasOverride {
		t.Fatal("origin must have override applied (configs/balance/origin.yaml present)")
	}

	// Building cost_base: nova и origin metal_mine 60/15/0 — совпадают.
	mmDef := def.Catalog.Buildings.Buildings["metal_mine"]
	mmOrigin := origin.Catalog.Buildings.Buildings["metal_mine"]
	if mmDef.CostBase.Metal != mmOrigin.CostBase.Metal {
		t.Errorf("metal_mine cost_base.metal: nova=%d origin=%d (плановое совпадение)",
			mmDef.CostBase.Metal, mmOrigin.CostBase.Metal)
	}

	// Lancer Ship: nova-cost 15000/35000/60000 + attack 5000;
	// origin-cost 2500/7500/15000 + attack 5500.
	lancerDef, ok := def.Catalog.Ships.Ships["lancer_ship"]
	if !ok {
		t.Skip("lancer_ship отсутствует в default ships.yml — skip integration check")
	}
	lancerOrigin := origin.Catalog.Ships.Ships["lancer_ship"]
	if lancerDef.Cost.Metal == lancerOrigin.Cost.Metal {
		t.Errorf("Lancer cost.metal должен отличаться: nova=%d origin=%d",
			lancerDef.Cost.Metal, lancerOrigin.Cost.Metal)
	}
	if lancerOrigin.Cost.Metal != 2500 {
		t.Errorf("Lancer origin cost.metal = %d, ожидался 2500 (na_construction.basic_metal)",
			lancerOrigin.Cost.Metal)
	}
	if lancerOrigin.Attack != 5500 {
		t.Errorf("Lancer origin attack = %d, ожидался 5500 (na_ship_datasheet.attack)",
			lancerOrigin.Attack)
	}

	// R0: default (nova) bundle не должен быть испорчен — Lancer cost
	// nova-default остался прежним.
	if lancerDef.Cost.Metal != 15000 {
		t.Errorf("nova default Lancer cost.metal = %d, ожидался 15000 (R0: не модифицировать)",
			lancerDef.Cost.Metal)
	}
}
