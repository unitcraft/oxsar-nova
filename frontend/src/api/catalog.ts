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
  cost?: Cost;
  cargo?: number;
}

export interface Cost { metal: number; silicon: number; hydrogen: number }

export interface BuildingEntry extends UnitEntry {
  costBase: Cost;
  costFactor: number;
}

// costForLevel: cost_base * cost_factor^(level-1), округление вниз.
export function costForLevel(base: Cost, factor: number, level: number): Cost {
  const m = factor ** (level - 1);
  return {
    metal:    Math.floor(base.metal    * m),
    silicon:  Math.floor(base.silicon  * m),
    hydrogen: Math.floor(base.hydrogen * m),
  };
}

export const BUILDINGS: BuildingEntry[] = [
  { id: 1,   key: 'metal_mine',       name: 'Рудник металла',                costBase: { metal: 60,    silicon: 15,   hydrogen: 0   }, costFactor: 1.5 },
  { id: 2,   key: 'silicon_lab',      name: 'Рудник по добыче кремния',      costBase: { metal: 48,    silicon: 24,   hydrogen: 0   }, costFactor: 1.6 },
  { id: 3,   key: 'hydrogen_lab',     name: 'Синтезатор водорода',            costBase: { metal: 225,   silicon: 75,   hydrogen: 0   }, costFactor: 1.5 },
  { id: 4,   key: 'solar_plant',      name: 'Солнечная электростанция',       costBase: { metal: 75,    silicon: 30,   hydrogen: 0   }, costFactor: 1.5 },
  { id: 5,   key: 'hydrogen_plant',   name: 'Термоядерная электростанция',    costBase: { metal: 900,   silicon: 360,  hydrogen: 180 }, costFactor: 1.8 },
  { id: 6,   key: 'robotic_factory',  name: 'Фабрика роботов',               costBase: { metal: 400,   silicon: 120,  hydrogen: 200 }, costFactor: 2.0 },
  { id: 8,   key: 'shipyard',         name: 'Верфь',                         costBase: { metal: 400,   silicon: 200,  hydrogen: 100 }, costFactor: 2.0 },
  { id: 9,   key: 'metal_storage',    name: 'Хранилище металла',             costBase: { metal: 1000,  silicon: 0,    hydrogen: 0   }, costFactor: 2.0 },
  { id: 10,  key: 'silicon_storage',  name: 'Хранилище кремния',             costBase: { metal: 1000,  silicon: 500,  hydrogen: 0   }, costFactor: 2.0 },
  { id: 11,  key: 'hydrogen_storage', name: 'Емкость для водорода',           costBase: { metal: 1000,  silicon: 1000, hydrogen: 0   }, costFactor: 2.0 },
  { id: 12,  key: 'research_lab',     name: 'Исследовательская лаборатория', costBase: { metal: 200,   silicon: 400,  hydrogen: 200 }, costFactor: 2.0 },
  { id: 13,  key: 'missile_silo',     name: 'Ракетная шахта',                costBase: { metal: 20000, silicon: 20000,hydrogen: 1000}, costFactor: 2.0 },
  { id: 100, key: 'repair_factory',   name: 'Ремонтный ангар',               costBase: { metal: 800,   silicon: 400,  hydrogen: 200 }, costFactor: 2.0 },
];

export interface ResearchEntry extends UnitEntry {
  costBase: Cost;
  costFactor: number;
  benefit: string;
}

