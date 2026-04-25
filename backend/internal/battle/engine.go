package battle

import (
	"errors"
	"math"

	"github.com/oxsar/nova/backend/pkg/rng"
)

// ErrInvalidInput — структурно некорректный вход (нет сторон, нет unit-ов).
var ErrInvalidInput = errors.New("battle: invalid input")

// Calculate — единственная публичная функция пакета. Чистая,
// детерминированная при фиксированном seed.
//
// СТАТУС: M4.1 (шиты + multi-channel, без rapidfire/ballistics/masking).
// Модель:
//   * каждый раунд: обе стороны стреляют по снимку начала раунда;
//   * unitState.turnShield восстанавливается до Shield × Quantity
//     в начале раунда (100% regen — Java default);
//   * урон сначала съедает turnShield, остаток — turnShell;
//   * при turnShell <= 0 юниты погибают целиком.
//
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
		// regen перед раундом (кроме первого — там значения и так
		// инициализированы в newState).
		if round > 0 {
			atk.regen()
			def.regen()
		}

		// Снимок «кто стреляет» — нужен, чтобы обе стороны имели шанс
		// выстрелить в этом раунде, даже если противник их убьёт
		// первым. Java хранит отдельное поле startTurnQuantity;
		// у нас — целиком клон.
		atkSnap := atk.snapshot()
		defSnap := def.snapshot()
		shootAtSides(r, atkSnap, def)
		shootAtSides(r, defSnap, atk)

		// Пересчёт quantity из turnShell (юнит жив, пока
		// turnShell > 0; точное целое вычисляется по Unit.Shell
		// × Quantity = totalShell).
		atk.commitDamage()
		def.commitDamage()

		trace := RoundTrace{
			Index:          round,
			AttackersAlive: atk.totalAlive(),
			DefendersAlive: def.totalAlive(),
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
	return report, nil
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
}

type sideState struct {
	userID   string
	username string
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
		ss := &sideState{userID: s.UserID, username: s.Username, tech: s.Tech}
		gunFactor := 1.0 + float64(s.Tech.Gun)*0.10
		shieldFactor := 1.0 + float64(s.Tech.Shield)*0.10
		shellFactor := 1.0 + float64(s.Tech.Shell)*0.10
		for ui, u := range s.Units {
			us := &unitState{
				tmpl:            u,
				idx:             ui,
				quantity:        u.Quantity,
				damaged:         clampDamaged(u.Damaged, u.Quantity),
				shellPercent:    clampPercent(u.ShellPercent),
				effectiveAttack: u.Attack * gunFactor,
				effectiveShell:  u.Shell * shellFactor,
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

// commitDamage пересчитывает quantity/damaged/shellPercent по
// оставшемуся turnShell. Модель M4.3 (ablation) совпадает с Java:
//
//   - fullRem = floor(turnShell / shell) — сколько осталось «полных»;
//   - если fullRem >= quantity: никто не умер, частичного damaged нет
//     (turnShell ровно на всех);
//   - иначе квантити = fullRem + 1 (если есть дробный остаток) или
//     fullRem (если точная граница). Один damaged-юнит с
//     shellPercent = (turnShell mod shell) / shell × 100.
//   - turnShell нормализуется к новому состоянию, чтобы в следующий
//     раунд не «таскать хвост».
func (b *battleState) commitDamage() {
	for _, s := range b.sides {
		for _, u := range s.units {
			if u.quantity <= 0 {
				continue
			}
			if u.tmpl.Shell <= 0 {
				continue
			}
			if u.turnShell <= 0 {
				u.quantity = 0
				u.damaged = 0
				u.shellPercent = 0
				continue
			}
			fullRem := int64(math.Floor(u.turnShell / u.tmpl.Shell))
			remainder := u.turnShell - float64(fullRem)*u.tmpl.Shell
			if remainder < 0 {
				remainder = 0
			}

			switch {
			case fullRem >= u.quantity:
				// Pool не перебил даже damaged-остаток начала раунда —
				// всё по-прежнему. damaged/shellPercent остаются как
				// были (если раунд не нанёс урон — инварианты живут).
				// turnShell нормализуется к текущему состоянию.
				u.turnShell = totalShell(u.tmpl.Shell, u.quantity, u.damaged, u.shellPercent)
			case remainder > 0:
				// Один damaged + fullRem здоровых. damaged не может
				// быть больше quantity.
				u.quantity = fullRem + 1
				u.damaged = 1
				u.shellPercent = (remainder / u.tmpl.Shell) * 100.0
				u.turnShell = totalShell(u.tmpl.Shell, u.quantity, u.damaged, u.shellPercent)
			default:
				// Точная граница — ровно fullRem полных юнитов, damaged нет.
				u.quantity = fullRem
				u.damaged = 0
				u.shellPercent = 0
				u.turnShell = float64(fullRem) * u.tmpl.Shell
			}
		}
	}
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
func shootAtSides(r *rng.R, shooters, targets *battleState) {
	_ = r

	// Собираем активные цели.
	var actives []*unitState
	// sideTechOf — по unitState'у находим Tech стороны-владельца,
	// чтобы взять masking при расчёте ballistics/masking-эффекта.
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
		return
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
				// shots (до rapidfire/masking) — доля от quantity стрелка,
				// распределяемая по данной цели.
				rawShots := int64(math.Round(float64(shooter.quantity) * portion))
				if rawShots <= 0 {
					rawShots = 1
				}
				// rapidfire: shooter×target — legacy таблица. Если
				// пары нет, считаем rf=1. Значение <1 не легально
				// (Java: Math.max(1, rf)) — принудительно 1.
				rf := rapidfireMult(shooters.rapidfire, shooter.tmpl.UnitID, tgt.tmpl.UnitID)
				shots := rawShots * int64(rf)

				// ballistics vs masking.
				tgtMasking := sideTechOf[tgt].Masking
				shots = applyMasking(shots, shooterBallistics, tgtMasking)
				if shots <= 0 {
					continue
				}
				applyShots(shooter, tgt, attack, shots)
			}
		}
	}
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
func applyShots(shooter, target *unitState, attack float64, shots int64) {
	if shots <= 0 || target.quantity <= 0 {
		return
	}
	_ = shooter

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
		return
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
		return
	}

	// Портировано из Java Units.processAttack строки 358-420.
	//
	// fullTurnShield = startTurnQuantity × shield (Java строка 358).
	// В Go startTurnShield уже хранит это значение.
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
			// Пересчёт фактического числа shots потраченных на щит (Java строка 389).
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
		sr := SideResult{UserID: before[i].UserID, Username: before[i].Username}
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
