package balance

import (
	"fmt"

	"oxsar/game-nova/internal/config"
)

// override — типизированная схема override-файла configs/balance/<id>.yaml.
//
// Поля все указатели или nullable map'ы — отсутствие = «оставить дефолт».
// Любой ключ внутри override-секции перекрывает соответствующий ключ
// дефолтного catalog (deep merge на уровне записей; merge внутри одной
// записи — поле в поле).
//
// Структура соответствует типам из config: BuildingSpec, ShipSpec,
// DefenseSpec, ResearchSpec и т. п. Ключи map — те же string-ключи,
// что в дефолтных configs/buildings.yml, units.yml и т. д.
type override struct {
	Version  int    `yaml:"version"`
	Universe string `yaml:"universe"`

	Globals      *globalsOverride                       `yaml:"globals,omitempty"`
	Buildings    map[string]*buildingOverride           `yaml:"buildings,omitempty"`
	Research     map[string]*researchOverride           `yaml:"research,omitempty"`
	Ships        map[string]*shipOverride               `yaml:"ships,omitempty"`
	Defense      map[string]*defenseOverride            `yaml:"defense,omitempty"`
	Rapidfire    map[int]map[int]int                    `yaml:"rapidfire,omitempty"`
	Requirements map[string][]config.Requirement        `yaml:"requirements,omitempty"`
	Units        *unitsOverride                         `yaml:"units,omitempty"`
}

// globalsOverride — все поля Globals в виде *float64, чтобы отличать
// «явный 0» от «не задано в YAML».
type globalsOverride struct {
	MetalMineBasicProd       *float64 `yaml:"metal_mine_basic_prod"`
	SiliconLabBasicProd      *float64 `yaml:"silicon_lab_basic_prod"`
	HydrogenLabBasicProd     *float64 `yaml:"hydrogen_lab_basic_prod"`
	SolarPlantBasicProd      *float64 `yaml:"solar_plant_basic_prod"`
	HydrogenPlantBasicProd   *float64 `yaml:"hydrogen_plant_basic_prod"`
	MoonHydrogenLabBasicProd *float64 `yaml:"moon_hydrogen_lab_basic_prod"`

	MetalMineTechFactor      *float64 `yaml:"metal_mine_tech_factor"`
	SiliconLabTechFactor     *float64 `yaml:"silicon_lab_tech_factor"`
	HydrogenLabTechFactor    *float64 `yaml:"hydrogen_lab_tech_factor"`
	SolarPlantTechFactor     *float64 `yaml:"solar_plant_tech_factor"`
	HydrogenPlantTechFactor  *float64 `yaml:"hydrogen_plant_tech_factor"`
	MineConsEnergyTechFactor *float64 `yaml:"mine_cons_energy_tech_factor"`

	HydrogenTempCoefficient *float64 `yaml:"hydrogen_temp_coefficient"`
	HydrogenTempIntercept   *float64 `yaml:"hydrogen_temp_intercept"`

	MetalMineConsEnergyBase    *float64 `yaml:"metal_mine_cons_energy_base"`
	SiliconLabConsEnergyBase   *float64 `yaml:"silicon_lab_cons_energy_base"`
	HydrogenLabConsEnergyBase  *float64 `yaml:"hydrogen_lab_cons_energy_base"`
	MoonHydrogenLabConsBase    *float64 `yaml:"moon_hydrogen_lab_cons_base"`
	HydrogenPlantConsHydroBase *float64 `yaml:"hydrogen_plant_cons_hydrogen_base"`
}

type costOverride struct {
	Metal    *int64 `yaml:"metal"`
	Silicon  *int64 `yaml:"silicon"`
	Hydrogen *int64 `yaml:"hydrogen"`
}

