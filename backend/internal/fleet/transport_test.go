package fleet

import (
	"testing"
	"time"
)

func TestTransportDuration_MinimumOneSecond(t *testing.T) {
	t.Parallel()
	// Нулевая дистанция — минимум 1 секунда (не 10 секунд базы, которая < 1s из-за деления).
	dur := transportDuration(0, 100, 100, 1)
	if dur < time.Second {
		t.Fatalf("minimum duration should be >= 1s, got %v", dur)
	}
}

func TestTransportDuration_ZeroSpeed(t *testing.T) {
	t.Parallel()
	// minSpeed=0 → fallback to 1 minute.
	dur := transportDuration(1000, 0, 100, 1)
	if dur != time.Minute {
		t.Fatalf("zero speed should return 1 minute, got %v", dur)
	}
}

func TestTransportDuration_IncreasesWithDistance(t *testing.T) {
	t.Parallel()
	d1 := transportDuration(1000, 100, 100, 1)
	d2 := transportDuration(5000, 100, 100, 1)
	if d1 >= d2 {
		t.Fatalf("longer distance should take longer: %v >= %v", d1, d2)
	}
}

func TestTransportDuration_SpeedPercentReduces(t *testing.T) {
	t.Parallel()
	dFast := transportDuration(5000, 100, 100, 1)
	dSlow := transportDuration(5000, 100, 50, 1)
	if dFast >= dSlow {
		t.Fatalf("100%% speed should be faster than 50%%: %v >= %v", dFast, dSlow)
	}
}

func TestTransportDuration_GameSpeedScales(t *testing.T) {
	t.Parallel()
	d1 := transportDuration(5000, 100, 100, 1)
	d2 := transportDuration(5000, 100, 100, 2)
	if d1 <= d2 {
		t.Fatalf("gameSpeed=2 should be faster than 1: %v <= %v", d1, d2)
	}
}

func TestTransportDuration_HigherMinSpeedFaster(t *testing.T) {
	t.Parallel()
	dFast := transportDuration(5000, 10000, 100, 1)
	dSlow := transportDuration(5000, 1000, 100, 1)
	if dFast >= dSlow {
		t.Fatalf("higher minSpeed should be faster: %v >= %v", dFast, dSlow)
	}
}
