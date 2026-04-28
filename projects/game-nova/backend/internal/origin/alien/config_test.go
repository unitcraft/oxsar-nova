package alien

import (
	"testing"
	"time"
)

// TestDefaultConfig_LegacyParity проверяет что значения DefaultConfig
// 1-в-1 с origin consts.php:752-770. Регрессия здесь — нарушение R15.
func TestDefaultConfig_LegacyParity(t *testing.T) {
	c := DefaultConfig()

	cases := []struct {
		name string
		got  any
		want any
	}{
		{"FleetsNumberNormal", c.FleetsNumberNormal, 50},
		{"FleetsNumberAttackTime", c.FleetsNumberAttackTime, 250},
		{"AttackInterval", c.AttackInterval, 6 * 24 * time.Hour},
		{"GrabCreditInterval", c.GrabCreditInterval, 10 * 24 * time.Hour},
		{"FlyMinTime", c.FlyMinTime, 15 * time.Hour},
		{"FlyMaxTime", c.FlyMaxTime, 24 * time.Hour},
		{"HaltingMinTime", c.HaltingMinTime, 12 * time.Hour},
		{"HaltingMaxTime", c.HaltingMaxTime, 24 * time.Hour},
		{"HaltingMaxRealTime", c.HaltingMaxRealTime, 15 * 24 * time.Hour},
		{"ChangeMissionMinTime", c.ChangeMissionMinTime, 8 * time.Hour},
		{"ChangeMissionMaxTime", c.ChangeMissionMaxTime, 10 * time.Hour},
		{"GrabMinCredit", c.GrabMinCredit, int64(100_000)},
		{"GrabCreditMinPercent", c.GrabCreditMinPercent, 0.08},
		{"GrabCreditMaxPercent", c.GrabCreditMaxPercent, 0.10},
		{"GiftCreditMinPercent", c.GiftCreditMinPercent, 5.0},
		{"GiftCreditMaxPercent", c.GiftCreditMaxPercent, 10.0},
		{"MaxGiftCredit", c.MaxGiftCredit, int64(500)},
		{"FleetMaxDebris", c.FleetMaxDebris, 1_000_000_000.0},
		{"ThursdayPowerMin", c.ThursdayPowerMin, 1.5},
		{"ThursdayPowerMax", c.ThursdayPowerMax, 2.0},
		{"PowerScaleMin", c.PowerScaleMin, 0.9},
		{"PowerScaleMax", c.PowerScaleMax, 1.1},
		// 2*3600/50 = 144 сек/оксар, R1 ADR-0009.
		{"HoldingPaySecondsPerCredit", c.HoldingPaySecondsPerCredit, 144.0},
		{"FindTargetUserShipsMin", c.FindTargetUserShipsMin, int64(1000)},
		{"FindTargetPlanetShipsMin", c.FindTargetPlanetShipsMin, int64(100)},
		{"FindCreditTargetUserShipsMin", c.FindCreditTargetUserShipsMin, int64(300_000)},
		{"FindCreditTargetPlanetShipsMin", c.FindCreditTargetPlanetShipsMin, int64(10_000)},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, tc.got, tc.want)
		}
	}
}
