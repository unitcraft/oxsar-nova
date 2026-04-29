package score

import (
	"math"
	"testing"
)

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

func TestCalcDmPoints(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name                                  string
		points, ePoints, maxPoints, wantApprox float64
	}{
		{
			name:        "points=0 → 0 (защита от NaN/деления)",
			points:      0,
			ePoints:     50,
			maxPoints:   100,
			wantApprox:  0,
		},
		{
			name:        "ePoints=0 → 0 (внутренний множитель = 0)",
			points:      1000,
			ePoints:     0,
			maxPoints:   1000,
			wantApprox:  0,
		},
		{
			name:        "max_points=0 → нормализуется до 1 (избегаем 0^0.9)",
			points:      100,
			ePoints:     50,
			maxPoints:   0,
			// mul=min(50,100)=50, base=50*100/1^0.9=5000, sqrt=70.71, *100=7071
			wantApprox: 7071.07,
		},
		{
			name:        "типовой случай: points=points==max_points",
			points:      10000,
			ePoints:     200,
			maxPoints:   10000,
			// pow(10000/4000,1.1)+200/100 = 2.5^1.1+2 ≈ 2.788+2 = 4.788
			// min(200,4.788)=4.788; min(200,100)=100; max=100
			// base=100*10000/10000^0.9 = 100*10000/3981.07 ≈ 251.19
			// sqrt(251.19)*100 ≈ 1584.89
			wantApprox: 1584.89,
		},
		{
			name:        "крупный игрок: points=1M",
			points:      1_000_000,
			ePoints:     5000,
			maxPoints:   2_000_000,
			// p/4000=250, pow(250,1.1)≈472.34, +5000/100=50 → 522.34
			// min(5000,522.34)=522.34; min(5000,100)=100; max=522.34
			// base=522.34*1e6/pow(2e6,0.9)
			// Проверено вычислением CalcDmPoints — результат ≈3214.17.
			wantApprox: 3214.17,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := CalcDmPoints(c.points, c.ePoints, c.maxPoints)
			if math.Abs(got-c.wantApprox) > 1.0 {
				t.Errorf("CalcDmPoints(%v, %v, %v) = %v, want ≈ %v",
					c.points, c.ePoints, c.maxPoints, got, c.wantApprox)
			}
		})
	}
}

// Свойство: dm_points монотонна по points при фиксированных
// ePoints/maxPoints (более крупный игрок → больше dm).
func TestCalcDmPoints_MonotonicByPoints(t *testing.T) {
	t.Parallel()
	prev := 0.0
	for p := 1000.0; p <= 1_000_000; p *= 2 {
		got := CalcDmPoints(p, 1000, 1_000_000)
		if got <= prev {
			t.Errorf("dm_points не монотонна: prev=%v, points=%v → %v",
				prev, p, got)
		}
		prev = got
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
