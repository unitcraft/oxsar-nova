<?php
/**
 * CastleMonsters HTML5 game

 */

defined('OXSAR_VERSION') or define('OXSAR_VERSION', "2.14.2");
defined('CLIENT_VERSION') or define('CLIENT_VERSION', 11);

define('OXSAR_RELEASED', strstr(OXSAR_VERSION, 'beta') || strstr(OXSAR_VERSION, 'dev') ? false : true);

if(1){
	define('SIMULATOR_ASSAULT_JAR', 'Assault.jar');
}else{
	define('SIMULATOR_ASSAULT_JAR', 'Assault-sim.jar');	
}

define("UNIT_METALMINE", 1); // mode - 1
define("UNIT_SILICON_LAB", 2); // mode - 1
define("UNIT_HYDROGEN_LAB", 3); // mode - 1
define("UNIT_SOLAR_PLANT", 4); // mode - 1
define("UNIT_HYDROGEN_PLANT", 5); // mode - 1
define("UNIT_ROBOTIC_FACTORY", 6); // mode - 1
define("UNIT_NANO_FACTORY", 7); // mode - 1
define("UNIT_SHIPYARD", 8); // mode - 1
define("UNIT_METAL_STORAGE", 9); // mode - 1
define("UNIT_SILICON_STORAGE", 10); // mode - 1
define("UNIT_HYDROGEN_STORAGE", 11); // mode - 1
define("UNIT_RESEARCH_LAB", 12); // mode - 1
define("UNIT_SPYWARE", 13); // mode - 2
define("UNIT_COMPUTER_TECH", 14); // mode - 2
define("UNIT_GUN_TECH", 15); // mode - 2
define("UNIT_SHIELD_TECH", 16); // mode - 2
define("UNIT_SHELL_TECH", 17); // mode - 2
define("UNIT_ENERGY_TECH", 18); // mode - 2
define("UNIT_HYPERSPACE_TECH", 19); // mode - 2
define("UNIT_COMBUSTION_ENGINE", 20); // mode - 2
define("UNIT_IMPULSE_ENGINE", 21); // mode - 2
define("UNIT_HYPERSPACE_ENGINE", 22); // mode - 2
define("UNIT_LASER_TECH", 23); // mode - 2
define("UNIT_ION_TECH", 24); // mode - 2
define("UNIT_PLASMA_TECH", 25); // mode - 2
define("UNIT_IGN", 26); // mode - 2
define("UNIT_EXPO_TECH", 27); // mode - 2
define("UNIT_GRAVI", 28); // mode - 2
define("UNIT_SMALL_TRANSPORTER", 29); // mode - 3
define("UNIT_LARGE_TRANSPORTER", 30); // mode - 3
define("UNIT_LIGHT_FIGHTER", 31); // mode - 3
define("UNIT_STRONG_FIGHTER", 32); // mode - 3
define("UNIT_CRUISER", 33); // mode - 3
define("UNIT_BATTLE_SHIP", 34); // mode - 3
define("UNIT_FRIGATE", 35); // mode - 3
define("UNIT_COLONY_SHIP", 36); // mode - 3
define("UNIT_RECYCLER", 37); // mode - 3
define("UNIT_ESPIONAGE_SENSOR", 38); // mode - 3
define("UNIT_SOLAR_SATELLITE", 39); // mode - 3
define("UNIT_BOMBER", 40); // mode - 3
define("UNIT_STAR_DESTROYER", 41); // mode - 3
define("UNIT_DEATH_STAR", 42); // mode - 3
define("UNIT_ROCKET_LAUNCHER", 43); // mode - 4
define("UNIT_LIGHT_LASER", 44); // mode - 4
define("UNIT_STRONG_LASER", 45); // mode - 4
define("UNIT_ION_GUN", 46); // mode - 4
define("UNIT_GAUSS_GUN", 47); // mode - 4
define("UNIT_PLASMA_GUN", 48); // mode - 4
define("UNIT_SMALL_SHIELD", 49); // mode - 4
define("UNIT_LARGE_SHIELD", 50); // mode - 4
define("UNIT_INTERCEPTOR_ROCKET", 51); // mode - 4
define("UNIT_INTERPLANETARY_ROCKET", 52); // mode - 4
define("UNIT_ROCKET_STATION", 53); // mode - 1
define("UNIT_MOON_BASE", 54); // mode - 5
define("UNIT_STAR_SURVEILLANCE", 55); // mode - 5
define("UNIT_STAR_GATE", 56); // mode - 5
define("UNIT_MOON_ROBOTIC_FACTORY", 57); // mode - 5
define("UNIT_TERRA_FORMER", 58); // mode - 1
define("UNIT_REPAIR_FACTORY", 100); // mode - 1
define("UNIT_DEFENSE_FACTORY", 101); // mode - 1
define("UNIT_LANCER_SHIP", 102); // mode - 3
define("UNIT_BALLISTICS_TECH", 103); // mode - 2
define("UNIT_MASKING_TECH", 104); // mode - 2
define("UNIT_ALIEN_TECH", 110); // mode - 2
define("UNIT_ARTEFACTS_TECH", 111); // mode - 2
define("UNIT_SHADOW_SHIP", 325); // mode - 3
define("UNIT_MOON_HYDROGEN_LAB", 326); // mode - 5
define("UNIT_MOON_LAB", 350); // mode - 5
define("UNIT_MOON_REPAIR_FACTORY", 351); // mode - 5
define("UNIT_SHIP_TRANSPLANTATOR", 352); // mode - 3
define("UNIT_SHIP_COLLECTOR", 353); // mode - 3
define("UNIT_SMALL_PLANET_SHIELD", 354); // mode - 4
define("UNIT_LARGE_PLANET_SHIELD", 355); // mode - 4
define("UNIT_SHIP_ARMORED_TERRAN", 358); // mode - 4

define("UNIT_EXCHANGE", 107); // mode - 1
define("UNIT_EXCH_OFFICE", 108); // mode - 1
define("UNIT_EXCH_SUPPORT_RANGE", 105); // mode - 3
define("UNIT_EXCH_SUPPORT_SLOT", 106); // mode - 4

define("UNIT_A_CORVETTE", 200);
define("UNIT_A_SCREEN", 201);
define("UNIT_A_PALADIN", 202);
define("UNIT_A_FRIGATE", 203);
define("UNIT_A_TORPEDOCARIER", 204);

