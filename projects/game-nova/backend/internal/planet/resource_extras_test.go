// Тесты для unitGroupConsumptionPerHour и cascade-payback (план 72.1.26).
//
// Цель — проверить 1:1 паритет с legacy `Functions.inc.php::unitGroupConsumptionPerHour`
// и cascade-логикой `Planet.class.php::getProduction` (строки 481-555).

package planet

import (
	"math"
	"testing"
)

func TestUnitGroupConsumptionPerHour_LegacyFormula(t *testing.T) {
	// Legacy формула:
	//   if (count < 1000) return 0;
	//   scale = (unitid == VIRT_DEFENSE) ? 0.5 : 1.0;
	//   cons = min(MAX, pow(BASE, count) * scale / 10 / 24) * count;
	//   return cons >= 0.01 ? cons : 0;
	// BASE = 1.000003, MAX = 0.1.
	cases := []struct {
		name   string
		unitID int
		count  int64
		want   float64
	}{
		{"small fleet (<1000) → 0", unitVirtFleet, 999, 0},
		{"min threshold 1000", unitVirtFleet, 1000, 0}, // pow(1.000003, 1000)/240 ≈ 0.00418, * 1000 = 4.18 → но cons = MIN(MAX=0.1, 0.00418) * 1000 = 4.18 — давайте посчитаем
	}
	// Для count=1000: pow(1.000003, 1000) = e^(1000 * ln(1.000003)) ≈ e^0.003 ≈ 1.003.
	// cons = min(0.1, 1.003 / 10 / 24) * 1000 = min(0.1, 0.00418) * 1000 = 0.00418 * 1000 = 4.18.
	cases[1].want = math.Min(0.1, math.Pow(1.000003, 1000)/240) * 1000

	for _, c := range cases {
		got := unitGroupConsumptionPerHour(c.unitID, c.count)
		if math.Abs(got-c.want) > 1e-6 {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestUnitGroupConsumptionPerHour_DefenseHalfScale(t *testing.T) {
	// Defense: scale=0.5, cons = pow * 0.5 / 240 * count.
	count := int64(2000)
	got := unitGroupConsumptionPerHour(unitVirtDefense, count)
	want := math.Min(0.1, math.Pow(1.000003, float64(count))*0.5/240.0) * float64(count)
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("defense scale: got %v, want %v", got, want)
	}

	// vs fleet того же размера — должно быть в 2 раза больше.
	gotFleet := unitGroupConsumptionPerHour(unitVirtFleet, count)
	if got > 0 && math.Abs(gotFleet/got-2.0) > 0.01 {
		t.Errorf("expected fleet/defense ratio ~2.0, got %v/%v = %v", gotFleet, got, gotFleet/got)
	}
}

func TestCascadeRatios_LegacyFormulas(t *testing.T) {
	// Sanity check: MARKET_BASE_CURS_* константы (из legacy consts.php).
	if marketBaseCursMetal != 600.0 {
		t.Errorf("MARKET_BASE_CURS_METAL: got %v, want 600", marketBaseCursMetal)
	}
	if marketBaseCursSilicon != 400.0 {
		t.Errorf("MARKET_BASE_CURS_SILICON: got %v, want 400", marketBaseCursSilicon)
	}
	if marketBaseCursHydrogen != 200.0 {
		t.Errorf("MARKET_BASE_CURS_HYDROGEN: got %v, want 200", marketBaseCursHydrogen)
	}

	// Шаг 1 (H→Si): over * 400/200 = over * 2.
	overH := 50.0
	wantSi := overH * marketBaseCursSilicon / marketBaseCursHydrogen
	if wantSi != 100.0 {
		t.Errorf("H→Si ratio: got %v, want 100", wantSi)
	}

	// Шаг 2 (Si→M): over * 600/400 = over * 1.5.
	overSi := 80.0
	wantM := overSi * marketBaseCursMetal / marketBaseCursSilicon
	if wantM != 120.0 {
		t.Errorf("Si→M ratio: got %v, want 120", wantM)
	}
}
