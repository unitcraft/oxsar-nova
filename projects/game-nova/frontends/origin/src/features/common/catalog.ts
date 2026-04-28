// Каталог юнитов origin-фронта (план 72 Ф.2 Spring 1).
//
// Минимальный справочник unit_id → группа + i18n-ключ имени. Полный
// каталог с уровнями/стоимостью/временем сейчас живёт на backend
// (configs/buildings.yml, research.yml, ships.yml, defense.yml) и
// **не экспортирован** через openapi.yaml в виде агрегированного
// endpoint'а вроде GET /api/catalog. Запись об отсутствии endpoint'а —
// в docs/simplifications.md (план 72 Ф.2 Spring 1, перечень дыр).
//
// i18n-ключи (R12 плана 72): максимально переиспользуем уже существующие
// в `projects/game-nova/configs/i18n/ru.yml` (info-namespace). Единичные
// добавки (heavyFighter, bigTransporter, naniteFactory и т.д.) внесены
// в тот же файл одной правкой.

export interface CatalogEntry {
  id: number;
  group: 'building' | 'research' | 'ship' | 'defense';
  /** namespace.key в i18n (configs/i18n/ru.yml) */
  i18n: string;
}

export const CATALOG: CatalogEntry[] = [
  // Buildings (id-ы из configs/buildings.yml)
  { id: 1,  group: 'building', i18n: 'info.metalmine' },
  { id: 2,  group: 'building', i18n: 'info.siliconLab' },
  { id: 3,  group: 'building', i18n: 'info.hydrogenLab' },
  { id: 4,  group: 'building', i18n: 'info.solarPlant' },
  { id: 12, group: 'building', i18n: 'info.hydrogenPlant' },
  { id: 14, group: 'building', i18n: 'info.roboticFactory' },
  { id: 15, group: 'building', i18n: 'info.nanoFactory' },
  { id: 21, group: 'building', i18n: 'info.shipyard' },
  { id: 22, group: 'building', i18n: 'info.metalStorage' },
  { id: 23, group: 'building', i18n: 'info.siliconStorage' },
  { id: 24, group: 'building', i18n: 'info.hydrogenStorage' },
  { id: 31, group: 'building', i18n: 'info.researchLab' },
  { id: 33, group: 'building', i18n: 'info.terraFormer' },
  { id: 34, group: 'building', i18n: 'info.allianceDepot' },

  // Research (id-ы из configs/research.yml)
  { id: 106, group: 'research', i18n: 'info.spyingTech' },
  { id: 108, group: 'research', i18n: 'info.computerTech' },
  { id: 109, group: 'research', i18n: 'info.gunTech' },
  { id: 110, group: 'research', i18n: 'info.shieldTech' },
  { id: 111, group: 'research', i18n: 'info.shellTech' },
  { id: 113, group: 'research', i18n: 'info.energyTech' },
  { id: 114, group: 'research', i18n: 'info.hyperspaceTech' },
  { id: 115, group: 'research', i18n: 'info.combustionEngine' },
  { id: 117, group: 'research', i18n: 'info.impulseEngine' },
  { id: 118, group: 'research', i18n: 'info.hyperspaceEngine' },

  // Ships (id-ы из configs/ships.yml)
  { id: 202, group: 'ship', i18n: 'info.smallTransporter' },
  { id: 203, group: 'ship', i18n: 'info.bigTransporter' },
  { id: 204, group: 'ship', i18n: 'info.lightFighter' },
  { id: 205, group: 'ship', i18n: 'info.heavyFighter' },
  { id: 206, group: 'ship', i18n: 'info.cruiser' },
  { id: 207, group: 'ship', i18n: 'info.battleShip' },
  { id: 208, group: 'ship', i18n: 'info.colonyShip' },
  { id: 209, group: 'ship', i18n: 'info.recycler' },
  { id: 210, group: 'ship', i18n: 'info.espionageSensor' },
  { id: 211, group: 'ship', i18n: 'info.bomber' },
  { id: 213, group: 'ship', i18n: 'info.starDestroyer' },
  { id: 214, group: 'ship', i18n: 'info.deathStar' },

  // Defense (id-ы из configs/defense.yml)
  { id: 401, group: 'defense', i18n: 'info.rocketLauncher' },
  { id: 402, group: 'defense', i18n: 'info.lightLaser' },
  { id: 403, group: 'defense', i18n: 'info.heavyLaser' },
  { id: 404, group: 'defense', i18n: 'info.gaussGun' },
  { id: 405, group: 'defense', i18n: 'info.ionGun' },
  { id: 406, group: 'defense', i18n: 'info.plasmaGun' },
  { id: 407, group: 'defense', i18n: 'info.smallShield' },
  { id: 408, group: 'defense', i18n: 'info.bigShield' },
  { id: 502, group: 'defense', i18n: 'info.interceptorRocket' },
  { id: 503, group: 'defense', i18n: 'info.interplanetaryRocket' },
];

export function catalogByGroup(
  group: CatalogEntry['group'],
): CatalogEntry[] {
  return CATALOG.filter((e) => e.group === group);
}

export function findCatalog(unitId: number): CatalogEntry | undefined {
  return CATALOG.find((e) => e.id === unitId);
}