define("ARTEFACT_MERCHANTS_MARK", 300); // mode - 6
define("ARTEFACT_CATALYST", 301); // mode - 6
define("ARTEFACT_POWER_GENERATOR", 302); // mode - 6
define("ARTEFACT_ATOMIC_DENSIFIER", 303); // mode - 6
define("ARTEFACT_ANNIHILATION_ENGINE", 304); // mode - 6
define("ARTEFACT_SUPERCOMPUTER", 305); // mode - 6
define("ARTEFACT_MEMORY_MODULE__28_1", 306); // mode - 6
define("ARTEFACT_MEMORY_MODULE__23_12", 307); // mode - 6
define("ARTEFACT_MEMORY_MODULE__15_15", 308); // mode - 6
define("ARTEFACT_MEMORY_MODULE__16_15", 309); // mode - 6
define("ARTEFACT_MEMORY_MODULE__17_15", 310); // mode - 6
define("ARTEFACT_MEMORY_MODULE__110_1", 311); // mode - 6
define("ARTEFACT_ASSEMBLY_MODULE__31_10", 312); // mode - 6
define("ARTEFACT_ASSEMBLY_MODULE__42_1", 313); // mode - 6
define("ARTEFACT_REPAIR_BOT", 314); // mode - 6
define("ARTEFACT_ROBOT_CONTROL_SYSTEM", 315); // mode - 6
define("ARTEFACT_BATTLE_SHELL_POWER", 316); // mode - 6 ???
define("ARTEFACT_BATTLE_SHIELD_POWER", 317); // mode - 6 ???
define("ARTEFACT_BATTLE_ATTACK_POWER", 318); // mode - 6 ???
define("ARTEFACT_BATTLE_NEUTRON_AFFECTOR", 356); // mode - 6 ???
define("ARTEFACT_MOON_CREATOR", 319); // mode - 6 ???
define("ARTEFACT_PLANET_CREATOR", 320); // mode - 6 ???
define("ARTEFACT_PACKED_BUILDING", 321); // mode - 6 ???
define("ARTEFACT_PACKED_RESEARCH", 322); // mode - 6 ???
define("ARTEFACT_PACKING_BUILDING", 323); // mode - 6 ???
define("ARTEFACT_PACKING_RESEARCH", 324); // mode - 6 ???
define("ARTEFACT_PLANET_TEMP_CREATOR", 347); // mode - 6 ???
define("ARTEFACT_IGLA_MORI", 349); // mode - 6 ???
define("ARTEFACT_PLANET_TELEPORTER", 357); // mode - 6 ???
define("ARTEFACT_BATTLE_SHELL_POWER_10", 359); // mode - 6 ???
define("ARTEFACT_BATTLE_SHIELD_POWER_10", 360); // mode - 6 ???
define("ARTEFACT_BATTLE_ATTACK_POWER_10", 361); // mode - 6 ???
define("ARTEFACT_ANNIHILATION_ENGINE_10", 362); // mode - 6
define("ARTEFACT_NANOBOT_REPAIR_SYSTEM", 363); // mode - 6
define("ARTEFACT_ALLY_IGN", 364); // mode - 6
define("ARTEFACT_BUG", 365); // mode - 6

define("ACHIEVEMENT_NEWBIE_END", 343); // mode - 7 ???

// bd_connect_info загружается в game.php до global.inc.php
defined('DB_HOST') || require_once(dirname(__FILE__)."/../src/bd_connect_info.php");
// @include_once(dirname(__FILE__)."/../../../global.local.php");

@include_once(dirname(__FILE__)."/consts.local.php");

defined('DEV_MODE') || define('DEV_MODE', true);
defined('PASSWORD_SALT') || define('PASSWORD_SALT', 'Ac5YemeiToy7htho');

defined('CLIENT_JS_VERSION') || define('CLIENT_JS_VERSION', OXSAR_VERSION.CLIENT_VERSION.rand());
defined('CLIENT_CSS_VERSION') || define('CLIENT_CSS_VERSION', OXSAR_VERSION.CLIENT_VERSION.rand());
defined('CLIENT_SOUNDS_VERSION') || define('CLIENT_SOUNDS_VERSION', OXSAR_VERSION.CLIENT_VERSION.rand());
defined('CLIENT_IMAGES_VERSION') || define('CLIENT_IMAGES_VERSION', OXSAR_VERSION.CLIENT_VERSION.rand());

defined('DEBUG_PASSWORD') || define('DEBUG_PASSWORD', 'quoYaMe1wHo4xaci');

defined('NUM_GALAXYS') || define('NUM_GALAXYS', 8); // sync with na_galaxy_new_active
defined('NUM_SYSTEMS') || define('NUM_SYSTEMS', 600); // sync with na_system_new_active
// planets - na_planet_new_active

defined('GAMESPEED') || define('GAMESPEED', 0.75);

defined('BLOCK_TARGET_BY_ALLY_ATTACK_USED_PERIOD') || define('BLOCK_TARGET_BY_ALLY_ATTACK_USED_PERIOD', 0);

defined('BASHING_PERIOD') || define('BASHING_PERIOD', 0);
defined('BASHING_MAX_ATTACKS') || define('BASHING_MAX_ATTACKS', 0);

defined('NEW_USER_OBSERVER') || define('NEW_USER_OBSERVER', 0);
defined('PROTECTION_PERIOD') || define('PROTECTION_PERIOD', 0);

defined('SEND_NEW_USER_MESSAGE') || define('SEND_NEW_USER_MESSAGE', 0);

defined('DEATHMATCH') or define('DEATHMATCH', false);
defined('UNIVERSE_NAME') or define('UNIVERSE_NAME', DEATHMATCH ? 'DeathMatch' : 'Niro');

defined('SHOW_PLEASE_ACTIVATE_ACCOUNT') or define('SHOW_PLEASE_ACTIVATE_ACCOUNT', false);

// defined('DM_POINTS_BATTLE_EXP_SCALE') or define('DM_POINTS_BATTLE_EXP_SCALE', 1);
// defined('DM_MAX_POINTS_POWER') or define('DM_MAX_POINTS_POWER', 0.3);

defined('UNITS_GROUP_CONSUMTION_POWER_BASE') or define('UNITS_GROUP_CONSUMTION_POWER_BASE', 1.000003);
defined('MAX_GROUP_UNIT_CONSUMTION_PER_HOUR') or define('MAX_GROUP_UNIT_CONSUMTION_PER_HOUR', 0.1);

defined('OBSERVER_OFF_CREDIT_COST') or define('OBSERVER_OFF_CREDIT_COST', 0);

defined('ALLOW_SEND_MESSAGE_POINTS') or define('ALLOW_SEND_MESSAGE_POINTS', 100);

define('VACATION_DISABLE_TIME', 60*60*24*30);
define('LAST_TIME_ON_VACATION_DISABLE', 60*60*24*20);

define("USE_CHAT_TEST", false);

// define("SKIN_TYPE", "facebook");

define("SKIN_TYPE_GENERIC", "");
define("SKIN_TYPE_FB", "facebook");
define("SKIN_TYPE_MOBI", "mobi");

//Here comes magic

defined('SHOW_USER_AGREEMENT') or define('SHOW_USER_AGREEMENT', !DEATHMATCH);
defined('SHOW_DM_POINTS') or define('SHOW_DM_POINTS', false);

defined('ACHIEVEMENTS_ENABLED') or define('ACHIEVEMENTS_ENABLED', !DEATHMATCH);
defined('TUTORIAL_ENABLED') or define('TUTORIAL_ENABLED', !DEATHMATCH);
defined('MISSION_HALTING_OTHER_ENABLED') or define('MISSION_HALTING_OTHER_ENABLED', !DEATHMATCH);

