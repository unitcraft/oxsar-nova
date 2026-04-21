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
