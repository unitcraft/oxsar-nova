package main

import "strings"

// Mode из na_construction:
//   1 = building
//   2 = research
//   3 = ship (fleet)
//   4 = defense
//   5+ = вспомогательные/служебные (часто moon-buildings)
//
// (Точные значения — d:/Sources/oxsar2/www/include/consts.php const MODE_*)

const (
	modeBuilding = 1
	modeResearch = 2
	modeShip     = 3
	modeDefense  = 4
)

// Mapping origin-CAPSCASE-имени → nova-snake_case-ключа делается через
// novaKeyForName (по name, не по ID): origin numeric ID не уникален между
// mode'ами (mode=1 ID=23 — moon_shipyard, mode=4 ID=23 — defense), поэтому
// lookup по имени надёжнее. R1: nova-стиль, не калька legacy.

// novaKeyForName: для buildings/research — origin name приходит
// в виде CAPSCASE/SCREAMING_SNAKE. Просто lower-case, и проверяем,
// существует ли такой ключ среди nova. Используется для buildings/research
// где numeric ID не уникален между mode'ами (mode=1 ID=23 — moon_shipyard;
// mode=4 ID=23 — defense). Поэтому в originIDToNovaKey есть конфликты,
// и для надёжности на mode-сегмент ID не полагаемся.
//
// При выборе ключа предпочтительнее lookup по name, чем по ID.
func novaKeyForName(originName string) string {
	switch strings.ToUpper(strings.TrimSpace(originName)) {
	case "METALMINE":
		return "metal_mine"
	case "SILICON_LAB":
		return "silicon_lab"
	case "HYDROGEN_LAB":
		return "hydrogen_lab"
	case "SOLAR_PLANT":
		return "solar_plant"
	case "HYDROGEN_PLANT":
		return "hydrogen_plant"
	case "ROBOTIC_FACTORY":
		return "robotic_factory"
	case "NANO_FACTORY":
		return "nano_factory"
	case "SHIPYARD":
		return "shipyard"
	case "DEFENSE_FACTORY":
		return "defense_factory"
	case "REPAIR_FACTORY":
		return "repair_factory"
	case "METAL_STORAGE":
		return "metal_storage"
	case "SILICON_STORAGE":
		return "silicon_storage"
	case "HYDROGEN_STORAGE":
		return "hydrogen_storage"
	case "RESEARCH_LAB":
		return "research_lab"
	case "EXCH_OFFICE":
		return "exch_office"
	case "EXCHANGE":
		return "exchange"
	case "ROCKET_STATION":
		return "rocket_station"
	case "TERRA_FORMER":
		return "terra_former"
	case "MOON_BASE":
		return "moon_base"
	case "STAR_SURVEILLANCE":
		return "star_surveillance"
	case "STAR_GATE":
		return "star_gate"
	case "MOON_LAB":
		return "moon_lab"
	case "MOON_REPAIR_FACTORY":
		return "moon_repair_factory"
	case "MOON_HYDROGEN_LAB":
		return "moon_hydrogen_lab"
	case "MOON_ROBOTIC_FACTORY":
		return "moon_robotic_factory"
	case "MOON_SHIPYARD":
		return "moon_shipyard"
	case "SPYWARE":
		return "spyware"

	// Research keys
	case "COMPUTER_TECH":
		return "computer_tech"
	case "GUN_TECH":
		return "gun_tech"
	case "SHIELD_TECH":
		return "shield_tech"
	case "SHELL_TECH":
		return "shell_tech"
	case "ENERGY_TECH":
		return "energy_tech"
	case "HYPERSPACE_TECH":
		return "hyperspace_tech"
	case "COMBUSTION_ENGINE":
		return "combustion_engine"
	case "IMPULSE_ENGINE":
		return "impulse_engine"
	case "HYPERSPACE_ENGINE":
		return "hyperspace_engine"
	case "LASER_TECH":
		return "laser_tech"
	case "ION_TECH":
		return "ion_tech"
	case "PLASMA_TECH":
		return "plasma_tech"
	case "IGN":
		return "ign"
	case "EXPO_TECH":
		return "expo_tech"
	case "GRAVI":
		return "gravi"
	case "BALLISTICS_TECH":
		return "ballistics_tech"
	case "MASKING_TECH":
		return "masking_tech"
	case "ASTRO_TECH":
		return "astro_tech"
	case "IGR_TECH":
		return "igr_tech"
	case "ARTEFACTS_TECH":
		return "artefacts_tech"

	// Standard ships
	case "SMALL_TRANSPORTER":
		return "small_transporter"
	case "LARGE_TRANSPORTER":
		return "large_transporter"
	case "LIGHT_FIGHTER":
		return "light_fighter"
	case "STRONG_FIGHTER":
		return "strong_fighter"
	case "CRUISER":
		return "cruiser"
	case "BATTLESHIP":
		return "battleship"
	case "COLONISATOR":
		return "colonisator"
	case "RECYCLER":
		return "recycler"
	case "ESPIONAGE_PROBE":
		return "espionage_probe"
	case "SOLAR_SATELLITE":
		return "solar_satellite"
	case "DEATH_STAR":
		return "death_star"
	case "STAR_DESTROYER":
		return "star_destroyer"
	case "BATTLE_CRUISER":
		return "battle_cruiser"
	case "BOMBER":
		return "bomber"

	// Defense
	case "ROCKET_LAUNCHER":
		return "rocket_launcher"
	case "LIGHT_LASER":
		return "light_laser"
	case "STRONG_LASER":
		return "strong_laser"
	case "ION_CANNON":
		return "ion_cannon"
	case "GAUSS_CANNON":
		return "gauss_cannon"
	case "PLASMA_STORM":
		return "plasma_storm"
	case "SMALL_SHIELD":
		return "small_shield"
	case "LARGE_SHIELD":
		return "large_shield"
	case "INTERCEPTOR_ROCKET":
		return "interceptor_rocket"
	case "INTERPLANETARY_ROCKET":
		return "interplanetary_rocket"

	// Spec — план 64 R0-исключение
	case "LANCER_SHIP":
		return "lancer_ship"
	case "SHADOW_SHIP":
		return "shadow_ship"
	case "SHIP_TRANSPLANTATOR":
		return "ship_transplantator"
	case "SHIP_COLLECTOR":
		return "ship_collector"
	case "SMALL_PLANET_SHIELD":
		return "small_planet_shield"
	case "LARGE_PLANET_SHIELD":
		return "large_planet_shield"
	case "SHIP_ARMORED_TERRAN":
		return "armored_terran"

	// Alien AI-флот
	case "UNIT_A_CORVETTE":
		return "alien_unit_1"
	case "UNIT_A_SCREEN":
		return "alien_unit_2"
	case "UNIT_A_PALADIN":
		return "alien_unit_3"
	case "UNIT_A_FRIGATE":
		return "alien_unit_4"
	case "UNIT_A_TORPEDOCARIER":
		return "alien_unit_5"

	default:
		// Fallback: lower + replace. Не идеально, но логируется в импортёре
		// для последующей ручной правки.
		return strings.ToLower(strings.ReplaceAll(originName, "-", "_"))
	}
}

