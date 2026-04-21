package battle

import (
	"math"
	"testing"
)

// Для тестов используем упрощённые юниты. Формулы M4.0 детерминированы
// при одинаковом входе (нет rand в shootAtSides), поэтому разные seed'ы
// дают одинаковый результат. Это OK — в M4.1 seed начнёт влиять на
// masking/ballistics-roll.

func simpleAttacker(q int64, attack float64, shell float64) Side {
	return Side{
		UserID: "att",
		Units: []Unit{{
			UnitID:   33,
			Quantity: q,
			Front:    0,
			Attack:   [3]float64{attack, 0, 0},
			Shell:    shell,
			Cost:     UnitCost{Metal: 20000, Silicon: 7000, Hydrogen: 2000},
		}},
	}
}

func simpleDefender(q int64, attack float64, shell float64) Side {
	return Side{
		UserID: "def",
		Units: []Unit{{
			UnitID:   31,
			Quantity: q,
			Front:    0,
			Attack:   [3]float64{attack, 0, 0},
			Shell:    shell,
			Cost:     UnitCost{Metal: 3000, Silicon: 1000},
		}},
	}
}

func TestCalculate_AttackersWin(t *testing.T) {
	t.Parallel()
	// 10 сильных атакующих (attack=1000, shell=10000) против 10 слабых
	// защитников (attack=50, shell=1000). В первом раунде атакующие
	// наносят 10*1000 = 10000 урона = уничтожают всех защитников.
	in := Input{
		Seed:      42,
		Attackers: []Side{simpleAttacker(10, 1000, 10000)},
		Defenders: []Side{simpleDefender(10, 50, 1000)},
	}
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Winner != "attackers" {
		t.Fatalf("expected attackers to win, got %q", rep.Winner)
	}
	if rep.Defenders[0].Units[0].QuantityEnd != 0 {
		t.Fatalf("expected defenders wiped, got %d remaining",
			rep.Defenders[0].Units[0].QuantityEnd)
	}
	if rep.Rounds != 1 {
		t.Fatalf("expected 1 round (early exit), got %d", rep.Rounds)
	}
}

func TestCalculate_DefendersWin(t *testing.T) {
	t.Parallel()
	// Ровно наоборот: 10 слабых атакующих, 10 сильных защитников.
	in := Input{
		Seed:      42,
		Attackers: []Side{simpleAttacker(10, 50, 1000)},
		Defenders: []Side{simpleDefender(10, 1000, 10000)},
	}
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Winner != "defenders" {
		t.Fatalf("expected defenders, got %q", rep.Winner)
	}
	if rep.Attackers[0].Units[0].QuantityEnd != 0 {
		t.Fatalf("expected attackers wiped, got %d", rep.Attackers[0].Units[0].QuantityEnd)
	}
}

func TestCalculate_Draw(t *testing.T) {
	t.Parallel()
	// Равные силы 100 vs 100, атака 100, shell 100. В одном раунде
	// обе стороны наносят по 100*100 = 10000 урона = уничтожают друг
	// друга (total shell = 100*100 = 10000). Итог — ничья.
	in := Input{
		Seed:      42,
		Attackers: []Side{simpleAttacker(100, 100, 100)},
		Defenders: []Side{simpleDefender(100, 100, 100)},
	}
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rep.Winner != "draw" {
		t.Fatalf("expected draw, got %q", rep.Winner)
	}
}

func TestCalculate_Deterministic(t *testing.T) {
	t.Parallel()
	build := func() Input {
		return Input{
			Seed:      12345,
			Attackers: []Side{simpleAttacker(50, 300, 5000)},
			Defenders: []Side{simpleDefender(50, 50, 1000)},
		}
	}
	a, err1 := Calculate(build())
	b, err2 := Calculate(build())
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v / %v", err1, err2)
	}
	if a.Winner != b.Winner ||
		a.Rounds != b.Rounds ||
		len(a.RoundsTrace) != len(b.RoundsTrace) {
		t.Fatalf("non-deterministic: %+v vs %+v", a, b)
	}
	for i := range a.Attackers[0].Units {
		if a.Attackers[0].Units[i].QuantityEnd != b.Attackers[0].Units[i].QuantityEnd {
			t.Fatalf("attacker unit %d end differs: %d vs %d",
				i, a.Attackers[0].Units[i].QuantityEnd, b.Attackers[0].Units[i].QuantityEnd)
		}
	}
}

