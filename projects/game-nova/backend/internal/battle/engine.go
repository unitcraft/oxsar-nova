package battle

import (
	"errors"
	"math"

	"oxsar/game-nova/pkg/rng"
)

// ErrInvalidInput — структурно некорректный вход (нет сторон, нет unit-ов).
var ErrInvalidInput = errors.New("battle: invalid input")

// Calculate — единственная публичная функция пакета. Чистая,
// детерминированная при фиксированном seed.
//
// План 72.1 ч.20.11.4: возвращает полный RoundTrace с tech-power,
// per-unit snapshot и Fight-stats (shots/power/shield_absorb/
// shell_destroyed/units_destroyed) — порт legacy oxsar2-java/Assault.java.
func Calculate(in Input) (Report, error) {
	if err := validate(in); err != nil {
		return Report{}, err
	}
	if in.Rounds <= 0 {
		in.Rounds = 6
	}

	r := rng.New(in.Seed)

	atk := newState(in.Attackers, in.Rapidfire)
	def := newState(in.Defenders, in.Rapidfire)

	report := Report{Seed: in.Seed}

	for round := 0; round < in.Rounds; round++ {
		if round > 0 {
			atk.regen()
			def.regen()
		}

		// startTurn — фиксируем startTurnQuantity/Damaged/ShellPercent/Shell
		// до выстрелов (Java: Units.startTurn). После shootAtSides у юнита
		// turnShell уменьшится, finishTurn рассчитает потери из дельты.
		atk.startTurn()
		def.startTurn()

		atkSnap := atk.snapshot()
		defSnap := def.snapshot()

		// Stats накапливаются для каждой стороны.
		atkStats := shootAtSides(r, atkSnap, def)
		defStats := shootAtSides(r, defSnap, atk)

		// Подсчёт уничтоженных юнитов до finishTurn.
		atkBeforeCount := totalUnits(atk)
		defBeforeCount := totalUnits(def)

		atk.finishTurn(r)
		def.finishTurn(r)

		atkUnitsDestroyed := atkBeforeCount - totalUnits(atk)
		defUnitsDestroyed := defBeforeCount - totalUnits(def)

		// Снимок «кто и сколько имел в начале раунда» — после finishTurn,
		// чтобы startTurnQuantityDiff содержал реальные потери раунда.
		// printParticipant в Java вызывается ПОСЛЕ finishTurn.
		atkUnits := snapshotRoundUnits(atk)
		defUnits := snapshotRoundUnits(def)

		// Tech-power: gunPower = level × 10%, ballistics/masking — int.
		atkSide := atk.firstSide()
		defSide := def.firstSide()

		// Fight-stats:
		//   - attackerSide.Shots/Power = атакующий стрелял в защитника = atkStats
		//   - attackerSide.UnitsDestroyed = убито у защитника от атак атакующего
		// Маппинг в Java одинаковый: attackerShots — выстрелы атакующего,
		// defenderShipsDestroyed — потери защитника (от атак атакующего).
		attackerSide := RoundSide{
			Username:       atkSide.username,
			Galaxy:         atkSide.galaxy,
			System:         atkSide.system,
			Position:       atkSide.position,
			IsMoon:         atkSide.isMoon,
			GunPowerPct:    techToPct(atkSide.tech.Gun),
			ShieldPowerPct: techToPct(atkSide.tech.Shield),
			ArmoringPct:    techToPct(atkSide.tech.Shell),
			BallisticsLvl:  atkSide.tech.Ballistics,
			MaskingLvl:     atkSide.tech.Masking,
			Shots:          atkStats.shots,
			Power:          atkStats.power,
			ShieldAbsorbed: atkStats.shieldAbsorbed,
			ShellDestroyed: atkStats.shellDestroyed,
			UnitsDestroyed: defUnitsDestroyed,
			Units:          atkUnits,
		}
		defenderSide := RoundSide{
			Username:       defSide.username,
			Galaxy:         defSide.galaxy,
			System:         defSide.system,
			Position:       defSide.position,
			IsMoon:         defSide.isMoon,
			GunPowerPct:    techToPct(defSide.tech.Gun),
			ShieldPowerPct: techToPct(defSide.tech.Shield),
			ArmoringPct:    techToPct(defSide.tech.Shell),
			BallisticsLvl:  defSide.tech.Ballistics,
			MaskingLvl:     defSide.tech.Masking,
			Shots:          defStats.shots,
			Power:          defStats.power,
			ShieldAbsorbed: defStats.shieldAbsorbed,
			ShellDestroyed: defStats.shellDestroyed,
			UnitsDestroyed: atkUnitsDestroyed,
			Units:          defUnits,
		}

		trace := RoundTrace{
			Index:          round,
			AttackersAlive: atk.totalAlive(),
			DefendersAlive: def.totalAlive(),
			AttackerSide:   attackerSide,
			DefenderSide:   defenderSide,
		}
		report.RoundsTrace = append(report.RoundsTrace, trace)
		report.Rounds = round + 1
		if trace.AttackersAlive == 0 || trace.DefendersAlive == 0 {
			break
		}
	}

	report.Attackers = summarize(in.Attackers, atk.toSides())
	report.Defenders = summarize(in.Defenders, def.toSides())
	report.Winner = decideWinner(report)

	// Опыт сторон — порт Java Assault.java:811-816:
	//   atterExperience  = min(20, max(0.1, defStartPower/atkStartPower)) × rounds
	//   defenderExperience = min(20, max(0.1, atkStartPower/defStartPower)) × rounds
	// startBattlePower считаем как сумма attack × startBattleQuantity по
	// всем юнитам стороны (соответствие Java startBattleAtterPower).
	atkPower := startBattlePower(atk)
	defPower := startBattlePower(def)
	if report.Rounds > 0 && atkPower > 0 && defPower > 0 {
		atkRatio := math.Max(0.1, math.Min(20, defPower/atkPower))
		defRatio := math.Max(0.1, math.Min(20, atkPower/defPower))
		report.AttackerExp = int(math.Round(atkRatio * float64(report.Rounds)))
		report.DefenderExp = int(math.Round(defRatio * float64(report.Rounds)))
	}

	return report, nil
}

