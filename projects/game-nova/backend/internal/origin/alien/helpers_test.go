package alien

import (
	"testing"
	"time"

	"oxsar/game-nova/pkg/rng"
)

func TestIsAttackTime(t *testing.T) {
	// Любой четверг — true; остальные дни — false.
	cases := []struct {
		t    time.Time
		want bool
	}{
		{time.Date(2026, 4, 30, 12, 0, 0, 0, time.UTC), true}, // четверг
		{time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC), false}, // вторник
		{time.Date(2026, 5, 1, 23, 59, 59, 0, time.UTC), false}, // пятница
		{time.Date(2026, 5, 7, 0, 0, 0, 1, time.UTC), true},   // четверг
	}
	for _, tc := range cases {
		if got := IsAttackTime(tc.t); got != tc.want {
			t.Errorf("IsAttackTime(%s [%s]): got %v, want %v",
				tc.t, tc.t.Weekday(), got, tc.want)
		}
	}
}

func TestRandRoundRange(t *testing.T) {
	r := rng.New(42)
	for i := 0; i < 1000; i++ {
		v := RandRoundRange(10, 20, r)
		if v < 10 || v > 20 {
			t.Fatalf("RandRoundRange out of bounds: %d", v)
		}
	}
	// min == max — детерминированный результат.
	if got := RandRoundRange(7, 7, r); got != 7 {
		t.Errorf("RandRoundRange(7,7) = %d, want 7", got)
	}
}

func TestRandRoundRangeDur(t *testing.T) {
	r := rng.New(42)
	min := 12 * time.Hour
	max := 24 * time.Hour
	for i := 0; i < 200; i++ {
		v := RandRoundRangeDur(min, max, r)
		if v < min || v > max {
			t.Fatalf("RandRoundRangeDur out of bounds: %v", v)
		}
	}
}

func TestRandFloatRange(t *testing.T) {
	r := rng.New(42)
	for i := 0; i < 1000; i++ {
		v := RandFloatRange(0.9, 1.1, r)
		if v < 0.9 || v >= 1.1 {
			t.Fatalf("RandFloatRange out of bounds: %v", v)
		}
	}
	if got := RandFloatRange(2.0, 2.0, r); got != 2.0 {
		t.Errorf("RandFloatRange(2,2) = %v, want 2.0", got)
	}
}

func TestPowerScale(t *testing.T) {
	cfg := DefaultConfig()
	r := rng.New(42)
	for i := 0; i < 50; i++ {
		s := PowerScaleNormal(cfg, r)
		if s < cfg.PowerScaleMin || s >= cfg.PowerScaleMax {
			t.Fatalf("PowerScaleNormal out: %v", s)
		}
		s = PowerScaleThursday(cfg, r)
		if s < cfg.ThursdayPowerMin || s >= cfg.ThursdayPowerMax {
			t.Fatalf("PowerScaleThursday out: %v", s)
		}
	}
}

func TestPowerScaleAfterControlTimes(t *testing.T) {
	// origin: 1 + control_times * 1.5 — растёт квадратично через цепочку.
	cases := []struct {
		ct   int
		want float64
	}{
		{0, 1.0},
		{1, 2.5},
		{2, 4.0},
		{5, 8.5},
	}
	for _, tc := range cases {
		got := PowerScaleAfterControlTimes(tc.ct)
		if got != tc.want {
			t.Errorf("PowerScaleAfterControlTimes(%d) = %v, want %v", tc.ct, got, tc.want)
		}
	}
}

func TestHoldingExtension(t *testing.T) {
	cfg := DefaultConfig()
	start := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	holds := start.Add(1 * time.Hour)

	// Платёж 0 — не меняет.
	if got := HoldingExtension(cfg, start, holds, 0); !got.Equal(holds) {
		t.Errorf("HoldingExtension(0) = %v, want %v", got, holds)
	}

	// 50 оксаров → +2 часа (origin: 60*60*2 * paid / 50).
	got := HoldingExtension(cfg, start, holds, 50)
	want := holds.Add(2 * time.Hour)
	if !got.Equal(want) {
		t.Errorf("HoldingExtension(50) = %v, want %v", got, want)
	}

	// Огромный платёж — clamp до max_real_time (15 дней от start).
	got = HoldingExtension(cfg, start, holds, 1_000_000)
	cap := start.Add(cfg.HaltingMaxRealTime)
	if !got.Equal(cap) {
		t.Errorf("HoldingExtension(huge) = %v, want %v (cap)", got, cap)
	}
}

func TestCalcGrabAmount(t *testing.T) {
	cfg := DefaultConfig()
	r := rng.New(42)

	// Под порогом — 0.
	if g := CalcGrabAmount(cfg, 50_000, r); g != 0 {
		t.Errorf("grab below threshold = %d, want 0", g)
	}
	if g := CalcGrabAmount(cfg, 100_000, r); g != 0 {
		t.Errorf("grab at threshold = %d, want 0", g)
	}

	// При credit=10M ожидаем grab в [8000, 10000].
	for i := 0; i < 100; i++ {
		g := CalcGrabAmount(cfg, 10_000_000, r)
		if g < 8000 || g > 10000 {
			t.Errorf("grab(10M) = %d, want in [8000,10000]", g)
		}
	}
}

func TestCalcGiftAmount(t *testing.T) {
	cfg := DefaultConfig()
	r := rng.New(42)

	// При очень богатом игроке — клампится по MaxGiftCredit≈500.
	g := CalcGiftAmount(cfg, 10_000_000, r)
	if g <= 0 || g > 600 { // допуск ±20% от 500
		t.Errorf("gift(10M) = %d, want ~500", g)
	}
	// При бедном (но > 0) — small gift.
	g = CalcGiftAmount(cfg, 1000, r)
	if g < 0 || g > 200 {
		t.Errorf("gift(1k) = %d, want small", g)
	}
}

func TestChangeMissionDelay_Bounds(t *testing.T) {
	cfg := DefaultConfig()
	r := rng.New(42)
	flight := 24 * time.Hour
	for i := 0; i < 200; i++ {
		d := ChangeMissionDelay(cfg, flight, r)
		if d < 0 || d > flight-10*time.Second {
			t.Fatalf("ChangeMissionDelay = %v out of [0, flight-10s]", d)
		}
	}
}

func TestHoldingAISubphaseDuration(t *testing.T) {
	cfg := DefaultConfig()
	r := rng.New(42)
	for ct := 1; ct < 50; ct++ {
		for i := 0; i < 20; i++ {
			d := HoldingAISubphaseDuration(cfg, ct, r)
			if d < 0 {
				t.Fatalf("subphase duration negative for ct=%d: %v", ct, d)
			}
		}
	}
}
