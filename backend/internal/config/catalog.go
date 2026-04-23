package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Catalog — загруженные YAML-справочники. Immutable, загружается один раз
// на старте процесса. Замена — только через передеплой (§10.1 ТЗ).
type Catalog struct {
	Units        UnitsCatalog        `yaml:"-"`
	Buildings    BuildingCatalog     `yaml:"-"`
	Research     ResearchCatalog     `yaml:"-"`
	Ships        ShipCatalog         `yaml:"-"`
	Defense      DefenseCatalog      `yaml:"-"`
	Rapidfire    RapidfireCatalog    `yaml:"-"`
	Requirements RequirementsCatalog `yaml:"-"`
	Artefacts    ArtefactCatalog     `yaml:"-"`
	Construction ConstructionCatalog `yaml:"-"`
}

// ConstructionCatalog — данные из legacy-таблицы na_construction,
// сгенерированные `import-datasheets` в configs/construction.yml
// (§1.4, §5.2.1 ТЗ).
//
// Источник истины для баланса. Если загружен — предпочтительнее
// BuildingCatalog: в BuildingCatalog живёт «приближение» из
// configs/buildings.yml, в ConstructionCatalog — исходные формулы
// legacy.
type ConstructionCatalog struct {
	Buildings map[string]ConstructionSpec `yaml:"buildings"`
}

// ConstructionSpec — одна запись na_construction в виде YAML.
// Поля prod/cons/charge хранят формулы как строки — только для документации;
// runtime-расчёт идёт через статические функции economy (план 16).
type ConstructionSpec struct {
	ID           int64          `yaml:"id"`
	Mode         int64          `yaml:"mode"`
	LegacyName   string         `yaml:"legacy_name"`
	Front        int64          `yaml:"front,omitempty"`
	Ballistics   int64          `yaml:"ballistics,omitempty"`
	Masking      int64          `yaml:"masking,omitempty"`
	Basic        ConstructionBasic    `yaml:"basic,omitempty"`
	Prod         ConstructionFormulas `yaml:"prod,omitempty"`
	Cons         ConstructionFormulas `yaml:"cons,omitempty"`
	Charge       ConstructionFormulas `yaml:"charge,omitempty"`
	ChargeCredit string               `yaml:"charge_credit,omitempty"`
	Demolish     float64              `yaml:"demolish,omitempty"`
	DisplayOrder int64                `yaml:"display_order,omitempty"`
}

type ConstructionBasic struct {
	Metal    int64 `yaml:"metal,omitempty"`
	Silicon  int64 `yaml:"silicon,omitempty"`
	Hydrogen int64 `yaml:"hydrogen,omitempty"`
	Energy   int64 `yaml:"energy,omitempty"`
	Credit   int64 `yaml:"credit,omitempty"`
}

type ConstructionFormulas struct {
	Metal    string `yaml:"metal,omitempty"`
	Silicon  string `yaml:"silicon,omitempty"`
	Hydrogen string `yaml:"hydrogen,omitempty"`
	Energy   string `yaml:"energy,omitempty"`
}

type UnitsCatalog struct {
	Buildings     []UnitEntry `yaml:"buildings"`
	MoonBuildings []UnitEntry `yaml:"moon_buildings"`
	Research      []UnitEntry `yaml:"research"`
	Fleet         []UnitEntry `yaml:"fleet"`
	Defense       []UnitEntry `yaml:"defense"`
}

type UnitEntry struct {
	ID   int    `yaml:"id"`
	Key  string `yaml:"key"`
	Name string `yaml:"name"`
}

type BuildingCatalog struct {
	Buildings map[string]BuildingSpec `yaml:"buildings"`
}

type BuildingSpec struct {
	ID                      int      `yaml:"id"`
	CostBase                ResCost  `yaml:"cost_base"`
	CostFactor              float64  `yaml:"cost_factor"`
	TimeBaseSeconds         int      `yaml:"time_base_seconds"`
	BaseRatePerHour         *float64 `yaml:"base_rate_per_hour,omitempty"`
	EnergyPerLevel          *float64 `yaml:"energy_per_level,omitempty"`
	EnergyOutputPerLevel    *float64 `yaml:"energy_output_per_level,omitempty"`
	CapacityBase            *int64   `yaml:"capacity_base,omitempty"`
	RocketCapacityPerLevel  *int64   `yaml:"rocket_capacity_per_level,omitempty"`
	MoonOnly                bool     `yaml:"moon_only,omitempty"`
}

type ResCost struct {
	Metal    int64 `yaml:"metal"`
	Silicon  int64 `yaml:"silicon"`
	Hydrogen int64 `yaml:"hydrogen"`
}

type ShipCatalog struct {
	Ships map[string]ShipSpec `yaml:"ships"`
}

type ShipSpec struct {
	ID     int     `yaml:"id"`
	Cost   ResCost `yaml:"cost"`
	Attack int     `yaml:"attack"`
	Shield int     `yaml:"shield"`
	Shell  int     `yaml:"shell"`
	Cargo  int64   `yaml:"cargo"`
	Speed  int     `yaml:"speed"`
	Fuel   int     `yaml:"fuel"`
}

