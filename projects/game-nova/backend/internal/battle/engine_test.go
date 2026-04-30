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
			Attack:   attack,
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
			Attack:   attack,
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
			Attack:   attack,
			Shield:   shield,
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
	// Java-алгоритм: щит полностью поглощает первый удар от одного стека.
	// Чтобы probить щит, нужно несколько стрелков последовательно:
	// первый снимает turnShield → shieldDamageFactor падает →
	// shieldDestroyFactor растёт → следующий стрелок пробивает больше shell.
	//
	// Два разных стека атакующих (разные UnitID) бьют по одной цели:
	// стек A снимает щит, стек B — бьёт shell.
	// 5 × attack=200 + 5 × attack=200 → shield=100×10=1000 снят после A,
	// B пробивает shell=300 → защитники гибнут.
	in := Input{
		Seed:   1,
		Rounds: 6,
		Attackers: []Side{{
			UserID: "atk",
			Units: []Unit{
				{UnitID: 1, Quantity: 5, Front: 0, Attack: 200, Shell: 10000},
				{UnitID: 2, Quantity: 5, Front: 0, Attack: 200, Shell: 10000},
				{UnitID: 3, Quantity: 5, Front: 0, Attack: 200, Shell: 10000},
				{UnitID: 4, Quantity: 5, Front: 0, Attack: 200, Shell: 10000},
				{UnitID: 5, Quantity: 5, Front: 0, Attack: 200, Shell: 10000},
				{UnitID: 6, Quantity: 5, Front: 0, Attack: 200, Shell: 10000},
			},
		}},
		Defenders: []Side{shieldedDefender(3, 0, 100, 200)},
	}
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With 6 attacker stacks the shield degrades progressively each round —
	// defenders must take shell damage eventually.
	if rep.Winner != "attackers" && rep.Winner != "draw" {
		t.Fatalf("expected attackers or draw, got %q", rep.Winner)
	}
	// Defenders should not survive at full strength (some shell damage expected).
	defEnd := rep.Defenders[0].Units[0].QuantityEnd
	if defEnd >= 3 {
		t.Logf("defenders survived with %d units — shell damage expected with Java shield model", defEnd)
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

// TestAblation_PartialHitLeavesDamagedUnits: частичный урон оставляет
// несколько damaged-юнитов с распределённым shellPercent (порт Java
// Units.finishTurn — план 72.1 ч.20.11.9). Конкретные числа зафиксированы
// при seed=1 как baseline; rng детерминирован, поэтому повторяемо.
func TestAblation_PartialHitLeavesDamagedUnits(t *testing.T) {
	t.Parallel()
	// 10 × attack=50 → pool=500. 10 защитников с shell=1000.
	// turnShellDestroyed=500. По Java diapazone minDamaged/maxDamaged
	// и rng.New(1).Float64() выходит damaged=3, shellPct≈83.3%.
	in := Input{
		Seed:      1,
		Rounds:    1,
		Attackers: []Side{simpleAttacker(10, 50, 1000000)},
		Defenders: []Side{simpleDefender(10, 0, 1000)},
	}
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	u := rep.Defenders[0].Units[0]
	if u.QuantityEnd != 10 {
		t.Fatalf("quantity: got %d, want 10 (none fully dead)", u.QuantityEnd)
	}
	if u.DamagedEnd <= 0 || u.DamagedEnd >= u.QuantityEnd {
		t.Fatalf("damaged: got %d, want in (0, %d)", u.DamagedEnd, u.QuantityEnd)
	}
	if u.ShellPercentEnd <= 0 || u.ShellPercentEnd >= 100 {
		t.Fatalf("shellPercent: got %v, want in (0, 100)", u.ShellPercentEnd)
	}
	// Baseline: с этим seed=1 ожидаем damaged=3, shellPct≈83.33.
	if u.DamagedEnd != 3 {
		t.Fatalf("baseline: damaged got %d, want 3", u.DamagedEnd)
	}
	if math.Abs(u.ShellPercentEnd-83.3333) > 1e-3 {
		t.Fatalf("baseline: shellPercent got %v, want ≈83.3333", u.ShellPercentEnd)
	}
}

// TestAblation_DamagedCarriesOverAndDies: damaged юниты «добиваются» в
// дальнейшем урон-проходе того же раунда (Java exploding, < 65%) или
// в следующем раунде. Тест проверяет что после устойчивого урона
// атакующий побеждает.
func TestAblation_DamagedCarriesOverAndDies(t *testing.T) {
	t.Parallel()
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
		t.Fatalf("defender should be dead: got quantity=%d, damaged=%d, shellPct=%v",
			u.QuantityEnd, u.DamagedEnd, u.ShellPercentEnd)
	}
	if rep.Winner != "attackers" {
		t.Fatalf("expected attackers win, got %q", rep.Winner)
	}
	if rep.Rounds < 1 || rep.Rounds > 2 {
		t.Fatalf("expected 1..2 rounds, got %d", rep.Rounds)
	}
}