export const RESEARCH: ResearchEntry[] = [
  { id: 13,  key: 'spyware',           name: 'Шпионаж',                              benefit: '+1 уровень шпионажа зонда',                costBase: { metal: 200,   silicon: 1000,  hydrogen: 200  }, costFactor: 2.0  },
  { id: 14,  key: 'computer_tech',     name: 'Компьютерная технология',              benefit: '+1 слот флота',                            costBase: { metal: 0,     silicon: 400,   hydrogen: 600  }, costFactor: 2.0  },
  { id: 15,  key: 'gun_tech',          name: 'Оружейная технология',                 benefit: '+2% атака флота и обороны',                costBase: { metal: 800,   silicon: 200,   hydrogen: 0    }, costFactor: 2.0  },
  { id: 16,  key: 'shield_tech',       name: 'Щитовая технология',                   benefit: '+2% щит флота и обороны',                  costBase: { metal: 200,   silicon: 600,   hydrogen: 0    }, costFactor: 2.0  },
  { id: 17,  key: 'shell_tech',        name: 'Броневая технология',                  benefit: '+2% броня флота и обороны',                costBase: { metal: 1000,  silicon: 0,     hydrogen: 0    }, costFactor: 2.0  },
  { id: 18,  key: 'energy_tech',       name: 'Энергетическая технология',            benefit: 'требование для высоких технологий',        costBase: { metal: 0,     silicon: 800,   hydrogen: 400  }, costFactor: 2.0  },
  { id: 19,  key: 'hyperspace_tech',   name: 'Гиперпространственная технология',     benefit: 'требование для гипердвигателя',            costBase: { metal: 0,     silicon: 4000,  hydrogen: 2000 }, costFactor: 2.0  },
  { id: 20,  key: 'combustion_engine', name: 'Реактивный двигатель',                 benefit: '+10% скорость транспортов и истребителей', costBase: { metal: 400,   silicon: 0,     hydrogen: 600  }, costFactor: 2.0  },
  { id: 21,  key: 'impulse_engine',    name: 'Импульсный двигатель',                 benefit: '+20% скорость крейсеров и зондов',         costBase: { metal: 2000,  silicon: 4000,  hydrogen: 600  }, costFactor: 2.0  },
  { id: 22,  key: 'hyperspace_engine', name: 'Гиперпространственный двигатель',      benefit: '+30% скорость линкоров и флагманов',       costBase: { metal: 10000, silicon: 20000, hydrogen: 6000 }, costFactor: 3.0  },
  { id: 23,  key: 'laser_tech',        name: 'Лазерная технология',                  benefit: 'требование для ионной технологии',         costBase: { metal: 200,   silicon: 100,   hydrogen: 0    }, costFactor: 2.0  },
  { id: 24,  key: 'ion_tech',          name: 'Ионная технология',                    benefit: 'требование для плазменной технологии',     costBase: { metal: 1000,  silicon: 300,   hydrogen: 100  }, costFactor: 2.0  },
  { id: 25,  key: 'plasma_tech',       name: 'Плазменная технология',                benefit: 'повышенный урон по ресурсам противника',   costBase: { metal: 2000,  silicon: 4000,  hydrogen: 1000 }, costFactor: 2.0  },
  { id: 27,  key: 'expo_tech',         name: 'Экспедиционная технология',            benefit: '+1 слот экспедиции за уровень',            costBase: { metal: 4000,  silicon: 8000,  hydrogen: 4000 }, costFactor: 1.75 },
  { id: 103, key: 'ballistics_tech',   name: 'Баллистическая технология',            benefit: '+1 ракета в шахте за уровень',             costBase: { metal: 4000,  silicon: 8000,  hydrogen: 4000 }, costFactor: 2.0  },
  { id: 104, key: 'masking_tech',      name: 'Маскировочная технология',             benefit: 'снижение видимости флота для шпионажа',    costBase: { metal: 4000,  silicon: 8000,  hydrogen: 4000 }, costFactor: 2.0  },
];

export const SHIPS: CombatEntry[] = [
  { id: 29, key: 'small_transporter', name: 'Малый транспорт',    attack: 5,      shield: 10,    shell: 4000,    cargo: 5000,    cost: { metal: 2000,    silicon: 2000,   hydrogen: 0       } },
  { id: 30, key: 'large_transporter', name: 'Большой транспорт',  attack: 5,      shield: 25,    shell: 12000,   cargo: 25000,   cost: { metal: 6000,    silicon: 6000,   hydrogen: 0       } },
  { id: 31, key: 'light_fighter',     name: 'Легкий истребитель', attack: 50,     shield: 10,    shell: 4000,    cargo: 50,      cost: { metal: 3000,    silicon: 1000,   hydrogen: 0       } },
  { id: 32, key: 'strong_fighter',    name: 'Тяжелый истребитель',attack: 150,    shield: 25,    shell: 10000,   cargo: 100,     cost: { metal: 6000,    silicon: 4000,   hydrogen: 0       } },
  { id: 33, key: 'cruiser',           name: 'Крейсер',            attack: 400,    shield: 50,    shell: 27000,   cargo: 800,     cost: { metal: 20000,   silicon: 7000,   hydrogen: 2000    } },
  { id: 34, key: 'battle_ship',       name: 'Линкор',             attack: 1000,   shield: 200,   shell: 60000,   cargo: 1500,    cost: { metal: 45000,   silicon: 15000,  hydrogen: 0       } },
  { id: 36, key: 'colony_ship',       name: 'Колонизатор',        attack: 50,     shield: 100,   shell: 30000,   cargo: 7500,    cost: { metal: 10000,   silicon: 20000,  hydrogen: 10000   } },
  { id: 37, key: 'recycler',          name: 'Переработчик',       attack: 1,      shield: 10,    shell: 16000,   cargo: 20000,   cost: { metal: 12500,   silicon: 2500,   hydrogen: 10000   } },
  { id: 38, key: 'espionage_sensor',  name: 'Шпионский зонд',     attack: 0,      shield: 0,     shell: 1000,    cargo: 5,       cost: { metal: 0,       silicon: 1000,   hydrogen: 0       } },
  { id: 39, key: 'solar_satellite',   name: 'Солнечный спутник',  attack: 1,      shield: 1,     shell: 2000,                    cost: { metal: 0,       silicon: 2000,   hydrogen: 500     } },
  { id: 40, key: 'bomber',            name: 'Бомбардировщик',     attack: 1000,   shield: 500,   shell: 75000,   cargo: 500,     cost: { metal: 50000,   silicon: 25000,  hydrogen: 15000   } },
  { id: 42, key: 'death_star',        name: 'Звезда смерти',      attack: 200000, shield: 50000, shell: 9000000, cargo: 1000000, cost: { metal: 5000000, silicon: 4000000,hydrogen: 1000000 } },
];

