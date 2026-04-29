package main

// Маппинг legacy unit-кодов (числовые ID из oxsar2 / na_construction +
// na_artefact_type) на nova unit_id (числа из projects/game-nova/configs/
// {buildings,research,units,defense,artefacts}.yml).
//
// План 64 копировал legacy-балансы 1-в-1, поэтому ID совпадают (1=metal_mine,
// 13=spyware и т.д.). Этот файл — явное закрепление контракта: если в
// configs/*.yml кто-то изменит id, импортёр продолжит ставить корректные
// nova unit_id. Mapping_test.go проверяет, что каждое значение существует
// в YAML-каталоге.
//
// Источник legacy ID: snapshot test-юзера + legacy oxsar2 docs.

// BuildingMapping — обычные здания на планетах (na_building2planet с
// buildingid в диапазонах 1-12, 53, 100-101, 107-108, 58).
var BuildingMapping = map[int]int{
	1:   1,   // metal_mine
	2:   2,   // silicon_lab
	3:   3,   // hydrogen_lab
	4:   4,   // solar_plant
	5:   5,   // hydrogen_plant
	6:   6,   // robotic_factory
	7:   7,   // nano_factory
	8:   8,   // shipyard
	9:   9,   // metal_storage
	10:  10,  // silicon_storage
	11:  11,  // hydrogen_storage
	12:  12,  // research_lab
	53:  53,  // rocket_station
	58:  58,  // terra_former
	100: 100, // repair_factory
	101: 101, // defense_factory
	107: 107, // exchange
	108: 108, // exch_office
}

// MoonBuildingMapping — лунные здания (54-57, 326, 350-351). В nova живут
// в той же таблице buildings и используются те же id, отличаются только
// тем что строятся на planet с is_moon=true.
var MoonBuildingMapping = map[int]int{
	54:  54,  // moon_base
	55:  55,  // star_surveillance
	56:  56,  // star_gate
	57:  57,  // moon_robotic_factory
	326: 326, // moon_hydrogen_lab
	350: 350, // moon_lab
	351: 351, // moon_repair_factory
}

// ResearchMapping — исследования (na_research2user.buildingid). 13-28,
// 103-104, 111-113. legacy id=111 устаревший, в nova нет (конструктор
// заменён на astro_tech=112); 111 → 112 для совместимости.
var ResearchMapping = map[int]int{
	13:  13,  // spyware
	14:  14,  // computer_tech
	15:  15,  // gun_tech
	16:  16,  // shield_tech
	17:  17,  // shell_tech
	18:  18,  // energy_tech
	19:  19,  // hyperspace_tech
	20:  20,  // combustion_engine
	21:  21,  // impulse_engine
	22:  22,  // hyperspace_engine
	23:  23,  // laser_tech
	24:  24,  // ion_tech
	25:  25,  // plasma_tech
	26:  26,  // ign (alliance network)
	27:  27,  // expo_tech
	28:  28,  // gravi
	103: 103, // ballistics_tech
	104: 104, // masking_tech
	111: 112, // legacy "astro_basic" → nova astro_tech (план 61)
	112: 112, // astro_tech (если уже новый id)
	113: 113, // igr_tech
}

// FleetMapping — корабли (na_unit2shipyard.unitid в диапазоне 29-42, 102,
// 200-204, 325, 352, 353, 358). 51, 52 — ракеты (см. RocketMapping ниже),
// 354, 355 — planet shields (DefenseMapping).
//
// Legacy unitid также обозначает оборону на тех же таблицах — 43-50.
// Различение: defense vs ship определяется по тому, в какой ветке кода
// (mode=4 для defense). Здесь FleetMapping даёт только id, импортёр
// решает по диапазону.
var FleetMapping = map[int]int{
	29:  29,  // small_transporter
	30:  30,  // large_transporter
	31:  31,  // light_fighter
	32:  32,  // strong_fighter
	33:  33,  // cruiser
	34:  34,  // battle_ship
	35:  35,  // frigate
	36:  36,  // colony_ship
	37:  37,  // recycler
	38:  38,  // espionage_sensor
	39:  39,  // solar_satellite
	40:  40,  // bomber
	41:  41,  // star_destroyer
	42:  42,  // death_star
	102: 102, // lancer_ship
	200: 200, // unit_a_corvette (alien)
	201: 201, // unit_a_screen
	202: 202, // unit_a_paladin
	203: 203, // unit_a_frigate
	204: 204, // unit_a_torpedocarier
	325: 325, // shadow_ship
	352: 352, // ship_transplantator
	353: 353, // ship_collector
	358: 358, // armored_terran
}

