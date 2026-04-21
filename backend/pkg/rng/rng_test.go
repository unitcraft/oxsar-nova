package rng

import "testing"

func TestDeterministic(t *testing.T) {
	t.Parallel()
	a := New(42)
	b := New(42)
	for i := 0; i < 1000; i++ {
		if a.Uint64() != b.Uint64() {
			t.Fatalf("rng diverged at iter %d", i)
		}
	}
}

func TestZeroSeedNotDegenerate(t *testing.T) {
	t.Parallel()
	r := New(0)
	seen := map[uint64]struct{}{}
	for i := 0; i < 10; i++ {
		seen[r.Uint64()] = struct{}{}
	}
	if len(seen) < 10 {
		t.Fatalf("rng collapsed with zero seed, unique=%d", len(seen))
	}
}

func TestIntNRange(t *testing.T) {
	t.Parallel()
	r := New(7)
	for i := 0; i < 10_000; i++ {
		v := r.IntN(100)
		if v < 0 || v >= 100 {
			t.Fatalf("IntN out of range: %d", v)
		}
	}
}

func TestIntNZeroOrNegative(t *testing.T) {
	t.Parallel()
	r := New(99)
	if r.IntN(0) != 0 {
		t.Fatal("IntN(0) must return 0")
	}
	if r.IntN(-5) != 0 {
		t.Fatal("IntN(-5) must return 0")
	}
}

func TestFloat64Range(t *testing.T) {
	t.Parallel()
	r := New(123)
	for i := 0; i < 10_000; i++ {
		v := r.Float64()
		if v < 0 || v >= 1 {
			t.Fatalf("Float64 out of [0,1): %v", v)
		}
	}
}

func TestRollAlwaysTrue(t *testing.T) {
	t.Parallel()
	r := New(55)
	for i := 0; i < 1000; i++ {
		if !r.Roll(1.0) {
			t.Fatal("Roll(1.0) must always be true")
		}
	}
}

func TestRollAlwaysFalse(t *testing.T) {
	t.Parallel()
	r := New(55)
	for i := 0; i < 1000; i++ {
		if r.Roll(0.0) {
			t.Fatal("Roll(0.0) must always be false")
		}
	}
}

func TestRollStatistical(t *testing.T) {
	t.Parallel()
	r := New(777)
	const n = 100_000
	count := 0
	for i := 0; i < n; i++ {
		if r.Roll(0.3) {
			count++
		}
	}
	ratio := float64(count) / n
	// Ожидаем ~30%, допускаем ±3%.
	if ratio < 0.27 || ratio > 0.33 {
		t.Fatalf("Roll(0.3) ratio = %.3f, want 0.27..0.33", ratio)
	}
}
