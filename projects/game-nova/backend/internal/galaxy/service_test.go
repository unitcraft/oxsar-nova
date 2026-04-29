package galaxy

import "testing"

const (
	testNumGalaxies = 8
	testNumSystems  = 600
)

func TestCoordsValidate(t *testing.T) {
	t.Parallel()
	bad := []Coords{
		{Galaxy: 0, System: 1, Position: 1},
		{Galaxy: testNumGalaxies + 1, System: 1, Position: 1},
		{Galaxy: 1, System: 0, Position: 1},
		{Galaxy: 1, System: testNumSystems + 1, Position: 1},
		{Galaxy: 1, System: 1, Position: 16},
	}
	for _, c := range bad {
		if err := c.Validate(testNumGalaxies, testNumSystems); err == nil {
			t.Fatalf("expected validation error for %+v", c)
		}
	}
	if err := (Coords{1, 1, 1, false}).Validate(testNumGalaxies, testNumSystems); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDistanceMonotonic(t *testing.T) {
	t.Parallel()
	base := Coords{1, 100, 5, false}
	near := Coords{1, 100, 8, false}
	far := Coords{1, 200, 8, false}
	galFar := Coords{2, 200, 8, false}
	dn := Distance(base, near, testNumSystems)
	df := Distance(base, far, testNumSystems)
	dg := Distance(base, galFar, testNumSystems)
	if !(dn < df && df < dg) {
		t.Fatalf("distance ordering broken: near=%d far=%d galFar=%d", dn, df, dg)
	}
}

func TestDistanceSymmetric(t *testing.T) {
	t.Parallel()
	a := Coords{1, 100, 5, false}
	b := Coords{3, 200, 10, false}
	if Distance(a, b, testNumSystems) != Distance(b, a, testNumSystems) {
		t.Fatalf("distance must be symmetric: %d != %d",
			Distance(a, b, testNumSystems), Distance(b, a, testNumSystems))
	}
}

func TestDistanceSamePlanet(t *testing.T) {
	t.Parallel()
	a := Coords{2, 50, 7, false}
	if Distance(a, a, testNumSystems) != 5 {
		t.Fatalf("same planet distance = %d, want 5", Distance(a, a, testNumSystems))
	}
}

func TestDistanceFormulas(t *testing.T) {
	t.Parallel()
	cases := []struct {
		a, b Coords
		want int
	}{
		// Разные галактики: 20000 * |dg|.
		{Coords{1, 1, 1, false}, Coords{3, 1, 1, false}, 20000 * 2},
		// Та же галактика, разные системы: 2700 + 95*|ds|.
		{Coords{1, 1, 1, false}, Coords{1, 11, 1, false}, 2700 + 95*10},
		// Та же система, разные позиции: 1000 + 5*|dp|.
		{Coords{1, 1, 1, false}, Coords{1, 1, 6, false}, 1000 + 5*5},
	}
	for _, c := range cases {
		if got := Distance(c.a, c.b, testNumSystems); got != c.want {
			t.Errorf("Distance(%+v, %+v) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

// TestDistanceWraparound — план 72.1 ч.12: системы образуют кольцо,
// расстояние = min(|s1-s2|, numSystems - |s1-s2|).
func TestDistanceWraparound(t *testing.T) {
	t.Parallel()
	// numSystems=600. Системы 1 и 600 — соседи (diff=1, не 599).
	a := Coords{1, 1, 1, false}
	b := Coords{1, 600, 1, false}
	want := 2700 + 95*1
	if got := Distance(a, b, 600); got != want {
		t.Fatalf("wraparound distance(1↔600) = %d, want %d", got, want)
	}
	// Системы 1 и 302 — линейная разница 301 (>300=600/2),
	// кольцевая разница 600-301=299. min(301, 299)=299.
	c := Coords{1, 302, 1, false}
	wantWrap := 2700 + 95*299
	if got := Distance(a, c, 600); got != wantWrap {
		t.Fatalf("wraparound distance(1↔302) = %d, want %d", got, wantWrap)
	}
}