// DefenseMapping — оборона (na_unit2shipyard.unitid в диапазоне 43-50).
// 51, 52 — ракеты-перехватчики/межпланетные (в nova относятся к units.fleet,
// см. RocketMapping). 354, 355 — planetary shields.
var DefenseMapping = map[int]int{
	43: 43, // rocket_launcher
	44: 44, // light_laser
	45: 45, // strong_laser
	46: 46, // ion_gun
	47: 47, // gauss_gun
	48: 48, // plasma_gun
	49: 49, // small_shield
	50: 50, // large_shield
}

// RocketMapping — межпланетные ракеты (id 51-52). В legacy лежат в
// na_unit2shipyard вместе с обороной, но по семантике летают как fleet.
// В nova каталог числит их в fleet (units.yml), таблица — ships (план 22).
var RocketMapping = map[int]int{
	51: 51, // interceptor_rocket
	52: 52, // interplanetary_rocket
}

// PlanetShieldMapping — планетарные щиты (354, 355). В legacy объявлены в
// каталоге, но не реализованы (план 22 Ф.2.2). В nova каталог имеет id,
// в таблице defense могут лежать строки. Переносим как defense.
var PlanetShieldMapping = map[int]int{
	354: 354, // small_planet_shield
	355: 355, // large_planet_shield
}

// ArtefactMapping — типы артефактов (na_artefact2user.typeid → nova unit_id
// в configs/artefacts.yml). Только подмножество, явно перечисленное в
// nova-каталоге. Legacy snapshot содержит typeid=300-364; для отсутствующих
// в nova типов импортёр пропускает строку с warning — это R15-trade-off,
// см. docs/simplifications.md (план 81/82 — расширение каталога артефактов).
var ArtefactMapping = map[int]int{
	300: 300, // merchants_mark
	301: 301, // catalyst
	302: 302, // power_generator
	303: 303, // atomic_densifier
	304: 304, // (legacy "iron_curtain" — в nova нет; legacy id, см. simplifications)
	305: 305, // supercomputer
	315: 315, // robot_control_system
	316: 316, // battle_shell_power
	317: 317, // battle_shield_power
	318: 318, // battle_attack_power
	// 319-364 — большинство в legacy, nova содержит только 359-361
	// (battle_*_power_10). Остальные пропускаем (см. log warning).
	359: 359, // battle_shell_power_10
	360: 360, // battle_shield_power_10
	361: 361, // battle_attack_power_10
}

// ProfessionMapping — числовой код na_user.profession → nova users.profession
// (TEXT). 0 = none, 1 = miner, 2 = attacker, 3 = defender, 4 = tank
// (см. legacy oxsar2 Profession.class.php + nova configs/professions.yml).
var ProfessionMapping = map[int]string{
	0: "none",
	1: "miner",
	2: "attacker",
	3: "defender",
	4: "tank",
}

// LegacyPlanetPicture → nova planet_type. legacy `na_planet.picture`
// хранит binary string ("dschjungelplanet05", "moon", "wasserplanet09"...).
// nova `planets.planet_type` — нормализованный тип без суффикса номера.
//
// Все типы из legacy oxsar2 / planet_pictures + ext-перекрытий.
var LegacyPlanetTypeMapping = map[string]string{
	"dschjungelplanet": "dschjungelplanet",
	"normaltempplanet": "normaltempplanet",
	"wasserplanet":     "wasserplanet",
	"wuestenplanet":    "wuestenplanet",
	"trockenplanet":    "trockenplanet",
	"eisplanet":        "eisplanet",
	"gasplanet":        "gasplanet",
	"mond":             "moon", // Луна → moon
}