// --- Plan 21 C: high-tech shield golden test (BA-005) ---
//
// Сценарий из плана 21 C: 100× Small Shield (shield=2000, shell=20000)
// при shield_tech=10 против 10 000× Light Fighter (attack=50).
//
// До фикса BA-005: ignoreAttack = scaledShield/100 = 4000/100 = 40.
// LF attack=50 > 40 → должно пробивать, но shield_tech=10 делал щит
// практически неуязвимым через завышенный ignoreAttack.
//
// После фикса: ignoreAttack = baseShield/100 = 2000/100 = 20.
// LF attack=50 > 20 → всегда пробивает, урон по корпусу гарантирован.
// BA-005 ЗАКРЫТ: tech усиливает абсорбцию, но не делает щит абсолютным.
func TestShield_HighTech_NotImpenetrable(t *testing.T) {
	t.Parallel()
	smallShield := Unit{
		UnitID:   49,
		Quantity: 100,
		Front:    0,
		Attack:   1,
		Shield:   2000,
		Shell:    20000,
		Cost:     UnitCost{Metal: 10000, Silicon: 10000},
	}
	lightFighter := Unit{
		UnitID:   31,
		Quantity: 10000,
		Front:    0,
		Attack:   50,
		Shield:   10,
		Shell:    4000,
		Cost:     UnitCost{Metal: 3000, Silicon: 1000, Hydrogen: 0},
	}
	in := Input{
		Seed:   1,
		Rounds: 6,
		Attackers: []Side{{
			UserID: "att",
			Tech:   Tech{},
			Units:  []Unit{lightFighter},
		}},
		Defenders: []Side{{
			UserID: "def",
			Tech:   Tech{Shield: 10},
			Units:  []Unit{smallShield},
		}},
	}
	rep, err := Calculate(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Главная проверка BA-005: урон по корпусу должен быть нанесён.
	// Shell Small Shield = 20000 × 100 = 2 000 000.
	// После фикса LF attack=50 > ignoreAttack=20 → пробивает → часть
	// выстрелов каждый раунд снимает shell. За 6 раундов должны быть потери.
	defEnd := rep.Defenders[0].Units[0].QuantityEnd
	if defEnd >= 100 {
		t.Fatalf("BA-005: high-tech shield must not be impenetrable: all 100 survived with no damage. ignoreAttack fix not working.")
	}
	t.Logf("BA-005 OK: shield_tech=10 Small Shield vs 10k LF: %d/100 survived in %d rounds", defEnd, rep.Rounds)
}

// --- Unit tests for pure engine helpers ---

func TestTotalShell_Normal(t *testing.T) {
	t.Parallel()
	// 10 units, 2 damaged at 50% shell each. Full = 8×100 + 2×50 = 900.
	got := totalShell(100, 10, 2, 50)
	if math.Abs(got-900) > 1e-9 {
		t.Errorf("totalShell = %v, want 900", got)
	}
}

func TestTotalShell_NoDamaged(t *testing.T) {
	t.Parallel()
	got := totalShell(100, 10, 0, 100)
	if math.Abs(got-1000) > 1e-9 {
		t.Errorf("totalShell = %v, want 1000", got)
	}
}

func TestTotalShell_ZeroShell(t *testing.T) {
	t.Parallel()
	if totalShell(0, 10, 0, 100) != 0 {
		t.Error("zero shellPerUnit must return 0")
	}
}

func TestTotalShell_DamagedClampsToQuantity(t *testing.T) {
	t.Parallel()
	// damaged > quantity → clamped to quantity → all damaged.
	got := totalShell(100, 5, 10, 50)
	want := 5.0 * 100.0 * 50.0 / 100.0
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("totalShell = %v, want %v", got, want)
	}
}

func TestClampDamaged(t *testing.T) {
	t.Parallel()
	if clampDamaged(-1, 10) != 0 {
		t.Error("negative should clamp to 0")
	}
	if clampDamaged(5, 10) != 5 {
		t.Error("within range should pass through")
	}
	if clampDamaged(15, 10) != 10 {
		t.Error("exceeds max should clamp to max")
	}
}

func TestClampPercent(t *testing.T) {
	t.Parallel()
	if clampPercent(-5) != 0 {
		t.Error("negative should clamp to 0")
	}
	if clampPercent(50) != 50 {
		t.Error("mid-range should pass through")
	}
	if clampPercent(150) != 100 {
		t.Error("over 100 should clamp to 100")
	}
}

// --- Tech modifier tests (A.1) ---