// startBattlePower — суммарная начальная огневая мощность стороны
// (Σ effectiveAttack × startBattleQuantity по всем юнитам). Используется
// в формуле опыта (Java Assault.startBattleAtterPower / startBattleDefenderPower).
func startBattlePower(b *battleState) float64 {
	var p float64
	for _, s := range b.sides {
		for _, u := range s.units {
			p += u.effectiveAttack * float64(u.startBattleQuantity)
		}
	}
	return p
}

// techToPct — gun_level × 10% (Java: attackLevel * 10.0).
func techToPct(level int) float64 {
	return float64(level) * 10.0
}

// totalUnits — сумма quantity всех юнитов всех сторон battleState.
func totalUnits(b *battleState) int64 {
	var n int64
	for _, s := range b.sides {
		for _, u := range s.units {
			if u.quantity > 0 {
				n += u.quantity
			}
		}
	}
	return n
}

// snapshotRoundUnits — снимок состояния юнитов на конец раунда (=
// «начало раунда» в Java printParticipant, т.к. Java рисует таблицу
// после finishTurn). Java: Units.getStartTurnQuantity, getStartTurnDamaged,
// getStartTurnDamagedShellPercent, getStartTurnQuantityDiff.
//
// При quantity=0, но startBattleQuantity>0 — юнит уничтожен в этом раунде,
// показываем строку с diff (Java делает то же).
func snapshotRoundUnits(b *battleState) []RoundUnit {
	var out []RoundUnit
	for _, s := range b.sides {
		for _, u := range s.units {
			if u.quantity <= 0 && u.startBattleQuantity == 0 {
				continue
			}
			// Скрываем юниты, которые уже уничтожены и не понесли
			// потерь в этом раунде (Java condition в printParticipant:
			// startTurnQuantity > 0 || startTurnQuantityDiff < 0).
			if u.quantity <= 0 && u.startTurnQuantityDiff >= 0 {
				continue
			}
			ru := RoundUnit{
				UnitID:                u.tmpl.UnitID,
				Name:                  u.tmpl.Name,
				StartTurnQuantity:     u.quantity,
				StartTurnQuantityDiff: u.startTurnQuantityDiff,
				StartTurnDamaged:      u.damaged,
				DamagedShellPercent:   int(math.Round(u.shellPercent)),
				Attack:                u.effectiveAttack,
				Shield:                u.tmpl.Shield,
				Shell:                 u.effectiveShell,
				Front:                 u.tmpl.Front,
				BallisticsLevel:       s.tech.Ballistics,
				MaskingLevel:          s.tech.Masking,
				StartBattleQuantity:   u.startBattleQuantity,
			}
			if u.startBattleQuantity > 0 {
				ru.AlivePercent = int(u.quantity * 100 / u.startBattleQuantity)
				if ru.AlivePercent > 100 {
					ru.AlivePercent = 100
				}
			}
			out = append(out, ru)
		}
	}
	return out
}

// firstSide — для simulator scenario (одна сторона на стороне).
// При ACS Java берёт «общую» tech, но для симулятора это первая сторона.
func (b *battleState) firstSide() *sideState {
	if len(b.sides) == 0 {
		return &sideState{}
	}
	return b.sides[0]
}

