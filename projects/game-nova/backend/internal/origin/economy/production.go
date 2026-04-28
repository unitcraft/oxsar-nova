package economy

import (
	"math"

	"oxsar/game-nova/internal/balance"
)

// MetalMineProduction — производство металла в час одной планеты.
// Соответствует origin DSL:
//
//	floor(30 * {level} * pow(1.1 + {tech=23}*0.0006, {level}))
//
// где tech=23 — laser_tech (LASER_TECH в legacy, по факту — производ-
// ственная техн. metal_mine; имя в legacy исторически «laser»).
//
// Коэффициенты 30 и 0.0006 берутся из bundle.Globals (Modern-вселенные
// используют ModernGlobals — числа совпадают с internal/economy/
// MetalmineProdMetal). Origin может override через configs/balance/
// origin.yaml::globals.metal_mine_basic_prod / metal_mine_tech_factor.
func MetalMineProduction(g balance.Globals, level, techLaser int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(g.MetalMineBasicProd * float64(level) *
		math.Pow(1.1+float64(techLaser)*g.MetalMineTechFactor, float64(level)))
}

// SiliconLabProduction — производство кремния в час.
//   floor(20 * {level} * pow(1.1 + {tech=24}*0.0007, {level}))
// tech=24 → silicon-tech (SILICON_TECH в legacy).
func SiliconLabProduction(g balance.Globals, level, techSilicon int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(g.SiliconLabBasicProd * float64(level) *
		math.Pow(1.1+float64(techSilicon)*g.SiliconLabTechFactor, float64(level)))
}

// HydrogenLabProduction — производство водорода в час с учётом
// температуры планеты (закрывает D-029):
//
//	floor(10 * {level} * pow(1.1 + {tech=25}*0.0008, {level}) *
//	      (-0.002 * {temp} + 1.28))
//
// Температура — целое число в °C (origin диапазон −200..+200, см.
// na_planet.temp_max). Холодные планеты производят больше — это
// сознательный gameplay-mechanism legacy oxsar2.
//
// tech=25 → hydrogen-tech.
func HydrogenLabProduction(g balance.Globals, level, techHydrogen, planetTempC int) float64 {
	if level <= 0 {
		return 0
	}
	tempFactor := g.HydrogenTempCoefficient*float64(planetTempC) + g.HydrogenTempIntercept
	return math.Floor(g.HydrogenLabBasicProd * float64(level) *
		math.Pow(1.1+float64(techHydrogen)*g.HydrogenLabTechFactor, float64(level)) *
		tempFactor)
}

// SolarPlantProduction — выработка энергии солнечной электростанцией.
//   floor(20 * {level} * pow(1.1 + {tech=18}*0.0005, {level}))
// tech=18 → energy-tech (ENERGY_TECH в legacy).
func SolarPlantProduction(g balance.Globals, level, techEnergy int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(g.SolarPlantBasicProd * float64(level) *
		math.Pow(1.1+float64(techEnergy)*g.SolarPlantTechFactor, float64(level)))
}

// HydrogenPlantProduction — выработка энергии «синтезатором».
//   floor(50 * {level} * pow(1.15 + {tech=18}*0.005, {level}))
func HydrogenPlantProduction(g balance.Globals, level, techEnergy int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(g.HydrogenPlantBasicProd * float64(level) *
		math.Pow(1.15+float64(techEnergy)*g.HydrogenPlantTechFactor, float64(level)))
}

// MineConsEnergy — потребление энергии шахтами/лабами.
//   floor(base * L * pow(1.1 - tech_energy*0.0005, L))
//
// base — base coefficient (10 для metal_mine/silicon_lab, 20 для
// hydrogen_lab; читается из bundle.Globals).
func MineConsEnergy(g balance.Globals, base float64, level, techEnergy int) float64 {
	if level <= 0 {
		return 0
	}
	return math.Floor(base * float64(level) *
		math.Pow(1.1-float64(techEnergy)*g.MineConsEnergyTechFactor, float64(level)))
}

// HydrogenPlantConsHydrogen — потребление водорода синтезатором.
//   floor(10 * L * pow(1.1 - tech_energy*0.0005, L))
func HydrogenPlantConsHydrogen(g balance.Globals, level, techEnergy int) float64 {
	return MineConsEnergy(g, g.HydrogenPlantConsHydroBase, level, techEnergy)
}
