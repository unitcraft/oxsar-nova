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

// План 72.1.57 Ф.6: validateBattleLevels — pure helper для be_points
// усилений (legacy Mission.class.php:1638-1671).
func TestValidateBattleLevels(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		lvls     BattleLevels
		bePoints int64
		wantCost int64
		wantErr  error
	}{
		{
			name:    "all zero → no cost no error",
			lvls:    BattleLevels{},
			wantCost: 0,
		},
		{
			name:     "negative gun → ErrBattleLevelNegative (advanced battle off)",
			lvls:     BattleLevels{Gun: -5},
			bePoints: 10000,
			wantErr:  ErrBattleLevelNegative,
		},
		{
			name:     "K=10, gun=15 → ErrBattleLevelExceedsCap",
			lvls:     BattleLevels{Gun: 15},
			bePoints: 1000,
			wantErr:  ErrBattleLevelExceedsCap,
		},
		{
			name:     "K=10, gun=5+shield=3+shell=4=12 > 10 → ErrBattleLevelsBudget",
			lvls:     BattleLevels{Gun: 5, Shield: 3, Shell: 4},
			bePoints: 1000,
			wantErr:  ErrBattleLevelsBudget,
		},
		{
			name:     "K=10, gun=5+shield=3 → cost = 800",
			lvls:     BattleLevels{Gun: 5, Shield: 3},
			bePoints: 1000,
			wantCost: 800,
		},
		{
			name:     "K=20 cap (be_points=999999), all-max=20+0+0+0+0 → cost=2000",
			lvls:     BattleLevels{Gun: 20},
			bePoints: 999999,
			wantCost: 2000,
		},
		{
			name:     "K=20 cap full budget: 20+0+0+0+0",
			lvls:     BattleLevels{Gun: 4, Shield: 4, Shell: 4, Ballistics: 4, Masking: 4},
			bePoints: 999999,
			wantCost: 2000,
		},
		{
			name:     "K=20 cap, 5+5+5+5+5 = 25 > 20 → budget exceeded",
			lvls:     BattleLevels{Gun: 5, Shield: 5, Shell: 5, Ballistics: 5, Masking: 5},
			bePoints: 999999,
			wantErr:  ErrBattleLevelsBudget,
		},
		{
			name:     "K=0 (be_points=50<100), gun=1 → ExceedsCap",
			lvls:     BattleLevels{Gun: 1},
			bePoints: 50,
			wantErr:  ErrBattleLevelExceedsCap,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cost, err := validateBattleLevels(tc.lvls, tc.bePoints)
			if tc.wantErr != nil {
				if err != tc.wantErr {
					t.Fatalf("err = %v, want %v", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if cost != tc.wantCost {
				t.Errorf("cost = %d, want %d", cost, tc.wantCost)
			}
		})
	}
}

func TestBattleLevels_Sum_IsZero(t *testing.T) {
	t.Parallel()
	if !(BattleLevels{}).IsZero() {
		t.Error("zero struct must be IsZero=true")
	}
	if (BattleLevels{Gun: 1}).IsZero() {
		t.Error("non-zero must be IsZero=false")
	}
	got := (BattleLevels{Gun: 1, Shield: 2, Shell: 3, Ballistics: 4, Masking: 5}).Sum()
	if got != 15 {
		t.Errorf("Sum = %d, want 15", got)
	}
}