defined('YII_CONSOLE_IS_RUNNING') or define('YII_CONSOLE_IS_RUNNING', true);
defined('TIMEZONE') or define('TIMEZONE', "Europe/Moscow");
defined('IPCHECK') or define('IPCHECK', false);
defined('ADMIN_PASSWORD') or define('ADMIN_PASSWORD', 'quoYaMe1wHo4xaci');

// Requirements section
defined('REQ_BUILDING') or define('REQ_BUILDING', 1);
defined('REQ_RESEARCH') or define('REQ_RESEARCH', 2);
defined('REQ_FLEET') or define('REQ_FLEET', 3);
defined('REQ_ACHIEVEMENT') or define('REQ_ACHIEVEMENT', 4);

defined('REQ_RELATION_GREATER') or define('REQ_RELATION_GREATER', 1);
defined('REQ_RELATION_LESSER') or define('REQ_RELATION_LESSER', 2);
defined('REQ_RELATION_EQUAL') or define('REQ_RELATION_EQUAL', 3);

defined('REQ_FLEET_TOTAL') or define('REQ_FLEET_TOTAL', 1);
defined('REQ_FLEET_PLANET') or define('REQ_FLEET_PLANET', 2);
// Tips section

// SN section
defined('SN_ONKL_IFRAME_SIZE') or define('SN_ONKL_IFRAME_SIZE', 600);
defined('SN_ONKL_MIN_HEIGHT') or define('SN_ONKL_MIN_HEIGHT', 600);
// defined('ODNOKLASSNIKI_SECRET_KEY') or define('ODNOKLASSNIKI_SECRET_KEY', 'B013759BB27839854508BD1A');

// User States Section
defined('STATE_ASSAULT_SIMULATION') or define('STATE_ASSAULT_SIMULATION', 1);
defined('STATE_RES_UPDATE_EXCHANGE') or define('STATE_RES_UPDATE_EXCHANGE', 2);

define("DEF_LANGUAGE_ID", 1);

if(!DEATHMATCH && !isset($GLOBALS["ADMIN_USERS"])){
    $GLOBALS["ADMIN_USERS"] = array(
        1, // craft
        3, // irina
    );
}
if(!isset($GLOBALS["ADMIN_USERS"])){
    $GLOBALS["ADMIN_USERS"] = array();
}

define("NEWBIE_PROTECTION_ENABLED", !DEATHMATCH); // OXSAR_RELEASED);
define("NEWBIE_PROTECTION_1_POINTS", 3000);
define("NEWBIE_PROTECTION_1_PERCENT", 30);
define("NEWBIE_PROTECTION_2_POINTS", 10000);
define("NEWBIE_PROTECTION_2_PERCENT", 20);
define("NEWBIE_PROTECTION_3_POINTS", 1000000);
define("NEWBIE_PROTECTION_3_PERCENT", 10);
define("NEWBIE_PROTECTION_MAX_POINTS_PERCENT", 5);

defined('RES_TO_UNIT_POINTS') or define("RES_TO_UNIT_POINTS", (1.0 / 1000) * 2.0);
defined('RES_TO_RESEARCH_POINTS') or define("RES_TO_RESEARCH_POINTS", (1.0 / 1000) * 1.0);
defined('RES_TO_BUILD_POINTS') or define("RES_TO_BUILD_POINTS", (1.0 / 1000) * 0.5);
define("POINTS_PRECISION", 2);

define("ATTACK_BY_ESPIONAGE_UNIT_ENABLED", true);

define("FLEET_BULK_INTO_DEBRIS", 0.50);
define("DEFENCE_BULK_INTO_DEBRIS", 0.01);
// define("BUILD_LEVEL_BULK_INTO_DEBRIS", 0.50);

define("EV_ABORT_SAVE_TIME", 15); // queue can be aborted in 15 secs with save all resourses
define("EV_ABORT_MAX_BUILD_PERCENT", 95); // construction & research aborting returns this number of percent
define("EV_ABORT_MAX_SHIPYARD_PERCENT", 70); // shipyard aborting returns this number of percent
define("EV_ABORT_MAX_REPAIR_PERCENT", 70); // repair aborting returns this number of percent
define("EV_ABORT_MAX_DISASSEMBLE_PERCENT", 70); // disassemble aborting returns this number of percent
define("EV_ABORT_MAX_FLY_PERCENT", 90); // max return fuel percent


// keep the order of UNIT_VIRT...
define("UNIT_VIRT_FLEET", 			1000000); // virtual fleet unit
define("UNIT_VIRT_STOCK_FLEET", 	1000001); // virtual stock fleet unit
define("UNIT_VIRT_DEFENSE", 		1000002); // virtual defence unit
define("UNIT_VIRT_HALTING_START",	1000003); // virtual halting fleet unit

define("MAX_BUILDING_LEVEL", DEATHMATCH ? 35 : 40);
define("MAX_RESEARCH_LEVEL", DEATHMATCH ? 35 : 40);

$GLOBALS["MAX_UNIT_LEVELS"] = array(
	UNIT_MOON_HYDROGEN_LAB => 10,
	UNIT_MOON_REPAIR_FACTORY => 9,
	UNIT_MOON_LAB => 5,
	UNIT_NANO_FACTORY => 12,
	UNIT_STAR_GATE => 15,
	UNIT_GRAVI => 10,
);

$GLOBALS["CANT_PACK_UNITS"] = array(
	UNIT_EXCHANGE,
	// UNIT_NANO_FACTORY,
	UNIT_GRAVI,
	UNIT_IGN,
	UNIT_MOON_LAB,
	UNIT_ALIEN_TECH,
);

$GLOBALS["BLOCKED_MARKET_UNITS"] = array(
	UNIT_EXCHANGE,
	UNIT_NANO_FACTORY,
	UNIT_ALIEN_TECH,
	UNIT_GRAVI,
	UNIT_IGN,
	UNIT_MOON_REPAIR_FACTORY,
	UNIT_MOON_LAB,
);

$GLOBALS["BLOCKED_STARGATE_UNITS"] = array(
	UNIT_SMALL_SHIELD,
	UNIT_LARGE_SHIELD,
	UNIT_SMALL_PLANET_SHIELD,
	UNIT_LARGE_PLANET_SHIELD,
	UNIT_INTERCEPTOR_ROCKET,
	UNIT_INTERPLANETARY_ROCKET,
	UNIT_EXCH_SUPPORT_SLOT,
    UNIT_EXCH_SUPPORT_RANGE,
);

$GLOBALS["PACKED_ARTEFACT_BUILDING_LEVELS"] = array(
	2,3,4,5,6,8,10,15
);
$GLOBALS["PACKED_ARTEFACT_RESEARCH_LEVELS"] = array(
	2,3,4,5,6,8,10,15
	// 2,3,5,10,15
);

