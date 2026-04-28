package alien

import (
	"math"

	"oxsar/game-nova/pkg/rng"
)

// ShipSpec — упрощённая запись характеристик корабля для generateFleet.
// Замена origin-таблиц `ship_datasheet` + `construction` (PHP:419-420).
//
// Caller (Ф.3) собирает []ShipSpec из config.Catalog; в Ф.1+Ф.2
// тесты передают spec'и напрямую.
type ShipSpec struct {
	UnitID int
	Name   string
	// AttackerAttack / AttackerShield — поля origin ship_datasheet
	// (a — alien-режим: они выше defender'ских; используются как сила
	// флота при атаке). В nova хранятся как Attack/Shield
	// (battle-asymmetric attacker-vs-defender = §14.5 ТЗ).
	Attack int
	Shield int
	// BasicMetal/BasicSilicon — стоимость постройки, используется для
	// debris (50%) и shell (10%×30%). Origin поля basic_metal/basic_silicon
	// из таблицы construction.
	BasicMetal   int
	BasicSilicon int
}

// Спец-юниты для generateFleet (origin:414-415, 542-543, 545-550).
// Используются для специальной обработки в алгоритме — лимиты,
// масштабирование 0.2x для special_ships, исключения.
const (
	UnitEspionageSensor   = 38
	UnitSolarSatellite    = 39
	UnitDeathStar         = 42
	UnitShipTransplantator = 352
	UnitShipArmoredTerran = 358
	UnitAScreen           = 201 // ALIEN_SCREEN
)

// generateFleetParams — параметры PHP-сигнатуры ($params).
type generateFleetParams struct {
	maxDebris float64
	findMode  bool // origin "find_mode" — для оборонительной симуляции
}

// GenerateFleetOption — функциональная опция для GenerateFleet.
type GenerateFleetOption func(*generateFleetParams)

// WithMaxDebris переопределяет cfg.FleetMaxDebris (origin: $params["max_derbis"]).
func WithMaxDebris(v float64) GenerateFleetOption {
	return func(p *generateFleetParams) { p.maxDebris = v }
}

// WithFindMode включает «find_mode» — origin использует его для
// поиска подходящего ответного флота при защите. См. PHP:409, 449-451.
func WithFindMode() GenerateFleetOption {
	return func(p *generateFleetParams) { p.findMode = true }
}

