package alien

// План 66 Ф.6: property-based докрытие helper'ов AlienAI.
//
// Цели (R4):
//   - PickAttackTarget: пустой список / все ineligible → nil.
//   - PickCreditTarget: монотонность (увеличение credit ineligible-цели
//     не делает её eligible — только если все условия пройдены).
//   - GenerateFleet: total power ≤ target_power*(1+ε) при scale=1
//     (генератор останавливается при ≥ target).
//   - ApplyShuffledTechWeakening: каждый weakened уровень ≤ исходный+1
//     (origin: hi=level<3?level:level+1).

import (
	"testing"

	"pgregory.net/rapid"

	"oxsar/game-nova/pkg/rng"
)

// TestProperty_PickAttackTarget_EmptyReturnsNil — пустой список →
// PickAttackTarget возвращает nil без panic.
func TestProperty_PickAttackTarget_EmptyReturnsNil(t *testing.T) {
	cfg := DefaultConfig()
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		got := PickAttackTarget(nil, cfg, rng.New(seed))
		if got != nil {
			t.Fatalf("empty candidates → got %+v, want nil", *got)
		}
		got2 := PickAttackTarget([]TargetCandidate{}, cfg, rng.New(seed))
		if got2 != nil {
			t.Fatalf("empty slice → got %+v, want nil", *got2)
		}
	})
}

// TestProperty_PickAttackTarget_AllIneligibleReturnsNil — если ни одна
// цель не проходит eligibility (umode / inactive / мало кораблей),
// PickAttackTarget возвращает nil.
func TestProperty_PickAttackTarget_AllIneligibleReturnsNil(t *testing.T) {
	cfg := DefaultConfig()
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 20).Draw(t, "n")
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		cands := make([]TargetCandidate, n)
		for i := range cands {
			cands[i] = TargetCandidate{
				UserID: "u", PlanetID: "p",
				InUmode: true, // отсев по umode
			}
		}
		got := PickAttackTarget(cands, cfg, rng.New(seed))
		if got != nil {
			t.Fatalf("all umode → got %+v, want nil", *got)
		}
	})
}

// TestProperty_PickCreditTarget_EmptyReturnsNil — sanity для credit.
func TestProperty_PickCreditTarget_EmptyReturnsNil(t *testing.T) {
	cfg := DefaultConfig()
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		got := PickCreditTarget(nil, cfg, rng.New(seed))
		if got != nil {
			t.Fatalf("empty → got %+v, want nil", *got)
		}
	})
}

// TestProperty_PickCreditTarget_MonotonicByCredit — для одного и того
// же кандидата увеличение credit (при прочих равных) НЕ может сделать
// его ineligible (нет нисходящих порогов). А кандидат с credit ≤
// GrabMinCredit гарантированно ineligible.
func TestProperty_PickCreditTarget_MonotonicByCredit(t *testing.T) {
	cfg := DefaultConfig()
	rapid.Check(t, func(t *rapid.T) {
		base := TargetCandidate{
			UserID:          "u",
			PlanetID:        "p",
			LastActiveSeconds: 60, // активный
			InUmode:         false,
			UserShipCount:   cfg.FindCreditTargetUserShipsMin + 1,
			PlanetShipCount: cfg.FindCreditTargetPlanetShipsMin + 1,
		}

		lowCredit := rapid.Int64Range(0, cfg.GrabMinCredit).Draw(t, "lowCredit")
		highCredit := rapid.Int64Range(cfg.GrabMinCredit+1, 1_000_000_000_000).Draw(t, "highCredit")

		low := base
		low.Credit = lowCredit
		high := base
		high.Credit = highCredit

		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		// low (≤ threshold) → должен быть ineligible.
		if got := PickCreditTarget([]TargetCandidate{low}, cfg, rng.New(seed)); got != nil {
			t.Fatalf("credit=%d ≤ threshold → got non-nil; want nil", lowCredit)
		}
		// high (> threshold) → eligible (единственный).
		if got := PickCreditTarget([]TargetCandidate{high}, cfg, rng.New(seed)); got == nil {
			t.Fatalf("credit=%d > threshold + valid → got nil; want eligible", highCredit)
		}
	})
}