$GLOBALS["RECYCLER_UNITS"] = array(
	UNIT_RECYCLER,
	UNIT_SHIP_TRANSPLANTATOR,
	UNIT_SHIP_COLLECTOR,
);

define("UNIT_TYPE_CONSTRUCTION", 1);
define("UNIT_TYPE_MOON_CONSTRUCTION", 5);
define("UNIT_TYPE_RESEARCH", 2);
define("UNIT_TYPE_FLEET", 3);
define("UNIT_TYPE_DEFENSE", 4);
define("UNIT_TYPE_ARTEFACT", 6);
define("UNIT_TYPE_ACHIEVEMENT", 7);

define("UNIT_TYPE_REPAIR", 100); // it is a not real unit type
define("UNIT_TYPE_DISASSEMBLE", 101); // it is a not real unit type

define("ARTEFACT_EFFECT_TYPE_PLANET", 0);
define("ARTEFACT_EFFECT_TYPE_EMPIRE", 1);
define("ARTEFACT_EFFECT_TYPE_FLEET",  2);
define("ARTEFACT_EFFECT_TYPE_BATTLE", 3);
define("ARTEFACT_EFFECT_TYPE_AUTO",   4);

define("MAX_ALIEN_TECH_FACTOR", 0.01);

define("USER_READY_PROB", 10);
define("LOW_LEVEL_PROB",  50);
define("MID_LEVEL_PROB",  20);
define("HIGH_LEVEL_PROB", 5);
define("ALIEN_PROB",      1);

define("MAX_SPEED_POSSIBLE", 999999999999);
define("MAX_SHIPS", 100000000);
define("MAX_SHIPS_GRADE", 9);

define("MAX_BUILDING_ORDER_UNITS", 1000000);
define("MAX_BUILDING_ORDER_UNITS_GRADE", 7);

define("ABANDONED_USER_ARTEFACT_CAPTURE_CHANCE_SCALE", 2.0);
define("ABANDONED_USER_MIN_ARTEFACT_CAPTURE_CHANCE", 10);
define("ABANDONED_USER_MAX_ARTEFACT_CAPTURE_CHANCE", 90);

define("STARGATE_TRANSPORT_SPEED", 60*60); // hours to secs
define("STARGATE_TRANSPORT_TIME_SCALE", 10.0/(60*60)); // hours to secs
define("STARGATE_TRANSPORT_CONSUMPTION_SCALE", 1.0/1000);

define("DESTROY_BUILD_RESULT_MIN_OFFS_LEVEL", -1);

define("EVENT_BATCH_PROCESS_TIME", 10);
define("EVENT_BATCH_CONSOLE_PROCESS_TIME", 20);

define("EVENT_PROCESSED_WAIT", 0);
define("EVENT_PROCESSED_START", 1);
define("EVENT_PROCESSED_ERROR", 2);
define("EVENT_PROCESSED_OK", 3);

define("EVENT_BUILD_CONSTRUCTION", 1); //  Construction
define("EVENT_DEMOLISH_CONSTRUCTION", 2); //  Demolish
define("EVENT_RESEARCH", 3); //  Research
define("EVENT_BUILD_FLEET", 4); //  Fleet
define("EVENT_BUILD_DEFENSE", 5); //  Defense
define("EVENT_POSITION", 6); //  Position
define("EVENT_TRANSPORT", 7); //  Transport
define("EVENT_COLONIZE", 8); //  Colonize
define("EVENT_RECYCLING", 9); //  Recycling
define("EVENT_ATTACK_SINGLE", 10); //  Attack
define("EVENT_SPY", 11); //  Spy
define("EVENT_ATTACK_ALLIANCE", 12); //  Alliance attack
define("EVENT_HALT", 13); //  Halt
define("EVENT_MOON_DESTRUCTION", 14); //  Moon destruction
define("EVENT_EXPEDITION", 15); //  Expedition
define("EVENT_ROCKET_ATTACK", 16); //  Rocket attack
define("EVENT_HOLDING", 17); //  Holding
define("EVENT_ALLIANCE_ATTACK_ADDITIONAL", 18); //  Serves as referer to alliance attack
define("EVENT_RETURN", 20); //  Return
define("EVENT_DELIVERY_UNITS", 21); //  Delivery units
define("EVENT_DELIVERY_RESOURSES", 22); //  Delivery resourses
define("EVENT_ATTACK_DESTROY_BUILDING", 23); //  Destroy a building attack
define("EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING", 24); //  Destroy a building attack
define("EVENT_ATTACK_DESTROY_MOON", 25); //  Moon destroy attack
define("EVENT_ATTACK_ALLIANCE_DESTROY_MOON", 27); //  Moon destroy alliance attack
define("EVENT_STARGATE_TRANSPORT", 28);
define("EVENT_DELIVERY_ARTEFACTS", 29);
define("EVENT_COLONIZE_RANDOM_PLANET", 30);
define("EVENT_COLONIZE_NEW_USER_PLANET", 31);
define("EVENT_STARGATE_JUMP", 32);
define("EVENT_ALIEN_FLY_UNKNOWN", 33);
define("EVENT_ALIEN_HOLDING", 34);
define("EVENT_ALIEN_ATTACK", 35);
define("EVENT_ALIEN_HALT", 36);
define("EVENT_ALIEN_GRAB_CREDIT", 37);
define("EVENT_ALIEN_ATTACK_CUSTOM", 38);
define("EVENT_TELEPORT_PLANET", 39);
define("EVENT_REPAIR", 50);
define("EVENT_DISASSEMBLE", 51);
define("EVENT_TEMP_PLANET_DISAPEAR", 52);
define("EVENT_RUN_SIM_ASSAULT", 53);
define("EVENT_ARTEFACT_EXPIRE", 60); //  Expiration of the artefact's effect
define("EVENT_ARTEFACT_DISAPPEAR", 61); //  Expiration of the artefact's lifetime
define("EVENT_ARTEFACT_DELAY", 63);         //  Expiration of delay
define("EVENT_EXCH_EXPIRE", 64); //Expiration of Exchange
define("EVENT_EXCH_BAN", 66);
define("EVENT_TOURNAMENT_SCHEDULE", 70);
define("EVENT_TOURNAMENT_RESCHEDULE", 71);
define("EVENT_TOURNAMENT_PARTICIPANT", 72);
define("EVENT_ALIEN_HOLDING_AI", 80);
define("EVENT_ALIEN_CHANGE_MISSION_AI", 81);

define("EVENT_MARK_LAST_BUILD", EVENT_BUILD_DEFENSE);
define("EVENT_MARK_FIRST_FLEET", EVENT_POSITION);
define("EVENT_MARK_LAST_FLEET", EVENT_TELEPORT_PLANET);

define("RETREAT_FLEET_OK", 0);
define("RETREAT_FLEET_ALREADY_DONE", 1);
define("RETREAT_FLEET_PLANET_UNDER_ATTACK", 2);
define("RETREAT_FLEET_NOT_OWNER", 3);