// --- runtime-состояние боя ---
//
// sideState и unitState — мутабельные версии Side/Unit на время
// расчёта. Публичные типы (types.go) остаются value-семантикой и
// не несут turn-specific полей (turnShell/turnShield) — это деталь
// реализации движка.

type unitState struct {
	// Зеркало входного Unit'а, нужное для итогового summarize.
	tmpl Unit
	// Индекс в исходном Side.Units — чтобы summarize читал по нему.
	idx int
	// startBattleQuantity — количество в самом начале боя (для % alive
	// в Java printParticipant: alivePercent = qty * 100 / startBattleQuantity).
	startBattleQuantity int64
	// turnShell = totalShell (учитывает damaged), уменьшается выстрелами.
	turnShell float64
	// turnShield = Shield × Quantity, восстанавливается regen.
	turnShield float64
	// startTurnShield — значение turnShield в начале раунда (до атак).
	// Используется в shieldDestroyFactor (Java startTurnQuantity × shield).
	startTurnShield float64
	// Quantity — текущее оставшееся количество юнитов (включая damaged).
	quantity int64
	// damaged — сколько из quantity повреждены. M4.3 упрощение: не более
	// одного damaged в пачке (Java делает так же — последний «частично
	// подбитый» юнит становится damaged, ShellPercent отражает остаток).
	damaged int64
	// shellPercent — 0..100, доля shell у damaged-юнита (у здоровых 100%).
	shellPercent float64
	// effectiveAttack — Attack с применённым gun tech (+10%/уровень).
	effectiveAttack float64
	// effectiveShell — shell на юнит с применённым shell tech.
	effectiveShell float64
	// baseShield — Shield БЕЗ tech-множителя. Используется для вычисления
	// ignoreAttack-порога: tech повышает абсорбцию, но не делает щит
	// абсолютным (BA-005).
	baseShield float64

	// Снимок «начало раунда» — порт Java Units.startTurn* полей
	// (план 72.1 ч.20.11.9). Заполняется в startTurn() в начале каждого
	// раунда, читается finishTurn(). Также используется в snapshotRoundUnits
	// для отрисовки таблиц printParticipant.
	startTurnQuantity            int64   // quantity на начало раунда
	startTurnDamaged             int64   // damaged на начало раунда
	startTurnDamagedShellPercent float64 // shellPercent damaged-юнитов на начало раунда
	startTurnShell               float64 // turnShell на начало раунда (до атак)
	startTurnQuantityDiff        int64   // (quantity_after_finishTurn - startTurnQuantity), отрицателен при потерях
	// turnFiredQuantity — сколько юнитов умерло «по абсолютному ablation»
	// (turnShell упал ниже уровня, при котором даже здоровые гарантированно
	// уничтожены). Java увеличивает это в applyShots, finishTurn использует
	// для вычисления minDamaged/maxDamaged. Сбрасывается в startTurn().
	turnFiredQuantity int64
}

type sideState struct {
	userID   string
	username string
	galaxy   int
	system   int
	position int
	isMoon   bool
	// tech — для ballistics/masking. Значение Participant-уровня:
	// ballistics + masking одинаковы для всех unit-ов одной стороны.
	tech  Tech
	units []*unitState
}

type battleState struct {
	sides []*sideState
	// rapidfire[shooterUnitID][targetUnitID] = multiplier (>=1).
	// nil-safe: если таблицы нет — rf считается = 1 для всех пар.
	rapidfire map[int]map[int]int
}

func newState(input []Side, rf map[int]map[int]int) *battleState {
	bs := &battleState{sides: make([]*sideState, len(input)), rapidfire: rf}
	for si, s := range input {
		ss := &sideState{
			userID:   s.UserID,
			username: s.Username,
			galaxy:   s.Galaxy,
			system:   s.System,
			position: s.Position,
			isMoon:   s.IsMoon,
			tech:     s.Tech,
		}
		gunFactor := 1.0 + float64(s.Tech.Gun)*0.10
		shieldFactor := 1.0 + float64(s.Tech.Shield)*0.10
		shellFactor := 1.0 + float64(s.Tech.Shell)*0.10
		for ui, u := range s.Units {
			us := &unitState{
				tmpl:                u,
				idx:                 ui,
				quantity:            u.Quantity,
				startBattleQuantity: u.Quantity,
				damaged:             clampDamaged(u.Damaged, u.Quantity),
				shellPercent:        clampPercent(u.ShellPercent),
				effectiveAttack:     u.Attack * gunFactor,
				effectiveShell:      u.Shell * shellFactor,
			}
			// effectiveShield хранится прямо в turnShield/regen через
			// scaledShield — считаем один раз.
			baseShield := u.Shield
			scaledShield := baseShield * shieldFactor
			us.baseShield = baseShield
			us.turnShell = totalShell(us.effectiveShell, us.quantity, us.damaged, us.shellPercent)
			us.turnShield = float64(u.Quantity) * scaledShield
			us.startTurnShield = us.turnShield
			// Сохраняем масштабированный shield обратно в tmpl для regen и applyShots.
			us.tmpl.Shield = scaledShield
			us.tmpl.Shell = us.effectiveShell
			ss.units = append(ss.units, us)
		}
		bs.sides[si] = ss
	}
	return bs
}

