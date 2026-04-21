// Мини-каталог юнитов для UI. Значения дублируют configs/ships.yml + defense.yml.
// TODO: сгенерировать из YAML на этапе gen:api (см. CLAUDE.md).

export interface UnitEntry {
  id: number;
  key: string;
  name: string;
}

// CombatEntry — юнит с боевыми характеристиками для симулятора.
export interface CombatEntry extends UnitEntry {
  attack: number;
  shield: number;
  shell: number;
}

export const BUILDINGS: UnitEntry[] = [
  { id: 1,   key: 'metal_mine',       name: 'Рудник металла' },
  { id: 2,   key: 'silicon_lab',      name: 'Рудник по добыче кремния' },
  { id: 3,   key: 'hydrogen_lab',     name: 'Синтезатор водорода' },
  { id: 4,   key: 'solar_plant',      name: 'Солнечная электростанция' },
  { id: 5,   key: 'hydrogen_plant',   name: 'Термоядерная электростанция' },
  { id: 6,   key: 'robotic_factory',  name: 'Фабрика роботов' },
  { id: 8,   key: 'shipyard',         name: 'Верфь' },
  { id: 9,   key: 'metal_storage',    name: 'Хранилище металла' },
  { id: 10,  key: 'silicon_storage',  name: 'Хранилище кремния' },
  { id: 11,  key: 'hydrogen_storage', name: 'Емкость для водорода' },
  { id: 12,  key: 'research_lab',     name: 'Исследовательская лаборатория' },
  { id: 13,  key: 'missile_silo',     name: 'Ракетная шахта' },
  { id: 100, key: 'repair_factory',   name: 'Ремонтный ангар' },
];

export const RESEARCH: UnitEntry[] = [
  { id: 13,  key: 'spyware',           name: 'Шпионаж' },
  { id: 14,  key: 'computer_tech',     name: 'Компьютерная технология' },
  { id: 15,  key: 'gun_tech',          name: 'Оружейная технология' },
  { id: 16,  key: 'shield_tech',       name: 'Щитовая технология' },
  { id: 17,  key: 'shell_tech',        name: 'Броневая технология' },
  { id: 18,  key: 'energy_tech',       name: 'Энергетическая технология' },
  { id: 19,  key: 'hyperspace_tech',   name: 'Гиперпространственная технология' },
  { id: 20,  key: 'combustion_engine', name: 'Реактивный двигатель' },
  { id: 21,  key: 'impulse_engine',    name: 'Импульсный двигатель' },
  { id: 22,  key: 'hyperspace_engine', name: 'Гиперпространственный двигатель' },
  { id: 23,  key: 'laser_tech',        name: 'Лазерная технология' },
  { id: 24,  key: 'ion_tech',          name: 'Ионная технология' },
  { id: 25,  key: 'plasma_tech',       name: 'Плазменная технология' },
  { id: 27,  key: 'expo_tech',         name: 'Экспедиционная технология' },
  { id: 103, key: 'ballistics_tech',   name: 'Баллистическая технология' },
  { id: 104, key: 'masking_tech',      name: 'Маскировочная технология' },
];

export const SHIPS: CombatEntry[] = [
  { id: 29, key: 'small_transporter', name: 'Малый транспорт',    attack: 5,      shield: 10,    shell: 4000 },
  { id: 30, key: 'large_transporter', name: 'Большой транспорт',  attack: 5,      shield: 25,    shell: 12000 },
  { id: 31, key: 'light_fighter',     name: 'Легкий истребитель', attack: 50,     shield: 10,    shell: 4000 },
  { id: 32, key: 'strong_fighter',    name: 'Тяжелый истребитель',attack: 150,    shield: 25,    shell: 10000 },
  { id: 33, key: 'cruiser',           name: 'Крейсер',            attack: 400,    shield: 50,    shell: 27000 },
  { id: 34, key: 'battle_ship',       name: 'Линкор',             attack: 1000,   shield: 200,   shell: 60000 },
  { id: 36, key: 'colony_ship',       name: 'Колонизатор',        attack: 50,     shield: 100,   shell: 30000 },
  { id: 37, key: 'recycler',          name: 'Переработчик',       attack: 1,      shield: 10,    shell: 16000 },
  { id: 38, key: 'espionage_sensor',  name: 'Шпионский зонд',     attack: 0,      shield: 0,     shell: 1000 },
  { id: 39, key: 'solar_satellite',   name: 'Солнечный спутник',  attack: 1,      shield: 1,     shell: 2000 },
  { id: 40, key: 'bomber',            name: 'Бомбардировщик',     attack: 1000,   shield: 500,   shell: 75000 },
  { id: 42, key: 'death_star',        name: 'Звезда смерти',      attack: 200000, shield: 50000, shell: 9000000 },
];

export const DEFENSE: CombatEntry[] = [
  { id: 43, key: 'rocket_launcher', name: 'Ракетная установка',   attack: 80,   shield: 20,    shell: 2000 },
  { id: 44, key: 'light_laser',     name: 'Легкий лазер',         attack: 100,  shield: 25,    shell: 2000 },
  { id: 45, key: 'strong_laser',    name: 'Тяжелый лазер',        attack: 250,  shield: 100,   shell: 8000 },
  { id: 47, key: 'gauss_gun',       name: 'Пушка Гаусса',         attack: 1100, shield: 200,   shell: 35000 },
  { id: 48, key: 'plasma_gun',      name: 'Плазменное орудие',    attack: 3000, shield: 300,   shell: 100000 },
  { id: 49, key: 'small_shield',    name: 'Малый щитовой купол',  attack: 1,    shield: 2000,  shell: 20000 },
  { id: 50, key: 'large_shield',    name: 'Большой щитовой купол',attack: 1,    shield: 10000, shell: 100000 },
];

// Артефакты — только те, что реально реализованы в M5.0.1 (факторы).
// Остальные 300-365 добавятся в M5.1 вместе с one_shot/battle_bonus.
export const ARTEFACTS: UnitEntry[] = [
  { id: 300, key: 'merchants_mark',      name: 'Знак торговца' },
  { id: 301, key: 'catalyst',            name: 'Катализатор' },
  { id: 302, key: 'power_generator',     name: 'Энерготранс' },
  { id: 303, key: 'atomic_densifier',    name: 'Атомный уплотнитель' },
  { id: 305, key: 'supercomputer',       name: 'Суперкомпьютер' },
  { id: 315, key: 'robot_control_system',name: 'Система управления роботами' },
];

// imageOf возвращает путь к иконке юнита из /images/units/ (legacy std skin).
export function imageOf(key: string): string {
  // legacy именует шахту металла как "metalmine"
  const legacyKey = key === 'metal_mine' ? 'metalmine' : key;
  return `/images/units/${legacyKey}.gif`;
}

export function nameOf(id: number): string {
  for (const c of [SHIPS, DEFENSE, RESEARCH, ARTEFACTS, BUILDINGS]) {
    const u = c.find((x) => x.id === id);
    if (u) return u.name;
  }
  return `#${id}`;
}

export function buildingName(id: number): string {
  return BUILDINGS.find((b) => b.id === id)?.name ?? `#${id}`;
}
