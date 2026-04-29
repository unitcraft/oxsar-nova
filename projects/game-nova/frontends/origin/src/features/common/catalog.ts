// Каталог юнитов origin-фронта (план 72.1 ч.20).
//
// Полный каталог зданий/исследований/кораблей/обороны: ID из
// configs/units.yml, i18n-ключи из configs/i18n/ru.yml (group=info),
// icon — имя файла в public/assets/origin/images/units/{icon}.gif
// (скопированы из projects/game-legacy-php/public/images/buildings/std/).
//
// icon отличается от key потому что в legacy ассетах исторические
// имена не совпадают с текущими ключами units.yml: например
// metal_mine → metalmine.gif, big_transporter → large_transporter.gif.

export interface CatalogEntry {
  id: number;
  group: 'building' | 'research' | 'ship' | 'defense';
  /** namespace.key в i18n (configs/i18n/ru.yml) */
  i18n: string;
  /** Имя файла иконки без расширения */
  icon: string;
  /** moon_only — здание видно только на лунах */
  moonOnly?: boolean;
}

export const CATALOG: CatalogEntry[] = [
  // ────────── Buildings ──────────
  { id: 1,  group: 'building', i18n: 'info.metalmine',         icon: 'metal_mine' },
  { id: 2,  group: 'building', i18n: 'info.siliconLab',        icon: 'silicon_lab' },
  { id: 3,  group: 'building', i18n: 'info.hydrogenLab',       icon: 'hydrogen_lab' },
  { id: 4,  group: 'building', i18n: 'info.solarPlant',        icon: 'solar_plant' },
  { id: 5,  group: 'building', i18n: 'info.hydrogenPlant',     icon: 'hydrogen_plant' },
  { id: 6,  group: 'building', i18n: 'info.roboticFactory',    icon: 'robotic_factory' },
  { id: 7,  group: 'building', i18n: 'info.nanoFactory',       icon: 'nano_factory' },
  { id: 8,  group: 'building', i18n: 'info.shipyard',          icon: 'shipyard' },
  { id: 9,  group: 'building', i18n: 'info.metalStorage',      icon: 'metal_storage' },
  { id: 10, group: 'building', i18n: 'info.siliconStorage',    icon: 'silicon_storage' },
  { id: 11, group: 'building', i18n: 'info.hydrogenStorage',   icon: 'hydrogen_storage' },
  { id: 12, group: 'building', i18n: 'info.researchLab',       icon: 'research_lab' },
  { id: 53, group: 'building', i18n: 'info.rocketStation',     icon: 'rocket_station' },
  { id: 58, group: 'building', i18n: 'info.terraFormer',       icon: 'terra_former' },
  { id: 100,group: 'building', i18n: 'info.repairFactory',     icon: 'repair_factory' },
  { id: 101,group: 'building', i18n: 'info.defenseFactory',    icon: 'defense_factory' },
  { id: 107,group: 'building', i18n: 'info.exchange',          icon: 'exchange' },
  { id: 108,group: 'building', i18n: 'info.exchOffice',        icon: 'exch_office' },
  { id: 54, group: 'building', i18n: 'info.moonBase',          icon: 'moon_base',           moonOnly: true },
  { id: 56, group: 'building', i18n: 'info.moonRoboticFactory',icon: 'moon_robotic_factory',moonOnly: true },
  { id: 57, group: 'building', i18n: 'info.moonHydrogenLab',   icon: 'moon_hydrogen_lab',   moonOnly: true },

  // ────────── Research (configs/research.yml) ──────────
  { id: 13, group: 'research', i18n: 'info.spyware',           icon: 'spyware' },
  { id: 14, group: 'research', i18n: 'info.computerTech',      icon: 'computer_tech' },
  { id: 15, group: 'research', i18n: 'info.gunTech',           icon: 'gun_tech' },
  { id: 16, group: 'research', i18n: 'info.shieldTech',        icon: 'shield_tech' },
  { id: 17, group: 'research', i18n: 'info.shellTech',         icon: 'shell_tech' },
  { id: 18, group: 'research', i18n: 'info.energyTech',        icon: 'energy_tech' },
  { id: 19, group: 'research', i18n: 'info.hyperspaceTech',    icon: 'hyperspace_tech' },
  { id: 20, group: 'research', i18n: 'info.combustionEngine',  icon: 'combustion_engine' },
  { id: 21, group: 'research', i18n: 'info.impulseEngine',     icon: 'impulse_engine' },
  { id: 22, group: 'research', i18n: 'info.hyperspaceEngine',  icon: 'hyperspace_engine' },
  { id: 23, group: 'research', i18n: 'info.laserTech',         icon: 'laser_tech' },
  { id: 24, group: 'research', i18n: 'info.ionTech',           icon: 'ion_tech' },
  { id: 25, group: 'research', i18n: 'info.plasmaTech',        icon: 'plasma_tech' },
  { id: 26, group: 'research', i18n: 'info.ign',               icon: 'ign' },
  { id: 27, group: 'research', i18n: 'info.expoTech',          icon: 'expo_tech' },
  { id: 28, group: 'research', i18n: 'info.gravi',             icon: 'gravi' },
  { id: 103,group: 'research', i18n: 'info.ballisticsTech',    icon: 'ballistics_tech' },
  { id: 104,group: 'research', i18n: 'info.maskingTech',       icon: 'masking_tech' },
  { id: 112,group: 'research', i18n: 'info.astroTech',         icon: 'astro_tech' },
  { id: 113,group: 'research', i18n: 'info.igrTech',           icon: 'igr_tech' },

  // ────────── Ships (configs/ships.yml) ──────────
  { id: 202, group: 'ship', i18n: 'info.smallTransporter',     icon: 'small_transporter' },
  { id: 203, group: 'ship', i18n: 'info.bigTransporter',       icon: 'large_transporter' },
  { id: 204, group: 'ship', i18n: 'info.lightFighter',         icon: 'light_fighter' },
  { id: 205, group: 'ship', i18n: 'info.heavyFighter',         icon: 'strong_fighter' },
  { id: 206, group: 'ship', i18n: 'info.cruiser',              icon: 'cruiser' },
  { id: 207, group: 'ship', i18n: 'info.battleShip',           icon: 'battle_ship' },
  { id: 208, group: 'ship', i18n: 'info.colonyShip',           icon: 'colony_ship' },
  { id: 209, group: 'ship', i18n: 'info.recycler',             icon: 'recycler' },
  { id: 210, group: 'ship', i18n: 'info.espionageSensor',      icon: 'espionage_sensor' },
  { id: 211, group: 'ship', i18n: 'info.bomber',               icon: 'bomber' },
  { id: 213, group: 'ship', i18n: 'info.starDestroyer',        icon: 'frigate' },
  { id: 214, group: 'ship', i18n: 'info.deathStar',            icon: 'death_star' },

  // ────────── Defense (configs/defense.yml) ──────────
  { id: 401, group: 'defense', i18n: 'info.rocketLauncher',    icon: 'rocket_launcher' },
  { id: 402, group: 'defense', i18n: 'info.lightLaser',        icon: 'light_laser' },
  { id: 403, group: 'defense', i18n: 'info.heavyLaser',        icon: 'heavy_laser' },
  { id: 404, group: 'defense', i18n: 'info.gaussGun',          icon: 'gauss_gun' },
  { id: 405, group: 'defense', i18n: 'info.ionGun',            icon: 'ion_gun' },
  { id: 406, group: 'defense', i18n: 'info.plasmaGun',         icon: 'plasma_gun' },
  { id: 407, group: 'defense', i18n: 'info.smallShield',       icon: 'small_shield' },
  { id: 408, group: 'defense', i18n: 'info.bigShield',         icon: 'large_shield' },
  { id: 502, group: 'defense', i18n: 'info.interceptorRocket', icon: 'interceptor_rocket' },
  { id: 503, group: 'defense', i18n: 'info.interplanetaryRocket', icon: 'interplanetary_rocket' },
];

export function catalogByGroup(
  group: CatalogEntry['group'],
): CatalogEntry[] {
  return CATALOG.filter((e) => e.group === group);
}

export function findCatalog(unitId: number): CatalogEntry | undefined {
  return CATALOG.find((e) => e.id === unitId);
}