// GenerateFleet — порт AlienAI::generateFleet (AlienAI.class.php:405-622).
//
// Параметры:
//   - target — флот цели (для расчёта target_power).
//   - available — доступные юниты для пришельцев (UNIT_A_*).
//     Caller обязан передать только те, что разрешены вселенной.
//   - scale — множитель силы (1.5..2.0 в четверг, 0.9..1.1 обычно).
//
// Возвращает Fleet, либо nil если не получилось ничего собрать.
//
// Реализация 1-в-1 с PHP, кроме:
//   - вместо `array_rand($available_ships)` берём детерминированный
//     `r.IntN(len(available))` — равномерное распределение, отличается
//     порядком от PHP (для бит-совместимости — Ф.6 / golden tests).
//   - target_ships — здесь []ShipSpec со значением quantity
//     (PHP передаёт `$target_ships[$id] = qty` — мы используем
//     отдельный []TargetUnit).
func GenerateFleet(
	target []TargetUnit,
	available []ShipSpec,
	scale float64,
	cfg Config,
	r *rng.R,
	opts ...GenerateFleetOption,
) Fleet {
	p := generateFleetParams{maxDebris: cfg.FleetMaxDebris}
	for _, o := range opts {
		o(&p)
	}
	if scale < 1 {
		scale = 1 // origin pow(max(1, $scale), 2) — нижняя граница
	}

	// === Шаг 1: target_power по target-флоту (PHP:411-472).
	// `use_shield_power=true`, `use_shell_power=false` (PHP:407-408).
	var targetAttack, targetShields float64
	var targetAvgQuantitySum, targetAvgQuantityCnt int64
	deathStarDebris := 0.0

	for _, tu := range target {
		spec := tu.Spec
		quantity := tu.Quantity
		if spec.UnitID == UnitDeathStar {
			deathStarDebris = float64(spec.BasicMetal+spec.BasicSilicon) * 0.5
		}
		addQ := false
		special := false
		switch spec.UnitID {
		case UnitEspionageSensor:
			// PHP:428-432: sensor — игнорируется (zeroes).
			spec.BasicMetal = 0
			spec.BasicSilicon = 0
			spec.Attack = 0
			spec.Shield = 0
		case UnitDeathStar, UnitShipTransplantator:
			special = true
			spec.BasicMetal = int(float64(spec.BasicMetal) * 0.2)
			spec.BasicSilicon = int(float64(spec.BasicSilicon) * 0.2)
			spec.Attack = int(float64(spec.Attack) * 0.2)
			spec.Shield = int(float64(spec.Shield) * 0.2)
		case UnitShipArmoredTerran:
			spec.BasicMetal = int(float64(spec.BasicMetal) * 0.001)
			spec.BasicSilicon = int(float64(spec.BasicSilicon) * 0.001)
			spec.Attack = int(float64(spec.Attack) * 0.001)
			spec.Shield = int(float64(spec.Shield) * 0.001)
		default:
			addQ = true
		}
		if p.findMode {
			cap := int64(50 + r.IntN(51)) // mt_rand(50, 100)
			if quantity > cap {
				quantity = cap
			}
		}
		if special {
			cap := int64(50 + r.IntN(51))
			if quantity > cap {
				quantity = cap
			}
		}
		if addQ {
			targetAvgQuantitySum += quantity
			targetAvgQuantityCnt++
		}
		targetAttack += float64(spec.Attack) * float64(quantity)
		targetShields += float64(spec.Shield) * float64(quantity)
	}

	avgTargetQty := 0.0
	if targetAvgQuantityCnt > 0 {
		avgTargetQty = float64(targetAvgQuantitySum) / float64(targetAvgQuantityCnt)
	}

	targetPower := targetAttack + targetShields
	targetPower *= scale
	if targetPower < 100 {
		targetPower = 100
	}

	// === Шаг 2: подготовка available_ships (PHP:476-531).
	// max_death_stars / armored_terran random — origin использует
	// mt_rand(0, 10) и mt_rand(0, 50). У нас детерминированный rng,
	// поэтому семантически: «ненулевое значение» с тем же распределением.
	maxDeathStars := 0
	avail := make(map[int]availShip, len(available))
	for _, s := range available {
		avail[s.UnitID] = mkAvailShip(s)
	}

	// Если у цели есть Death Stars — добавляем им свои (PHP:479-487).
	for _, tu := range target {
		if tu.Spec.UnitID == UnitDeathStar {
			ensureSpec(avail, tu.Spec)
			roll := r.IntN(11) // mt_rand(0, 10)
			limFromMass := 0
			if deathStarDebris > 0 {
				if roll != 0 {
					limFromMass = ceilDiv(p.maxDebris*0.5, deathStarDebris)
				} else {
					limFromMass = ceilDiv(p.maxDebris*0.9, deathStarDebris)
				}
			}
			limFromQty := 0
			if roll != 0 {
				limFromQty = int(math.Ceil(float64(tu.Quantity) * 0.3))
			} else {
				limFromQty = int(math.Ceil(float64(tu.Quantity) * 0.9))
			}
			maxDeathStars = minInt(100, limFromQty)
			if limFromMass > 0 {
				maxDeathStars = minInt(maxDeathStars, limFromMass)
			}
			break
		}
	}
	if maxDeathStars == 0 && r.IntN(100) < 10 { // PHP:489 — 10%
		// 10% — добавить 1 DS в available даже если у цели нет
		// (но ShipSpec у нас нет; пропускаем — caller может передать).
		// Найдём DS среди available — если есть, выставим лимит 1.
		if _, ok := avail[UnitDeathStar]; ok {
			maxDeathStars = 1
		}
	}

	// PHP:494 — `mt_rand(0, 50)` ненулевое → удалить ARMORED_TERRAN.
	if r.IntN(51) != 0 {
		delete(avail, UnitShipArmoredTerran)
	}
	if len(avail) == 0 {
		return nil
	}

	// === Шаг 3: итеративный подбор (PHP:538-622).
	power := 0.0
	debris := 0.0
	fleet := make(map[int]*FleetUnit)
	maxSingleUnits := int64(0)

	ignoreCountUnits := map[int]struct{}{
		UnitDeathStar:          {},
		UnitShipArmoredTerran:  {},
		UnitShipTransplantator: {},
		UnitEspionageSensor:    {},
		UnitSolarSatellite:     {},
		UnitAScreen:            {},
	}
	maxSpecialScale := 1.0 / scale
	maxShipMass := map[int]float64{
		UnitDeathStar:          ifFloat(p.findMode, 0.03, 0.3) * maxSpecialScale,
		UnitShipTransplantator: ifFloat(p.findMode, 0.05, 0.1) * maxSpecialScale,
		UnitShipArmoredTerran:  0.001,
		UnitEspionageSensor:    ifFloat(p.findMode, 0.2, 0.5) * maxSpecialScale,
		UnitAScreen:            0.5 / maxSpecialScale,
	}

	// targetByID — для быстрого lookup в цикле (PHP:568-575).
	targetByID := make(map[int]int64, len(target))
	for _, tu := range target {
		targetByID[tu.Spec.UnitID] = tu.Quantity
	}

	const guardMaxIter = 20000
	for iter := 0; iter < guardMaxIter; iter++ {
		if power >= targetPower || debris >= p.maxDebris || len(avail) == 0 {
			break
		}
		// PHP `array_rand` — равномерный выбор ключа из map.
		id := pickRandomKey(avail, r)
		ship := avail[id]
		fu := fleet[id]
		if fu == nil {
			fu = &FleetUnit{
				UnitID:       id,
				Name:         ship.name,
				Quantity:     0,
				ShellPercent: 100,
			}
			fleet[id] = fu
		}

		// PHP:562 cap stack: rand(18000, 20000)
		stackCap := int64(18000 + r.IntN(2001))
		if fu.Quantity >= stackCap {
			power += targetPower * 0.05
			if ship.realAttack > 10 {
				if _, ignore := ignoreCountUnits[id]; !ignore {
					if fu.Quantity > maxSingleUnits {
						maxSingleUnits = fu.Quantity
					}
				}
			}
			continue
		}

		// targetQuantity — мягкий лимит, зависит от target-флота
		// (PHP:567-575).
		targetQuantity := int64(0)
		if tq, ok := targetByID[id]; ok && tq > 0 {
			s := math.Pow(scale, 2) // max(1, scale)^2 — мы выше нормировали scale≥1
			cap := int64(avgTargetQty*s) + int64(10+r.IntN(91))
			if tq < cap {
				targetQuantity = tq
			} else {
				targetQuantity = cap
			}
		}

		// PHP:578 inc_unit вычисление.
		lo := maxInt64(1, ceilInt64(float64(maxSingleUnits)*0.1))
		hi := maxInt64(5, ceilInt64(float64(maxSingleUnits)*0.3))
		if hi < lo {
			hi = lo
		}
		incUnit := lo
		if hi > lo {
			incUnit = lo + int64(r.IntN(int(hi-lo+1)))
		}
		if incUnit < 1 {
			incUnit = 1
		}

		// hard limits (PHP:579-594).
		hard := false
		switch {
		case targetQuantity > 0 && fu.Quantity+incUnit > targetQuantity*3:
			hard = true
		case fu.Quantity+incUnit > 15000:
			hard = true
		case id == UnitShipArmoredTerran:
			// PHP:585-590 — выставляем 1 и breakим в финальный флот.
			fleet = map[int]*FleetUnit{
				id: {
					UnitID:       id,
					Name:         ship.name,
					Quantity:     1,
					ShellPercent: 100,
				},
			}
			return fleetMapToSlice(fleet)
		case id == UnitDeathStar && fu.Quantity+incUnit > int64(maxDeathStars):
			hard = true
		case id == UnitShipTransplantator && fu.Quantity+incUnit > int64(1+maxDeathStars*2):
			hard = true
		}
		if hard {
			if fu.Quantity == 0 {
				incUnit = 1
			} else {
				incUnit = 0
			}
			delete(avail, id)
		} else if mass, ok := maxShipMass[id]; ok &&
			fu.Quantity+incUnit > ceilInt64(float64(maxSingleUnits)*mass) {
			incUnit = 1
		} else if ship.realAttack <= 10 &&
			fu.Quantity+incUnit > int64(float64(maxSingleUnits)*0.01) {
			incUnit = 1
		}
		fu.Quantity += incUnit

		power += ship.shield * float64(incUnit)
		power += ship.attack * float64(incUnit)
		debris += ship.debris * float64(incUnit)

		if ship.realAttack > 10 {
			if _, ignore := ignoreCountUnits[id]; !ignore {
				if fu.Quantity > maxSingleUnits {
					maxSingleUnits = fu.Quantity
				}
			}
		}
	}

	if len(fleet) == 0 {
		return nil
	}
	return fleetMapToSlice(fleet)
}

