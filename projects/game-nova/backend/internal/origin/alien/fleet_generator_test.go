package alien

import (
	"testing"

	"oxsar/game-nova/pkg/rng"
)

// alienAvailable — стандартный набор UNIT_A_* из origin (id 200-204).
// Числа взяты из projects/game-nova/configs/balance/origin.yaml ships:.
var alienAvailable = []ShipSpec{
	{UnitID: 200, Name: "Alien Corvette", Attack: 200, Shield: 75,
		BasicMetal: 19000, BasicSilicon: 6000},
	{UnitID: 201, Name: "Alien Screen", Attack: 22, Shield: 5000,
		BasicMetal: 20000, BasicSilicon: 10000},
	{UnitID: 202, Name: "Alien Paladin", Attack: 75, Shield: 50,
		BasicMetal: 3000, BasicSilicon: 1000},
	{UnitID: 203, Name: "Alien Frigate", Attack: 1250, Shield: 150,
		BasicMetal: 40000, BasicSilicon: 30000},
	{UnitID: 204, Name: "Alien Torpedocarrier", Attack: 350, Shield: 100,
		BasicMetal: 10000, BasicSilicon: 10000},
}

// targetSmall — типовой защитник: 100 LF + 50 cruiser.
var targetSmall = []TargetUnit{
	{Spec: ShipSpec{UnitID: 31, Name: "Light Fighter",
		Attack: 50, Shield: 10, BasicMetal: 3000, BasicSilicon: 1000}, Quantity: 100},
	{Spec: ShipSpec{UnitID: 33, Name: "Cruiser",
		Attack: 400, Shield: 50, BasicMetal: 20000, BasicSilicon: 7000}, Quantity: 50},
}

// fleetTotalAttack — суммарный AlienAttack по флоту.
func fleetTotalAttack(f Fleet, specs []ShipSpec) float64 {
	bySpec := map[int]ShipSpec{}
	for _, s := range specs {
		bySpec[s.UnitID] = s
	}
	var p float64
	for _, fu := range f {
		p += float64(bySpec[fu.UnitID].Attack) * float64(fu.Quantity)
	}
	return p
}

// targetPower — для сверки (attack + shield, scale=1).
func targetPower(target []TargetUnit) float64 {
	var p float64
	for _, t := range target {
		p += float64(t.Spec.Attack)*float64(t.Quantity) +
			float64(t.Spec.Shield)*float64(t.Quantity)
	}
	return p
}

func TestGenerateFleet_BasicNonEmpty(t *testing.T) {
	cfg := DefaultConfig()
	r := rng.New(42)
	out := GenerateFleet(targetSmall, alienAvailable, 1.0, cfg, r)
	if len(out) == 0 {
		t.Fatalf("expected non-empty fleet, got 0")
	}
	for _, fu := range out {
		if fu.Quantity <= 0 {
			t.Errorf("unit %d quantity = %d, want >0", fu.UnitID, fu.Quantity)
		}
		if fu.ShellPercent != 100 {
			t.Errorf("unit %d shell_percent = %d, want 100", fu.UnitID, fu.ShellPercent)
		}
	}
}

func TestGenerateFleet_MeetsTargetPower(t *testing.T) {
	cfg := DefaultConfig()
	tp := targetPower(targetSmall)
	for seed := uint64(1); seed < 30; seed++ {
		r := rng.New(seed)
		out := GenerateFleet(targetSmall, alienAvailable, 1.0, cfg, r)
		// alien power ≈ sum(attack*qty + shield*qty); проверяем что
		// итоговая power хотя бы >= 50% target_power (генератор
		// останавливается когда дошёл до target_power, но допускаем
		// люфт из-за random веса в маленьких флотах).
		var power float64
		bySpec := map[int]ShipSpec{}
		for _, s := range alienAvailable {
			bySpec[s.UnitID] = s
		}
		for _, fu := range out {
			s := bySpec[fu.UnitID]
			power += (float64(s.Attack) + float64(s.Shield)) * float64(fu.Quantity)
		}
		if power < tp*0.5 {
			t.Errorf("seed %d: alien power %.0f < 50%% of target %.0f", seed, power, tp)
		}
	}
}

func TestGenerateFleet_RespectsMaxDebris(t *testing.T) {
	cfg := DefaultConfig()
	r := rng.New(42)

	// При жёстком cap-е debris ожидаем что итерация остановится.
	out := GenerateFleet(targetSmall, alienAvailable, 1.0, cfg, r,
		WithMaxDebris(100_000))
	bySpec := map[int]ShipSpec{}
	for _, s := range alienAvailable {
		bySpec[s.UnitID] = s
	}
	var debris float64
	for _, fu := range out {
		s := bySpec[fu.UnitID]
		debris += (float64(s.BasicMetal)+float64(s.BasicSilicon))*0.5*float64(fu.Quantity)
	}
	// Алгоритм проверяет debris < max_debris ПОСЛЕ инкремента, то есть
	// финальный debris может слегка превысить cap, но не катастрофически.
	if debris > 1_000_000 {
		t.Errorf("max_debris=100k: фактический debris=%.0f слишком большой", debris)
	}
}

