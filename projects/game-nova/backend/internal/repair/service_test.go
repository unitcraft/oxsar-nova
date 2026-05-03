package repair

import (
	"testing"

	"oxsar/game-nova/internal/config"
)

func TestCeil10(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   float64
		want int64
	}{
		{0, 0},
		{-5, 0},
		{1, 10},
		{10, 10},
		{11, 20},
		{99, 100},
		{100, 100},
		{101, 110},
	}
	for _, c := range cases {
		got := ceil10(c.in)
		if got != c.want {
			t.Errorf("ceil10(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestScalePerUnit_Disassemble(t *testing.T) {
	t.Parallel()
	// Крейсер: cost 20000M/7000Si/2000H (legacy ships.yml).
	base := config.ResCost{Metal: 20000, Silicon: 7000, Hydrogen: 2000}

	// required = ceil(base * 0.2 / 10) * 10
	req := scalePerUnit(base, 0.2)
	if req.Metal != 4000 { // 20000 * 0.2 = 4000 → ceil(4000/10)*10 = 4000
		t.Errorf("req.Metal = %d, want 4000", req.Metal)
	}
	if req.Silicon != 1400 { // 7000 * 0.2 = 1400 → ceil(1400/10)*10 = 1400
		t.Errorf("req.Silicon = %d, want 1400", req.Silicon)
	}
	if req.Hydrogen != 400 { // 2000 * 0.2 = 400 → ceil(400/10)*10 = 400
		t.Errorf("req.Hydrogen = %d, want 400", req.Hydrogen)
	}

	// return = ceil(base * 0.9 / 10) * 10
	ret := scalePerUnit(base, 0.9)
	if ret.Metal != 18000 { // 20000 * 0.9 = 18000 → ceil(18000/10)*10 = 18000
		t.Errorf("ret.Metal = %d, want 18000", ret.Metal)
	}
	if ret.Silicon != 6300 { // 7000 * 0.9 = 6300 → ceil(6300/10)*10 = 6300
		t.Errorf("ret.Silicon = %d, want 6300", ret.Silicon)
	}
}

func TestScalePerUnit_ZeroBase(t *testing.T) {
	t.Parallel()
	base := config.ResCost{Metal: 0, Silicon: 0, Hydrogen: 0}
	got := scalePerUnit(base, 0.9)
	if got.Metal != 0 || got.Silicon != 0 || got.Hydrogen != 0 {
		t.Errorf("zero base should give zero: %+v", got)
	}
}

// План 72.1.56 B6: legacy 1:1 partial-repair (ExtRepair.class.php:510).
func TestClampRepairQuantity(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name              string
		requested, damaged int64
		want              int64
	}{
		{"requested zero → all damaged", 0, 10, 10},
		{"requested negative → all damaged", -5, 10, 10},
		{"requested less than damaged → as-is", 3, 10, 3},
		{"requested equals damaged → all", 10, 10, 10},
		{"requested greater than damaged → clamp to damaged", 99, 10, 10},
		{"damaged zero → zero", 5, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := clampRepairQuantity(tc.requested, tc.damaged)
			if got != tc.want {
				t.Errorf("clampRepairQuantity(%d,%d) = %d, want %d",
					tc.requested, tc.damaged, got, tc.want)
			}
		})
	}
}
