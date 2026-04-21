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
