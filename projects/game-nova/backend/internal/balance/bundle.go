package balance

import "oxsar/game-nova/internal/config"

// Bundle — балансовый bundle конкретной вселенной (план 64).
//
// Содержит:
//   - Catalog — все справочники (buildings, units, rapidfire, ...) после
//     применения override (если был). Совместим по API с config.Catalog,
//     поэтому существующие сервисы продолжают использовать те же поля.
//   - Globals — глобальные коэффициенты для динамических формул
//     (basic-prod, температурные коэффициенты водорода, ...). Modern-
//     вселенные используют ModernGlobals(); origin-вселенная может
//     переопределить через override-файл.
//   - UniverseID — идентификатор вселенной, для которой загружен bundle.
//   - HasOverride — true если был применён override-файл
//     (configs/balance/<id>.yaml). Для modern-вселенных — false.
//
// Bundle — immutable: после загрузки не модифицируется. Замена баланса
// — только через перезапуск процесса (см. план 64, секция «Что меняем»
// №4 — кеш per-universe in-memory, инвалидация через restart; будущая
// admin-ручка — см. risks таблицу).
type Bundle struct {
	UniverseID  string
	HasOverride bool
	Catalog     *config.Catalog
	Globals     Globals
}

// Globals — глобальные балансовые коэффициенты (план 64 §B3).
//
// Используются динамическими формулами production, которые зависят от
// runtime-контекста (tech-уровни, температура планеты). Числа взяты
// из economy/formulas.go (источник истины — modern nova) и хранятся
// здесь, чтобы:
//
//  1. Для origin-вселенной их можно было override через configs/
//     balance/origin.yaml (R0: modern-числа в коде не меняются, всё
//     отличие — через bundle).
//  2. Динамические Go-функции (internal/origin/economy/*) читали
//     коэффициенты из bundle, а не хардкодили.
//
// Имена полей соответствуют формулам oxsar2 (см. docs/research/origin-vs-
// nova/formula-dsl.md):
//
//   - MetalMineBasicProd          = 30  (formula: 30*L*pow(1.1+tech*0.0006,L))
//   - SiliconLabBasicProd         = 20  (20*L*pow(1.1+tech*0.0007,L))
//   - HydrogenLabBasicProd        = 10  (10*L*pow(1.1+tech*0.0008,L)*temp_factor)
//   - SolarPlantBasicProd         = 20  (20*L*pow(1.1+tech*0.0005,L))
//   - HydrogenPlantBasicProd      = 50  (50*L*pow(1.15+tech*0.005,L))
//   - MoonHydrogenLabBasicProd    = 100 (для лунного синтезатора)
//
//   - MetalMineTechFactor         = 0.0006 (множитель к 1.1 от уровня
//                                          лазерной техн.)
//   - SiliconLabTechFactor        = 0.0007 (от уровня кремниевой техн.)
//   - HydrogenLabTechFactor       = 0.0008 (от уровня водородной техн.)
//   - SolarPlantTechFactor        = 0.0005 (от уровня энерго-техн.)
//   - HydrogenPlantTechFactor     = 0.005  (от уровня энерго-техн.)
//
//   - HydrogenTempCoefficient     = -0.002 (per-degree множитель в формуле
//                                          водорода)
//   - HydrogenTempIntercept       = 1.28   (свободный член формулы
//                                          водорода — temp_factor =
//                                          coefficient*temp + intercept)
//
//   - MineConsBase                = (10, 10, 20) — базовое потребление
//                                  энергии (metal_mine, silicon_lab,
//                                  hydrogen_lab; формула:
//                                  base*L*pow(1.1-tech_energy*0.0005,L))
//   - MineConsEnergyTechFactor    = 0.0005
type Globals struct {
	MetalMineBasicProd       float64
	SiliconLabBasicProd      float64
	HydrogenLabBasicProd     float64
	SolarPlantBasicProd      float64
	HydrogenPlantBasicProd   float64
	MoonHydrogenLabBasicProd float64

	MetalMineTechFactor       float64
	SiliconLabTechFactor      float64
	HydrogenLabTechFactor     float64
	SolarPlantTechFactor      float64
	HydrogenPlantTechFactor   float64
	MineConsEnergyTechFactor  float64

	HydrogenTempCoefficient float64
	HydrogenTempIntercept   float64

	MetalMineConsEnergyBase    float64
	SiliconLabConsEnergyBase   float64
	HydrogenLabConsEnergyBase  float64
	MoonHydrogenLabConsBase    float64
	HydrogenPlantConsHydroBase float64
}

// ModernGlobals — глобальные коэффициенты modern-вселенных (uni01,
// uni02 и любых будущих modern). Соответствуют existing nova-формулам
// в internal/economy/formulas.go. R0: эти числа не меняются.
func ModernGlobals() Globals {
	return Globals{
		MetalMineBasicProd:       30,
		SiliconLabBasicProd:      20,
		HydrogenLabBasicProd:     10,
		SolarPlantBasicProd:      20,
		HydrogenPlantBasicProd:   50,
		MoonHydrogenLabBasicProd: 100,

		MetalMineTechFactor:      0.0006,
		SiliconLabTechFactor:     0.0007,
		HydrogenLabTechFactor:    0.0008,
		SolarPlantTechFactor:     0.0005,
		HydrogenPlantTechFactor:  0.005,
		MineConsEnergyTechFactor: 0.0005,

		HydrogenTempCoefficient: -0.002,
		HydrogenTempIntercept:   1.28,

		MetalMineConsEnergyBase:    10,
		SiliconLabConsEnergyBase:   10,
		HydrogenLabConsEnergyBase:  20,
		MoonHydrogenLabConsBase:    200,
		HydrogenPlantConsHydroBase: 10,
	}
}
