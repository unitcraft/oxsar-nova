<?php

define('DEV_MODE', true);

defined('CLIENT_JS_VERSION') || define('CLIENT_JS_VERSION', OXSAR_VERSION.CLIENT_VERSION);
defined('CLIENT_CSS_VERSION') || define('CLIENT_CSS_VERSION', OXSAR_VERSION.CLIENT_VERSION);
defined('CLIENT_SOUNDS_VERSION') || define('CLIENT_SOUNDS_VERSION', OXSAR_VERSION.CLIENT_VERSION);
defined('CLIENT_IMAGES_VERSION') || define('CLIENT_IMAGES_VERSION', OXSAR_VERSION.CLIENT_VERSION);

$GLOBALS["ADMIN_USERS"] = array(
    1, // craft
);

define('DEATHMATCH', true);

define('DEATHMATCH_START_TIME', strtotime('2012-04-04 21:00')); // 22
define('DEATHMATCH_END_TIME', strtotime('2012-04-11 21:00'));

$GLOBALS['PRIZE_INFO'] = array(
    strtotime('2012-04-16 14:00') => 500,
    strtotime('2012-04-16 22:00') => 900,
    strtotime('2012-04-17 10:00') => 950,
    strtotime('2012-04-17 22:00') => 1200,
    strtotime('2012-04-18 10:00') => 1250,
    strtotime('2012-04-18 22:00') => 1300,
    /*
    strtotime('2012-04-01 14:00') => 0,
    strtotime('2012-04-04 23:00') => 1000,
    strtotime('2012-04-05 09:00') => 1050,
    strtotime('2012-04-06') => 1500,
    strtotime('2012-04-06 10:00') => 1550,
    strtotime('2012-04-07') => 2000,
    strtotime('2012-04-07 10:00') => 2050,
    strtotime('2012-04-07 14:00') => 2200,
    strtotime('2012-04-07 14:20') => 2200,
    strtotime('2012-04-08') => 2300,
    strtotime('2012-04-08 10:00') => 2350,
    strtotime('2012-04-09') => 2500,
    strtotime('2012-04-09 10:00') => 2550,
    strtotime('2012-04-10') => 3000,
    strtotime('2012-04-10 10:00') => 3050,
    strtotime('2012-04-10 21:00') => 3050,
    strtotime('2012-04-11 9:00') => 3200,
    strtotime('2012-04-11 12:10') => 3415,
    strtotime('2012-04-11 17:00') => 3900,
    strtotime('2012-04-11 21:00') => 4000,

    strtotime('2012-03-11') => 0,
    strtotime('2012-03-12') => 700,
    strtotime('2012-03-13') => 1000,
    strtotime('2012-03-13 9:00') => 1050,
    strtotime('2012-03-14') => 1500,
    strtotime('2012-03-14 9:00') => 1550,
    strtotime('2012-03-15') => 2000,
    strtotime('2012-03-15 9:00') => 2050,
    strtotime('2012-03-16') => 3000,
    strtotime('2012-03-16 9:00') => 3050,
    strtotime('2012-03-17') => 4000,
    strtotime('2012-03-17 9:00') => 4050,
    strtotime('2012-03-17 13:00') => 4300,
    strtotime('2012-03-18') => 4500,
    strtotime('2012-03-18 1:15') => 4500,
    strtotime('2012-03-18 9:00') => 4550,
    strtotime('2012-03-18 14:00') => 4700,
    strtotime('2012-03-18 16:12') => 4700,
    strtotime('2012-03-19') => 5050,
    strtotime('2012-03-19 9:00') => 5100,
    strtotime('2012-03-19 21:40') => 5500,
     *
     */
);

define('OBSERVER_OFF_CREDIT_COST', 1000);

define('NUM_GALAXYS', 1); // sync with na_galaxy_new_active
define('NUM_SYSTEMS', 200); // sync with na_system_new_active
// planets - na_planet_new_active

define('SHOW_USER_AGREEMENT', false);
define('SHOW_DM_POINTS', true);

define('BLOCK_TARGET_BY_ALLY_ATTACK_USED_PERIOD', 60*60*24*3);

