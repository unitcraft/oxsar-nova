package miner

import "testing"

func TestNeedPoints(t *testing.T) {
	t.Parallel()
	cases := []struct {
		level int
		want  int64
	}{
		{-1, 100},
		{0, 100},
		{1, 200},                     // round(pow(1.5, 0) * 200) = 200
		{2, 300},                     // round(pow(1.5, 1) * 200) = 300
		{3, 450},                     // round(pow(1.5, 2) * 200) = 450
		{4, 675},                     // round(pow(1.5, 3) * 200) = 675
		{5, 1013},                    // round(pow(1.5, 4) * 200) ≈ 1012.5 → 1013
		{10, int64(7689)},            // round(pow(1.5, 9) * 200) ≈ 7688.67 → 7689
	}
	for _, c := range cases {
		if got := NeedPoints(c.level); got != c.want {
			t.Errorf("NeedPoints(%d) = %d, want %d", c.level, got, c.want)
		}
	}
}

func TestLevelUp_NoChange(t *testing.T) {
	t.Parallel()
	r := LevelUp(0, 0, 0)
	if r.NewLevel != 0 || r.NewPoints != 0 || r.CreditsAwarded != 0 || r.LevelUps != 0 {
		t.Errorf("expected zero result, got %+v", r)
	}
}

func TestLevelUp_BelowThreshold(t *testing.T) {
	t.Parallel()
	r := LevelUp(0, 0, 50)
	if r.NewLevel != 0 || r.NewPoints != 50 {
		t.Errorf("level should not change, got %+v", r)
	}
}

func TestLevelUp_FirstThreshold(t *testing.T) {
	t.Parallel()
	// 100 points → level 1, остаток 0, награда 10.
	r := LevelUp(0, 0, 100)
	if r.NewLevel != 1 || r.NewPoints != 0 || r.CreditsAwarded != 10 || r.LevelUps != 1 {
		t.Errorf("expected (1,0,10,1), got %+v", r)
	}
}

func TestLevelUp_MultipleAtOnce(t *testing.T) {
	t.Parallel()
	// 0 → 1 (100, награда 10) → 2 (200, награда 25) → 3 (300, награда 50)
	// → 4 (450, награда 75). Если добавить ровно 100+200+300+450 = 1050,
	// получим уровень 4, остаток 0, суммарная награда 10+25+50+75=160.
	r := LevelUp(0, 0, 1050)
	if r.NewLevel != 4 {
		t.Errorf("expected level 4, got %d", r.NewLevel)
	}
	if r.NewPoints != 0 {
		t.Errorf("expected 0 points remaining, got %d", r.NewPoints)
	}
	if r.CreditsAwarded != 160 {
		t.Errorf("expected 160 credits, got %d", r.CreditsAwarded)
	}
	if r.LevelUps != 4 {
		t.Errorf("expected 4 level-ups, got %d", r.LevelUps)
	}
}

func TestLevelUp_PartialResidue(t *testing.T) {
	t.Parallel()
	// От 0/0 +250: уровень 0 → 1 (тратим 100), осталось 150.
	// уровень 1 → 2? нужно 200, у нас 150 — стоп. Финал level=1, pts=150.
	r := LevelUp(0, 0, 250)
	if r.NewLevel != 1 || r.NewPoints != 150 {
		t.Errorf("expected (1, 150), got level=%d points=%d", r.NewLevel, r.NewPoints)
	}
}

func TestLevelUp_AboveLevel13RewardCap(t *testing.T) {
	t.Parallel()
	// Стартуем с level=13, points=0. Добавляем достаточно для +1.
	// NeedPoints(13) = round(pow(1.5,12)*200) = round(25940.34) ≈ 25940.
	need13 := NeedPoints(13)
	r := LevelUp(13, 0, need13)
	if r.NewLevel != 14 {
		t.Errorf("expected level 14, got %d", r.NewLevel)
	}
	if r.CreditsAwarded != 300 {
		t.Errorf("expected 300 credits (cap), got %d", r.CreditsAwarded)
	}
}

func TestLevelUp_NegativeInputsClampedToZero(t *testing.T) {
	t.Parallel()
	r := LevelUp(-5, -100, -50)
	if r.NewLevel != 0 || r.NewPoints != 0 || r.CreditsAwarded != 0 {
		t.Errorf("negative inputs must clamp to 0, got %+v", r)
	}
}

// Свойство: LevelUp(L, P, A) == LevelUp(LevelUp(L, P, A/2).{L,P}, 0, A/2)
// (т.е. порядок не важен). Проверяем для нескольких pivot-точек.
func TestLevelUp_Associative(t *testing.T) {
	t.Parallel()
	cases := [][2]int64{
		{0, 1500},
		{0, 5000},
		{0, 10000},
	}
	for _, c := range cases {
		startPts := c[0]
		add := c[1]
		full := LevelUp(0, startPts, add)
		half1 := LevelUp(0, startPts, add/2)
		half2 := LevelUp(half1.NewLevel, half1.NewPoints, add-add/2)
		if full.NewLevel != half2.NewLevel || full.NewPoints != half2.NewPoints {
			t.Errorf("not associative for add=%d: full=%+v, half=%+v",
				add, full, half2)
		}
		if full.CreditsAwarded != half1.CreditsAwarded+half2.CreditsAwarded {
			t.Errorf("credits not associative for add=%d", add)
		}
	}
}
