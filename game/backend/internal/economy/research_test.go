package economy

import "testing"

// Косвенный тест формулы исследования: CostForLevel + понимание, что
// для research используется тот же CostForLevel, что и для зданий.
func TestResearchCost_ScaleBy2(t *testing.T) {
	t.Parallel()
	// factor=2, level=3 => base * 4
	got := CostForLevel(Cost{Metal: 100, Silicon: 50}, 2.0, 3)
	if got.Metal != 400 || got.Silicon != 200 {
		t.Fatalf("expected m=400 s=200, got %+v", got)
	}
}