define("RETREAT_EVENT_BLOCK_END_TIME", 5);

$GLOBALS["RETREAT_FLEET_EVENTS"] = array(
	EVENT_POSITION,
	EVENT_TRANSPORT,
	EVENT_COLONIZE,
	EVENT_RECYCLING,
	EVENT_ATTACK_SINGLE,
	EVENT_SPY,
	EVENT_ATTACK_ALLIANCE,
	EVENT_HALT,
	EVENT_MOON_DESTRUCTION,
	EVENT_EXPEDITION,
	EVENT_ROCKET_ATTACK,
	EVENT_HOLDING,
	EVENT_ALLIANCE_ATTACK_ADDITIONAL,
	// EVENT_RETURN,
	// EVENT_DELIVERY_UNITS,
	// EVENT_DELIVERY_RESOURSES,
	EVENT_ATTACK_DESTROY_BUILDING,
	EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING,
	EVENT_ATTACK_DESTROY_MOON,
	EVENT_ATTACK_ALLIANCE_DESTROY_MOON,
	EVENT_STARGATE_TRANSPORT,
	EVENT_TELEPORT_PLANET,
	// EVENT_DELIVERY_ARTEFACTS,
	EVENT_COLONIZE_RANDOM_PLANET,
	// EVENT_COLONIZE_NEW_USER_PLANET,
	// EVENT_STARGATE_JUMP,
	// EVENT_ALIEN_FLY_UNKNOWN,
	// EVENT_ALIEN_HOLDING,
	// EVENT_ALIEN_ATTACK,
	// EVENT_ALIEN_HALT,
	// EVENT_ALIEN_GRAB_CREDIT,
	// EVENT_ALIEN_ATTACK_CUSTOM,
);

define("MSG_FOLDER_INBOX",	1);
define("MSG_FOLDER_SENT",	2);
define("MSG_FOLDER_FLEET",	3);
define("MSG_FOLDER_SPY",	4);
define("MSG_FOLDER_BATTLE_REPORTS",	5);
define("MSG_FOLDER_ALLIANCE",	6);
define("MSG_FOLDER_ARTEFACTS",	7);
define("MSG_FOLDER_CREDIT",		8);
define("MSG_FOLDER_EXPEDITION", 9);
define("MSG_FOLDER_RECYCLER",   10);
define("MSG_FOLDER_SURVEILLANCE_DETECTED", 11);

define("MSG_POSITION_REPORT",   6);
define("MSG_TRANSPORT_REPORT",  7);
define("MSG_COLONIZE_REPORT",   8);
define("MSG_RECYCLING_REPORT",  9);
define("MSG_ESPIONAGE_COMMITTED", 11);
define("MSG_RETURN_REPORT",     20);
define("MSG_TRANSPORT_REPORT_OTHER", 21);
define("MSG_ASTEROID",          22);
define("MSG_ALLY_ABANDONED",    23);
define("MSG_MEMBER_RECEIPTED",  24);
define("MSG_MEMBER_REFUSED",    25);
define("MSG_NEW_MEMBER",        100);
define("MSG_MEMBER_LEFT",       101);
define("MSG_MEMBER_KICKED",     102);
define("MSG_GRASPED_REPORT",    30);
define("MSG_DELIVERY_UNITS",    31);
define("MSG_DELIVERY_RESOURSES",32);
define("MSG_EXPEDITION_REPORT", 50);
define("MSG_EXPEDITION_SENSOR", 51);
define("MSG_RETREAT_OTHER",     52);
define("MSG_BUILDING_DESTROYED",53);
define("MSG_LOST_ARTEFACTS",    54);
define("MSG_GRASPED_ARTEFACTS", 55);
define("MSG_MOON_DESTROYED",	56);
define("MSG_TRANSPORT_REPORT_ARTEFACT", 57);
define("MSG_DELAY_ARTEFACT", 			58);
define("MSG_EXPIRE_ARTEFACT", 			59);
define("MSG_DISAPEAR_ARTEFACT", 		60);
define("MSG_ACTIVATE_ARTEFACT", 		61);
define("MSG_DEACTIVATE_ARTEFACT", 		62);
define("MSG_CAPTURE_ARTEFACT", 			63);
define("MSG_ARTEFACT", 					64);
define("MSG_CREDIT",					65);
define("MSG_EXPEDITION_NEW_PLANET",		66);
define("MSG_UFO_PLANET_DIE",			67);
define("MSG_TEMP_PLANET_DIE",			68);
define("MSG_RETREAT_TRANSPORT", 		69);
define("MSG_STARGATE_JUMP_REPORT", 		70);
define("MSG_SURVEILLANCE_DETECTED", 	71);
define("MSG_ALIEN_HALTING", 			72);
define("MSG_EXCH_LOT_BACK_RESOURSES",	73);
define("MSG_ALIEN_RESOURSES_GIFT",		74);
define("MSG_PLANET_TELEPORTED",			75);
define("MSG_PLANET_NOT_TELEPORTED",		76);

define("RES_UPDATE_PLANET_PRODUCTION",  1);
define("RES_UPDATE_EXCHANGE",           2);
define("RES_UPDATE_COST",               3);
define("RES_UPDATE_CANCEL",             4);
define("RES_UPDATE_DISASSEMBLE",        5);
define("RES_UPDATE_UNLOAD",             6);
define("RES_UPDATE_VIEW_GALAXY",        7);
define("RES_UPDATE_MONITOR_PLANET",     8);
define("RES_UPDATE_EXPEDITION_CREDITS", 9);
define("RES_UPDATE_BUY_CREDITS",        10);
define("RES_UPDATE_EXCH_LOT_UNLOAD",    11);
define("RES_UPDATE_EXCH_LOT_RESERVE",   12);
define("RES_UPDATE_EXCH_LOT_BUY",       13);
define("RES_UPDATE_EXCH_LOT_SELL",      14);
define("RES_UPDATE_EXCH_OWNER_PROFIT",    15);
define("RES_UPDATE_EXCH_DEFENDER_PROFIT", 16);
define("RES_UPDATE_FIX",                17);
define("RES_UPDATE_EXCH_LOT_COMISSION", 18);
define("RES_UPDATE_BUY_ARTEFACT",       19);
define("RES_UPDATE_VIP_START",          20);
define("RES_UPDATE_ACHIEVEMENT",		21);
define("RES_UPDATE_EXCH_FUEL_REST",		22);
define("RES_UPDATE_ATTACKER",    		24);
define("RES_UPDATE_LOAD_FLEET",         25);
define("RES_UPDATE_UNLOAD_FLEET",       26);
define("RES_UPDATE_ALIEN_GRAB_CREDIT",  27);
define("RES_UPDATE_ALIEN_GIFT_CREDIT",  28);
define("RES_UPDATE_EXCH_LOT_PREMIUM",   29);
define("RES_UPDATE_EXCH_LOT_PRICE_EXT",	30);
define("RES_UPDATE_ALIEN_GIFT_RESOURSES",  31);
define("RES_UPDATE_CHANGE_PROFESSION",  32);

