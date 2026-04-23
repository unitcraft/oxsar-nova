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

// CostForLevelFloor — как CostForLevel, но с floor() вместо ceil().
// Используется для зданий, где legacy-YAML явно пишет floor({basic}*pow(f,L-1)).
func CostForLevelFloor(base Cost, factor float64, targetLevel int) Cost {
	if targetLevel <= 0 {
		return Cost{}
	}
	m := float64(targetLevel - 1)
	scale := math.Pow(factor, m)
	return Cost{
		Metal:    int64(math.Floor(float64(base.Metal) * scale)),
		Silicon:  int64(math.Floor(float64(base.Silicon) * scale)),
		Hydrogen: int64(math.Floor(float64(base.Hydrogen) * scale)),
	}
}

// ---- Prod-функции: производство ресурса за час ----

// MetalmineProdMetal — производство металла шахтой.
// Формула: floor(30 * L * pow(1.1 + techLaser*0.0006, L))
func MetalmineProdMetal(level, techLaser int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(30 * float64(level) * math.Pow(1.1+float64(techLaser)*0.0006, float64(level)))
}

// SiliconLabProdSilicon — производство кремния.
// Формула: floor(20 * L * pow(1.1 + techSilicon*0.0007, L))
func SiliconLabProdSilicon(level, techSilicon int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(20 * float64(level) * math.Pow(1.1+float64(techSilicon)*0.0007, float64(level)))
}

// HydrogenLabProdHydrogen — производство водорода синтезатором.
// Формула: floor(10 * L * pow(1.1 + techHydrogen*0.0008, L) * (-0.002*temp + 1.28))
func HydrogenLabProdHydrogen(level, techHydrogen, tempC int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(10 * float64(level) * math.Pow(1.1+float64(techHydrogen)*0.0008, float64(level)) * (-0.002*float64(tempC) + 1.28))
}

// MoonHydrogenLabProdHydrogen — производство водорода лунным синтезатором.
// Формула: floor(100 * L * pow(1.1 + techHydrogen*0.0008, L) * (-0.002*temp + 1.28))
func MoonHydrogenLabProdHydrogen(level, techHydrogen, tempC int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(100 * float64(level) * math.Pow(1.1+float64(techHydrogen)*0.0008, float64(level)) * (-0.002*float64(tempC) + 1.28))
}

// SolarPlantProdEnergy — выработка энергии солнечной станцией.
// Формула: floor(20 * L * pow(1.1 + techEnergy*0.0005, L))
func SolarPlantProdEnergy(level, techEnergy int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(20 * float64(level) * math.Pow(1.1+float64(techEnergy)*0.0005, float64(level)))
}

// HydrogenPlantProdEnergy — выработка энергии синтезатором водорода.
// Формула: floor(50 * L * pow(1.15 + techEnergy*0.005, L))
func HydrogenPlantProdEnergy(level, techEnergy int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(50 * float64(level) * math.Pow(1.15+float64(techEnergy)*0.005, float64(level)))
}

// SolarSatelliteProdEnergy — выработка энергии одним солнечным спутником.
// Формула: max(10, min(floor(temp/4 + 20), 50)) * pow(1.05, techEnergy)
// Умножить на количество спутников (count) нужно снаружи.
func SolarSatelliteProdEnergy(tempC, techEnergy int) float64 {
	perUnit := math.Max(10, math.Min(math.Floor(float64(tempC)/4+20), 50))
	return perUnit * math.Pow(1.05, float64(techEnergy))
}

// GraviProdEnergy — выработка энергии гравитрона.
// Формула: basicEnergy * pow(3, L-1)
func GraviProdEnergy(level int, basicEnergy int64) float64 {
	if level <= 0 {
		return 0
	}
	return float64(basicEnergy) * math.Pow(3, float64(level-1))
}

// ---- Cons-функции: потребление энергии зданием ----

// MineConsEnergy — общая формула потребления энергии шахтами/лабораториями.
// Используется для: metalmine (base=10), silicon_lab (base=10), hydrogen_lab (base=20),
// moon_hydrogen_lab (base=200).
// Формула: floor(base * L * pow(1.1 - techEnergy*0.0005, L))
func MineConsEnergy(base float64, level, techEnergy int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(base * float64(level) * math.Pow(1.1-float64(techEnergy)*0.0005, float64(level)))
}

// HydrogenPlantConsHydrogen — потребление водорода силовой установкой.
// Формула: floor(10 * L * pow(1.1 - techEnergy*0.0005, L))
func HydrogenPlantConsHydrogen(level, techEnergy int) float64 {
	return MineConsEnergy(10, level, techEnergy)
}

// Округление в пользу системы для стоимостей (§18.9 ТЗ).
func roundUp(v float64) int64 {
	return int64(math.Ceil(v))
}
