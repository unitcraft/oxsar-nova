package alien

import (
	"testing"

	"oxsar/game-nova/pkg/rng"
)

func makeAttackable() TargetCandidate {
	return TargetCandidate{
		UserID:            "u1",
		PlanetID:          "p1",
		UserShipCount:     5000,
		PlanetShipCount:   500,
		LastActiveSeconds: 60,
	}
}

func TestPickAttackTarget_Empty(t *testing.T) {
	cfg := DefaultConfig()
	r := rng.New(7)
	if got := PickAttackTarget(nil, cfg, r); got != nil {
		t.Errorf("nil candidates → got %+v, want nil", got)
	}
}

func TestPickAttackTarget_Eligibility(t *testing.T) {
	cfg := DefaultConfig()

	// Один валидный normal-кандидат (HasOnlySatellites=false).
	// SolarSatelliteTargetChance=10% — в 90% случаев должен пикнуть,
	// в 10% — вернуть nil (satellite-only mode не нашёл ничего).
	cands := []TargetCandidate{makeAttackable()}
	picked := false
	for seed := uint64(1); seed < 30; seed++ {
		r := rng.New(seed)
		if got := PickAttackTarget(cands, cfg, r); got != nil {
			picked = true
			break
		}
	}
	if !picked {
		t.Errorf("expected at least one pick across 30 seeds; never picked")
	}

	tests := []struct {
		name  string
		mod   func(c *TargetCandidate)
	}{
		{"in_umode", func(c *TargetCandidate) { c.InUmode = true }},
		{"inactive >30m", func(c *TargetCandidate) { c.LastActiveSeconds = 31 * 60 }},
		{"too few user ships", func(c *TargetCandidate) { c.UserShipCount = cfg.FindTargetUserShipsMin }},
		{"too few planet ships", func(c *TargetCandidate) { c.PlanetShipCount = cfg.FindTargetPlanetShipsMin }},
		{"recent alien event", func(c *TargetCandidate) { c.HasRecentAlienEvent = true }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := makeAttackable()
			tc.mod(&c)
			r := rng.New(1)
			if got := PickAttackTarget([]TargetCandidate{c}, cfg, r); got != nil {
				t.Errorf("expected nil for %s; got %+v", tc.name, got)
			}
		})
	}
}

func TestPickCreditTarget_Eligibility(t *testing.T) {
	cfg := DefaultConfig()

	good := TargetCandidate{
		UserID:            "u1",
		PlanetID:          "p1",
		Credit:            500_000,
		UserShipCount:     1_000_000,
		PlanetShipCount:   50_000,
		LastActiveSeconds: 60,
	}

	r := rng.New(1)
	if got := PickCreditTarget([]TargetCandidate{good}, cfg, r); got == nil {
		t.Errorf("expected good candidate; got nil")
	}

	// Под порогом credit.
	bad := good
	bad.Credit = cfg.GrabMinCredit
	r = rng.New(1)
	if got := PickCreditTarget([]TargetCandidate{bad}, cfg, r); got != nil {
		t.Errorf("credit at threshold → expected nil; got %+v", got)
	}

	// recent grab event — отбрасывается.
	bad = good
	bad.HasRecentGrabEvent = true
	r = rng.New(1)
	if got := PickCreditTarget([]TargetCandidate{bad}, cfg, r); got != nil {
		t.Errorf("recent grab → expected nil; got %+v", got)
	}
}

// TestPickAttackTarget_SatelliteFilter — проверка satellite-фильтра.
// 10% — берём только satellite-only, 90% — наоборот. Метод
// детерминирован относительно r, поэтому крутим много seed'ов.
func TestPickAttackTarget_SatelliteFilter(t *testing.T) {
	cfg := DefaultConfig()
	cfg.SolarSatelliteTargetChance = 50 // искусственно 50/50 для измеряемости

	// Два кандидата: один satellite-only, один normal.
	normal := makeAttackable()
	sat := makeAttackable()
	sat.PlanetID = "p2"
	sat.HasOnlySatellites = true

	cands := []TargetCandidate{normal, sat}

	picksNormal, picksSat := 0, 0
	for seed := uint64(1); seed < 200; seed++ {
		r := rng.New(seed)
		got := PickAttackTarget(cands, cfg, r)
		if got == nil {
			continue
		}
		if got.HasOnlySatellites {
			picksSat++
		} else {
			picksNormal++
		}
	}
	// Ожидаем оба >0 (распределение примерно 50/50).
	if picksNormal == 0 || picksSat == 0 {
		t.Errorf("satellite filter не балансирует: normal=%d sat=%d",
			picksNormal, picksSat)
	}
}