// totalShell — суммарный shell пачки: (quantity-damaged) полных юнитов
// + damaged × shellPercent/100.
func totalShell(shellPerUnit float64, quantity, damaged int64, shellPct float64) float64 {
	if shellPerUnit <= 0 || quantity <= 0 {
		return 0
	}
	if damaged < 0 {
		damaged = 0
	}
	if damaged > quantity {
		damaged = quantity
	}
	full := quantity - damaged
	return float64(full)*shellPerUnit + float64(damaged)*shellPerUnit*shellPct/100.0
}

func clampDamaged(d, q int64) int64 {
	if d < 0 {
		return 0
	}
	if d > q {
		return q
	}
	return d
}

func clampPercent(p float64) float64 {
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}

// regen восстанавливает щиты до полного объёма на начало раунда.
// Java делает так же: turnShield = startTurnQuantity × shield (если
// shield > 0 и юниты живы).
func (b *battleState) regen() {
	for _, s := range b.sides {
		for _, u := range s.units {
			if u.quantity <= 0 {
				u.turnShield = 0
				continue
			}
			u.turnShield = float64(u.quantity) * u.tmpl.Shield
			u.startTurnShield = u.turnShield
		}
	}
}

// startTurn — снапшот «начало раунда» (план 72.1 ч.20.11.9). Java
// Units перед каждым раундом запоминает startTurnQuantity / Damaged
// / DamagedShellPercent / Shell и обнуляет turnFiredQuantity. Эти
// значения читает finishTurn() и snapshotRoundUnits().
func (b *battleState) startTurn() {
	for _, s := range b.sides {
		for _, u := range s.units {
			u.startTurnQuantity = u.quantity
			u.startTurnDamaged = u.damaged
			u.startTurnDamagedShellPercent = u.shellPercent
			u.startTurnShell = u.turnShell
			u.startTurnQuantityDiff = 0
			u.turnFiredQuantity = 0
		}
	}
}

// snapshot возвращает копию состояния — используется как стрелок,
// чтобы мутация целей в shootAtSides не влияла на количество выстрелов.
func (b *battleState) snapshot() *battleState {
	out := &battleState{
		sides:     make([]*sideState, len(b.sides)),
		rapidfire: b.rapidfire,
	}
	for si, s := range b.sides {
		ss := &sideState{userID: s.userID, username: s.username, tech: s.tech}
		for _, u := range s.units {
			cp := *u
			ss.units = append(ss.units, &cp)
		}
		out.sides[si] = ss
	}
	return out
}

// finishTurn — порт Java oxsar2-java/Units.finishTurn (план 72.1
// ч.20.11.9). Модель ablation:
//
//  1. turnShellDestroyed = startTurnShell - turnShell (urон, нанесённый
//     за раунд).
//  2. Если есть startTurnDamaged > 0 — сначала «добиваем» уже подбитые
//     юниты (Java строки 509-518). damagedUnitShell = shell × DPP / 100.
//     damagedUnitsDestroyed = ceil(min(turnDamaged,
//     floor(turnShellDestroyed / damagedUnitShell)) × 0.85).
//  3. Затем уничтожаем здоровых: maxUnitsDestroyed = floor(turnShellDestroyed
//     / shell), unitsDestroyed = ceil(maxUnitsDestroyed × 0.85).
//  4. minDamaged/maxDamaged — диапазон для «частично подбитых» по остатку
//     shell, выбираем случайно через rng.
//  5. turnShellPercent = (turnShell - (turnQuantity - turnDamaged) × shell)
//     × 100 / (turnDamaged × shell).
//  6. Дополнительный «exploding» проход (turnShellPercent < 20|65|99):
//     юниты с критически малым shell гибнут с шансом, оставшиеся остаются
//     damaged. >99 → damaged=0 (округление в полные).
//
// rng детерминированный (rng.New(seed)), поэтому результат полностью
// повторяем для тех же входов.
func (b *battleState) finishTurn(r *rng.R) {
	for _, s := range b.sides {
		for _, u := range s.units {
			u.applyFinishTurn(r)
		}
	}
}