define('EXPEDITION_ENABLED', !DEATHMATCH);

define("EXPED_TYPE_ARTEFACT",	1);
define("EXPED_TYPE_ASTEROID",	2);
define("EXPED_TYPE_BATTLEFIELD",3);
define("EXPED_TYPE_BLACK_HOLE",	4);
define("EXPED_TYPE_CREDIT",		5);
define("EXPED_TYPE_DELAY",		6);
define("EXPED_TYPE_LOST",		7);
define("EXPED_TYPE_FAST",		8);
define("EXPED_TYPE_NOTHING",	9);
define("EXPED_TYPE_PIRATES",	10);
define("EXPED_TYPE_RESOURCE",	11);
define("EXPED_TYPE_SHIP",		12);
define("EXPED_TYPE_UNKNOWN",	13);

define("EXPED_LOST_ENABLED",    true);

define("EXPED_PLANET_LIFETIME_MIN",	60*60*12);
define("EXPED_PLANET_LIFETIME_MAX",	60*60*24);

define("TEMP_PLANET_LIFETIME",	60*60*24*21);
define("TEMP_MOON_ENABLED",		true);
define("TEMP_MOON_SIZE_MIN",	2000);
define("TEMP_MOON_SIZE_MAX",	2500);

define("MAX_NORMAL_PLANET_POSITION", 30);
define("EXPED_PLANET_POSITION",		MAX_NORMAL_PLANET_POSITION + 1);
define("EXPED_START_CREATE_PLANET",	EXPED_PLANET_POSITION + 10);
define("EXPED_END_CREATE_PLANET",	EXPED_START_CREATE_PLANET + 30);
define("MAX_POSITION", 				EXPED_END_CREATE_PLANET);

defined('MAX_PLANETS') or define("MAX_PLANETS", 10);
defined('ADDITIONAL_ARTEFACT_PLANETS_NUMBER') or define("ADDITIONAL_ARTEFACT_PLANETS_NUMBER", 3);
defined('TEMP_PLANETS_NUMBER') or define("TEMP_PLANETS_NUMBER", 5);

define("POSITION_TO_CALC_EXP_TO",	150);

define("COLONIZE_NEW_USER_PLANET_TIME", 3);
define("COLONIZE_NEW_USER_PLANET_TIME_MAX_DELTA", 2);
define("NEW_USER_ALIEN_ATTACK_FLYTIME", 60*60*3);

define("ALLIANCE_FOUND_USER_MIN_POINTS", 100);

define("PLANET_TELEPORT_MIN_INTERVAL_TIME",	60*60*24*1);

define("EXPED_LOG_TIME",	604800);

define("MARKET_BASE_CURS_METAL", 600);
define("MARKET_BASE_CURS_SILICON", 400);
define("MARKET_BASE_CURS_HYDROGEN", 200);
define("MARKET_BASE_CURS_CREDIT", 1);
define("MARKET_PROD_HOURS", 24);
define("MARKET_PROD_CREDITS", 100);
define("MARKET_MIN_PLANET_RATIO", 1.0);

define("RACE_HUMAN", 1);
define("RACE_ALIEN", 2);

define("ESTATUS_OK", 1);  //��� ��������
define("ESTATUS_SOLD", 2); //��� ������
define("ESTATUS_RECALL", 3); //��� ������� �������������
define("ESTATUS_REMOVED", 4); //��� ���� �� ���������� TTL
define("ESTATUS_SUSPENDED", 5); //������� ��������������
define("ESTATUS_BANNED", 6); //������� ��������������

define("ETYPE_RESOURCE", 1); //��������� �������
define("ETYPE_FLEET", 2); //�������� ����
define("ETYPE_ARTEFACT", 3); //�������� ��������

define("EXCH_ENABLED", true);
defined("EXCH_NEW_PROFIT_TYPE") or define("EXCH_NEW_PROFIT_TYPE", true); // isset($_SERVER['REMOTE_ADDR']) ? $_SERVER['REMOTE_ADDR'] == '95.221.98.239' : false);

if(EXCH_NEW_PROFIT_TYPE){
    defined("EXCH_MERCHANT_PREMIUM_COMMISSION") or define("EXCH_MERCHANT_PREMIUM_COMMISSION", 16);
    defined("EXCH_MERCHANT_COMMISSION") or define("EXCH_MERCHANT_COMMISSION", 19);

    defined("EXCH_NO_MERCHANT_PREMIUM_COMMISSION") or define("EXCH_NO_MERCHANT_PREMIUM_COMMISSION", 22);
    defined("EXCH_NO_MERCHANT_COMMISSION") or define("EXCH_NO_MERCHANT_COMMISSION", 25);

    define("EXCH_COMMISSION_BASE_UNIT", 1 - EXCH_MERCHANT_PREMIUM_COMMISSION*0.01);

    define("EXCH_PREMIUM_PERCENT", 0.5); // ������� ���, ������� �� ����
    define("EXCH_PREMIUM_MIN_COST", 10); // ����������� ���� ��� ������� ����
}else{
    defined("EXCH_MERCHANT_PREMIUM_COMMISSION") or define("EXCH_MERCHANT_PREMIUM_COMMISSION", 2);
    defined("EXCH_MERCHANT_COMMISSION") or define("EXCH_MERCHANT_COMMISSION", 3);

    defined("EXCH_NO_MERCHANT_PREMIUM_COMMISSION") or define("EXCH_NO_MERCHANT_PREMIUM_COMMISSION", 15);
    defined("EXCH_NO_MERCHANT_COMMISSION") or define("EXCH_NO_MERCHANT_COMMISSION", 20);

    define("EXCH_PREMIUM_LOT_COST", 10); // ������� ���
}

defined("EXCH_INVIOLABLE") or define("EXCH_INVIOLABLE", DEATHMATCH ? 0 : 2); //id ���������������� �����
define("EXCH_LEVEL_SLOTS", 15); //���������� ������ �����, ���. ���� ������ �� �������
define("EXCH_MAX_LOTS_FACTOR", 0.5);
define("EXCH_RADIUS_FACTOR", 1);
define("EXCH_RADIUS_SYSTEMS_PER_GALAXY", 300);
define("EXCH_MAX_TTL", 7); //���� �� ���������� ����
define("EXCH_MIN_TTL", 3); //����� ����� ����� ��� ����� ������
define("EXCH_DEF_TTL", 4); // ����� ����� ���� �����������
define("EXCH_FEE_MIN", 1); //����������� % ����� �� ������� ����
define("EXCH_FEE_MAX", 30); //������������ % ����� �� ������� ����
define("EXCH_SELLER_MIN_PROFIT", -30); //����������� ������� ��������� � % �� ������� ����
define("EXCH_SELLER_DEF_PROFIT", 10); // ����������� ������� ��������� � % �� ������� ���� ��� ������� ���� �� ���������
define("EXCH_SELLER_MAX_PROFIT", 1000); //������������ ������� ��������� � % �� ������� ����
define("EXCH_SELLER_ART_MAX_PROFIT", 0); //������������ ������� ��������� � % �� ������� ���������
define("EXCH_COMMISSION_MIN", 1); //����������� ����� �� ������ ������
define("EXCH_COMMISSION_MAX", 10); //������������ ����� �� ������ ������
define("EXCH_MIN_DISCOUNT", 1);
define("EXCH_DEF_DISCOUNT", 3); //������ ��� �������� ������ �� ���������
define("EXCH_MAX_DISCOUNT", 30);

