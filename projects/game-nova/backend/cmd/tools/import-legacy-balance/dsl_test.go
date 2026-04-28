package main

import (
	"math"
	"testing"
)

func intp(v int) *int        { return &v }
func float64p(v float64) *float64 { return &v }

func TestParse_Numbers(t *testing.T) {
	cases := []struct {
		src  string
		want float64
	}{
		{"42", 42},
		{"3.14", 3.14},
		{"0", 0},
	}
	for _, c := range cases {
		got, err := EvalNumber(c.src, VarBinding{})
		if err != nil {
			t.Errorf("%q: %v", c.src, err)
			continue
		}
		if got != c.want {
			t.Errorf("%q = %v, want %v", c.src, got, c.want)
		}
	}
}

func TestParse_Operators(t *testing.T) {
	cases := []struct {
		src  string
		want float64
	}{
		{"2 + 3", 5},
		{"10 - 4", 6},
		{"2 * 3 + 4", 10},
		{"2 + 3 * 4", 14},     // precedence
		{"(2 + 3) * 4", 20},   // grouping
		{"10 / 2", 5},
		{"-5 + 3", -2},
		{"2 ** 3", 8},
		{"2 ** 3 ** 2", 512},  // right-assoc
		{"-(2 + 3)", -5},
	}
	for _, c := range cases {
		got, err := EvalNumber(c.src, VarBinding{})
		if err != nil {
			t.Errorf("%q: %v", c.src, err)
			continue
		}
		if got != c.want {
			t.Errorf("%q = %v, want %v", c.src, got, c.want)
		}
	}
}

func TestParse_Variables(t *testing.T) {
	bind := VarBinding{
		Level: intp(5),
		Basic: float64p(60),
		Temp:  intp(100),
		Tech:  map[int]int{23: 12},
	}
	cases := []struct {
		src  string
		want float64
	}{
		{"{level}", 5},
		{"{basic}", 60},
		{"{temp}", 100},
		{"{tech=23}", 12},
		{"{tech=99}", 0},          // не открыт
		{"{level} * 2", 10},
		{"{basic} + {level}", 65},
	}
	for _, c := range cases {
		got, err := EvalNumber(c.src, bind)
		if err != nil {
			t.Errorf("%q: %v", c.src, err)
			continue
		}
		if got != c.want {
			t.Errorf("%q = %v, want %v", c.src, got, c.want)
		}
	}
}

func TestParse_Functions(t *testing.T) {
	cases := []struct {
		src  string
		want float64
	}{
		{"floor(3.7)", 3},
		{"floor(-3.2)", -4},
		{"ceil(3.2)", 4},
		{"round(3.5)", 4},
		{"round(2.5)", 3}, // half away from zero, не half-to-even
		{"round(-2.5)", -3},
		{"abs(-7)", 7},
		{"min(3, 5)", 3},
		{"max(3, 5)", 5},
		{"min(3, 5, 1, 4)", 1},
		{"max(3, 5, 1, 4)", 5},
		{"pow(2, 10)", 1024},
		{"pow(1.5, 2)", 2.25},
	}
	for _, c := range cases {
		got, err := EvalNumber(c.src, VarBinding{})
		if err != nil {
			t.Errorf("%q: %v", c.src, err)
			continue
		}
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%q = %v, want %v", c.src, got, c.want)
		}
	}
}

// TestEvalLegacyChargeMetalMine — проверка реальной формулы charge_metal
// METALMINE из na_construction: floor({basic} * pow(1.5, ({level} - 1))).
// {basic}=60. Эталоны взяты MySQL FLOOR(60 * POW(1.5, level-1)) на
// docker-mysql-1 — это и есть «истина» origin (PHP eval идёт через
// тот же float64 path что MySQL POW).
func TestEvalLegacyChargeMetalMine(t *testing.T) {
	src := "floor({basic} * pow(1.5, ({level} - 1)))"
	want := []float64{60, 90, 135, 202, 303, 455, 683, 1025, 1537, 2306}
	for i, w := range want {
		level := i + 1
		got, err := EvalNumber(src, VarBinding{Level: &level, Basic: float64p(60)})
		if err != nil {
			t.Fatalf("level %d: %v", level, err)
		}
		if got != w {
			t.Errorf("level %d: got %v, want %v", level, got, w)
		}
	}
}