// applyFinishTurn — порт Java Units.finishTurn() per-unit. Все имена
// переменных сохранены из Java для удобства сверки.
func (u *unitState) applyFinishTurn(r *rng.R) {
	if u.startTurnQuantity <= 0 {
		u.startTurnQuantityDiff = 0
		return
	}
	shell := u.tmpl.Shell
	if shell <= 0 {
		// Юнит без shell (защ. зона?) — ablation не применим, состояние
		// не меняется. Diff = 0.
		u.startTurnQuantityDiff = 0
		return
	}

	startTurnQuantity := u.startTurnQuantity
	startTurnShell := u.startTurnShell
	turnShell := u.turnShell

	// turnFiredQuantity = 0 (см. unitState.turnFiredQuantity комментарий) —
	// у нас applyShots не «убивает целиком», всё в turnShell.
	turnQuantity := float64(startTurnQuantity - u.turnFiredQuantity)
	turnDamaged := float64(u.startTurnDamaged)
	turnShellPercent := u.startTurnDamagedShellPercent

	if turnShell < startTurnShell {
		turnShellDestroyed := startTurnShell - turnShell

		// 1. Добиваем уже подбитых (Java 509-518).
		if turnDamaged > 0 && u.startTurnDamagedShellPercent > 0 {
			damagedUnitShell := shell * u.startTurnDamagedShellPercent / 100
			if damagedUnitShell > 0 {
				maxDamagedUnitsDestroyed := math.Min(turnDamaged, math.Floor(turnShellDestroyed/damagedUnitShell))
				damagedUnitsDestroyed := math.Ceil(maxDamagedUnitsDestroyed * 0.85)
				if damagedUnitsDestroyed < 0 {
					damagedUnitsDestroyed = 0
				}
				turnQuantity -= damagedUnitsDestroyed
				turnDamaged -= damagedUnitsDestroyed
				turnShellDestroyed -= damagedUnitsDestroyed * damagedUnitShell
				if turnShellDestroyed < 0 {
					turnShellDestroyed = 0
				}
			}
		}

		// 2. Уничтожаем здоровых (Java 519-521).
		maxUnitsDestroyed := math.Min(turnQuantity, math.Floor(turnShellDestroyed/shell))
		unitsDestroyed := math.Ceil(maxUnitsDestroyed * 0.85)
		if unitsDestroyed < 0 {
			unitsDestroyed = 0
		}
		turnQuantity -= unitsDestroyed

		// 3. Диапазон damaged по остатку (Java 523-535).
		residualShell := turnShellDestroyed - unitsDestroyed*shell
		var minDamaged, maxDamaged float64
		if residualShell > 0 {
			minDamaged = residualShell / (shell * 0.99)
			maxDamaged = residualShell / (shell * 0.1)
		}
		minDamaged = clampVal(minDamaged, turnDamaged, turnQuantity)
		maxDamaged = clampVal(maxDamaged, minDamaged, turnQuantity)
		deltaDamaged := (maxDamaged - minDamaged) * 0.5
		minDamaged += deltaDamaged * 0.49
		maxDamaged -= deltaDamaged * 0.49
		turnDamaged = math.Round(randDouble(r, minDamaged, maxDamaged))
		if turnDamaged == 0 && turnShellDestroyed > 0 && turnQuantity > 0 {
			turnDamaged = 1
		}

		// 4. shellPercent (Java 536).
		if turnDamaged > 0 {
			turnShellPercent = (turnShell - (turnQuantity-turnDamaged)*shell) * 100 / (turnDamaged * shell)
		} else {
			turnShellPercent = 0
		}

		// 5. Exploding (Java 538-565).
		remainTurnShellDestroyed := turnShellDestroyed - (maxUnitsDestroyed-unitsDestroyed)*shell
		if remainTurnShellDestroyed < 0 {
			remainTurnShellDestroyed = 0
		}
		// accurateExploding = true в Java.
		var maxExplode float64
		spClamped := clampVal(turnShellPercent, 1, 99)
		if spClamped > 0 {
			maxExplode = math.Ceil(remainTurnShellDestroyed / (shell * spClamped / 100))
		}
		if maxExplode > turnDamaged {
			maxExplode = turnDamaged
		}

		switch {
		case turnShellPercent < 20:
			turnQuantity -= maxExplode
			turnDamaged -= maxExplode
		case turnShellPercent < 65:
			explodingChance := 1 - turnShellPercent/100
			explodingUnits := math.Ceil(maxExplode * explodingChance)
			turnQuantity -= explodingUnits
			turnDamaged -= explodingUnits
		case turnShellPercent > 99:
			turnDamaged = 0
		}
	} else if turnShell < 1 {
		turnQuantity = 0
		turnDamaged = 0
		turnShellPercent = 0
	}

	turnQuantity = clampVal(turnQuantity, 0, float64(startTurnQuantity))
	turnDamaged = clampVal(turnDamaged, 0, turnQuantity)
	turnShellPercent = clampVal(turnShellPercent, 0, 100)

	newQty := int64(turnQuantity)
	u.startTurnQuantityDiff = newQty - startTurnQuantity
	u.quantity = newQty
	u.damaged = int64(turnDamaged)
	u.shellPercent = turnShellPercent
	// turnShell нормализуем к новому состоянию, чтобы regen+следующий раунд
	// видел правильное startTurnShell.
	u.turnShell = totalShell(shell, u.quantity, u.damaged, u.shellPercent)
}

