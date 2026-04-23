package score

import "testing"

func TestColumnFor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		want  string
	}{
		{"b", "b_points"},
		{"r", "r_points"},
		{"u", "u_points"},
		{"a", "a_points"},
		{"total", "points"},
		{"", "points"},
		{"unknown", "points"},
	}
	for _, c := range cases {
		if got := columnFor(c.input); got != c.want {
			t.Errorf("columnFor(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestRoundPts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input float64
		want  float64
	}{
		{0, 0},
		{1.234, 1.23},
		{1.235, 1.24},
		{100.0, 100.0},
		{3.14159, 3.14},
	}
	for _, c := range cases {
		if got := roundPts(c.input); got != c.want {
			t.Errorf("roundPts(%v) = %v, want %v", c.input, got, c.want)
		}
	}
}

func TestScoreCoefficients_Defaults(t *testing.T) {
	t.Parallel()
	// Dominator defaults — must not change without ADR.
	svc := NewService(nil, nil)
	if svc.kBld != 0.00005 {
		t.Errorf("kBld = %v, want 0.00005", svc.kBld)
	}
	if svc.kRes != 0.0005 {
		t.Errorf("kRes = %v, want 0.0005", svc.kRes)
	}
	if svc.kUnt != 0.002 {
		t.Errorf("kUnt = %v, want 0.002", svc.kUnt)
	}
}