define('BASHING_PERIOD', 60*60*5);
define('BASHING_MAX_ATTACKS', 4);

define('NEW_USER_OBSERVER', 1);
define('PROTECTION_PERIOD', 60*60*24);

define("EXCH_INVIOLABLE", 0);

define("EXCH_NEW_PROFIT_TYPE", true);

define("EXCH_MERCHANT_PREMIUM_COMMISSION", 7);
define("EXCH_MERCHANT_COMMISSION", 10);

define("EXCH_NO_MERCHANT_PREMIUM_COMMISSION", 13);
define("EXCH_NO_MERCHANT_COMMISSION", 16);

// define('UNITS_GROUP_CONSUMTION_POWER_BASE', 1.000004);

define("RES_TO_UNIT_POINTS", (1.0 / 1000) * 2.0);
define("RES_TO_RESEARCH_POINTS", (1.0 / 1000) * 1.0 * 0.5);
define("RES_TO_BUILD_POINTS", (1.0 / 1000) * 0.5 * 0.1);

$GLOBALS['DISABLED_ARTEFACTS'] = array(
    ARTEFACT_BATTLE_SHELL_POWER_10,
    ARTEFACT_BATTLE_SHIELD_POWER_10,
    ARTEFACT_BATTLE_ATTACK_POWER_10,
    ARTEFACT_ANNIHILATION_ENGINE_10,
    ARTEFACT_BATTLE_NEUTRON_AFFECTOR,
    ARTEFACT_PACKED_BUILDING,
    ARTEFACT_PACKED_RESEARCH,
);

$GLOBALS["INITIAL_BUILDINGS"] = array(
    UNIT_METALMINE => 2,
    UNIT_SILICON_LAB => 2,
    UNIT_HYDROGEN_LAB => 2,
    UNIT_SOLAR_PLANT => 4,
    UNIT_ROBOTIC_FACTORY => 2,
    UNIT_SHIPYARD => 2,
    UNIT_RESEARCH_LAB => 2,
    UNIT_DEFENSE_FACTORY => 1,
    UNIT_REPAIR_FACTORY => 1,
);

$GLOBALS["INITIAL_RESEARCHES"] = array(
    UNIT_COMPUTER_TECH => 1,
    UNIT_ENERGY_TECH => 1,
    UNIT_COMBUSTION_ENGINE => 2,

);

$GLOBALS["INITIAL_UNITS"] = array(
    UNIT_SMALL_TRANSPORTER => 20,
    // UNIT_LARGE_TRANSPORTER => 10,
    UNIT_LIGHT_FIGHTER => 10,
    // UNIT_STRONG_FIGHTER => 5,
    UNIT_COLONY_SHIP => 3,
    UNIT_ESPIONAGE_SENSOR => 10,
);

define('GAMESPEED_SCALE', 8);
define('GAMESPEED', 0.75 / GAMESPEED_SCALE);

define('FLY_SPEED_SCALE', 2);
define('ROCKET_SPEED_FACTOR', GAMESPEED_SCALE * FLY_SPEED_SCALE);

for($i = 1; $i <= NUM_GALAXYS; $i++){
    $GLOBALS["GALAXY_SPEC_PARAMS"][$i] = array(
        "ADVANCED_BATTLE" => 0,
        "PLANET_CONSTRUCTION_SPEED_FACTOR" => 1,
        "MOON_CONSTRUCTION_SPEED_FACTOR" => 2,
        "TEMP_MOON_CONSTRUCTION_SPEED_FACTOR" => 4,
        "FLEET_BUILDING_SPEED_FACTOR" => 1,
        "DEFENSE_BUILDING_SPEED_FACTOR" => 1,
        "RESEARCH_SPEED_FACTOR" => 2,
        "RESOURCES_PRODUCTION_FACTOR" => 5,
        "ENEGRY_PRODUCTION_FACTOR" => 1,
        "STORAGE_FACTOR" => 5,
        "FLEET_SPEED_FACTOR" => FLY_SPEED_SCALE,
        "ALLOW_DESTROY_BUILDING" => 0,
        "ALLOW_DESTROY_MOON" => 1,
        "ALLOW_STARGATE_TRANSPORT" => 0,
    );
}