type buildingOverride struct {
	ID                     *int          `yaml:"id"`
	CostBase               *costOverride `yaml:"cost_base"`
	CostFactor             *float64      `yaml:"cost_factor"`
	TimeBaseSeconds        *int          `yaml:"time_base_seconds"`
	BaseRatePerHour        *float64      `yaml:"base_rate_per_hour"`
	EnergyPerLevel         *float64      `yaml:"energy_per_level"`
	EnergyOutputPerLevel   *float64      `yaml:"energy_output_per_level"`
	CapacityBase           *int64        `yaml:"capacity_base"`
	RocketCapacityPerLevel *int64        `yaml:"rocket_capacity_per_level"`
	MoonOnly               *bool         `yaml:"moon_only"`
	MaxLevel               *int          `yaml:"max_level"`
	DisplayOrder           *int          `yaml:"display_order"`
	Demolish               *float64      `yaml:"demolish"`
	ChargeCredit           *string       `yaml:"charge_credit"`
}

type researchOverride struct {
	ID         *int          `yaml:"id"`
	CostBase   *costOverride `yaml:"cost_base"`
	CostFactor *float64      `yaml:"cost_factor"`
}

type shipOverride struct {
	ID     *int          `yaml:"id"`
	Attack *int          `yaml:"attack"`
	Shield *int          `yaml:"shield"`
	Shell  *int          `yaml:"shell"`
	Cargo  *int64        `yaml:"cargo"`
	Speed  *int          `yaml:"speed"`
	Fuel   *int          `yaml:"fuel"`
	Cost   *costOverride `yaml:"cost"`
	Front  *int          `yaml:"front"`
}

type defenseOverride struct {
	ID     *int          `yaml:"id"`
	Cost   *costOverride `yaml:"cost"`
	Attack *int          `yaml:"attack"`
	Shield *int          `yaml:"shield"`
	Shell  *int          `yaml:"shell"`
	Front  *int          `yaml:"front"`
}

// unitsOverride — добавки в реестр units.yml. Не поддерживает удаление —
// только append entries в группы (R0: дефолт не теряется). Используется
// для R0-исключения: алиен/спец-юниты добавляются в дефолтный реестр,
// а не в origin override (см. план 64 Ф.2 — там этого нет, но если
// потребуется чисто-origin юнит, можно завести через эту секцию).
type unitsOverride struct {
	Buildings     []config.UnitEntry `yaml:"buildings"`
	MoonBuildings []config.UnitEntry `yaml:"moon_buildings"`
	Research      []config.UnitEntry `yaml:"research"`
	Fleet         []config.UnitEntry `yaml:"fleet"`
	Defense       []config.UnitEntry `yaml:"defense"`
}

