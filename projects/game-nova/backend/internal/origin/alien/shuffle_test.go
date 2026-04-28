package alien

import (
	"testing"

	"oxsar/game-nova/internal/economy"
	"oxsar/game-nova/pkg/rng"
)

// TestShuffleKeyValues_PreservesMultiset проверяет инвариант:
// перетасовка не теряет и не дублирует значения, только меняет
// привязку к ключам.
func TestShuffleKeyValues_PreservesMultiset(t *testing.T) {
	r := rng.New(123)
	tech := TechProfile{
		economy.IDTechGun:    10,
		economy.IDTechShield: 7,
		economy.IDTechShell:  3,
	}
	keys := []int{economy.IDTechGun, economy.IDTechShield, economy.IDTechShell}

	for i := 0; i < 100; i++ {
		out := ShuffleKeyValues(tech, keys, r)
		// мультимножество значений сохранено
		seen := map[int]int{}
		for _, k := range keys {
			seen[out[k]]++
		}
		expected := map[int]int{10: 1, 7: 1, 3: 1}
		for v, n := range expected {
			if seen[v] != n {
				t.Errorf("iter %d: value %d count = %d, want %d (out=%v)",
					i, v, seen[v], n, out)
			}
		}
	}
}

// TestShuffleKeyValues_NotMutatesInput — pure function требование (R15).
func TestShuffleKeyValues_NotMutatesInput(t *testing.T) {
	r := rng.New(7)
	tech := TechProfile{
		economy.IDTechGun:    10,
		economy.IDTechShield: 7,
		economy.IDTechShell:  3,
	}
	snapshot := tech.Clone()
	_ = ShuffleKeyValues(tech, []int{economy.IDTechGun, economy.IDTechShield, economy.IDTechShell}, r)
	for k, v := range snapshot {
		if tech[k] != v {
			t.Errorf("input mutated: key=%d, before=%d, after=%d", k, v, tech[k])
		}
	}
	if len(tech) != len(snapshot) {
		t.Errorf("input length changed: %d → %d", len(snapshot), len(tech))
	}
}

func TestShuffleKeyValues_MissingKeyAsZero(t *testing.T) {
	r := rng.New(7)
	tech := TechProfile{economy.IDTechGun: 10} // только Gun
	keys := []int{economy.IDTechGun, economy.IDTechShield, economy.IDTechShell}
	out := ShuffleKeyValues(tech, keys, r)
	// Должны быть значения {10, 0, 0} в каком-то порядке.
	count := map[int]int{}
	for _, k := range keys {
		count[out[k]]++
	}
	if count[10] != 1 || count[0] != 2 {
		t.Errorf("missing-key shuffle: count=%v, want {10:1, 0:2}", count)
	}
}

func TestShuffleAllAlienTechGroups_AllGroupsTouched(t *testing.T) {
	r := rng.New(7)
	tech := TechProfile{
		economy.IDTechGun:        5,
		economy.IDTechShield:     6,
		economy.IDTechShell:      7,
		economy.IDTechBallistics: 1,
		economy.IDTechMasking:    2,
		economy.IDTechLaser:      3,
		economy.IDTechSilicon:    4,
		economy.IDTechHydrogen:   5,
	}
	out := ShuffleAllAlienTechGroups(tech, r)
	// Тотальная сумма сохраняется внутри каждой группы.
	sum := func(keys ...int) (in, ot int) {
		for _, k := range keys {
			in += tech[k]
			ot += out[k]
		}
		return
	}
	if a, b := sum(economy.IDTechGun, economy.IDTechShield, economy.IDTechShell); a != b {
		t.Errorf("group1 sum changed: %d → %d", a, b)
	}
	if a, b := sum(economy.IDTechBallistics, economy.IDTechMasking); a != b {
		t.Errorf("group2 sum changed: %d → %d", a, b)
	}
	if a, b := sum(economy.IDTechLaser, economy.IDTechSilicon, economy.IDTechHydrogen); a != b {
		t.Errorf("group3 sum changed: %d → %d", a, b)
	}
}

func TestApplyShuffledTechWeakening(t *testing.T) {
	r := rng.New(7)
	// origin: max(0, rand(floor(level*0.7), level<3 ? level : level+1))
	// для level=10 диапазон [7, 11]
	for trial := 0; trial < 200; trial++ {
		out := ApplyShuffledTechWeakening(TechProfile{1: 10}, r)
		if out[1] < 7 || out[1] > 11 {
			t.Fatalf("weakening(10) = %d, want [7,11]", out[1])
		}
	}
	// level=2 — level<3 → diapazon [1, 2]
	for trial := 0; trial < 200; trial++ {
		out := ApplyShuffledTechWeakening(TechProfile{1: 2}, r)
		if out[1] < 1 || out[1] > 2 {
			t.Fatalf("weakening(2) = %d, want [1,2]", out[1])
		}
	}
	// level=0 — всегда 0
	for trial := 0; trial < 5; trial++ {
		out := ApplyShuffledTechWeakening(TechProfile{1: 0}, r)
		if out[1] != 0 {
			t.Fatalf("weakening(0) = %d, want 0", out[1])
		}
	}
}