define("EXCH_MIN_UNIT_PRICE", 0.00001);

define("EXCH_PREMIUM_LIST_MAX_SIZE", 5); // ������������ ���������� ����� � ������ �������
define("EXCH_PREMIUM_LOT_EXPIRY_TIME", 60*60*2); // ���� ��� ������� � ������� ����� �������, ��� ������ ������� �������
define("EXCH_PREMIUM_LOT_FUEL_COST_MULT", 0.5); // ����. ���� ������� ��� ������� �����
define("EXCH_PREMIUM_LOT_FLY_TIME_MULT", 0.5); // ����. ������� �������� ������� ����

define("EXCH_TYPE_RESOURCES", 1);
define("EXCH_TYPE_FLEET", 2);
define("EXCH_TYPE_ARTEFACT", 3);
define("EXCH_ART_MIN_PERCENT", 5);
define("EXCH_ART_DEF_PERCENT", 50);

// keep achiev state order
define("ACHIEV_STATE_ALERT", 0);
define("ACHIEV_STATE_HIDDEN", 1);
define("ACHIEV_STATE_BONUS_GIVEN", 2);
define("ACHIEV_STATE_PROCESSED", 3);

define("ACHIEV_BONUS_BUILD_TYPE_ANY", 0);
define("ACHIEV_BONUS_BUILD_TYPE_PLANET", 1<<0);
define("ACHIEV_BONUS_BUILD_TYPE_MOON", 1<<1);
define("ACHIEV_BONUS_BUILD_TYPE_ERROR", 1<<2);

// define("FOURTH_GALAXY_MOON_CONSTRUCTION_SPEED", 2); //��������� �������� ������������� �� ���� � 4 ���������
// define("FOURTH_GALAXY_RESOURCES_GATHER", 1.1); //��������� ������ �������� ��� ��������� ���������
// define("FOURTH_GALAXY_FLEET_SPEED", 1.1); //��������� �������� ������ ������ ��� ��������� ���������

define("METAL", -1);
define("SILICON", -2);
define("HYDROGEN", -3);

define("POINTS_PER_ADD_TECH_LEVEL", 100);
define("MAX_ADD_TECH_LEVELS", 20);
define("MAX_MAILS", 50);
define("MAIL_SENDER_NAME", 'oxsar.ru');
define("MAIL_SENDER", 'no-reply@oxsar.ru');

define("FLEET_FUEL_CONSUMPTION", 0.5); // 0.8);
define("MAX_HEIGHT", '480');
define("MAX_MENU_ITEMS", '10');
define("MAX_PLANET_ITEMS", '6');
define("MAX_FB_EMPIRE_PLANETS", '3');

define('CHAT_REFRESH_RATE', 10000); //ms
define('CHAT_NEWS_REFRESH_RATE', 20000); //ms

define('STORAGE_SAVE_FACTOR', 0.01); // /*2.6*/

define('MOON_CREATION_USER_INTERVAL', 3600 * 24 * 7);
// define('MOON_CREATION_PLANET_INTERVAL', 3600 * 24 * 7);
define('MOON_CREATION_SYSTEM_INTERVAL', 3600 * 24 * 7);

define("ALIEN_ENABLED", !DEATHMATCH);
define("ALIEN_NORMAL_FLEETS_NUMBER", 50);
define("ALIEN_ATTACK_INTERVAL", 60*60*24*6);
define("ALIEN_GRAB_CREDIT_INTERVAL", 60*60*24*10);
define("ALIEN_ATTACK_TIME_FLEETS_NUMBER", ALIEN_NORMAL_FLEETS_NUMBER * 5);
define("ALIEN_FLY_MIN_TIME", 60*60*15);
define("ALIEN_FLY_MAX_TIME", 60*60*24);
define("ALIEN_HALTING_MIN_TIME", 60*60*12);
define("ALIEN_HALTING_MAX_TIME", 60*60*24);
define("ALIEN_HALTING_MAX_REAL_TIME", 60*60*24*15);
define("ALIEN_CHANGE_MISSION_MIN_TIME", 60*60*8);
define("ALIEN_CHANGE_MISSION_MAX_TIME", 60*60*10);
define("ALIEN_GRAB_MIN_CREDIT", 100*1000);
define("ALIEN_GRAB_CREDIT_MIN_PERCENT", 0.08);
define("ALIEN_GRAB_CREDIT_MAX_PERCENT", 0.1);
define("ALIEN_GIFT_CREDIT_MIN_PERCENT", 5);
define("ALIEN_GIFT_CREDIT_MAX_PERCENT", 10);
define("ALIEN_MAX_GIFT_CREDIT", 500);
define("ALIEN_FLEET_MAX_DERBIS", 1000*1000*1000);

define("PROFESSION_CHANGE_MIN_DAYS", DEATHMATCH ? 7 : 14);
define("PROFESSION_CHANGE_COST", 1000);

$GLOBALS["PROFESSIONS"] = array(
    0 => array(
        'name' => 'PROFESSION_UNKNOWN',
        'tech_special' => array(),
    ),
    1 => array(
        'name' => 'PROFESSION_MINER',
        'tech_special' => array(
            UNIT_METALMINE => 1,
            UNIT_SILICON_LAB => 1,
            UNIT_SOLAR_PLANT => 2,
            UNIT_SHIPYARD => -2,
            UNIT_GUN_TECH => -2,
            UNIT_SHIELD_TECH => -2,
            UNIT_SHELL_TECH => -2,
            UNIT_BALLISTICS_TECH => -1,
            UNIT_COMPUTER_TECH => -1,
        ),
    ),
    2 => array(
        'name' => 'PROFESSION_ATTACKER',
        'tech_special' => array(
            // UNIT_COMPUTER_TECH => 1,
            UNIT_GUN_TECH => 1,
            UNIT_SHIELD_TECH => 1,
            UNIT_SHELL_TECH => 1,
            UNIT_BALLISTICS_TECH => 1,
            UNIT_SHIPYARD => 1,
            UNIT_METALMINE => -1,
            UNIT_SILICON_LAB => -1,
            UNIT_IGN => -1,
            UNIT_DEFENSE_FACTORY => -3,
        ),
    ),
    3 => array(
        'name' => 'PROFESSION_DEFENDER',
        'tech_special' => array(
            UNIT_MASKING_TECH => 1,
            UNIT_SHIELD_TECH => 1,
            UNIT_SHELL_TECH => 1,
            UNIT_DEFENSE_FACTORY => 1,
            UNIT_ROCKET_STATION => 1,
            UNIT_COMPUTER_TECH => -1,
            UNIT_GUN_TECH => -1,
            UNIT_SHIPYARD => -3,
        ),
    ),
    4 => array(
        'name' => 'PROFESSION_TANK',
        'tech_special' => array(
            UNIT_GUN_TECH => 2,
            UNIT_SHIELD_TECH => 2,
            UNIT_SHELL_TECH => 2,
            UNIT_GRAVI => -2,
            UNIT_COMBUSTION_ENGINE => -2,
            UNIT_IMPULSE_ENGINE => -2,
            UNIT_HYPERSPACE_ENGINE => -2,
        ),
    ),
    /*
    4 => array(
        'name' => 'PROFESSION_RESEARCHER',
        'tech_special' => array(
            // UNIT_COMPUTER_TECH => 1,
            UNIT_RESEARCH_LAB => 1,
            // UNIT_SPYWARE => 1,
            UNIT_EXPO_TECH => 1,
            UNIT_GUN_TECH => -1,
            UNIT_DEFENSE_FACTORY => -1,
        ),
    ),
     *
     */
);