func TestCalculate_LostResources(t *testing.T) {
	t.Parallel()
	// Убедимся, что lost metal/silicon считается корректно:
	// 10 уничтоженных атакующих × cost (20000/7000/2000) = 200000/70000/20000.
	// Подбираем сценарий, где ровно 10 атакующих умирают: 10 vs 10,
	// атака/shell равны.
	in := Input{
		Seed:      7,
		Attackers: []Side{simpleAttacker(10, 100, 100)},
		Defenders: []Side{simpleDefender(10, 100, 100)},
	}
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := rep.Attackers[0].LostMetal
	want := int64(10) * 20000
	if got != want {
		t.Fatalf("LostMetal: got %d, want %d", got, want)
	}
}

func TestCalculate_InvalidInput(t *testing.T) {
	t.Parallel()
	in := Input{Seed: 1, Defenders: []Side{simpleDefender(1, 10, 10)}}
	if _, err := Calculate(in); err == nil {
		t.Fatalf("expected error on missing attackers")
	}
}

// --- M4.1 shield tests ---

// shieldedDefender — защитник с полностью заполненным щитом и корпусом.
// Использует primary-канал 0 (normal).
func shieldedDefender(q int64, attack, shield, shell float64) Side {
	return Side{
		UserID: "def",
		Units: []Unit{{
			UnitID:   31,
			Quantity: q,
			Front:    0,
			Attack:   [3]float64{attack, 0, 0},
			Shield:   [3]float64{shield, 0, 0},
			Shell:    shell,
			Cost:     UnitCost{Metal: 3000, Silicon: 1000},
		}},
	}
}

func TestCalculate_ShieldIgnoreThreshold(t *testing.T) {
	t.Parallel()
	// 10 атакующих с attack=50 стреляют по 10 защитников с
	// shield=10000 (огромный щит). ignoreThreshold = shield/100 = 100,
	// а attack=50 < 100 → выстрелы полностью абсорбируются щитом,
	// корпус нетронут. За 6 раундов никто не гибнет.
	in := Input{
		Seed:      1,
		Attackers: []Side{simpleAttacker(10, 50, 1000)},
		Defenders: []Side{shieldedDefender(10, 0, 10000, 1000)},
	}
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := rep.Defenders[0].Units[0].QuantityEnd; got != 10 {
		t.Fatalf("defender quantity: got %d, want 10 (shield should block)", got)
	}
	if got := rep.Attackers[0].Units[0].QuantityEnd; got != 10 {
		t.Fatalf("attacker quantity: got %d, want 10 (def has attack=0)", got)
	}
}

func TestCalculate_ShieldPierced(t *testing.T) {
	t.Parallel()
	// Shield небольшой и быстро падает; корпус тоже лёгкий.
	// 10 × attack=500 за один раунд дают пул 5000 на общую защиту
	// 10×100 (щит) + 10×300 (shell) = 4000 → все цели гибнут в 1 раунде.
	in := Input{
		Seed:      1,
		Attackers: []Side{simpleAttacker(10, 500, 100000)},
		Defenders: []Side{shieldedDefender(10, 0, 100, 300)},
	}
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := rep.Defenders[0].Units[0].QuantityEnd; got != 0 {
		t.Fatalf("defender wiped expected, got %d", got)
	}
	if rep.Winner != "attackers" {
		t.Fatalf("expected attackers win, got %q", rep.Winner)
	}
	if rep.Rounds != 1 {
		t.Fatalf("expected rounds=1, got %d", rep.Rounds)
	}
}

