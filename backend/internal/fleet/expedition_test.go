package fleet

import (
	"math"
	"testing"

	"github.com/oxsar/nova/backend/pkg/rng"
)

func TestCalcExpPower(t *testing.T) {
	t.Parallel()
	// No tech, no hours, no probes → 0
	if p := calcExpPower(0, 0, 0, 0); p != 0 {
		t.Fatalf("expected 0, got %v", p)
	}
	// expoTech=5, hours=3, no spy → 5 + 6 = 11
	got := calcExpPower(5, 0, 0, 3)
	if got != 11 {
		t.Fatalf("expected 11, got %v", got)
	}
	// spy_tech=10, 100 probes → 10/10 * pow(100, 0.4) ≈ 6.31
	got = calcExpPower(0, 10, 100, 0)
	want := float64(10) / 10.0 * math.Pow(100, 0.4)
	if math.Abs(got-want) > 0.001 {
		t.Fatalf("expected ~%.3f, got %.3f", want, got)
	}
}

func TestCalcVisitedScale(t *testing.T) {
	t.Parallel()
	// Нет посещений — штрафа нет, но hours_scale влияет.
	// hours=0 → hours_scale = pow(1/6, 1.5) ≈ 0.068
	vs := calcVisitedScale(0, 0)
	want := math.Pow(1.0/6.0, 1.5)
	if math.Abs(vs-want) > 0.001 {
		t.Fatalf("visits=0,hours=0: expected ~%.4f, got %.4f", want, vs)
	}
	// 20 посещений, hours=6 → vs = pow(20,-0.7) * pow(7/6, 1.5)
	vs = calcVisitedScale(20, 6)
	if vs >= 0.3 {
		t.Fatalf("expected < 0.3 with 20 visits, got %.4f", vs)
	}
	// hours=5 → scale должен быть выше нуля
	if calcVisitedScale(0, 5) <= 0 {
		t.Fatal("visitedScale must be > 0")
	}
}

func TestCalcExpWeights_ZeroPower(t *testing.T) {
	t.Parallel()
	r := rng.New(42)
	w := calcExpWeights(0, 0, 0, 0, 0, r)

	// При expPower=0, hours=0 — ship и battlefield отсутствуют (hours < 1/2)
	if w["ship"] != 0 {
		t.Fatalf("ship weight should be 0 at hours=0, got %.2f", w["ship"])
	}
	if w["battlefield"] != 0 {
		t.Fatalf("battlefield weight should be 0 at hours=0, got %.2f", w["battlefield"])
	}
	// artefact/credit тоже отсутствуют (hours < 4)
	if w["artefact"] != 0 {
		t.Fatalf("artefact weight should be 0 at hours=0, got %.2f", w["artefact"])
	}
	// resource, asteroid, delay, fast, nothing, lost ненулевые (после jitter могут быть ≥ 0)
	if w["resource"] <= 0 {
		t.Fatalf("resource weight should be > 0, got %.2f", w["resource"])
	}
	if w["nothing"] <= 0 {
		t.Fatalf("nothing weight should be > 0, got %.2f", w["nothing"])
	}
}

func TestCalcExpWeights_HighPower(t *testing.T) {
	t.Parallel()
	r := rng.New(99)
	w := calcExpWeights(5, 4, 0, 0, 0, r)
	// При hours=4 — artefact и credit должны быть ненулевыми (если jitter не обнулил)
	// Проверяем, что ключи присутствуют в карте.
	if _, ok := w["artefact"]; !ok {
		t.Fatal("artefact key missing from weights map")
	}
	if _, ok := w["credit"]; !ok {
		t.Fatal("credit key missing from weights map")
	}
}

func TestWeightedChoice_Deterministic(t *testing.T) {
	t.Parallel()
	// Единственный ненулевой вес для resource — всегда должен возвращать "resource".
	w := map[string]float64{"resource": 100}
	r := rng.New(1)
	for i := 0; i < 20; i++ {
		got := weightedChoice(w, r)
		if got != "resource" {
			t.Fatalf("expected 'resource', got %q", got)
		}
	}
}

func TestWeightedChoice_Empty(t *testing.T) {
	t.Parallel()
	w := map[string]float64{}
	r := rng.New(1)
	got := weightedChoice(w, r)
	if got != "nothing" {
		t.Fatalf("expected 'nothing' for empty weights, got %q", got)
	}
}

func TestCalcResK_Basic(t *testing.T) {
	t.Parallel()
	r := rng.New(7)
	// expPower=5, hours=5, visitedScale=1.0
	k := calcResK(r, 5, 5, 1.0)
	if k <= 0 {
		t.Fatalf("res_k must be > 0, got %d", k)
	}
	// Проверяем порядок величины: при expPower=5, hours=5
	// base = max(0.5, (1+pow(5,1.1))*5/40) ≈ max(0.5, 0.76) = 0.76
	// res_k ≈ 500_000..1_000_000 * 0.76 * ~1 * 2 → ≈ 760_000..1_520_000
	if k < 100_000 || k > 200_000_000 {
		t.Fatalf("res_k out of expected range: %d", k)
	}
}

func TestCalcPirateCount_Bounds(t *testing.T) {
	t.Parallel()
	// Слабый флот → min 3.
	r := rng.New(5)
	cnt := calcPirateCount(nil, nil, 0, r)
	if cnt < 3 {
		t.Fatalf("min pirate count should be 3, got %d", cnt)
	}
	if cnt > 500 {
		t.Fatalf("pirate count exceeds max 500, got %d", cnt)
	}
}
