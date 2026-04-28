package alien

import (
	"oxsar/game-nova/internal/economy"
	"oxsar/game-nova/pkg/rng"
)

// Tech-группы, которые AlienAI shuffle'ит перед записью в Mission.AddTech.
// Семантика origin (AlienAI.class.php:133-135):
//
//   shuffleKeyValues($research, [GUN_TECH, SHIELD_TECH, SHELL_TECH]);
//   shuffleKeyValues($research, [BALLISTICS_TECH, MASKING_TECH]);
//   shuffleKeyValues($research, [LASER_TECH, ION_TECH, PLASMA_TECH]);
//
// Цель: alien не знает точно какие техи у игрока, но «сборная» сила
// сохраняется (значения внутри группы перетасованы, не обнулены).
//
// Затем для каждой техи применяется:
//   add_tech_$id = max(0, mt_rand(floor(level*0.7), level<3?level:level+1))
// (origin:138). Это финальный «угадываемый» уровень для Assault.
//
// В Ф.1 реализован только сам shuffle. Применение random ослабления
// (ApplyShuffledTech) — отдельная функция здесь же.

// shuffleGroups — группы взаимозаменяемых техник.
var shuffleGroups = [][]int{
	{economy.IDTechGun, economy.IDTechShield, economy.IDTechShell},
	{economy.IDTechBallistics, economy.IDTechMasking},
	// IDTechSilicon=24=ion_tech, IDTechHydrogen=25=plasma_tech (см. economy/ids.go).
	// IDTechLaser=23=laser_tech.
	{economy.IDTechLaser, economy.IDTechSilicon, economy.IDTechHydrogen},
}

// ShuffleKeyValues — порт PHP AlienAI::shuffleKeyValues
// (AlienAI.class.php:251-264).
//
// Перетасовывает значения внутри одной группы tech ID:
//
//	for keys = [G,S,Sh]:
//	  values = [t[G], t[S], t[Sh]]
//	  shuffle(values)
//	  t[G], t[S], t[Sh] = values
//
// Если ключ отсутствует — берётся 0. Выходной map содержит все
// перечисленные keys (вход без них теряет их, но это совпадает
// с PHP, где `isset` возвращает 0).
//
// Не мутирует входной tech (R15: helper'ы pure без побочек).
func ShuffleKeyValues(tech TechProfile, keys []int, r *rng.R) TechProfile {
	out := tech.Clone()
	if len(keys) == 0 {
		return out
	}

	values := make([]int, len(keys))
	for i, k := range keys {
		values[i] = tech[k] // missing → 0
	}
	shuffleInts(values, r)
	for i, k := range keys {
		out[k] = values[i]
	}
	return out
}

// ShuffleAllAlienTechGroups — применяет ShuffleKeyValues ко всем 3
// каноническим группам (origin AlienAI.class.php:133-135).
func ShuffleAllAlienTechGroups(tech TechProfile, r *rng.R) TechProfile {
	out := tech
	for _, group := range shuffleGroups {
		out = ShuffleKeyValues(out, group, r)
	}
	return out
}

// ApplyShuffledTechWeakening — для каждого тех-уровня применяет
// origin-формулу ослабления (AlienAI.class.php:138):
//
//	level_visible = max(0, rand(floor(level*0.7), level<3 ? level : level+1))
//
// Возвращает новую карту, не мутирует вход.
func ApplyShuffledTechWeakening(tech TechProfile, r *rng.R) TechProfile {
	out := make(TechProfile, len(tech))
	for k, level := range tech {
		out[k] = weakenedTechLevel(level, r)
	}
	return out
}

func weakenedTechLevel(level int, r *rng.R) int {
	if level <= 0 {
		return 0
	}
	lo := int(float64(level) * 0.7) // floor
	hi := level
	if level >= 3 {
		hi = level + 1
	}
	if lo > hi {
		lo = hi
	}
	v := lo
	if hi > lo {
		v += r.IntN(hi - lo + 1)
	}
	if v < 0 {
		return 0
	}
	return v
}

// shuffleInts — Fisher-Yates shuffle на детерминированном RNG.
// Семантически совместим с PHP shuffle() (порядок отличается,
// но равномерность распределения та же; для бит-совместимости
// потребуется порт PHP'шной mt_rand — это R8 / Ф.6).
func shuffleInts(a []int, r *rng.R) {
	for i := len(a) - 1; i > 0; i-- {
		j := r.IntN(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}