// TargetUnit — пара ShipSpec + quantity (вместо PHP `$target_ships[$id]=qty`).
type TargetUnit struct {
	Spec     ShipSpec
	Quantity int64
}

// availShip — состав available_ships из PHP (после нормализации
// special-кораблей).
type availShip struct {
	id         int
	name       string
	debris     float64
	shell      float64
	shield     float64
	attack     float64
	realAttack float64
}

func mkAvailShip(s ShipSpec) availShip {
	bm := float64(s.BasicMetal)
	bs := float64(s.BasicSilicon)
	att := float64(s.Attack)
	shi := float64(s.Shield)
	switch s.UnitID {
	case UnitEspionageSensor:
		bm, bs, att, shi = 0, 0, 0, 0
	case UnitDeathStar, UnitShipTransplantator:
		bm *= 0.2
		bs *= 0.2
		att *= 0.2
		shi *= 0.2
	}
	return availShip{
		id:         s.UnitID,
		name:       s.Name,
		debris:     (float64(s.BasicMetal) + float64(s.BasicSilicon)) * 0.5,
		shell:      math.Max(10, (bm+bs)/10) * 0.3,
		shield:     math.Max(10, shi),
		attack:     math.Max(10, att),
		realAttack: float64(s.Attack),
	}
}

func ensureSpec(m map[int]availShip, s ShipSpec) {
	if _, ok := m[s.UnitID]; !ok {
		m[s.UnitID] = mkAvailShip(s)
	}
}

// pickRandomKey — равномерный выбор ключа карты на детерминированном rng.
// Эмулирует PHP `array_rand($map)`.
func pickRandomKey(m map[int]availShip, r *rng.R) int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// сортируем для детерминизма (порядок в map не определён).
	sortInts(keys)
	return keys[r.IntN(len(keys))]
}

func sortInts(a []int) {
	// маленький insertion sort — в типичном случае len(a) <= 6
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j-1] > a[j]; j-- {
			a[j-1], a[j] = a[j], a[j-1]
		}
	}
}

func fleetMapToSlice(m map[int]*FleetUnit) Fleet {
	if len(m) == 0 {
		return nil
	}
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortInts(keys)
	out := make(Fleet, 0, len(keys))
	for _, k := range keys {
		fu := m[k]
		if fu.Quantity <= 0 {
			continue
		}
		out = append(out, *fu)
	}
	return out
}

func ceilDiv(a, b float64) int {
	if b == 0 {
		return 0
	}
	return int(math.Ceil(a / b))
}

func ceilInt64(v float64) int64 {
	return int64(math.Ceil(v))
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func ifFloat(cond bool, a, b float64) float64 {
	if cond {
		return a
	}
	return b
}