// applyOverride — deep merge override поверх дефолта. Возвращает новый
// Bundle (вход не модифицируется, важная инвариант для in-memory cache).
func applyOverride(def *Bundle, ov *override) (*Bundle, error) {
	cat := cloneCatalog(def.Catalog)
	g := def.Globals

	if ov.Globals != nil {
		applyGlobalsOverride(&g, ov.Globals)
	}

	for key, bo := range ov.Buildings {
		if bo == nil {
			continue
		}
		spec, ok := cat.Buildings.Buildings[key]
		if !ok {
			return nil, fmt.Errorf("buildings.%s: not in default catalog (override can only modify existing keys)", key)
		}
		applyBuildingOverride(&spec, bo)
		cat.Buildings.Buildings[key] = spec
	}

	for key, ro := range ov.Research {
		if ro == nil {
			continue
		}
		spec, ok := cat.Research.Research[key]
		if !ok {
			return nil, fmt.Errorf("research.%s: not in default catalog", key)
		}
		applyResearchOverride(&spec, ro)
		cat.Research.Research[key] = spec
	}

	for key, so := range ov.Ships {
		if so == nil {
			continue
		}
		spec, ok := cat.Ships.Ships[key]
		if !ok {
			return nil, fmt.Errorf("ships.%s: not in default catalog", key)
		}
		applyShipOverride(&spec, so)
		cat.Ships.Ships[key] = spec
	}

	for key, dov := range ov.Defense {
		if dov == nil {
			continue
		}
		spec, ok := cat.Defense.Defense[key]
		if !ok {
			return nil, fmt.Errorf("defense.%s: not in default catalog", key)
		}
		applyDefenseOverride(&spec, dov)
		cat.Defense.Defense[key] = spec
	}

	for shooterID, targets := range ov.Rapidfire {
		dst, ok := cat.Rapidfire.Rapidfire[shooterID]
		if !ok {
			dst = make(map[int]int, len(targets))
			cat.Rapidfire.Rapidfire[shooterID] = dst
		}
		for targetID, mult := range targets {
			dst[targetID] = mult
		}
	}

	for targetKey, reqs := range ov.Requirements {
		if reqs == nil {
			continue
		}
		// requirements в override — полная замена списка для targetKey.
		copied := make([]config.Requirement, len(reqs))
		copy(copied, reqs)
		cat.Requirements.Requirements[targetKey] = copied
	}

	if ov.Units != nil {
		appendUnitEntries(&cat.Units.Buildings, ov.Units.Buildings)
		appendUnitEntries(&cat.Units.MoonBuildings, ov.Units.MoonBuildings)
		appendUnitEntries(&cat.Units.Research, ov.Units.Research)
		appendUnitEntries(&cat.Units.Fleet, ov.Units.Fleet)
		appendUnitEntries(&cat.Units.Defense, ov.Units.Defense)
	}

	return &Bundle{
		Catalog: cat,
		Globals: g,
	}, nil
}

func applyGlobalsOverride(g *Globals, o *globalsOverride) {
	if o.MetalMineBasicProd != nil {
		g.MetalMineBasicProd = *o.MetalMineBasicProd
	}
	if o.SiliconLabBasicProd != nil {
		g.SiliconLabBasicProd = *o.SiliconLabBasicProd
	}
	if o.HydrogenLabBasicProd != nil {
		g.HydrogenLabBasicProd = *o.HydrogenLabBasicProd
	}
	if o.SolarPlantBasicProd != nil {
		g.SolarPlantBasicProd = *o.SolarPlantBasicProd
	}
	if o.HydrogenPlantBasicProd != nil {
		g.HydrogenPlantBasicProd = *o.HydrogenPlantBasicProd
	}
	if o.MoonHydrogenLabBasicProd != nil {
		g.MoonHydrogenLabBasicProd = *o.MoonHydrogenLabBasicProd
	}
	if o.MetalMineTechFactor != nil {
		g.MetalMineTechFactor = *o.MetalMineTechFactor
	}
	if o.SiliconLabTechFactor != nil {
		g.SiliconLabTechFactor = *o.SiliconLabTechFactor
	}
	if o.HydrogenLabTechFactor != nil {
		g.HydrogenLabTechFactor = *o.HydrogenLabTechFactor
	}
	if o.SolarPlantTechFactor != nil {
		g.SolarPlantTechFactor = *o.SolarPlantTechFactor
	}
	if o.HydrogenPlantTechFactor != nil {
		g.HydrogenPlantTechFactor = *o.HydrogenPlantTechFactor
	}
	if o.MineConsEnergyTechFactor != nil {
		g.MineConsEnergyTechFactor = *o.MineConsEnergyTechFactor
	}
	if o.HydrogenTempCoefficient != nil {
		g.HydrogenTempCoefficient = *o.HydrogenTempCoefficient
	}
	if o.HydrogenTempIntercept != nil {
		g.HydrogenTempIntercept = *o.HydrogenTempIntercept
	}
	if o.MetalMineConsEnergyBase != nil {
		g.MetalMineConsEnergyBase = *o.MetalMineConsEnergyBase
	}
	if o.SiliconLabConsEnergyBase != nil {
		g.SiliconLabConsEnergyBase = *o.SiliconLabConsEnergyBase
	}
	if o.HydrogenLabConsEnergyBase != nil {
		g.HydrogenLabConsEnergyBase = *o.HydrogenLabConsEnergyBase
	}
	if o.MoonHydrogenLabConsBase != nil {
		g.MoonHydrogenLabConsBase = *o.MoonHydrogenLabConsBase
	}
	if o.HydrogenPlantConsHydroBase != nil {
		g.HydrogenPlantConsHydroBase = *o.HydrogenPlantConsHydroBase
	}
}

