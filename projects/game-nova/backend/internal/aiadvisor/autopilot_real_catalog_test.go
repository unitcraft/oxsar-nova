package aiadvisor

import (
	"path/filepath"
	"runtime"
	"testing"

	"oxsar/game-nova/internal/config"
)

// TestMeetsRequirements_RealCatalog — sanity-check что meetsRequirements
// корректно блокирует plasma_tech при недостаточном laser_tech.
//
// Использует реальный configs/requirements.yml, чтобы поймать
// рассогласование между snake_case ключами в requirements.yml и
// фактическими unit-ID, которые читает meetsRequirements.
func TestMeetsRequirements_RealCatalog(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	configsDir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "configs")
	cat, err := config.LoadCatalog(configsDir)
	if err != nil {
		t.Skipf("real catalog unavailable: %v", err)
		return
	}

	// Получаем ID интересующих research-юнитов.
	laserTech, ok := cat.Research.Research["laser_tech"]
	if !ok {
		t.Fatal("laser_tech not found in catalog")
	}
	plasmaReqs := cat.Requirements.Requirements["plasma_tech"]
	if len(plasmaReqs) == 0 {
		t.Fatal("plasma_tech has no requirements in catalog")
	}
	t.Logf("plasma_tech requirements: %d items", len(plasmaReqs))
	for _, r := range plasmaReqs {
		t.Logf("  - %s %s level %d", r.Kind, r.Key, r.Level)
	}

	// Сценарий: laser_tech ур.5 (нужно ≥8 для plasma).
	research := map[int]int{
		laserTech.ID: 5, // недостаточно (нужно 8)
	}
	// Energy_tech и ion_tech уровень достаточный.
	if energy, ok := cat.Research.Research["energy_tech"]; ok {
		research[energy.ID] = 10
	}
	if ion, ok := cat.Research.Research["ion_tech"]; ok {
		research[ion.ID] = 6
	}

	// Building-prereq не проверяется (planetBuildings=nil) — research-only.
	got := meetsRequirements("plasma_tech", nil, research, cat)
	if got {
		t.Errorf("meetsRequirements(plasma_tech) returned TRUE despite laser_tech=5<8; expected FALSE")
	}

	// Если поднять laser_tech до 8 — должно стать true.
	research[laserTech.ID] = 8
	got = meetsRequirements("plasma_tech", nil, research, cat)
	if !got {
		t.Errorf("meetsRequirements(plasma_tech) with all reqs met returned FALSE")
	}
}
