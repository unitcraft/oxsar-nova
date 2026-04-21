package galaxy

import "testing"

func TestCoordsValidate(t *testing.T) {
	t.Parallel()
	bad := []Coords{
		{Galaxy: 0, System: 1, Position: 1},
		{Galaxy: 1, System: 0, Position: 1},
		{Galaxy: 1, System: 1, Position: 16},
	}
	for _, c := range bad {
		if err := c.Validate(); err == nil {
			t.Fatalf("expected validation error for %+v", c)
		}
	}
	if err := (Coords{1, 1, 1, false}).Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDistanceMonotonic(t *testing.T) {
	t.Parallel()
	base := Coords{1, 100, 5, false}
	near := Coords{1, 100, 8, false}
	far := Coords{1, 200, 8, false}
	galFar := Coords{2, 200, 8, false}
	if !(Distance(base, near) < Distance(base, far) && Distance(base, far) < Distance(base, galFar)) {
		t.Fatalf("distance ordering broken: near=%d far=%d galFar=%d",
			Distance(base, near), Distance(base, far), Distance(base, galFar))
	}
}

func TestDistanceSymmetric(t *testing.T) {
	t.Parallel()
	a := Coords{1, 100, 5, false}
	b := Coords{3, 200, 10, false}
	if Distance(a, b) != Distance(b, a) {
		t.Fatalf("distance must be symmetric: %d != %d", Distance(a, b), Distance(b, a))
	}
}

func TestDistanceSamePlanet(t *testing.T) {
	t.Parallel()
	a := Coords{2, 50, 7, false}
	if Distance(a, a) != 5 {
		t.Fatalf("same planet distance = %d, want 5", Distance(a, a))
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
		if got := Distance(c.a, c.b); got != c.want {
			t.Errorf("Distance(%+v, %+v) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}