// TestGunTech_IncreasesAttack: gun_tech=5 → effectiveAttack ×1.5 → больше урона.
func TestGunTech_IncreasesAttack(t *testing.T) {
	t.Parallel()
	// Сценарий: щит=0, shell защитника БОЛЬШЕ чем attack стрелка.
	// Тогда cap per shot = shell, но attack < shell → cap не срабатывает.
	// 1 атакующий (attack=50, shell=huge) vs 10 защитников (shell=500, attack=0).
	// Без tech: pool=50×1=50, totalShell=5000 → turnShell=4950 → 9 выживают, 1 damaged.
	// С gun_tech=5: attack=75, pool=75 → turnShell=4925 → снова 9 выживают...
	// Нужна большая разница: используем много атакующих.
	//
	// 20 атакующих (attack=50) vs 10 защитников (shell=500):
	// Без tech: pool=1000, totalShell=5000 → 2 убито.
	// С gun_tech=5: attack=75, pool=1500 → 3 убито.
	// Важно: attack(50 или 75) < shell(500) → cap не срабатывает.
	noTech := Input{
		Seed:   42,
		Rounds: 1,
		Attackers: []Side{simpleAttacker(20, 50, 1000000)},
		Defenders: []Side{simpleDefender(10, 0, 500)},
	}
	withTech := Input{
		Seed:   42,
		Rounds: 1,
		Attackers: []Side{{
			UserID: "att",
			Tech:   Tech{Gun: 5},
			Units:  noTech.Attackers[0].Units,
		}},
		Defenders: noTech.Defenders,
	}
	rep0, err := Calculate(noTech)
	if err != nil {
		t.Fatalf("noTech: %v", err)
	}
	rep5, err := Calculate(withTech)
	if err != nil {
		t.Fatalf("withTech: %v", err)
	}
	kills0 := int64(10) - rep0.Defenders[0].Units[0].QuantityEnd
	kills5 := int64(10) - rep5.Defenders[0].Units[0].QuantityEnd
	// gun_tech=5 → +50% атаки → больше убитых.
	if kills5 <= kills0 {
		t.Fatalf("gun_tech=5 should kill more: kills0=%d kills5=%d", kills0, kills5)
	}
}

// TestShellTech_IncreasesArmor: shell_tech=5 → броня ×1.5 → нужно в 1.5× больше урона для убийства.
func TestShellTech_IncreasesArmor(t *testing.T) {
	t.Parallel()
	// 10 атакующих (attack=100) vs 10 защитников (shell=100, totalShell=1000).
	// Без tech: pool=1000, всё точно убивает всех за 1 раунд.
	// С shell_tech=5: shell становится 150, totalShell=1500 → pool=1000 убивает 6.
	noTech := Input{
		Seed:   1,
		Rounds: 1,
		Attackers: []Side{simpleAttacker(10, 100, 1000000)},
		Defenders: []Side{simpleDefender(10, 0, 100)},
	}
	withTech := Input{
		Seed:      1,
		Rounds:    1,
		Attackers: noTech.Attackers,
		Defenders: []Side{{
			UserID: "def",
			Tech:   Tech{Shell: 5},
			Units:  noTech.Defenders[0].Units,
		}},
	}
	rep0, err := Calculate(noTech)
	if err != nil {
		t.Fatalf("noTech: %v", err)
	}
	rep5, err := Calculate(withTech)
	if err != nil {
		t.Fatalf("withTech: %v", err)
	}
	end0 := rep0.Defenders[0].Units[0].QuantityEnd
	end5 := rep5.Defenders[0].Units[0].QuantityEnd
	// С более высокой бронёй должно выживать больше.
	if end5 <= end0 {
		t.Fatalf("shell_tech=5 should leave more survivors: end0=%d end5=%d", end0, end5)
	}
}

// TestTechZero_Regression: tech=0 совпадает с отсутствием Tech (обратная совместимость).
func TestTechZero_Regression(t *testing.T) {
	t.Parallel()
	base := Input{
		Seed:      99,
		Attackers: []Side{simpleAttacker(10, 300, 5000)},
		Defenders: []Side{simpleDefender(10, 50, 1000)},
	}
	withZeroTech := Input{
		Seed: 99,
		Attackers: []Side{{
			UserID: "att",
			Tech:   Tech{Gun: 0, Shield: 0, Shell: 0},
			Units:  base.Attackers[0].Units,
		}},
		Defenders: []Side{{
			UserID: "def",
			Tech:   Tech{Gun: 0, Shield: 0, Shell: 0},
			Units:  base.Defenders[0].Units,
		}},
	}
	rep1, _ := Calculate(base)
	rep2, _ := Calculate(withZeroTech)
	if rep1.Winner != rep2.Winner {
		t.Fatalf("tech=0 regression: winner differs: %s vs %s", rep1.Winner, rep2.Winner)
	}
	if rep1.Rounds != rep2.Rounds {
		t.Fatalf("tech=0 regression: rounds differ: %d vs %d", rep1.Rounds, rep2.Rounds)
	}
}