$GLOBALS["GALAXY_CLUSTERS"] = array(
	"NIRO" => array(
		// "NAME" => "Niro",
		"START" => 1,
		"END" => 8,
		"SYSTEMS" => 600,
	),
	"UNKNOWN" => array(
		// "NAME" => "Hrad",
		"START" => 20,
		"END" => 25,
		"SYSTEMS" => 400,
	),
);

$GLOBALS["STARGATE"] = array(
	"START_DISABLE" => true,
	"END_DISABLE" 	=> false,
);

defined('ROCKET_SPEED_FACTOR') or define('ROCKET_SPEED_FACTOR', 1);

if(!isset($GLOBALS["GALAXY_SPEC_PARAMS"])){
    $GLOBALS["GALAXY_SPEC_PARAMS"] = array(
        4 => array(
            "ADVANCED_BATTLE" => 1,
            "PLANET_CONSTRUCTION_SPEED_FACTOR" => 1.1, // ��������� �������� ������������� �������� �� �������
            "MOON_CONSTRUCTION_SPEED_FACTOR" => 2, // ��������� �������� ������������� �������� �� ����
            "TEMP_MOON_CONSTRUCTION_SPEED_FACTOR" => 4, // ��������� �������� ������������� �������� �� ��������� ����
            "FLEET_BUILDING_SPEED_FACTOR" => 1.1, // ��������� �������� ������������� �����
            "DEFENSE_BUILDING_SPEED_FACTOR" => 1.1, // ��������� �������� ������������� �������
            "RESEARCH_SPEED_FACTOR" => 1.1, // ��������� �������� ������������
            "RESOURCES_PRODUCTION_FACTOR" => 1.1, // ��������� ������ ��������
            "ENEGRY_PRODUCTION_FACTOR" => 1.1, // ��������� ������ �������
            "STORAGE_FACTOR" => 1.1, // ��������� ������� ��������
            "FLEET_SPEED_FACTOR" => 1.1, // ��������� �������� ������ �����
            "ALLOW_DESTROY_BUILDING" => 0, // 1, // ��������� ��������� ���������
            "ALLOW_DESTROY_MOON" => 1, // ��������� ��������� ����
            "ALLOW_STARGATE_TRANSPORT" => 1, // ��������� ������ ��������������������� ������ � ������
        ),
    );

    $GLOBALS["GALAXY_SPEC_PARAMS"][5] = $GLOBALS["GALAXY_SPEC_PARAMS"][4];

    for($i = $GLOBALS["GALAXY_CLUSTERS"]["NIRO"]["START"]; $i <= $GLOBALS["GALAXY_CLUSTERS"]["NIRO"]["END"]; $i++)
    {
        if( !isset($GLOBALS["GALAXY_SPEC_PARAMS"][$i]["FLEET_SPEED_FACTOR"]) )
        {
            $GLOBALS["GALAXY_SPEC_PARAMS"][$i]["FLEET_SPEED_FACTOR"] = 1;
        }
        $GLOBALS["GALAXY_SPEC_PARAMS"][$i]["FLEET_SPEED_FACTOR"] *= 2;
        $GLOBALS["GALAXY_SPEC_PARAMS"][$i]["TEMP_MOON_CONSTRUCTION_SPEED_FACTOR"] = 4; // ��������� �������� ������������� �������� �� ��������� ����
    }
}

$GLOBALS["POINTS_PER_MINNING_LEVEL"] = array(
	1 => 10,
	2 => 25,
	3 => 50,
	4 => 75,
	5 => 100,
	6 => 125,
	7 => 150,
	8 => 175,
	9 => 200,
   10 => 225,
   11 => 250,
   12 => 275,
   13 => 300,
   /*
   14 => 325,
   15 => 350,
   16 => 375,
   17 => 400,
   */
);
$GLOBALS["VACATION_BLOCKING_EVENTS"] = array(
	EVENT_BUILD_FLEET,
	EVENT_BUILD_DEFENSE,
	EVENT_TRANSPORT,
	EVENT_POSITION,
	EVENT_ATTACK_SINGLE,
	EVENT_ATTACK_ALLIANCE,
	EVENT_MOON_DESTRUCTION,
	EVENT_ROCKET_ATTACK,
	EVENT_ALLIANCE_ATTACK_ADDITIONAL,
	EVENT_ATTACK_DESTROY_BUILDING,
	EVENT_ATTACK_ALLIANCE_DESTROY_BUILDING,
	EVENT_ATTACK_DESTROY_MOON,
	EVENT_ATTACK_ALLIANCE_DESTROY_MOON,
	EVENT_REPAIR,
	EVENT_DISASSEMBLE,
	EVENT_RECYCLING,
	EVENT_COLONIZE,
	EVENT_COLONIZE_RANDOM_PLANET,
	EVENT_COLONIZE_NEW_USER_PLANET,
	EVENT_SPY,
	EVENT_HALT,
	EVENT_EXPEDITION,
	EVENT_HOLDING,
	// EVENT_RETURN,
	EVENT_STARGATE_TRANSPORT,
	EVENT_STARGATE_JUMP,
	EVENT_BUILD_CONSTRUCTION,
	EVENT_DEMOLISH_CONSTRUCTION,
	EVENT_RESEARCH,
	EVENT_TELEPORT_PLANET,
//EVENT_DELIVERY_UNITS,
//EVENT_DELIVERY_RESOURSES,
//EVENT_DELIVERY_ARTEFACTS,
//EVENT_ARTEFACT_EXPIRE,
//EVENT_ARTEFACT_DISAPPEAR,
//EVENT_ARTEFACT_DELAY,
//EVENT_EXCH_EXPIRE,
//EVENT_EXCH_BAN,
);