export const DEFENSE: CombatEntry[] = [
  { id: 43, key: 'rocket_launcher', name: 'Ракетная установка',   attack: 80,   shield: 20,    shell: 2000,    cost: { metal: 2000,  silicon: 0,    hydrogen: 0    } },
  { id: 44, key: 'light_laser',     name: 'Легкий лазер',         attack: 100,  shield: 25,    shell: 2000,    cost: { metal: 1500,  silicon: 500,  hydrogen: 0    } },
  { id: 45, key: 'strong_laser',    name: 'Тяжелый лазер',        attack: 250,  shield: 100,   shell: 8000,    cost: { metal: 6000,  silicon: 2000, hydrogen: 0    } },
  { id: 47, key: 'gauss_gun',       name: 'Пушка Гаусса',         attack: 1100, shield: 200,   shell: 35000,   cost: { metal: 20000, silicon: 15000,hydrogen: 2000 } },
  { id: 48, key: 'plasma_gun',      name: 'Плазменное орудие',    attack: 3000, shield: 300,   shell: 100000,  cost: { metal: 50000, silicon: 50000,hydrogen: 30000} },
  { id: 49, key: 'small_shield',    name: 'Малый щитовой купол',  attack: 1,    shield: 2000,  shell: 20000,   cost: { metal: 10000, silicon: 10000,hydrogen: 0    } },
  { id: 50, key: 'large_shield',    name: 'Большой щитовой купол',attack: 1,    shield: 10000, shell: 100000,  cost: { metal: 50000, silicon: 50000,hydrogen: 0    } },
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

const KEY_MAP: Record<string, string> = {
  metal_mine:   'metalmine',
  missile_silo: 'rocket_station',
};

// imageOf возвращает путь к иконке юнита из /images/units/ (legacy std skin).
export function imageOf(key: string): string {
  return `/images/units/${KEY_MAP[key] ?? key}.gif`;
}

// imageOfId возвращает путь к иконке юнита по его числовому id.
export function imageOfId(id: number): string {
  for (const c of [SHIPS, DEFENSE, RESEARCH, ARTEFACTS, BUILDINGS]) {
    const u = c.find((x) => x.id === id);
    if (u) return imageOf(u.key);
  }
  return '';
}

// Тип планеты по позиции в системе (из PlanetPictures.xml legacy).
const PLANET_TYPES: Array<{ name: string; from: number; to: number; count: number }> = [
  { name: 'trockenplanet',    from: 1,  to: 4,  count: 10 },
  { name: 'wuestenplanet',    from: 1,  to: 3,  count: 4  },
  { name: 'dschjungelplanet', from: 3,  to: 7,  count: 10 },
  { name: 'normaltempplanet', from: 6,  to: 10, count: 7  },
  { name: 'wasserplanet',     from: 9,  to: 13, count: 9  },
  { name: 'eisplanet',        from: 12, to: 15, count: 10 },
  { name: 'gasplanet',        from: 13, to: 15, count: 8  },
];

// planetImageOf возвращает путь к картинке планеты.
// Если передан planetType — используется напрямую (из БД).
// Иначе — детерминировано из позиции слота и хэша id планеты.
export function planetImageOf(position: number, planetId: string, planetType?: string): string {
  // простой хэш из id планеты
  let h = 0;
  for (let i = 0; i < planetId.length; i++) h = (h * 31 + planetId.charCodeAt(i)) >>> 0;

  let type: (typeof PLANET_TYPES)[number];
  if (planetType && planetType !== 'moon') {
    const found = PLANET_TYPES.find((t) => t.name === planetType);
    type = found ?? PLANET_TYPES[3]!;
  } else {
    const eligible = PLANET_TYPES.filter((t) => position >= t.from && position <= t.to);
    const types = eligible.length > 0 ? eligible : [PLANET_TYPES[3]!];
    type = types[h % types.length]!;
  }
  const num = (h % type.count) + 1;
  return `/images/planets/${type.name}${String(num).padStart(2, '0')}.jpg`;
}

// planetImageSize возвращает размер в пикселях для отображения планеты по диаметру.
// Диапазон диаметров: ~2000 (луна) до ~17000. Планеты: 32px..64px.
export function planetImageSize(diameter?: number): number {
  if (!diameter) return 48;
  return 32 + Math.round((Math.min(diameter, 17000) / 17000) * 32);
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
