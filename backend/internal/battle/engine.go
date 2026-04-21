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
//   * выстрел выбирает primary channel (канал с max Attack),
//     бьёт по соответствующему Shield[channel];
//   * урон сначала съедает turnShield, остаток — turnShell;
//   * при turnShell <= 0 юниты погибают целиком.
//
// Что не реализовано (пойдёт в M4.2+):
//   - rapidfire (in.Rapidfire уже есть, но не используется);
//   - ballistics/masking roll (RNG подготовлен);
//   - частичный regen щитов (damageFactor при массовом пробитии);
//   - multi-канальное распределение одного выстрела;
//   - ablation (damaged/ShellPercent) — пока работает «целиком жив
//     или мёртв».
func Calculate(in Input) (Report, error) {
	if err := validate(in); err != nil {
		return Report{}, err
	}
	if in.Rounds <= 0 {
		in.Rounds = 6
	}

	r := rng.New(in.Seed)

	atk := newState(in.Attackers)
	def := newState(in.Defenders)

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
	// turnShell = totalShell, уменьшается выстрелами.
	turnShell float64
	// turnShield = Shield[primaryCh] × Quantity, восстанавливается regen.
	turnShield float64
	// Quantity — текущее оставшееся количество юнитов.
	quantity int64
	// primaryChannel — канал с max Attack; выбираем один раз, не
	// пересчитываем каждый выстрел.
	primaryChannel int
}

type sideState struct {
	userID   string
	username string
	units    []*unitState
}

type battleState struct {
	sides []*sideState
}

func newState(input []Side) *battleState {
	bs := &battleState{sides: make([]*sideState, len(input))}
	for si, s := range input {
		ss := &sideState{userID: s.UserID, username: s.Username}
		for ui, u := range s.Units {
			us := &unitState{
				tmpl:           u,
				idx:            ui,
				quantity:       u.Quantity,
				primaryChannel: primaryChannel(u.Attack),
			}
			us.turnShell = float64(u.Quantity) * u.Shell
			us.turnShield = float64(u.Quantity) * u.Shield[us.primaryChannel]
			ss.units = append(ss.units, us)
		}
		bs.sides[si] = ss
	}
	return bs
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
			u.turnShield = float64(u.quantity) * u.tmpl.Shield[u.primaryChannel]
		}
	}
}

// snapshot возвращает копию состояния — используется как стрелок,
// чтобы мутация целей в shootAtSides не влияла на количество выстрелов.
func (b *battleState) snapshot() *battleState {
	out := &battleState{sides: make([]*sideState, len(b.sides))}
	for si, s := range b.sides {
		ss := &sideState{userID: s.userID, username: s.username}
		for _, u := range s.units {
			cp := *u
			ss.units = append(ss.units, &cp)
		}
		out.sides[si] = ss
	}
	return out
}

// commitDamage пересчитывает quantity по оставшемуся turnShell.
// Если turnShell ≤ 0 → все юниты этой пачки мертвы.
func (b *battleState) commitDamage() {
	for _, s := range b.sides {
		for _, u := range s.units {
			if u.quantity <= 0 {
				continue
			}
			if u.turnShell <= 0 {
				u.quantity = 0
				continue
			}
			if u.tmpl.Shell <= 0 {
				continue
			}
			// Сколько целых юнитов осталось по оставшемуся shell.
			newQty := int64(math.Floor(u.turnShell / u.tmpl.Shell))
			if newQty > u.quantity {
				newQty = u.quantity
			}
			if newQty < 0 {
				newQty = 0
			}
			u.quantity = newQty
			// Нормализуем turnShell к целым юнитам, чтобы следующий
			// раунд не потянул «дробный хвост» шелла.
			u.turnShell = float64(newQty) * u.tmpl.Shell
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
			// Сохраняем оригинал, перетираем только quantity/damaged/
			// ShellPercent (summarize читает именно эти три поля).
			side.Units[i] = u.tmpl
			side.Units[i].Quantity = u.quantity
		}
		out[si] = side
	}
	return out
}

// shootAtSides — shooters бьёт по targets (один полу-раунд).
// Распределение выстрелов по целям — пропорционально их weight
// (2^Front × Quantity, как Java Units.getStartTurnWeight).
func shootAtSides(r *rng.R, shooters, targets *battleState) {
	_ = r // M4.2: ballistics/masking roll

	// Собираем активные цели.
	var actives []*unitState
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
			totalWeight += w
		}
	}
	if len(actives) == 0 || totalWeight == 0 {
		return
	}

	for _, s := range shooters.sides {
		for _, shooter := range s.units {
			if shooter.quantity <= 0 {
				continue
			}
			attack := shooter.tmpl.Attack[shooter.primaryChannel]
			if attack <= 0 {
				continue
			}
			shots := shooter.quantity // rapidfire=1 в M4.1
			for _, tgt := range actives {
				if tgt.quantity <= 0 {
					continue
				}
				w := unitWeight(*tgt)
				portion := w / totalWeight
				targetShots := int64(math.Round(float64(shots) * portion))
				if targetShots <= 0 {
					targetShots = 1
				}
				applyShots(shooter, tgt, attack, targetShots)
			}
		}
	}
}

// applyShots — применяет shots выстрелов мощности attack к target.
//
// Модель щитов (M4.1, упрощённая vs Java):
//   - pool = attack × shots (общая мощность пачки выстрелов);
//   - shieldAbsorb = min(pool, turnShield) — сколько щит поглотит;
//   - turnShield -= shieldAbsorb;
//   - остаток pool идёт в turnShell;
//   - turnShell -= pool_remainder.
//
// Важное отличие от Java: мы НЕ моделируем «одинокий щит защищает от
// многих слабых выстрелов». В Java каждый выстрел меньше unit.shield
// полностью поглощается без урона в shell (ignoreAttack = shield/100).
// В M4.1 мы суммируем мощность, что даёт более «линейное» поведение.
// Для OGame-паритета это придёт в M4.2.
func applyShots(shooter, target *unitState, attack float64, shots int64) {
	if shots <= 0 || target.quantity <= 0 {
		return
	}
	// M4.1 упрощение «ignoreAttack»: если урон одного выстрела меньше,
	// чем shield[primary] / 100 (Java threshold), выстрелы
	// суммарно НЕ пробивают щит. Это поведение важно для мелких
	// пулемётов против больших линкоров.
	unitShield := target.tmpl.Shield[target.primaryChannel]
	ignoreThreshold := unitShield / 100.0
	if attack > 0 && attack < ignoreThreshold {
		// Весь пул сбивает щит до 0, но ничего не идёт в shell.
		pool := attack * float64(shots)
		if pool > target.turnShield {
			pool = target.turnShield
		}
		target.turnShield -= pool
		return
	}

	pool := attack * float64(shots)
	if target.turnShield > 0 {
		absorb := pool
		if absorb > target.turnShield {
			absorb = target.turnShield
		}
		target.turnShield -= absorb
		pool -= absorb
	}
	if pool > 0 && target.turnShell > 0 {
		target.turnShell -= pool
		if target.turnShell < 0 {
			target.turnShell = 0
		}
	}
	_ = shooter // shooter-специфичный tracking придёт в M4.3 (stats)
}

// primaryChannel — возвращает индекс канала (0/1/2) с максимальным
// Attack. Если все нули — 0.
func primaryChannel(attack [3]float64) int {
	maxIdx := 0
	for i := 1; i < 3; i++ {
		if attack[i] > attack[maxIdx] {
			maxIdx = i
		}
	}
	return maxIdx
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