func applyCostOverride(c *config.ResCost, o *costOverride) {
	if o.Metal != nil {
		c.Metal = *o.Metal
	}
	if o.Silicon != nil {
		c.Silicon = *o.Silicon
	}
	if o.Hydrogen != nil {
		c.Hydrogen = *o.Hydrogen
	}
}

func applyBuildingOverride(s *config.BuildingSpec, o *buildingOverride) {
	if o.ID != nil {
		s.ID = *o.ID
	}
	if o.CostBase != nil {
		applyCostOverride(&s.CostBase, o.CostBase)
	}
	if o.CostFactor != nil {
		s.CostFactor = *o.CostFactor
	}
	if o.TimeBaseSeconds != nil {
		s.TimeBaseSeconds = *o.TimeBaseSeconds
	}
	if o.BaseRatePerHour != nil {
		v := *o.BaseRatePerHour
		s.BaseRatePerHour = &v
	}
	if o.EnergyPerLevel != nil {
		v := *o.EnergyPerLevel
		s.EnergyPerLevel = &v
	}
	if o.EnergyOutputPerLevel != nil {
		v := *o.EnergyOutputPerLevel
		s.EnergyOutputPerLevel = &v
	}
	if o.CapacityBase != nil {
		v := *o.CapacityBase
		s.CapacityBase = &v
	}
	if o.RocketCapacityPerLevel != nil {
		v := *o.RocketCapacityPerLevel
		s.RocketCapacityPerLevel = &v
	}
	if o.MoonOnly != nil {
		s.MoonOnly = *o.MoonOnly
	}
	if o.MaxLevel != nil {
		s.MaxLevel = *o.MaxLevel
	}
	if o.DisplayOrder != nil {
		s.DisplayOrder = *o.DisplayOrder
	}
	if o.Demolish != nil {
		s.Demolish = *o.Demolish
	}
	if o.ChargeCredit != nil {
		s.ChargeCredit = *o.ChargeCredit
	}
}

func applyResearchOverride(s *config.ResearchSpec, o *researchOverride) {
	if o.ID != nil {
		s.ID = *o.ID
	}
	if o.CostBase != nil {
		applyCostOverride(&s.CostBase, o.CostBase)
	}
	if o.CostFactor != nil {
		s.CostFactor = *o.CostFactor
	}
}

func applyShipOverride(s *config.ShipSpec, o *shipOverride) {
	if o.ID != nil {
		s.ID = *o.ID
	}
	if o.Attack != nil {
		s.Attack = *o.Attack
	}
	if o.Shield != nil {
		s.Shield = *o.Shield
	}
	if o.Shell != nil {
		s.Shell = *o.Shell
	}
	if o.Cargo != nil {
		s.Cargo = *o.Cargo
	}
	if o.Speed != nil {
		s.Speed = *o.Speed
	}
	if o.Fuel != nil {
		s.Fuel = *o.Fuel
	}
	if o.Cost != nil {
		applyCostOverride(&s.Cost, o.Cost)
	}
	if o.Front != nil {
		s.Front = *o.Front
	}
}

func applyDefenseOverride(s *config.DefenseSpec, o *defenseOverride) {
	if o.ID != nil {
		s.ID = *o.ID
	}
	if o.Cost != nil {
		applyCostOverride(&s.Cost, o.Cost)
	}
	if o.Attack != nil {
		s.Attack = *o.Attack
	}
	if o.Shield != nil {
		s.Shield = *o.Shield
	}
	if o.Shell != nil {
		s.Shell = *o.Shell
	}
	if o.Front != nil {
		s.Front = *o.Front
	}
}

