package market

import "testing"

func TestValidResource(t *testing.T) {
	t.Parallel()
	for _, r := range []string{"metal", "silicon", "hydrogen"} {
		if !validResource(r) {
			t.Errorf("expected %q to be valid", r)
		}
	}
	for _, r := range []string{"", "gold", "Metal", "METAL"} {
		if validResource(r) {
			t.Errorf("expected %q to be invalid", r)
		}
	}
}

func TestResourceCostRatios(t *testing.T) {
	t.Parallel()
	// Metal:Silicon:Hydrogen = 1:2:4 (OGame classic).
	if resourceCost["metal"] != 1.0 {
		t.Errorf("metal cost = %v, want 1.0", resourceCost["metal"])
	}
	if resourceCost["silicon"] != 2.0 {
		t.Errorf("silicon cost = %v, want 2.0", resourceCost["silicon"])
	}
	if resourceCost["hydrogen"] != 4.0 {
		t.Errorf("hydrogen cost = %v, want 4.0", resourceCost["hydrogen"])
	}
}

// TestExchangePreview проверяет формулу предварительного расчёта обмена
// (без вызова БД — только арифметика).
func TestExchangePreview(t *testing.T) {
	t.Parallel()
	// 1000 metal → silicon при userRate=1.2:
	// toAmount = 1000 * 1.0 / 2.0 / 1.2 = 416.67 → floor = 416.
	fromAmount := int64(1000)
	fromCost := resourceCost["metal"]
	toCost := resourceCost["silicon"]
	userRate := 1.2
	got := int64(float64(fromAmount) * fromCost / toCost / userRate)
	if got != 416 {
		t.Errorf("exchange preview metal→silicon 1000 @ 1.2 = %d, want 416", got)
	}

	// При userRate=1.0 честный паритет: 1000 * 1 / 2 = 500.
	userRate = 1.0
	got = int64(float64(fromAmount) * fromCost / toCost / userRate)
	if got != 500 {
		t.Errorf("exchange preview metal→silicon 1000 @ 1.0 = %d, want 500", got)
	}
}
