package rocket

import (
	"testing"
)

func TestApplyRocketDamage_NoStacks(t *testing.T) {
	t.Parallel()
	losses := applyRocketDamage(10, nil, 0)
	if len(losses) != 0 {
		t.Fatalf("expected no losses, got %v", losses)
	}
}

func TestApplyRocketDamage_ZeroRockets(t *testing.T) {
	t.Parallel()
	stacks := []defStack{{UnitID: 43, Count: 100, Shell: 2000}}
	losses := applyRocketDamage(0, stacks, 0)
	if len(losses) != 0 {
		t.Fatalf("expected no losses, got %v", losses)
	}
}

func TestApplyRocketDamage_Proportional(t *testing.T) {
	t.Parallel()
	// 1 ракета = 12000 урона. Два одинаковых стека — каждый должен получить ~50%.
	stacks := []defStack{
		{UnitID: 43, Count: 100, Shell: 2000},
		{UnitID: 44, Count: 100, Shell: 2000},
	}
	losses := applyRocketDamage(1, stacks, 0)
	if len(losses) != 2 {
		t.Fatalf("expected 2 loss entries, got %d", len(losses))
	}
	// Каждый стек получает 6000 урона / 2000 shell = 3 юнита.
	for _, l := range losses {
		if l.Lost != 3 {
			t.Errorf("unit %d: expected 3 killed, got %d", l.UnitID, l.Lost)
		}
	}
}

func TestApplyRocketDamage_PriorityTarget(t *testing.T) {
	t.Parallel()
	// 1 ракета = 12000 урона. Приоритет — unit 44 (shell=2000, count=10 → макс убить 10).
	// 12000 / 2000 = 6 убито с приоритетного; остаток = 12000 - 6×2000 = 0.
	stacks := []defStack{
		{UnitID: 43, Count: 100, Shell: 2000},
		{UnitID: 44, Count: 10, Shell: 2000},
	}
	losses := applyRocketDamage(1, stacks, 44)

	var lost43, lost44 int64
	for _, l := range losses {
		switch l.UnitID {
		case 43:
			lost43 = l.Lost
		case 44:
			lost44 = l.Lost
		}
	}
	if lost44 != 6 {
		t.Errorf("priority unit 44: expected 6 killed, got %d", lost44)
	}
	if lost43 != 0 {
		t.Errorf("unit 43 should not be hit (no overflow), got %d", lost43)
	}
}

func TestApplyRocketDamage_PriorityOverflow(t *testing.T) {
	t.Parallel()
	// Priority unit имеет только 2 единицы (shell=2000, урон 4000).
	// 1 ракета = 12000. Убили 2 из приоритетного (4000), остаток 8000 → ещё 4 из unit 43.
	stacks := []defStack{
		{UnitID: 43, Count: 100, Shell: 2000},
		{UnitID: 44, Count: 2, Shell: 2000},
	}
	losses := applyRocketDamage(1, stacks, 44)

	var lost43, lost44 int64
	for _, l := range losses {
		switch l.UnitID {
		case 43:
			lost43 = l.Lost
		case 44:
			lost44 = l.Lost
		}
	}
	if lost44 != 2 {
		t.Errorf("priority unit 44: expected 2 killed, got %d", lost44)
	}
	if lost43 != 4 {
		t.Errorf("overflow to unit 43: expected 4 killed, got %d", lost43)
	}
}

func TestApplyRocketDamage_CannotKillMoreThanCount(t *testing.T) {
	t.Parallel()
	// 100 ракет = 1,200,000 урона. Только 5 юнитов с shell=2000 — не более 5 убито.
	stacks := []defStack{{UnitID: 43, Count: 5, Shell: 2000}}
	losses := applyRocketDamage(100, stacks, 0)
	if len(losses) != 1 || losses[0].Lost > 5 {
		t.Fatalf("expected at most 5 killed, got %v", losses)
	}
}
