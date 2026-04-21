package planet

import "testing"

func ptr64(v int64) *int64 { return &v }
func ptrF64(v float64) *float64 { return &v }

func TestInt64OrDefault(t *testing.T) {
	t.Parallel()
	if got := int64OrDefault(nil, 42); got != 42 {
		t.Errorf("nil pointer: got %d, want 42", got)
	}
	v := int64(100)
	if got := int64OrDefault(&v, 42); got != 100 {
		t.Errorf("non-nil pointer: got %d, want 100", got)
	}
}

func TestFloatOr(t *testing.T) {
	t.Parallel()
	if got := floatOr(nil, 3.14); got != 3.14 {
		t.Errorf("nil pointer: got %v, want 3.14", got)
	}
	v := 2.71
	if got := floatOr(&v, 3.14); got != 2.71 {
		t.Errorf("non-nil pointer: got %v, want 2.71", got)
	}
}

func TestClampAdd(t *testing.T) {
	t.Parallel()
	cases := []struct {
		cur, delta, max, want float64
		name                  string
	}{
		{50, 30, 100, 80, "normal add"},
		{80, 30, 100, 100, "clamp to max"},
		{10, -20, 100, 0, "clamp to zero"},
		{0, 0, 100, 0, "zero delta"},
		{50, 0, 100, 50, "no change"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := clampAdd(c.cur, c.delta, c.max); got != c.want {
				t.Errorf("clampAdd(%v, %v, %v) = %v, want %v", c.cur, c.delta, c.max, got, c.want)
			}
		})
	}
}