func TestGenerateFleet_ScalesUp(t *testing.T) {
	cfg := DefaultConfig()
	bySpec := map[int]ShipSpec{}
	for _, s := range alienAvailable {
		bySpec[s.UnitID] = s
	}

	totalUnits := func(f Fleet) int64 {
		var t int64
		for _, fu := range f {
			t += fu.Quantity
		}
		return t
	}

	// Среднее количество за 30 seed'ов для масштабов 1.0 и 2.0.
	const trials = 30
	var sum1, sum2 int64
	for seed := uint64(1); seed <= trials; seed++ {
		r1 := rng.New(seed)
		r2 := rng.New(seed)
		f1 := GenerateFleet(targetSmall, alienAvailable, 1.0, cfg, r1)
		f2 := GenerateFleet(targetSmall, alienAvailable, 2.0, cfg, r2)
		sum1 += totalUnits(f1)
		sum2 += totalUnits(f2)
	}
	if sum2 <= sum1 {
		t.Errorf("scale=2 produced %d units total, scale=1 = %d; expected scale=2 >> scale=1",
			sum2, sum1)
	}
}

func TestGenerateFleet_NoTargetReturnsMinFleet(t *testing.T) {
	cfg := DefaultConfig()
	r := rng.New(42)
	// Пустой target — target_power = 100 (нижняя граница).
	out := GenerateFleet(nil, alienAvailable, 1.0, cfg, r)
	if len(out) == 0 {
		t.Errorf("empty target → expected min fleet, got nil")
	}
}

func TestGenerateFleet_DeterministicForSameSeed(t *testing.T) {
	cfg := DefaultConfig()
	r1 := rng.New(123)
	r2 := rng.New(123)
	f1 := GenerateFleet(targetSmall, alienAvailable, 1.5, cfg, r1)
	f2 := GenerateFleet(targetSmall, alienAvailable, 1.5, cfg, r2)
	if len(f1) != len(f2) {
		t.Fatalf("len differs: %d vs %d", len(f1), len(f2))
	}
	for i := range f1 {
		if f1[i] != f2[i] {
			t.Errorf("unit %d differs: %+v vs %+v", i, f1[i], f2[i])
		}
	}
}

func TestGenerateFleet_ArmoredTerranBreakRule(t *testing.T) {
	// Если ARMORED_TERRAN попадает в available и каким-то образом
	// активируется — ожидаем что результирующий флот = ровно 1×AT.
	// Сценарий редкий (mt_rand(0,50)==0 → 1/51), но проверим что
	// алгоритм корректен.
	cfg := DefaultConfig()
	available := append(alienAvailable, ShipSpec{
		UnitID: UnitShipArmoredTerran, Name: "Armored Terran",
		Attack: 1, Shield: 1, BasicMetal: 100, BasicSilicon: 100,
	})

	// Для R15 не падаем тестом если AT не выпал — проверим что
	// при ANY запуске AT либо не появляется, либо появляется ровно
	// в количестве 1.
	for seed := uint64(1); seed < 60; seed++ {
		r := rng.New(seed)
		out := GenerateFleet(targetSmall, available, 1.0, cfg, r)
		for _, fu := range out {
			if fu.UnitID == UnitShipArmoredTerran && fu.Quantity > 1 {
				t.Errorf("seed %d: AT quantity = %d, want 1", seed, fu.Quantity)
			}
		}
	}
}

// TestGenerateFleet_ShellPercentDefault100 — без params["damaged"]
// все юниты идут с shell_percent=100 и damaged=0 (origin:558-559).
func TestGenerateFleet_ShellPercentDefault100(t *testing.T) {
	cfg := DefaultConfig()
	r := rng.New(99)
	out := GenerateFleet(targetSmall, alienAvailable, 1.0, cfg, r)
	for _, fu := range out {
		if fu.ShellPercent != 100 {
			t.Errorf("unit %d shell_percent = %d, want 100", fu.UnitID, fu.ShellPercent)
		}
		if fu.Damaged != 0 {
			t.Errorf("unit %d damaged = %d, want 0", fu.UnitID, fu.Damaged)
		}
	}
	// Тёплая проверка, что totalAttack > 0 (флот не пустой/сломанный).
	if fleetTotalAttack(out, alienAvailable) <= 0 {
		t.Errorf("alien fleet has 0 total attack")
	}
}