func appendUnitEntries(dst *[]config.UnitEntry, src []config.UnitEntry) {
	if len(src) == 0 {
		return
	}
	seen := make(map[string]struct{}, len(*dst))
	for _, e := range *dst {
		seen[e.Key] = struct{}{}
	}
	for _, e := range src {
		if _, dup := seen[e.Key]; dup {
			continue
		}
		*dst = append(*dst, e)
		seen[e.Key] = struct{}{}
	}
}

// cloneCatalog — shallow-deep клон Catalog. Внутренние map'ы пере-
// создаются (чтобы override не модифицировал shared default-bundle),
// но values map'а — копируются по значению (структуры, не ссылки).
//
// Это критично для immutability default-bundle: LoadDefaults() возвращает
// один и тот же объект всем consumer'ам, override-merge не должен его
// портить.
func cloneCatalog(src *config.Catalog) *config.Catalog {
	dst := &config.Catalog{
		Units: config.UnitsCatalog{
			Buildings:     append([]config.UnitEntry(nil), src.Units.Buildings...),
			MoonBuildings: append([]config.UnitEntry(nil), src.Units.MoonBuildings...),
			Research:      append([]config.UnitEntry(nil), src.Units.Research...),
			Fleet:         append([]config.UnitEntry(nil), src.Units.Fleet...),
			Defense:       append([]config.UnitEntry(nil), src.Units.Defense...),
		},
		Buildings:    config.BuildingCatalog{Buildings: cloneMap(src.Buildings.Buildings)},
		Research:     config.ResearchCatalog{Research: cloneMap(src.Research.Research)},
		Ships:        config.ShipCatalog{Ships: cloneMap(src.Ships.Ships)},
		Defense:      config.DefenseCatalog{Defense: cloneMap(src.Defense.Defense)},
		Rapidfire:    config.RapidfireCatalog{Rapidfire: cloneRapidfire(src.Rapidfire.Rapidfire)},
		Requirements: config.RequirementsCatalog{Requirements: cloneRequirements(src.Requirements.Requirements)},
		Artefacts:    config.ArtefactCatalog{Artefacts: cloneMap(src.Artefacts.Artefacts)},
		Professions:  config.ProfessionCatalog{Professions: cloneProfessions(src.Professions.Professions)},
	}
	return dst
}

func cloneMap[V any](src map[string]V) map[string]V {
	if src == nil {
		return nil
	}
	dst := make(map[string]V, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func cloneRapidfire(src map[int]map[int]int) map[int]map[int]int {
	if src == nil {
		return nil
	}
	dst := make(map[int]map[int]int, len(src))
	for k, inner := range src {
		copy := make(map[int]int, len(inner))
		for ik, iv := range inner {
			copy[ik] = iv
		}
		dst[k] = copy
	}
	return dst
}

func cloneRequirements(src map[string][]config.Requirement) map[string][]config.Requirement {
	if src == nil {
		return nil
	}
	dst := make(map[string][]config.Requirement, len(src))
	for k, v := range src {
		copied := make([]config.Requirement, len(v))
		copy(copied, v)
		dst[k] = copied
	}
	return dst
}

func cloneProfessions(src map[string]config.ProfessionSpec) map[string]config.ProfessionSpec {
	if src == nil {
		return nil
	}
	dst := make(map[string]config.ProfessionSpec, len(src))
	for k, v := range src {
		// Bonus/Malus map'ы — потенциально shared, клонируем для безопасности.
		bonus := make(map[string]int, len(v.Bonus))
		for bk, bv := range v.Bonus {
			bonus[bk] = bv
		}
		malus := make(map[string]int, len(v.Malus))
		for mk, mv := range v.Malus {
			malus[mk] = mv
		}
		dst[k] = config.ProfessionSpec{
			Label: v.Label,
			Bonus: bonus,
			Malus: malus,
		}
	}
	return dst
}
