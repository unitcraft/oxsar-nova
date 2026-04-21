// Package economy содержит формулы производства, стоимостей и времени.
//
// Все формулы должны быть чистыми функциями (никаких IO, никаких
// глобальных переменных), чтобы их было легко покрыть table-driven
// тестами и выверять баланс против oxsar2.
package economy

import (
	"math"
	"time"
)

// Cost — ресурсная стоимость.
type Cost struct {
	Metal    int64
	Silicon  int64
	Hydrogen int64
}

// CostForLevel возвращает стоимость апгрейда здания/исследования до
// уровня targetLevel. Формула: base * factor^(target-1).
func CostForLevel(base Cost, factor float64, targetLevel int) Cost {
	if targetLevel <= 0 {
		return Cost{}
	}
	m := float64(targetLevel - 1)
	scale := math.Pow(factor, m)
	return Cost{
		Metal:    roundUp(float64(base.Metal) * scale),
		Silicon:  roundUp(float64(base.Silicon) * scale),
		Hydrogen: roundUp(float64(base.Hydrogen) * scale),
	}
}

// BuildDuration возвращает время постройки с учётом Robotic/Nano фабрик
// и множителя скорости вселенной (GAMESPEED > 1 ускоряет).
func BuildDuration(baseSeconds int, cost Cost, roboticLevel, nanoLevel int, gameSpeed float64) time.Duration {
	if baseSeconds <= 0 {
		baseSeconds = 60
	}
	resSum := float64(cost.Metal + cost.Silicon)
	// Базовая формула похожа на OGame: t = (m+s) / (2500 * (1+robo) * 2^nano) секунд * baseScale
	baseScale := float64(baseSeconds) / 60.0
	raw := resSum / (2500.0 * (1.0 + float64(roboticLevel)) * math.Pow(2, float64(nanoLevel)))
	raw *= baseScale
	if gameSpeed > 0 {
		raw /= gameSpeed
	}
	if raw < 1 {
		raw = 1
	}
	return time.Duration(raw * float64(time.Second))
}

// ProductionPerHour возвращает добычу ресурса для конкретного уровня
// шахты. Формула: base * level * 1.1^level * factor.
//
// factor — это составной множитель (planet.produce_factor × research ×
// artefact bonus × officer bonus × energy ratio).
func ProductionPerHour(baseRate float64, level int, factor float64) float64 {
	if level <= 0 || baseRate <= 0 {
		return 0
	}
	return baseRate * float64(level) * math.Pow(1.1, float64(level)) * factor
}

// EnergyDemand — потребление энергии шахтой с базовой потребностью на
// уровень.
func EnergyDemand(perLevel float64, level int) float64 {
	if level <= 0 {
		return 0
	}
	return perLevel * float64(level) * math.Pow(1.1, float64(level))
}

// EnergyOutput — выход энергии здания (Solar Plant).
func EnergyOutput(perLevel float64, level int) float64 {
	if level <= 0 {
		return 0
	}
	return perLevel * float64(level) * math.Pow(1.1, float64(level))
}

// StorageCapacity возвращает ёмкость хранилища ресурсов.
//
// Формула OGame classic: cap(level) = base * round(2.5 * exp(level * 20 / 33)).
// При level=0 берём базу (5000), при level=1 — ~10000, при level=20 — миллионы.
// factor — это storage_factor планеты (§5.10.1): активный ATOMIC_DENSIFIER
// увеличивает на 0.15.
//
// baseCap = 5000 (точное совпадение с legacy — подлежит сверке с
// oxsar2 на M1, см. TODO).
func StorageCapacity(baseCap int64, level int, factor float64) float64 {
	if factor <= 0 {
		factor = 1
	}
	if level <= 0 {
		// Уровень 0 — минимальная ёмкость (стартовая).
		return float64(baseCap) * factor
	}
	exp := math.Exp(float64(level) * 20.0 / 33.0)
	return float64(baseCap) * math.Round(2.5*exp) * factor
}

// EnergyRatio возвращает долю удовлетворённой энергии (0..1). Если
// производство ≥ потребности, возвращает 1.
func EnergyRatio(output, demand float64) float64 {
	if demand <= 0 {
		return 1
	}
	if output >= demand {
		return 1
	}
	return output / demand
}

// Округление в пользу системы для стоимостей (§18.9 ТЗ).
func roundUp(v float64) int64 {
	return int64(math.Ceil(v))
}
