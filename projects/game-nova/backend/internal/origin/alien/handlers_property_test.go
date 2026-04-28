package alien

import (
	"testing"

	"pgregory.net/rapid"

	"oxsar/game-nova/pkg/rng"
)

// TestProperty_CalcGrabAmount_Determinism — детерминизм решения грабежа
// (R4): один и тот же seed + cfg + credit даёт ровно тот же grab.
// Защищает handler от random-расхождений между retry'ами.
func TestProperty_CalcGrabAmount_Determinism(t *testing.T) {
	cfg := DefaultConfig()
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		credit := rapid.Int64Range(0, 1_000_000_000_000).Draw(t, "credit")

		r1 := rng.New(seed)
		r2 := rng.New(seed)
		got1 := CalcGrabAmount(cfg, credit, r1)
		got2 := CalcGrabAmount(cfg, credit, r2)
		if got1 != got2 {
			t.Fatalf("non-deterministic grab: seed=%d credit=%d → %d vs %d",
				seed, credit, got1, got2)
		}

		// Bound: при credit > GrabMinCredit grab ∈ [credit*0.0008, credit*0.001];
		// при credit <= GrabMinCredit grab == 0.
		if credit <= cfg.GrabMinCredit {
			if got1 != 0 {
				t.Fatalf("below threshold: got %d want 0", got1)
			}
		} else {
			lo := int64(float64(credit) * 0.0008 * 0.999) // допуск round
			hi := int64(float64(credit)*0.001*1.001) + 1
			if got1 < lo || got1 > hi {
				t.Fatalf("grab(%d) = %d, want in [%d, %d]", credit, got1, lo, hi)
			}
		}
	})
}

// TestProperty_CalcGiftAmount_Determinism — то же для CalcGiftAmount.
func TestProperty_CalcGiftAmount_Determinism(t *testing.T) {
	cfg := DefaultConfig()
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		credit := rapid.Int64Range(0, 1_000_000_000_000).Draw(t, "credit")

		r1 := rng.New(seed)
		r2 := rng.New(seed)
		got1 := CalcGiftAmount(cfg, credit, r1)
		got2 := CalcGiftAmount(cfg, credit, r2)
		if got1 != got2 {
			t.Fatalf("non-deterministic gift: seed=%d credit=%d → %d vs %d",
				seed, credit, got1, got2)
		}

		// Bound: gift ≤ MaxGiftCredit * 1.02 (origin: 500 * rand(0.98, 1.02)).
		// На очень низких credit gift тоже клампится в обе стороны.
		if got1 < 0 {
			t.Fatalf("gift negative: %d", got1)
		}
		maxAllowed := int64(float64(cfg.MaxGiftCredit)*1.02) + 1
		if got1 > maxAllowed {
			t.Fatalf("gift too large: %d > %d (cap)", got1, maxAllowed)
		}
	})
}

// TestProperty_HoldingExtension_Monotonic — каждое новое продление
// не уменьшает holds_until_at, и cap на HaltingMaxRealTime соблюдён.
func TestProperty_HoldingExtension_Monotonic(t *testing.T) {
	cfg := DefaultConfig()
	rapid.Check(t, func(t *rapid.T) {
		paid1 := rapid.Int64Range(0, 1000).Draw(t, "paid1")
		paid2 := rapid.Int64Range(0, 1000).Draw(t, "paid2")

		// Произвольный момент старта.
		now := nowFn()
		holds := now.Add(cfg.HaltingMinTime)

		ext1 := HoldingExtension(cfg, now, holds, paid1)
		ext2 := HoldingExtension(cfg, now, ext1, paid2)

		if ext1.Before(holds) {
			t.Fatalf("extension regressed: %v → %v", holds, ext1)
		}
		if ext2.Before(ext1) {
			t.Fatalf("second extension regressed: %v → %v", ext1, ext2)
		}
		// Cap: ext не превышает now + HaltingMaxRealTime.
		cap := now.Add(cfg.HaltingMaxRealTime)
		if ext2.After(cap) {
			t.Fatalf("extension exceeds cap: %v > %v", ext2, cap)
		}
	})
}