type DefenseCatalog struct {
	Defense map[string]DefenseSpec `yaml:"defense"`
}

type DefenseSpec struct {
	ID     int     `yaml:"id"`
	Cost   ResCost `yaml:"cost"`
	Attack int     `yaml:"attack"`
	Shield int     `yaml:"shield"`
	Shell  int     `yaml:"shell"`
}

// RapidfireCatalog — table[shooter][target] = multiplier.
type RapidfireCatalog struct {
	Rapidfire map[int]map[int]int `yaml:"rapidfire"`
}

// ResearchCatalog — баланс исследований. Формула стоимостей та же, что
// и у зданий: cost = base * factor^(level-1).
type ResearchCatalog struct {
	Research map[string]ResearchSpec `yaml:"research"`
}

type ResearchSpec struct {
	ID         int     `yaml:"id"`
	CostBase   ResCost `yaml:"cost_base"`
	CostFactor float64 `yaml:"cost_factor"`
}

// RequirementsCatalog — зависимости юнитов. key — ключ цели (например,
// "cruiser"), значение — список требований.
type RequirementsCatalog struct {
	Requirements map[string][]Requirement `yaml:"requirements"`
}

// Requirement — одно требование.
//
// kind = building|research (fleet-требования тут тоже можно задать,
// но в OGame-механике они не нужны: нужны только здания/исследования).
type Requirement struct {
	Kind  string `yaml:"kind"`
	Key   string `yaml:"key"`
	Level int    `yaml:"level"`
}

// ArtefactCatalog — описания артефактов и их эффектов.
// Подробно семантика полей — в configs/artefacts.yml и §5.10 ТЗ.
type ArtefactCatalog struct {
	Artefacts map[string]ArtefactSpec `yaml:"artefacts"`
}

// ArtefactSpec — один артефакт. Содержит идентификатор, эффект и
// метаданные жизненного цикла.
type ArtefactSpec struct {
	ID              int            `yaml:"id"`
	Name            string         `yaml:"name"`
	Effect          ArtefactEffect `yaml:"effect"`
	Stackable       bool           `yaml:"stackable"`
	LifetimeSeconds int            `yaml:"lifetime_seconds"`
	DelaySeconds    int            `yaml:"delay_seconds,omitempty"`
}

// ArtefactEffect — как артефакт меняет состояние.
//
// type = factor_user | factor_planet | factor_all_planets |
//        one_shot | battle_bonus
// op   = set | add
//
// Для op=set используются ActiveValue/InactiveValue (set-эффект на
// всё время активации; деактивация возвращает InactiveValue).
// Для op=add используется Value (прибавка при активации, вычитание
// при деактивации — зеркальная операция).
type ArtefactEffect struct {
	Type           string  `yaml:"type"`
	Field          string  `yaml:"field,omitempty"`
	Op             string  `yaml:"op,omitempty"`
	Value          float64 `yaml:"value,omitempty"`
	ActiveValue    float64 `yaml:"active_value,omitempty"`
	InactiveValue  float64 `yaml:"inactive_value,omitempty"`
	BattleAttack   float64 `yaml:"battle_attack,omitempty"`
	BattleShield   float64 `yaml:"battle_shield,omitempty"`
	BattleShell    float64 `yaml:"battle_shell,omitempty"`
}

// LoadCatalog читает все YAML-справочники из dir. Если файла нет —
// возвращает ошибку (конфиг обязателен, а не optional).
func LoadCatalog(dir string) (*Catalog, error) {
	cat := &Catalog{}
	type pair struct {
		file string
		into any
	}
	for _, p := range []pair{
		{"units.yml", &cat.Units},
		{"buildings.yml", &cat.Buildings},
		{"research.yml", &cat.Research},
		{"ships.yml", &cat.Ships},
		{"defense.yml", &cat.Defense},
		{"rapidfire.yml", &cat.Rapidfire},
		{"requirements.yml", &cat.Requirements},
		{"artefacts.yml", &cat.Artefacts},
	} {
		path := filepath.Join(dir, p.file)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p.file, err)
		}
		if err := yaml.Unmarshal(data, p.into); err != nil {
			return nil, fmt.Errorf("parse %s: %w", p.file, err)
		}
	}

	// Опциональные generated-файлы. Если их нет — экономика работает
	// на приближённых формулах buildings.yml. Появляются после первого
	// прогона `cmd/tools/import-datasheets` (§1.4, §5.2.1 ТЗ).
	optional := []pair{
		{"construction.yml", &cat.Construction},
	}
	for _, p := range optional {
		path := filepath.Join(dir, p.file)
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("read %s: %w", p.file, err)
		}
		if err := yaml.Unmarshal(data, p.into); err != nil {
			return nil, fmt.Errorf("parse %s: %w", p.file, err)
		}
	}
	return cat, nil
}