func TestCalculate_ShieldRegensBetweenRounds(t *testing.T) {
	t.Parallel()
	// В каждом раунде атакующие наносят ровно столько, чтобы снять
	// щит (без урона корпусу). Shield regen должен возвращать щит
	// на следующий раунд. Итог — бой уходит в draw за Rounds раундов,
	// никто не гибнет.
	//
	// attack=1000, shots=10 → pool=10000 на раунд.
	// shield=1000, quantity=10 → totalShield=10000.
	// Точное равенство → весь пул тратится на щит, shell не страдает.
	// regen после раунда возвращает totalShield=10000.
	in := Input{
		Seed:      1,
		Rounds:    3,
		Attackers: []Side{simpleAttacker(10, 1000, 1000000)},
		Defenders: []Side{shieldedDefender(10, 0, 1000, 5000)},
	}
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := rep.Defenders[0].Units[0].QuantityEnd; got != 10 {
		t.Fatalf("defender should survive regen: got %d, want 10", got)
	}
	if rep.Rounds != 3 {
		t.Fatalf("expected full 3 rounds, got %d", rep.Rounds)
	}
	if rep.Winner != "draw" {
		t.Fatalf("expected draw, got %q", rep.Winner)
	}
}

// --- M4.2 rapidfire / masking tests ---

// TestRapidfire_TriplesDamage: атакующие с rf=3 наносят в 3 раза
// больше урона, чем без rf (при прочих равных).
func TestRapidfire_TriplesDamage(t *testing.T) {
	t.Parallel()
	noRF := Input{
		Seed:      1,
		Rounds:    1,
		Attackers: []Side{simpleAttacker(10, 100, 1000000)},
		Defenders: []Side{simpleDefender(10, 0, 1000)},
	}
	withRF := noRF
	withRF.Rapidfire = map[int]map[int]int{
		33: {31: 3}, // shooter 33 has rf=3 on target 31
	}

	rep1, err := Calculate(noRF)
	if err != nil {
		t.Fatalf("noRF: %v", err)
	}
	rep2, err := Calculate(withRF)
	if err != nil {
		t.Fatalf("withRF: %v", err)
	}
	killsNoRF := int64(10) - rep1.Defenders[0].Units[0].QuantityEnd
	killsRF := int64(10) - rep2.Defenders[0].Units[0].QuantityEnd
	if killsRF != 3*killsNoRF {
		t.Fatalf("rapidfire: kills no-rf=%d, with-rf=%d (expected 3x)", killsNoRF, killsRF)
	}
}

// TestMasking_ReducesDamage: если masking цели > ballistics стрелка,
// часть выстрелов промахивается → урон падает.
//
// Масштабы подобраны так, чтобы разница в shots давала разное число
// kills (иначе floor() поглощает эффект): attack=500, 10 shooters,
// base pool=5000 → 5 kills. С masking=10: factor=0.667, shots=10-6=4,
// pool=2000, kills=floor((10000-2000)/1000) → но без ablation даже
// частичный урон убивает целого юнита: kills = 10 - floor(8000/1000) = 2.
func TestMasking_ReducesDamage(t *testing.T) {
	t.Parallel()
	base := Input{
		Seed:      1,
		Rounds:    1,
		Attackers: []Side{simpleAttacker(10, 500, 1000000)},
		Defenders: []Side{simpleDefender(10, 0, 1000)},
	}
	rep1, err := Calculate(base)
	if err != nil {
		t.Fatalf("base: %v", err)
	}

	masked := base
	masked.Defenders = []Side{{
		UserID: "def",
		Tech:   Tech{Masking: 10},
		Units:  base.Defenders[0].Units,
	}}
	rep2, err := Calculate(masked)
	if err != nil {
		t.Fatalf("masked: %v", err)
	}
	kills1 := int64(10) - rep1.Defenders[0].Units[0].QuantityEnd
	kills2 := int64(10) - rep2.Defenders[0].Units[0].QuantityEnd
	if kills2 >= kills1 {
		t.Fatalf("masking should reduce damage: kills base=%d, masked=%d", kills1, kills2)
	}
}