// clampVal — порт Java Assault.clampVal(v, min, max).
func clampVal(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// randDouble — равномерное случайное в [min, max], детерминированно
// через rng.R. Java: Assault.randDouble(double min, double max).
func randDouble(r *rng.R, lo, hi float64) float64 {
	if hi <= lo {
		return lo
	}
	return lo + r.Float64()*(hi-lo)
}

func (b *battleState) totalAlive() int64 {
	var sum int64
	for _, s := range b.sides {
		for _, u := range s.units {
			sum += u.quantity
		}
	}
	return sum
}

// toSides возвращает []Side с текущим состоянием — для summarize.
// Не может называться sides(), потому что у battleState уже есть
// поле sides.
func (b *battleState) toSides() []Side {
	out := make([]Side, len(b.sides))
	for si, s := range b.sides {
		side := Side{UserID: s.userID, Username: s.username}
		side.Units = make([]Unit, len(s.units))
		for i, u := range s.units {
			// Сохраняем оригинал, перетираем quantity/damaged/
			// ShellPercent (summarize читает именно эти три поля).
			side.Units[i] = u.tmpl
			side.Units[i].Quantity = u.quantity
			side.Units[i].Damaged = u.damaged
			side.Units[i].ShellPercent = u.shellPercent
		}
		out[si] = side
	}
	return out
}

// shootAtSides — shooters бьёт по targets (один полу-раунд).
// Распределение выстрелов по целям — пропорционально их weight
// (2^Front × Quantity, как Java Units.getStartTurnWeight).
//
//        missed = floor(shots * factor)
//     где masking берётся со стороны ЦЕЛИ, ballistics — со СТОРОНЫ-
//     стрелка. Семантика: «я прячу свой флот, враг ищет сквозь помехи».
// roundSideStats — агрегаты Java Fight таблицы для одной стороны
// в одном раунде (план 72.1 ч.20.11.4).
type roundSideStats struct {
	shots          int64   // Java attackerShots
	power          float64 // Java attackerPower (сумма attack × shots до cap)
	shieldAbsorbed float64 // Java defenderShield (поглощено щитом цели)
	shellDestroyed float64 // Java defenderShellDestroyed (вырвано из брони цели)
}

func shootAtSides(r *rng.R, shooters, targets *battleState) roundSideStats {
	_ = r
	stats := roundSideStats{}

	// Собираем активные цели.
	var actives []*unitState
	sideTechOf := make(map[*unitState]Tech)
	var totalWeight float64
	for _, s := range targets.sides {
		for _, u := range s.units {
			if u.quantity <= 0 {
				continue
			}
			w := unitWeight(*u)
			if w <= 0 {
				continue
			}
			actives = append(actives, u)
			sideTechOf[u] = s.tech
			totalWeight += w
		}
	}
	if len(actives) == 0 || totalWeight == 0 {
		return stats
	}

	for _, s := range shooters.sides {
		shooterBallistics := s.tech.Ballistics
		for _, shooter := range s.units {
			if shooter.quantity <= 0 {
				continue
			}
			attack := shooter.effectiveAttack
			if attack <= 0 {
				continue
			}
			for _, tgt := range actives {
				if tgt.quantity <= 0 {
					continue
				}
				w := unitWeight(*tgt)
				portion := w / totalWeight
				rawShots := int64(math.Round(float64(shooter.quantity) * portion))
				if rawShots <= 0 {
					rawShots = 1
				}
				rf := rapidfireMult(shooters.rapidfire, shooter.tmpl.UnitID, tgt.tmpl.UnitID)
				shots := rawShots * int64(rf)
				tgtMasking := sideTechOf[tgt].Masking
				shots = applyMasking(shots, shooterBallistics, tgtMasking)
				if shots <= 0 {
					continue
				}
				stats.shots += shots
				stats.power += attack * float64(shots)
				absorbed, destroyed := applyShots(shooter, tgt, attack, shots)
				stats.shieldAbsorbed += absorbed
				stats.shellDestroyed += destroyed
			}
		}
	}
	return stats
}

// rapidfireMult — вернуть множитель, min = 1.
func rapidfireMult(table map[int]map[int]int, shooterID, targetID int) int {
	if table == nil {
		return 1
	}
	row, ok := table[shooterID]
	if !ok {
		return 1
	}
	v := row[targetID]
	if v < 1 {
		return 1
	}
	return v
}

// applyMasking — детерминированная формула Java processAttack:
//
//	maskingEffect = max(0, masking - ballistics)
//	factor = 1 - 1 / (1 + maskingEffect * 2/10)
//	missed = floor(shots * factor)
//
// Если ballistics >= masking — эффекта нет, возвращаем исходные shots.
func applyMasking(shots int64, ballistics, masking int) int64 {
	if shots <= 0 {
		return 0
	}
	effect := masking - ballistics
	if effect <= 0 {
		return shots
	}
	factor := 1.0 - 1.0/(1.0+float64(effect)*2.0/10.0)
	missed := int64(math.Floor(float64(shots) * factor))
	if missed < 0 {
		missed = 0
	}
	if missed > shots {
		missed = shots
	}
	return shots - missed
}

// applyShots — применяет shots выстрелов мощности attack к target.
//
// Портировано из Java Units.processAttack (строки 315–427):
//
//  1. ignoreAttack = shield / 100. Если attack ≤ ignoreAttack —
//     выстрелы поглощаются щитом без урона в shell.
//
//  2. shieldDestroyFactor = clamp(1 - turnShield/fullTurnShield, 0.01, 1.0)
//     Чем больше щит разрушен, тем больше shots «проходят сквозь».
//
//  3. shieldDestroyUnits = floor(turnShield × shieldDestroyFactor / shield),
//     capped at shots. Эти shots проходят к shell.
//
//  4. Оставшиеся shots бьют щит: cap power per shot ≤ shield,
//     cap total ≤ turnShield. Вычитаем из turnShield.
//
//  5. Shots дошедшие до shell: cap power per shot ≤ shell,
//     cap total ≤ turnShell. Вычитаем из turnShell.
// applyShots возвращает (shieldAbsorbed, shellDestroyed) — для
// агрегации в roundSideStats (Java attackerShield / defenderShield
// и attackerShellDestroyed / defenderShellDestroyed).
func applyShots(shooter, target *unitState, attack float64, shots int64) (float64, float64) {
	if shots <= 0 || target.quantity <= 0 {
		return 0, 0
	}
	_ = shooter

	shieldBefore := target.turnShield
	shellBefore := target.turnShell

	unitShield := target.tmpl.Shield
	// ignoreAttack вычисляется по базовому (до tech) щиту — BA-005.
	// Исключение (Java строки 348-350): планетарные щиты (Small/Large Shield,
	// id 49/50) имеют ignoreAttack=0 — любая атака их пробивает.
	var ignoreAttack float64
	if target.tmpl.UnitID != 49 && target.tmpl.UnitID != 50 {
		ignoreAttack = target.baseShield / 100.0
	}

	// Shots weaker than ignoreAttack don't penetrate to shell.
	if attack > 0 && attack <= ignoreAttack {
		pool := attack * float64(shots)
		if pool > target.turnShield {
			pool = target.turnShield
		}
		target.turnShield -= pool
		return shieldBefore - target.turnShield, 0
	}

	shotsF := float64(shots)

	if unitShield <= 0 {
		// No shield — all shots go directly to shell.
		shellPower := attack * shotsF
		maxShellPower := target.tmpl.Shell * shotsF
		if shellPower > maxShellPower {
			shellPower = maxShellPower
		}
		if shellPower > target.turnShell {
			shellPower = target.turnShell
		}
		target.turnShell -= shellPower
		if target.turnShell < 0 {
			target.turnShell = 0
		}
		return 0, shellBefore - target.turnShell
	}

	// Портировано из Java Units.processAttack строки 358-420.
	fullTurnShield := target.startTurnShield
	var shieldDamageFactor float64
	if fullTurnShield > 0 {
		shieldDamageFactor = target.turnShield / fullTurnShield
	}
	shieldDestroyFactor := 1.0 - shieldDamageFactor
	if shieldDestroyFactor < 0.01 {
		shieldDestroyFactor = 0.01
	} else if shieldDestroyFactor > 1.0 {
		shieldDestroyFactor = 1.0
	}

	// shieldDestroyUnits — сколько щитов уже «сломано» (Java строки 366-372).
	shieldDestroy := target.turnShield * shieldDestroyFactor
	shieldDestroyUnits := math.Floor(shieldDestroy / unitShield)
	if shieldDestroyUnits > shotsF {
		shieldDestroyUnits = shotsF
	}
	shieldDestroy = shieldDestroyUnits * unitShield
	if target.turnShield > 0 {
		shieldDestroyFactor = shieldDestroy / target.turnShield
	}
	shieldExistFactor := 1.0 - shieldDestroyFactor

	// shieldShotsNumber — выстрелы, которые бьют в щит (Java строка 377).
	shieldShotsNumber := math.Ceil(shotsF * shieldExistFactor)
	remainingShots := shotsF

	if shieldShotsNumber > 0 && target.turnShield > 0 {
		shieldShotsPower := attack * shieldShotsNumber
		// При attack > ignoreAttack cap power по unitShield (Java строки 380-388).
		if attack > ignoreAttack {
			maxShieldShotsPower := shieldShotsNumber * unitShield
			if shieldShotsPower > maxShieldShotsPower {
				shieldShotsPower = maxShieldShotsPower
			}
			shieldExist := target.turnShield * shieldExistFactor
			if shieldShotsPower > shieldExist {
				shieldShotsPower = shieldExist
			}
			target.turnShield -= shieldShotsPower
			shieldShotsNumber = math.Round(shieldShotsPower / attack)
		}
		remainingShots -= shieldShotsNumber
	}

	// Оставшиеся выстрелы идут в shell (Java строки 409-420).
	if remainingShots > 0 && target.turnShell > 0 {
		shellPower := attack * remainingShots
		maxShellPower := target.tmpl.Shell * remainingShots
		if shellPower > maxShellPower {
			shellPower = maxShellPower
		}
		if shellPower > target.turnShell {
			shellPower = target.turnShell
		}
		target.turnShell -= shellPower
		if target.turnShell < 0 {
			target.turnShell = 0
		}
	}

	return shieldBefore - target.turnShield, shellBefore - target.turnShell
}

// unitWeight — 2^Front × Quantity. Java: getStartTurnWeight.
func unitWeight(u unitState) float64 {
	if u.quantity <= 0 {
		return 0
	}
	front := u.tmpl.Front
	if front < 0 {
		front = 0
	}
	if front > 30 {
		front = 30
	}
	return math.Pow(2, float64(front)) * float64(u.quantity)
}

func validate(in Input) error {
	if len(in.Attackers) == 0 || len(in.Defenders) == 0 {
		return ErrInvalidInput
	}
	for _, s := range append(append([]Side{}, in.Attackers...), in.Defenders...) {
		if len(s.Units) == 0 {
			return ErrInvalidInput
		}
		for _, u := range s.Units {
			if u.Quantity < 0 {
				return ErrInvalidInput
			}
		}
	}
	return nil
}

func summarize(before []Side, after []Side) []SideResult {
	out := make([]SideResult, len(before))
	for i := range before {
		sr := SideResult{
			UserID:   before[i].UserID,
			Username: before[i].Username,
			IsAliens: before[i].IsAliens,
		}
		for j, u := range before[i].Units {
			endQ := int64(0)
			var endDamaged int64
			var endShellPct float64
			if i < len(after) && j < len(after[i].Units) {
				endQ = after[i].Units[j].Quantity
				endDamaged = after[i].Units[j].Damaged
				endShellPct = after[i].Units[j].ShellPercent
			}
			lost := u.Quantity - endQ
			if lost < 0 {
				lost = 0
			}
			sr.LostMetal += lost * u.Cost.Metal
			sr.LostSilicon += lost * u.Cost.Silicon
			sr.LostHydrogen += lost * u.Cost.Hydrogen
			// Очки за потерянные юниты — порт Java Units.java:113-115:
			// pointsPerUnit = (metal + silicon + hydrogen) / 1000 × 2.
			// LostPoints = Σ qty_lost × pointsPerUnit (Participant.java:766).
			pointsPerUnit := float64(u.Cost.Metal+u.Cost.Silicon+u.Cost.Hydrogen) / 1000.0 * 2.0
			sr.LostPoints += float64(lost) * pointsPerUnit
			sr.LostUnits += lost
			sr.Units = append(sr.Units, UnitResult{
				UnitID: u.UnitID, QuantityStart: u.Quantity,
				QuantityEnd: endQ, DamagedEnd: endDamaged,
				ShellPercentEnd: endShellPct,
			})
		}
		out[i] = sr
	}
	return out
}

func decideWinner(r Report) string {
	if len(r.RoundsTrace) == 0 {
		return "draw"
	}
	last := r.RoundsTrace[len(r.RoundsTrace)-1]
	switch {
	case last.AttackersAlive > 0 && last.DefendersAlive == 0:
		return "attackers"
	case last.DefendersAlive > 0 && last.AttackersAlive == 0:
		return "defenders"
	default:
		return "draw"
	}
}