// TestProperty_GenerateFleet_PowerNotExcessive — alien power не
// превышает target_power*(1+ε) слишком сильно. Origin останавливает
// итерацию когда power >= target_power; допускается превышение на
// последнем инкременте, но не катастрофическое (cap множитель ×3).
func TestProperty_GenerateFleet_PowerNotExcessive(t *testing.T) {
	cfg := DefaultConfig()
	bySpec := map[int]ShipSpec{}
	for _, s := range alienAvailable {
		bySpec[s.UnitID] = s
	}
	tp := targetPower(targetSmall)
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		scale := 1.0
		f := GenerateFleet(targetSmall, alienAvailable, scale, cfg, rng.New(seed))
		var power float64
		for _, fu := range f {
			s := bySpec[fu.UnitID]
			power += (float64(s.Attack) + float64(s.Shield)) * float64(fu.Quantity)
		}
		// Допустимый максимум — 3x scale*target_power. Этого достаточно
		// чтобы отлавливать unbounded growth (например, бесконечный
		// цикл или забытый break при power >= target).
		cap := tp * scale * 3.0
		if power > cap {
			t.Fatalf("seed=%d alien power %.0f > 3×target %.0f", seed, power, cap)
		}
	})
}

// TestProperty_GenerateFleet_DeterministicSameSeed — два запуска с
// одинаковым seed дают идентичный флот.
func TestProperty_GenerateFleet_DeterministicSameSeed(t *testing.T) {
	cfg := DefaultConfig()
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		f1 := GenerateFleet(targetSmall, alienAvailable, 1.0, cfg, rng.New(seed))
		f2 := GenerateFleet(targetSmall, alienAvailable, 1.0, cfg, rng.New(seed))
		if len(f1) != len(f2) {
			t.Fatalf("seed=%d len %d vs %d", seed, len(f1), len(f2))
		}
		for i := range f1 {
			if f1[i] != f2[i] {
				t.Fatalf("seed=%d unit[%d]: %+v vs %+v", seed, i, f1[i], f2[i])
			}
		}
	})
}

// TestProperty_ApplyShuffledTechWeakening_NotAbove — каждый weakened
// уровень не превышает исходный+1 (origin:138, hi = level<3 ? level :
// level+1). Таким образом «угаданный» уровень ≤ 100% реального+1.
func TestProperty_ApplyShuffledTechWeakening_NotAbove(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		// Сгенерируем tech-карту с уровнями 0..30.
		nKeys := rapid.IntRange(1, 8).Draw(t, "nKeys")
		profile := make(TechProfile, nKeys)
		levels := make([]int, nKeys)
		for i := 0; i < nKeys; i++ {
			lvl := rapid.IntRange(0, 30).Draw(t, "lvl")
			profile[100+i] = lvl
			levels[i] = lvl
		}
		got := ApplyShuffledTechWeakening(profile, rng.New(seed))
		for i := 0; i < nKeys; i++ {
			gv, ok := got[100+i]
			if !ok {
				t.Fatalf("key %d missing in output", 100+i)
			}
			lvl := levels[i]
			expectedHi := lvl
			if lvl >= 3 {
				expectedHi = lvl + 1
			}
			if gv < 0 {
				t.Fatalf("key %d weakened=%d, want ≥0", 100+i, gv)
			}
			if gv > expectedHi {
				t.Fatalf("key %d weakened=%d > hi=%d (level=%d)",
					100+i, gv, expectedHi, lvl)
			}
		}
	})
}

// TestProperty_ShuffleKeyValues_PreservesMultiset — shuffle сохраняет
// мультимножество значений (origin:251-264, shuffle лишь меняет порядок).
func TestProperty_ShuffleKeyValues_PreservesMultiset(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		// Группа 3 ключей с произвольными уровнями.
		profile := TechProfile{
			201: rapid.IntRange(0, 30).Draw(t, "lvl1"),
			202: rapid.IntRange(0, 30).Draw(t, "lvl2"),
			203: rapid.IntRange(0, 30).Draw(t, "lvl3"),
		}
		keys := []int{201, 202, 203}
		shuffled := ShuffleKeyValues(profile, keys, rng.New(seed))

		// Multi-set сохраняется.
		want := []int{profile[201], profile[202], profile[203]}
		got := []int{shuffled[201], shuffled[202], shuffled[203]}
		if !sameMultiset(want, got) {
			t.Fatalf("shuffle changed multiset: want %v, got %v", want, got)
		}
	})
}

func sameMultiset(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[int]int)
	for _, x := range a {
		counts[x]++
	}
	for _, x := range b {
		counts[x]--
		if counts[x] < 0 {
			return false
		}
	}
	return true
}