// novaResearchKeys — ключи исследований, существующие в nova default
// configs/research.yml. Origin может содержать research-ключи, которых
// в nova нет (например ARTEFACTS_TECH); override бессмыслен — skip.
// novaBuildingKeys — ключи зданий, существующие в nova default
// configs/buildings.yml. Origin может содержать moon_lab/moon_repair_factory
// и т.п. — известные knownOrphans (см. internal/config/catalog_validate_test.go),
// override для них не имеет смысла. Также SPYWARE из origin — это
// research-юнит mode=2, не building.
var novaBuildingKeys = map[string]bool{
	"metal_mine":           true,
	"silicon_lab":          true,
	"hydrogen_lab":         true,
	"solar_plant":          true,
	"hydrogen_plant":       true,
	"robotic_factory":      true,
	"nano_factory":         true,
	"shipyard":             true,
	"metal_storage":        true,
	"silicon_storage":      true,
	"hydrogen_storage":     true,
	"research_lab":         true,
	"rocket_station":       true,
	"repair_factory":       true,
	"defense_factory":      true,
	"terra_former":         true,
	"exch_office":          true,
	"exchange":             true,
	"moon_base":            true,
	"star_surveillance":    true,
	"star_gate":            true,
	"moon_robotic_factory": true,
}

func isKnownNovaBuilding(key string) bool {
	return novaBuildingKeys[key]
}

var novaResearchKeys = map[string]bool{
	"computer_tech":     true,
	"gun_tech":          true,
	"shield_tech":       true,
	"shell_tech":        true,
	"energy_tech":       true,
	"hyperspace_tech":   true,
	"combustion_engine": true,
	"impulse_engine":    true,
	"hyperspace_engine": true,
	"laser_tech":        true,
	"ion_tech":          true,
	"plasma_tech":       true,
	"ign":               true,
	"expo_tech":         true,
	"gravi":             true,
	"ballistics_tech":   true,
	"masking_tech":      true,
	"astro_tech":        true,
	"igr_tech":          true,
}

func isKnownNovaResearch(key string) bool {
	return novaResearchKeys[key]
}

// alienUnitIDs — origin-IDs алиен-флота (UNIT_A_*). Используются при
// генерации добавок в дефолтный configs/units.yml + ships.yml +
// rapidfire.yml.
var alienUnitIDs = []int{200, 201, 202, 203, 204}

// specialUnitIDs — origin-IDs спец-юнитов, которые игроки могут строить
// (план 64 R0-исключение, после посещения AlienAI-флотом).
var specialUnitIDs = []int{102, 325, 352, 353, 354, 355, 358}

// isAlienOrSpecial возвращает true если unitid — алиен (UNIT_A_*) или
// спец (Lancer/Shadow/...). Эти юниты идут в дефолтные конфиги, не в
// origin override (R0-исключение).
func isAlienOrSpecial(unitID int) bool {
	for _, id := range alienUnitIDs {
		if id == unitID {
			return true
		}
	}
	for _, id := range specialUnitIDs {
		if id == unitID {
			return true
		}
	}
	return false
}