// TestBallistics_OffsetsMasking: если ballistics == masking, эффекта
// нет — результат совпадает с отсутствием masking.
func TestBallistics_OffsetsMasking(t *testing.T) {
	t.Parallel()
	base := Input{
		Seed:      1,
		Rounds:    1,
		Attackers: []Side{simpleAttacker(10, 100, 1000000)},
		Defenders: []Side{simpleDefender(10, 0, 1000)},
	}
	rep1, _ := Calculate(base)

	balanced := base
	balanced.Attackers = []Side{{
		UserID: "att",
		Tech:   Tech{Ballistics: 10},
		Units:  base.Attackers[0].Units,
	}}
	balanced.Defenders = []Side{{
		UserID: "def",
		Tech:   Tech{Masking: 10},
		Units:  base.Defenders[0].Units,
	}}
	rep2, _ := Calculate(balanced)

	k1 := int64(10) - rep1.Defenders[0].Units[0].QuantityEnd
	k2 := int64(10) - rep2.Defenders[0].Units[0].QuantityEnd
	if k1 != k2 {
		t.Fatalf("ballistics should offset masking: k1=%d, k2=%d", k1, k2)
	}
}

// --- M4.3 ablation tests ---

// TestAblation_PartialHitLeavesDamagedUnit: частичный урон (меньше
// unit.Shell) оставляет один damaged-юнит, остальные здоровы.
func TestAblation_PartialHitLeavesDamagedUnit(t *testing.T) {
	t.Parallel()
	// 10 × attack=100 → pool=1000. 10 защитников с shell=1000,
	// totalShell=10000. После удара turnShell=9000 →
	// fullRem=9, remainder=0 → точная граница, kills=1, без damaged.
	// Проверим с non-aligned числами.
	in := Input{
		Seed:      1,
		Rounds:    1,
		Attackers: []Side{simpleAttacker(10, 50, 1000000)},
		Defenders: []Side{simpleDefender(10, 0, 1000)},
	}
	// pool = 500. totalShell 10000 → 9500. fullRem=9, remainder=500.
	// → quantity=10, damaged=1, shellPercent=50.
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	u := rep.Defenders[0].Units[0]
	if u.QuantityEnd != 10 {
		t.Fatalf("quantity: got %d, want 10 (none fully dead)", u.QuantityEnd)
	}
	if u.DamagedEnd != 1 {
		t.Fatalf("damaged: got %d, want 1", u.DamagedEnd)
	}
	if u.ShellPercentEnd <= 0 || u.ShellPercentEnd >= 100 {
		t.Fatalf("shellPercent: got %v, want in (0, 100)", u.ShellPercentEnd)
	}
	if math.Abs(u.ShellPercentEnd-50) > 1e-6 {
		t.Fatalf("shellPercent: got %v, want 50", u.ShellPercentEnd)
	}
}

// TestAblation_DamagedCarriesOverAndDies: damaged юнит переносит
// пониженный shell на следующий раунд. Небольшой pool, который ранее
// только ранил, во втором раунде уже добивает.
func TestAblation_DamagedCarriesOverAndDies(t *testing.T) {
	t.Parallel()
	// 1 защитник, shell=1000. pool=500/раунд.
	// Round 1: turnShell 1000 → 500 → quantity=1, damaged=1, pct=50.
	// Round 2 regen не влияет на shell (regen только у щитов),
	//          turnShell=500 → -500 = 0 → defender dead.
	in := Input{
		Seed:      1,
		Rounds:    2,
		Attackers: []Side{simpleAttacker(10, 50, 1000000)},
		Defenders: []Side{simpleDefender(1, 0, 1000)},
	}
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	u := rep.Defenders[0].Units[0]
	if u.QuantityEnd != 0 {
		t.Fatalf("defender should be dead after 2 rounds of 50%% damage: got quantity=%d, damaged=%d, shellPct=%v",
			u.QuantityEnd, u.DamagedEnd, u.ShellPercentEnd)
	}
	if rep.Winner != "attackers" {
		t.Fatalf("expected attackers win, got %q", rep.Winner)
	}
	if rep.Rounds != 2 {
		t.Fatalf("expected 2 rounds, got %d", rep.Rounds)
	}
}
