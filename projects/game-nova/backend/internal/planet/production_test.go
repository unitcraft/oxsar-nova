package planet

import (
	"math"
	"testing"

	"oxsar/game-nova/internal/config"
)

// TestProductionRatesMetal проверяет, что productionRatesDSL возвращает
// правильное производство металла по статической формуле metalmine
// (§5.2.1 ТЗ, buildingid=1).
func TestProductionRatesMetal(t *testing.T) {
	t.Parallel()

	cat := &config.Catalog{
		Buildings: config.BuildingCatalog{Buildings: map[string]config.BuildingSpec{
			"metal_mine":       {ID: 1},
			"silicon_lab":      {ID: 2},
			"hydrogen_lab":     {ID: 3},
			"solar_plant":      {ID: 4},
			"metal_storage":    {ID: 9},
			"silicon_storage":  {ID: 10},
			"hydrogen_storage": {ID: 11},
		}},
	}

	svc := NewService(nil, nil, cat)

	p := &Planet{
		TempMin:       20,
		TempMax:       60,
		EnergyFactor:  1,
		ProduceFactor: 1,
	}
	buildingLvl := map[int]int{
		1: 10, // metal_mine
		4: 10, // solar_plant (чтобы хватало энергии)
	}
	techLvl := map[int]int{
		23: 5, // laser_tech
		18: 3, // energy_tech
	}

	r := svc.productionRatesDSL(p, buildingLvl, techLvl)

	// Ожидаемое производство metal по legacy-формуле:
	// floor(30 * level * pow(1.1 + tech*0.0006, level))
	want := math.Floor(30 * 10 * math.Pow(1.1+5*0.0006, 10))
	// energyRatio=1 при достаточной энергии (10 уровней солнечки >>
	// потребления шахты).
	got := r.metalPerSec * 3600.0 // per-hour для сравнения с формулой
	if math.Abs(got-want) > 1e-6 {
		t.Fatalf("metal production: got %v, want %v", got, want)
	}
}