// TestEvalLegacyProdMetalMine — динамическая формула: floor(30*L*pow(1.1+tech*0.0006,L)).
// Эталон при level=1, tech=0: 30 * 1 * 1.1^1 = 33.
// При level=10, tech=0: 30 * 10 * 1.1^10 ≈ 778.12 → floor=778.
func TestEvalLegacyProdMetalMine(t *testing.T) {
	src := "floor(30 * {level} * pow(1.1+{tech=23}*0.0006, {level}))"

	cases := []struct {
		level, tech int
		want        float64
	}{
		{1, 0, 33},
		{10, 0, 778},
		{20, 12, math.Floor(30 * 20 * math.Pow(1.1+12*0.0006, 20))},
	}
	for _, c := range cases {
		got, err := EvalNumber(src, VarBinding{
			Level: &c.level,
			Tech:  map[int]int{23: c.tech},
		})
		if err != nil {
			t.Errorf("level=%d tech=%d: %v", c.level, c.tech, err)
			continue
		}
		if got != c.want {
			t.Errorf("level=%d tech=%d: got %v, want %v", c.level, c.tech, got, c.want)
		}
	}
}

// TestEvalLegacyHydrogenLab — формула водорода с температурой
// (закрывает D-029): floor(10*L*pow(1.1+tech*0.0008,L)*(-0.002*temp+1.28)).
func TestEvalLegacyHydrogenLab(t *testing.T) {
	src := "floor(10 * {level} * pow(1.1+{tech=25}*0.0008, {level} )* (-0.002 * {temp} + 1.28))"
	level, tech, temp := 5, 0, 0
	got, err := EvalNumber(src, VarBinding{
		Level: &level,
		Tech:  map[int]int{25: tech},
		Temp:  &temp,
	})
	if err != nil {
		t.Fatal(err)
	}
	// 10 * 5 * 1.1^5 * 1.28 = 50 * 1.61051 * 1.28 = 103.07 → floor=103
	want := math.Floor(10 * 5 * math.Pow(1.1, 5) * 1.28)
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestIsDynamic(t *testing.T) {
	staticCases := []string{
		"42",
		"floor({basic} * pow(1.5, {level} - 1))",
		"50 * pow(1.5, {level})",
		"{basic} * pow(2, ({level} - 1))",
	}
	for _, s := range staticCases {
		ok, err := IsDynamic(s)
		if err != nil {
			t.Errorf("IsDynamic(%q): %v", s, err)
			continue
		}
		if ok {
			t.Errorf("%q must be STATIC, got dynamic", s)
		}
	}

	dynCases := []string{
		"floor(30 * {level} * pow(1.1+{tech=23}*0.0006, {level}))",
		"floor(10 * {level} * pow(1.1+{tech=25}*0.0008, {level})*(-0.002 * {temp} + 1.28))",
		"{building=14} + 5",
	}
	for _, s := range dynCases {
		ok, err := IsDynamic(s)
		if err != nil {
			t.Errorf("IsDynamic(%q): %v", s, err)
			continue
		}
		if !ok {
			t.Errorf("%q must be DYNAMIC, got static", s)
		}
	}
}

func TestParse_EmptyAndWhitespace(t *testing.T) {
	for _, src := range []string{"", "   ", "\t\n"} {
		e, err := Parse(src)
		if err != nil {
			t.Errorf("Parse(%q): unexpected error %v", src, err)
		}
		if e != nil {
			t.Errorf("Parse(%q): want nil expr, got %T", src, e)
		}
	}
}

func TestParse_Errors(t *testing.T) {
	bad := []string{
		"2 +",
		"(1 + 2",
		"sqrt(4)", // not in whitelist — parses but eval fails
		"unknown_var",
	}
	for _, src := range bad {
		_, err := EvalNumber(src, VarBinding{})
		if err == nil {
			t.Errorf("EvalNumber(%q): expected error, got nil", src)
		}
	}
}

func TestRoundHalfAwayFromZero(t *testing.T) {
	cases := []struct{ in, want float64 }{
		{0.5, 1},
		{1.5, 2},
		{2.5, 3}, // half-away-from-zero: 2.5 → 3 (PHP), не 2 (Go math.Round half-to-even ?)
		{-0.5, -1},
		{-2.5, -3},
		{0.4, 0},
		{0.6, 1},
	}
	for _, c := range cases {
		if got := roundHalfAwayFromZero(c.in); got != c.want {
			t.Errorf("roundHalfAwayFromZero(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
