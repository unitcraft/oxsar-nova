package formula

import (
	"math"
	"testing"
)

func mustEval(t *testing.T, src string, c Context) float64 {
	t.Helper()
	e, err := Parse(src)
	if err != nil {
		t.Fatalf("parse %q: %v", src, err)
	}
	v, err := e.Eval(c)
	if err != nil {
		t.Fatalf("eval %q: %v", src, err)
	}
	return v
}

func TestEmpty(t *testing.T) {
	t.Parallel()
	v := mustEval(t, "", Context{})
	if v != 0 {
		t.Fatalf("empty formula must eval to 0, got %v", v)
	}
}

func TestArithmetic(t *testing.T) {
	t.Parallel()
	cases := map[string]float64{
		"1 + 2":              3,
		"2 * 3 + 1":          7,
		"2 * (3 + 1)":        8,
		"10 / 4":             2.5,
		"-3 + 5":             2,
		"+7":                 7,
		"pow(2, 10)":         1024,
		"floor(3.9)":         3,
		"ceil(3.1)":          4,
		"round(3.5)":         4,
		"sqrt(16)":           4,
		"abs(-5)":            5,
		"min(3, 7) + max(1,2)": 5,
	}
	for src, want := range cases {
		got := mustEval(t, src, Context{})
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("%q = %v, want %v", src, got, want)
		}
	}
}

func TestVariables(t *testing.T) {
	t.Parallel()
	c := Context{
		Level:       5,
		Basic:       60,
		Temperature: 20,
		Tech:        map[int]int{23: 10, 18: 3},
	}
	// Формула metalmine: floor(30 * {level} * pow(1.1+{tech=23}*0.0006, {level}))
	got := mustEval(t, "floor(30 * {level} * pow(1.1+{tech=23}*0.0006, {level}))", c)
	// Вычислим вручную.
	want := math.Floor(30 * 5 * math.Pow(1.1+10*0.0006, 5))
	if got != want {
		t.Fatalf("metalmine formula: got %v, want %v", got, want)
	}
}

func TestVariables_TechMissingZero(t *testing.T) {
	t.Parallel()
	c := Context{Level: 1} // Tech = nil
	v := mustEval(t, "{tech=99}", c)
	if v != 0 {
		t.Fatalf("missing tech must be 0, got %v", v)
	}
}

func TestVariables_TempAffectsHydrogenLab(t *testing.T) {
	t.Parallel()
	// hydrogen_lab: factor = (-0.002 * {temp} + 1.28)
	got := mustEval(t, "-0.002 * {temp} + 1.28", Context{Temperature: 40})
	want := -0.002*40 + 1.28
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("hydrogen_lab temp factor: got %v want %v", got, want)
	}
}

func TestParseErrors(t *testing.T) {
	t.Parallel()
	bad := []string{
		"1 +",             // trailing op
		"1 + * 2",         // double op
		"(1 + 2",          // unbalanced
		"foo(1)",          // unknown function
		"pow(1)",          // wrong arity
		"{unknown}",       // unknown variable
		"{tech=}",         // empty tech id
		"{tech=abc}",      // non-numeric tech id
		"{tech=1",         // unterminated variable
		"1 ^ 2",           // unsupported op
		"System.exit()",   // unknown function (защита от фантазии)
	}
	for _, src := range bad {
		if _, err := Parse(src); err == nil {
			t.Errorf("expected parse error for %q", src)
		}
	}
}

func TestEvalErrors(t *testing.T) {
	t.Parallel()
	e, err := Parse("1 / 0")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if _, err := e.Eval(Context{}); err == nil {
		t.Fatalf("expected division-by-zero error")
	}
}

func TestDeterminism(t *testing.T) {
	t.Parallel()
	src := "floor({basic} * pow(1.5, ({level} - 1)))"
	c := Context{Level: 7, Basic: 60}
	var first float64
	for i := 0; i < 100; i++ {
		v := mustEval(t, src, c)
		if i == 0 {
			first = v
		} else if v != first {
			t.Fatalf("non-deterministic at iter %d: %v vs %v", i, v, first)
		}
	}
}

// TestRealLegacyFormulas — проверяет, что настоящие формулы из
// oxsar2/sql/table_dump/na_construction.sql парсятся и вычисляются.
// Это не golden-тест (точные ожидаемые значения мы проверим в
// import-datasheets на M0.1), здесь только валидность синтаксиса.
func TestRealLegacyFormulas(t *testing.T) {
	t.Parallel()
	legacy := []string{
		// METALMINE
		"floor(30 * {level} * pow(1.1+{tech=23}*0.0006, {level}))",
		// SILICON_LAB
		"floor(20 * {level} * pow(1.1+{tech=24}*0.0007, {level}))",
		// HYDROGEN_LAB
		"floor(10 * {level} * pow(1.1+{tech=25}*0.0008, {level} )* (-0.002 * {temp} + 1.28))",
		// SOLAR_PLANT
		"floor(20 * {level} * pow(1.1+{tech=18}*0.0005, {level}))",
		// MINE cons_energy
		"floor(10 * {level} * pow(1.1-{tech=18}*0.0005, {level}))",
		// charge_metal шахт
		"floor({basic} * pow(1.5, ({level} - 1)))",
		// ROBOTIC_FACTORY charge_*
		"{basic} * pow(2, ({level} - 1))",
		// METAL_STORAGE charge_credit
		"(1 + ceil(pow(1.6, {level}))) * 50000",
		// HYDROGEN_PLANT charge
		"floor({basic} * pow(1.8, ({level} - 1)))",
		// Hydrogen lab cons_energy
		"floor(20 * {level} * pow(1.1-{tech=18}*0.0005, {level}))",
	}
	ctx := Context{Level: 5, Basic: 60, Temperature: 20, Tech: map[int]int{18: 2, 23: 3, 24: 4, 25: 5}}
	for _, f := range legacy {
		v := mustEval(t, f, ctx)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Errorf("formula %q produced invalid value %v", f, v)
		}
	}
}
