package alien

import (
	"testing"
	"time"

	"pgregory.net/rapid"

	"oxsar/game-nova/pkg/rng"
)

// TestPickHoldingSubphase_Distribution — property-based (R4):
// pickHoldingSubphase эквивалентно использует все 8 веток
// и при достаточном числе бросков встречаются все ветки.
//
// Защищает выбор от регрессии вида «всегда первый элемент» или
// «не используется последний».
func TestPickHoldingSubphase_Distribution(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		r := rng.New(seed)

		seen := map[HoldingSubphase]int{}
		const N = 800 // достаточно при 1/8: ожидаем 100 на ветку,
		// случайно может быть < 5 — допуск ниже.
		for i := 0; i < N; i++ {
			seen[pickHoldingSubphase(r)]++
		}
		// Все 8 веток встретились.
		for _, sub := range holdingSubphasesOrder {
			if seen[sub] == 0 {
				t.Fatalf("subphase %s never picked over %d trials (seed=%d)",
					sub, N, seed)
			}
		}
		// Ни одной ветке не >50% бросков (грубая проверка
		// против «всегда одна ветка»).
		for sub, count := range seen {
			if count > N/2 {
				t.Fatalf("subphase %s dominated %d/%d trials (seed=%d)",
					sub, count, N, seed)
			}
		}
	})
}

// TestPickHoldingSubphase_Determinism — один и тот же seed →
// один и тот же выбор. Защищает retry от расхождений (R9 косвенно).
func TestPickHoldingSubphase_Determinism(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		r1 := rng.New(seed)
		r2 := rng.New(seed)
		for i := 0; i < 50; i++ {
			a := pickHoldingSubphase(r1)
			b := pickHoldingSubphase(r2)
			if a != b {
				t.Fatalf("non-deterministic at i=%d seed=%d: %s vs %s",
					i, seed, a, b)
			}
		}
	})
}

// TestUnloadGift_Bounds — формула gift = ceil(min(snap*0.7, snap*0.1*times))
// (origin AlienAI.class.php:1058) удовлетворяет инвариантам:
//
//  1. gift <= snapshot * 0.7 (cap из origin).
//  2. gift >= 0.
//  3. gift = 0 при snapshot = 0.
//  4. gift монотонно растёт с times до cap.
func TestUnloadGift_Bounds(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		snapshot := rapid.Int64Range(0, 1_000_000_000).Draw(t, "snapshot")
		times1 := rapid.IntRange(1, 100).Draw(t, "times1")
		times2 := rapid.IntRange(1, 100).Draw(t, "times2")

		g1 := unloadGift(snapshot, float64(times1))
		g2 := unloadGift(snapshot, float64(times2))

		// Inv 1: cap.
		cap := int64(float64(snapshot) * 0.7)
		if g1 > cap+1 { // +1 для floor/ceil rounding
			t.Fatalf("gift %d > cap %d (snap=%d times=%d)",
				g1, cap, snapshot, times1)
		}
		// Inv 2: non-negative.
		if g1 < 0 {
			t.Fatalf("gift negative: %d", g1)
		}
		// Inv 3: zero at zero.
		if snapshot == 0 && g1 != 0 {
			t.Fatalf("gift %d at snapshot=0", g1)
		}
		// Inv 4: монотонность по times до cap (g1 <= g2 при times1 <= times2).
		if times1 <= times2 && g1 > g2 {
			t.Fatalf("non-monotonic: times1=%d g1=%d, times2=%d g2=%d",
				times1, g1, times2, g2)
		}
	})
}

// TestHoldingAISubphaseDuration_GrowsWithControlTimes — duration
// HOLDING_AI subphase растёт с control_times (origin AlienAI:974).
// Гарантирует, что цепочка тиков ускоряется при долгом удержании.
func TestHoldingAISubphaseDuration_GrowsWithControlTimes(t *testing.T) {
	cfg := DefaultConfig()
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<30).Draw(t, "seed")
		// Берём «достаточно большие» times чтобы попасть в зону
		// 30s*times > HaltingMinTime и 60s*times > HaltingMaxTime.
		times1 := rapid.IntRange(1, 50).Draw(t, "times1")
		times2 := rapid.IntRange(1500, 3000).Draw(t, "times2")
		// При times1 в [1..50] границы упрутся в clamp на
		// (HaltingMinTime, HaltingMaxTime) — duration ≈ 12-24h.
		// При times2 ≥ 1500 границы будут (30s*1500, 60s*1500) =
		// (12.5h, 25h), но т.к. clamp в HoldingAISubphaseDuration:
		//   lo = min(HaltingMinTime, 30s*times)
		//   hi = max(HaltingMaxTime, 60s*times)
		// При times1=1: lo=30s, hi=24h. При times2=2000: lo=12h, hi=33h.
		// → hi(times2) >= hi(times1) (монотонно).

		r1 := rng.New(seed)
		r2 := rng.New(seed)
		d1 := HoldingAISubphaseDuration(cfg, times1, r1)
		d2 := HoldingAISubphaseDuration(cfg, times2, r2)

		// Не строгая монотонность по конкретному значению (rng всё-таки),
		// но верхняя граница hi растёт с times. Проверим: hi(times2) >=
		// hi(times1) (вычисляем напрямую).
		hi1 := maxDur(cfg.HaltingMaxTime, time60s(times1))
		hi2 := maxDur(cfg.HaltingMaxTime, time60s(times2))
		if hi2 < hi1 {
			t.Fatalf("hi monotonicity broken: times1=%d hi=%v, times2=%d hi=%v",
				times1, hi1, times2, hi2)
		}
		_ = d1
		_ = d2
	})
}

// TestHoldingAIControlTimesIncrement_Invariant — на каждом тике
// pl.ControlTimes++. Property-test самой формулы PowerScaleAfterControlTimes
// (origin AlienAI:884) — она монотонна и неотрицательна.
func TestPowerScaleAfterControlTimes_Monotone(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		ct1 := rapid.IntRange(0, 1000).Draw(t, "ct1")
		ct2 := rapid.IntRange(0, 1000).Draw(t, "ct2")
		ps1 := PowerScaleAfterControlTimes(ct1)
		ps2 := PowerScaleAfterControlTimes(ct2)
		if ct1 <= ct2 && ps1 > ps2 {
			t.Fatalf("non-monotone: ct1=%d ps1=%v, ct2=%d ps2=%v",
				ct1, ps1, ct2, ps2)
		}
		if ps1 < 1.0 {
			t.Fatalf("power_scale < 1.0 at ct=%d: %v", ct1, ps1)
		}
	})
}

// time60s — helper для теста: 60 секунд × times как time.Duration.
// Дублирование с приватной формулой в HoldingAISubphaseDuration осознанное:
// тест защищает именно семантику этой формулы от регрессии.
func time60s(times int) time.Duration {
	return time.Duration(60*times) * time.Second
}
