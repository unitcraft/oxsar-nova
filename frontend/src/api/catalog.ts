// Мини-каталог юнитов для UI. Значения дублируют configs/ships.yml + defense.yml.
// TODO: сгенерировать из YAML на этапе gen:api (см. CLAUDE.md).

export interface UnitEntry {
  id: number;
  key: string;
  name: string;
  tKey: string; // key in the 'info' i18n group for the unit name
}

// CombatEntry — юнит с боевыми характеристиками для симулятора.
export interface CombatEntry extends UnitEntry {
  attack: number;
  shield: number;
  shell: number;
  cost?: Cost;
  cargo?: number;
  speed?: number;
  fuel?: number;
  requires?: Req[];
  rapidfire?: Record<number, number>;
  // Из legacy ship_datasheet: боевые параметры для UnitInfoScreen
  front?: number;           // приоритет цели (выше = чаще атакуют первым)
  ballistics?: number;      // точность (уровень баллистики)
  masking?: number;         // маскировка (уровень)
  // attacker_* — значения когда юнит выступает в роли атакующего
  attacker_front?: number;
  attacker_ballistics?: number;
  attacker_masking?: number;
}

export interface Cost { metal: number; silicon: number; hydrogen: number }

export interface Req { kind: 'building' | 'research'; key: string; level: number }

export interface BuildingEntry extends UnitEntry {
  costBase: Cost;
  costFactor: number;
  maxLevel?: number;
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

// Форматирование чисел с сокращениями: 1.5M, 2k
export function formatNum(v: number): string {
  if (v >= 1_000_000) return (v / 1_000_000).toFixed(1) + 'M';
  if (v >= 1_000) return (v / 1_000).toFixed(0) + 'K';
  return Math.floor(v).toLocaleString('ru-RU');
}

export const BUILDINGS: BuildingEntry[] = [
  { id: 1,   key: 'metal_mine',       tKey: 'metalmine',        name: 'metal_mine',       costBase: { metal: 60,      silicon: 15,     hydrogen: 0      }, costFactor: 1.5 },
  { id: 2,   key: 'silicon_lab',      tKey: 'siliconLab',       name: 'silicon_lab',      costBase: { metal: 48,      silicon: 24,     hydrogen: 0      }, costFactor: 1.6 },
  { id: 3,   key: 'hydrogen_lab',     tKey: 'hydrogenLab',      name: 'hydrogen_lab',     costBase: { metal: 225,     silicon: 75,     hydrogen: 0      }, costFactor: 1.5 },
  { id: 4,   key: 'solar_plant',      tKey: 'solarPlant',       name: 'solar_plant',      costBase: { metal: 75,      silicon: 30,     hydrogen: 0      }, costFactor: 1.5 },
  { id: 5,   key: 'hydrogen_plant',   tKey: 'hydrogenPlant',    name: 'hydrogen_plant',   costBase: { metal: 900,     silicon: 360,    hydrogen: 180    }, costFactor: 1.8 },
  { id: 6,   key: 'robotic_factory',  tKey: 'roboticFactory',   name: 'robotic_factory',  costBase: { metal: 400,     silicon: 120,    hydrogen: 200    }, costFactor: 2.0 },
  { id: 7,   key: 'nano_factory',     tKey: 'nanoFactory',      name: 'nano_factory',     costBase: { metal: 1000000, silicon: 500000, hydrogen: 100000 }, costFactor: 2.0 },
  { id: 8,   key: 'shipyard',         tKey: 'shipyard',         name: 'shipyard',         costBase: { metal: 400,     silicon: 200,    hydrogen: 100    }, costFactor: 2.0 },
  { id: 9,   key: 'metal_storage',    tKey: 'metalStorage',     name: 'metal_storage',    costBase: { metal: 1000,    silicon: 0,      hydrogen: 0      }, costFactor: 2.0 },
  { id: 10,  key: 'silicon_storage',  tKey: 'siliconStorage',   name: 'silicon_storage',  costBase: { metal: 1000,    silicon: 500,    hydrogen: 0      }, costFactor: 2.0 },
  { id: 11,  key: 'hydrogen_storage', tKey: 'hydrogenStorage',  name: 'hydrogen_storage', costBase: { metal: 1000,    silicon: 1000,   hydrogen: 0      }, costFactor: 2.0 },
  { id: 12,  key: 'research_lab',     tKey: 'researchLab',      name: 'research_lab',     costBase: { metal: 200,     silicon: 400,    hydrogen: 200    }, costFactor: 2.0 },
  { id: 53,  key: 'missile_silo',     tKey: 'rocketStation',    name: 'missile_silo',     costBase: { metal: 20000,   silicon: 20000,  hydrogen: 1000   }, costFactor: 2.0 },
  { id: 100, key: 'repair_factory',   tKey: 'repairFactory',    name: 'repair_factory',   costBase: { metal: 800,     silicon: 400,    hydrogen: 200    }, costFactor: 2.0 },
  { id: 101, key: 'defense_factory',  tKey: 'defenseFactory',   name: 'defense_factory',  costBase: { metal: 350,     silicon: 200,    hydrogen: 100    }, costFactor: 2.0 },
  { id: 107, key: 'exchange',         tKey: 'exchange',         name: 'exchange',         costBase: { metal: 0,       silicon: 1000,   hydrogen: 500    }, costFactor: 2.0 },
  { id: 108, key: 'exch_office',      tKey: 'exchOffice',       name: 'exch_office',      costBase: { metal: 500000,  silicon: 500000, hydrogen: 500000 }, costFactor: 2.0 },
  { id: 58,  key: 'terra_former',     tKey: 'terraFormer',      name: 'terra_former',     costBase: { metal: 50000,   silicon: 50000,  hydrogen: 100000 }, costFactor: 2.0 },
];

export const MOON_BUILDINGS: BuildingEntry[] = [
  { id: 54, key: 'moon_base',            tKey: 'moonBase',           name: 'moon_base',            costBase: { metal: 50000,    silicon: 20000,   hydrogen: 10000   }, costFactor: 2.0 },
  { id: 55, key: 'star_surveillance',    tKey: 'starSurveillance',   name: 'star_surveillance',    costBase: { metal: 100000,   silicon: 20000,   hydrogen: 50000   }, costFactor: 2.0 },
  { id: 56, key: 'star_gate',            tKey: 'starGate',           name: 'star_gate',            costBase: { metal: 4000000,  silicon: 2000000, hydrogen: 1000000 }, costFactor: 2.0 },
  { id: 57, key: 'moon_robotic_factory', tKey: 'moonRoboticFactory', name: 'moon_robotic_factory', costBase: { metal: 10000,    silicon: 6000,    hydrogen: 4000    }, costFactor: 2.0 },
];

export interface ResearchEntry extends UnitEntry {
  costBase: Cost;
  costFactor: number;
  requires?: Req[];
}

export const RESEARCH: ResearchEntry[] = [
  { id: 13,  key: 'spyware',           tKey: 'spyware',           name: 'spyware',           costBase: { metal: 200,    silicon: 1000,   hydrogen: 200    }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 3 }] },
  { id: 14,  key: 'computer_tech',     tKey: 'computerTech',      name: 'computer_tech',     costBase: { metal: 0,      silicon: 400,    hydrogen: 600    }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 1 }] },
  { id: 15,  key: 'gun_tech',          tKey: 'gunTech',           name: 'gun_tech',          costBase: { metal: 800,    silicon: 200,    hydrogen: 0      }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 4 }] },
  { id: 16,  key: 'shield_tech',       tKey: 'shieldTech',        name: 'shield_tech',       costBase: { metal: 200,    silicon: 600,    hydrogen: 0      }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 6 }] },
  { id: 17,  key: 'shell_tech',        tKey: 'shellTech',         name: 'shell_tech',        costBase: { metal: 1000,   silicon: 0,      hydrogen: 0      }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 2 }] },
  { id: 26,  key: 'ign',               tKey: 'ign',               name: 'ign',               costBase: { metal: 0,      silicon: 0,      hydrogen: 0      }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 1 }] },
  { id: 28,  key: 'gravi',             tKey: 'gravi',             name: 'gravi',             costBase: { metal: 0,      silicon: 0,      hydrogen: 0      }, costFactor: 3.0,  requires: [{ kind: 'building', key: 'research_lab', level: 12 }, { kind: 'research', key: 'energy_tech', level: 12 }] },
  { id: 112, key: 'astro_tech',        tKey: 'astroTech',         name: 'astro_tech',        costBase: { metal: 4000,   silicon: 8000,   hydrogen: 4000   }, costFactor: 1.75, requires: [{ kind: 'building', key: 'research_lab', level: 3 }, { kind: 'research', key: 'impulse_engine', level: 3 }, { kind: 'research', key: 'spyware', level: 4 }] },
  { id: 113, key: 'igr_tech',          tKey: 'igrTech',           name: 'igr_tech',          costBase: { metal: 240000, silicon: 400000, hydrogen: 160000 }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 10 }, { kind: 'research', key: 'computer_tech', level: 8 }, { kind: 'research', key: 'hyperspace_tech', level: 8 }] },
  { id: 18,  key: 'energy_tech',       tKey: 'energyTech',        name: 'energy_tech',       costBase: { metal: 0,      silicon: 800,    hydrogen: 400    }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 3 }] },
  { id: 19,  key: 'hyperspace_tech',   tKey: 'hyperspaceTech',    name: 'hyperspace_tech',   costBase: { metal: 0,      silicon: 4000,   hydrogen: 2000   }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 7 }, { kind: 'research', key: 'energy_tech', level: 5 }, { kind: 'research', key: 'shield_tech', level: 5 }] },
  { id: 20,  key: 'combustion_engine', tKey: 'combustionEngine',  name: 'combustion_engine', costBase: { metal: 400,    silicon: 0,      hydrogen: 600    }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 1 }, { kind: 'research', key: 'energy_tech', level: 1 }] },
  { id: 21,  key: 'impulse_engine',    tKey: 'impulseEngine',     name: 'impulse_engine',    costBase: { metal: 2000,   silicon: 4000,   hydrogen: 600    }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 2 }, { kind: 'research', key: 'energy_tech', level: 1 }] },
  { id: 22,  key: 'hyperspace_engine', tKey: 'hyperspaceEngine',  name: 'hyperspace_engine', costBase: { metal: 10000,  silicon: 20000,  hydrogen: 6000   }, costFactor: 3.0,  requires: [{ kind: 'building', key: 'research_lab', level: 7 }, { kind: 'research', key: 'hyperspace_tech', level: 3 }] },
  { id: 23,  key: 'laser_tech',        tKey: 'laserTech',         name: 'laser_tech',        costBase: { metal: 200,    silicon: 100,    hydrogen: 0      }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 1 }, { kind: 'research', key: 'energy_tech', level: 2 }] },
  { id: 24,  key: 'ion_tech',          tKey: 'ionTech',           name: 'ion_tech',          costBase: { metal: 1000,   silicon: 300,    hydrogen: 100    }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 4 }, { kind: 'research', key: 'laser_tech', level: 5 }, { kind: 'research', key: 'energy_tech', level: 4 }] },
  { id: 25,  key: 'plasma_tech',       tKey: 'plasmaTech',        name: 'plasma_tech',       costBase: { metal: 2000,   silicon: 4000,   hydrogen: 1000   }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 4 }, { kind: 'research', key: 'ion_tech', level: 5 }, { kind: 'research', key: 'laser_tech', level: 8 }, { kind: 'research', key: 'energy_tech', level: 8 }] },
  { id: 27,  key: 'expo_tech',         tKey: 'expoTech',          name: 'expo_tech',         costBase: { metal: 4000,   silicon: 8000,   hydrogen: 4000   }, costFactor: 1.75, requires: [{ kind: 'building', key: 'research_lab', level: 3 }, { kind: 'research', key: 'impulse_engine', level: 3 }, { kind: 'research', key: 'spyware', level: 4 }] },
  { id: 103, key: 'ballistics_tech',   tKey: 'ballisticsTech',    name: 'ballistics_tech',   costBase: { metal: 4000,   silicon: 8000,   hydrogen: 4000   }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 5 }, { kind: 'research', key: 'gun_tech', level: 5 }] },
  { id: 104, key: 'masking_tech',      tKey: 'maskingTech',       name: 'masking_tech',      costBase: { metal: 4000,   silicon: 8000,   hydrogen: 4000   }, costFactor: 2.0,  requires: [{ kind: 'building', key: 'research_lab', level: 5 }, { kind: 'research', key: 'hyperspace_tech', level: 3 }] },
];

export const SHIPS: CombatEntry[] = [
  { id: 29,  key: 'small_transporter',      tKey: 'smallTransporter',    name: 'small_transporter',      attack: 5,      shield: 10,    shell: 4000,    cargo: 5000,    speed: 5000,      fuel: 10,   cost: { metal: 2000,    silicon: 2000,    hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 2 }, { kind: 'research', key: 'combustion_engine', level: 2 }], front: 10, ballistics: 0, masking: 0 },
  { id: 30,  key: 'large_transporter',      tKey: 'largeTransporter',    name: 'large_transporter',      attack: 5,      shield: 25,    shell: 12000,   cargo: 25000,   speed: 7500,      fuel: 50,   cost: { metal: 6000,    silicon: 6000,    hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 4 }, { kind: 'research', key: 'combustion_engine', level: 6 }], front: 10, ballistics: 0, masking: 0 },
  { id: 31,  key: 'light_fighter',          tKey: 'lightFighter',        name: 'light_fighter',          attack: 50,     shield: 10,    shell: 4000,    cargo: 50,      speed: 12500,     fuel: 20,   cost: { metal: 3000,    silicon: 1000,    hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 1 }, { kind: 'research', key: 'combustion_engine', level: 1 }], front: 10, ballistics: 0, masking: 0 },
  { id: 32,  key: 'strong_fighter',         tKey: 'strongFighter',       name: 'strong_fighter',         attack: 150,    shield: 25,    shell: 10000,   cargo: 100,     speed: 10000,     fuel: 75,   cost: { metal: 6000,    silicon: 4000,    hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 3 }, { kind: 'research', key: 'impulse_engine', level: 2 }, { kind: 'research', key: 'shell_tech', level: 2 }], front: 10, ballistics: 0, masking: 0 },
  { id: 33,  key: 'cruiser',                tKey: 'cruiser',             name: 'cruiser',                attack: 400,    shield: 50,    shell: 27000,   cargo: 800,     speed: 15000,     fuel: 300,  cost: { metal: 20000,   silicon: 7000,    hydrogen: 2000    }, requires: [{ kind: 'building', key: 'shipyard', level: 5 }, { kind: 'research', key: 'impulse_engine', level: 4 }, { kind: 'research', key: 'ion_tech', level: 2 }], front: 10, ballistics: 0, masking: 0, rapidfire: { 31: 6, 43: 10 } },
  { id: 34,  key: 'battle_ship',            tKey: 'battleShip',          name: 'battle_ship',            attack: 1000,   shield: 200,   shell: 60000,   cargo: 1500,    speed: 10000,     fuel: 500,  cost: { metal: 45000,   silicon: 15000,   hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 7 }, { kind: 'research', key: 'hyperspace_engine', level: 4 }], front: 10, ballistics: 0, masking: 0, rapidfire: { 38: 5, 39: 5 } },
  { id: 35,  key: 'frigate',                tKey: 'frigate',             name: 'frigate',                attack: 700,    shield: 400,   shell: 70000,   cargo: 750,     speed: 10000,     fuel: 250,  cost: { metal: 30000,   silicon: 12500,   hydrogen: 500     }, requires: [{ kind: 'building', key: 'shipyard', level: 6 }, { kind: 'research', key: 'impulse_engine', level: 4 }, { kind: 'research', key: 'shield_tech', level: 5 }], front: 10, ballistics: 0, masking: 0 },
  { id: 36,  key: 'colony_ship',            tKey: 'colonyShip',          name: 'colony_ship',            attack: 50,     shield: 100,   shell: 30000,   cargo: 7500,    speed: 2500,      fuel: 1000, cost: { metal: 10000,   silicon: 20000,   hydrogen: 10000   }, requires: [{ kind: 'building', key: 'shipyard', level: 4 }, { kind: 'research', key: 'impulse_engine', level: 3 }], front: 10, ballistics: 0, masking: 0 },
  { id: 37,  key: 'recycler',               tKey: 'recycler',            name: 'recycler',               attack: 1,      shield: 10,    shell: 16000,   cargo: 20000,   speed: 2000,      fuel: 300,  cost: { metal: 12500,   silicon: 2500,    hydrogen: 10000   }, requires: [{ kind: 'building', key: 'shipyard', level: 4 }, { kind: 'research', key: 'combustion_engine', level: 6 }, { kind: 'research', key: 'shield_tech', level: 2 }], front: 10, ballistics: 0, masking: 0 },
  { id: 38,  key: 'espionage_sensor',       tKey: 'espionageSensor',     name: 'espionage_sensor',       attack: 0,      shield: 0,     shell: 1000,    cargo: 5,       speed: 100000000, fuel: 1,    cost: { metal: 0,       silicon: 1000,    hydrogen: 0       }, requires: [{ kind: 'building', key: 'shipyard', level: 3 }, { kind: 'research', key: 'combustion_engine', level: 3 }, { kind: 'research', key: 'spyware', level: 2 }], front: 10, ballistics: 0, masking: 0 },
  { id: 39,  key: 'solar_satellite',        tKey: 'solarSatellite',      name: 'solar_satellite',        attack: 1,      shield: 1,     shell: 2000,                    speed: 5000,      fuel: 0,    cost: { metal: 0,       silicon: 2000,    hydrogen: 500     }, requires: [{ kind: 'building', key: 'shipyard', level: 1 }], front: 10, ballistics: 0, masking: 0 },
  { id: 40,  key: 'bomber',                 tKey: 'bomber',              name: 'bomber',                 attack: 1000,   shield: 500,   shell: 75000,   cargo: 500,     speed: 4000,      fuel: 700,  cost: { metal: 50000,   silicon: 25000,   hydrogen: 15000   }, requires: [{ kind: 'building', key: 'shipyard', level: 8 }, { kind: 'research', key: 'impulse_engine', level: 6 }, { kind: 'research', key: 'plasma_tech', level: 5 }], front: 10, ballistics: 0, masking: 0 },
  { id: 41,  key: 'star_destroyer',         tKey: 'starDestroyer',       name: 'star_destroyer',         attack: 2000,   shield: 500,   shell: 110000,  cargo: 2000,    speed: 5000,      fuel: 1000, cost: { metal: 60000,   silicon: 50000,   hydrogen: 15000   }, requires: [{ kind: 'building', key: 'shipyard', level: 9 }, { kind: 'research', key: 'hyperspace_tech', level: 5 }, { kind: 'research', key: 'hyperspace_engine', level: 6 }, { kind: 'research', key: 'gun_tech', level: 7 }], front: 10, ballistics: 0, masking: 0 },
  { id: 42,  key: 'death_star',             tKey: 'deathStar',           name: 'death_star',             attack: 200000, shield: 50000, shell: 9000000, cargo: 1000000, speed: 100,       fuel: 1,    cost: { metal: 5000000, silicon: 4000000, hydrogen: 1000000 }, requires: [{ kind: 'building', key: 'shipyard', level: 12 }, { kind: 'research', key: 'hyperspace_tech', level: 6 }, { kind: 'research', key: 'hyperspace_engine', level: 7 }], front: 10, ballistics: 4, masking: 0, attacker_front: 9, attacker_ballistics: 4, attacker_masking: 0, rapidfire: { 29: 250, 30: 250, 31: 200, 32: 100, 33: 33, 34: 30, 37: 250, 38: 1250, 39: 1250, 40: 25, 41: 5, 43: 200, 44: 200, 45: 100, 46: 100, 47: 50, 48: 50 } },
  { id: 52,  key: 'interplanetary_missile', tKey: 'interplanetaryRocket', name: 'interplanetary_missile', attack: 0,      shield: 0,     shell: 1500,    cargo: 0,       speed: 12000,     fuel: 1,    cost: { metal: 12500,   silicon: 2500,    hydrogen: 10000   }, requires: [{ kind: 'building', key: 'missile_silo', level: 1 }, { kind: 'research', key: 'impulse_engine', level: 1 }], front: 10, ballistics: 0, masking: 0 },
  { id: 102, key: 'lancer_ship',            tKey: 'lancerShip',          name: 'lancer_ship',            attack: 5500,   shield: 200,   shell: 10000,   cargo: 200,     speed: 8000,      fuel: 100,  cost: { metal: 15000,   silicon: 35000,   hydrogen: 60000   }, requires: [{ kind: 'building', key: 'shipyard', level: 7 }, { kind: 'research', key: 'hyperspace_tech', level: 5 }], front: 8,  ballistics: 0, masking: 0, rapidfire: { 42: 3 } },
  { id: 325, key: 'shadow_ship',            tKey: 'shadowShip',          name: 'shadow_ship',            attack: 200,    shield: 30,    shell: 4000,    cargo: 75,      speed: 13000,     fuel: 35,   cost: { metal: 1000,    silicon: 3000,    hydrogen: 1000    }, requires: [{ kind: 'building', key: 'shipyard', level: 4 }, { kind: 'research', key: 'masking_tech', level: 3 }], front: 7,  ballistics: 5, masking: 5, rapidfire: { 30: 15, 31: 5, 35: 20, 36: 20, 37: 30, 40: 25, 41: 15, 42: 70 } },
  // Корабли пришельцев (200-204) — не строятся игроком, появляются в AI-атаках.
  { id: 200, key: 'unit_a_corvette',        tKey: 'unitACorvette',       name: 'unit_a_corvette',        attack: 200,    shield: 75,    shell: 25500,   cargo: 300,     speed: 20000,     fuel: 150,  cost: { metal: 6000,    silicon: 3000,    hydrogen: 1500    }, front: 10, ballistics: 0, masking: 0 },
  { id: 201, key: 'unit_a_screen',          tKey: 'unitAScreen',         name: 'unit_a_screen',          attack: 22,     shield: 5000,  shell: 30000,   cargo: 800,     speed: 10000,     fuel: 75,   cost: { metal: 8000,    silicon: 6000,    hydrogen: 2000    }, front: 10, ballistics: 0, masking: 0 },
  { id: 202, key: 'unit_a_paladin',         tKey: 'unitAPaladin',        name: 'unit_a_paladin',         attack: 75,     shield: 50,    shell: 4200,    cargo: 50,      speed: 8000,      fuel: 20,   cost: { metal: 1500,    silicon: 1000,    hydrogen: 0       }, front: 10, ballistics: 0, masking: 0 },
  { id: 203, key: 'unit_a_frigate',         tKey: 'unitAFrigate',        name: 'unit_a_frigate',         attack: 1250,   shield: 150,   shell: 70000,   cargo: 2000,    speed: 10000,     fuel: 300,  cost: { metal: 25000,   silicon: 15000,   hydrogen: 5000    }, front: 10, ballistics: 0, masking: 0 },
  { id: 204, key: 'unit_a_torpedocarier',   tKey: 'unitATorpedocarier',  name: 'unit_a_torpedocarier',   attack: 350,    shield: 100,   shell: 11000,   cargo: 200,     speed: 13000,     fuel: 100,  cost: { metal: 4000,    silicon: 2500,    hydrogen: 500     }, front: 10, ballistics: 0, masking: 0 },
];

export const DEFENSE: CombatEntry[] = [
  { id: 43, key: 'rocket_launcher', tKey: 'rocketLauncher', name: 'rocket_launcher', attack: 80,   shield: 20,    shell: 2000,   cost: { metal: 2000,  silicon: 0,     hydrogen: 0     }, requires: [{ kind: 'building', key: 'shipyard', level: 1 }], front: 10, ballistics: 0, masking: 0 },
  { id: 44, key: 'light_laser',     tKey: 'lightLaser',     name: 'light_laser',     attack: 100,  shield: 25,    shell: 2000,   cost: { metal: 1500,  silicon: 500,   hydrogen: 0     }, requires: [{ kind: 'building', key: 'shipyard', level: 2 }, { kind: 'research', key: 'laser_tech', level: 3 }], front: 10, ballistics: 0, masking: 0 },
  { id: 45, key: 'strong_laser',    tKey: 'strongLaser',    name: 'strong_laser',    attack: 250,  shield: 100,   shell: 8000,   cost: { metal: 6000,  silicon: 2000,  hydrogen: 0     }, requires: [{ kind: 'building', key: 'shipyard', level: 4 }, { kind: 'research', key: 'laser_tech', level: 6 }, { kind: 'research', key: 'energy_tech', level: 3 }], front: 10, ballistics: 0, masking: 1 },
  { id: 46, key: 'ion_gun',         tKey: 'ionGun',         name: 'ion_gun',         attack: 150,  shield: 500,   shell: 8000,   cost: { metal: 8000,  silicon: 2000,  hydrogen: 500   }, requires: [{ kind: 'building', key: 'defense_factory', level: 4 }, { kind: 'research', key: 'ion_tech', level: 4 }], front: 10, ballistics: 0, masking: 0 },
  { id: 47, key: 'gauss_gun',       tKey: 'gaussGun',       name: 'gauss_gun',       attack: 1100, shield: 200,   shell: 35000,  cost: { metal: 20000, silicon: 15000, hydrogen: 2000  }, requires: [{ kind: 'building', key: 'shipyard', level: 6 }, { kind: 'research', key: 'gun_tech', level: 3 }, { kind: 'research', key: 'shield_tech', level: 1 }, { kind: 'research', key: 'energy_tech', level: 6 }], front: 10, ballistics: 1, masking: 2 },
  { id: 48, key: 'plasma_gun',      tKey: 'plasmaGun',      name: 'plasma_gun',      attack: 3000, shield: 300,   shell: 100000, cost: { metal: 50000, silicon: 50000, hydrogen: 30000 }, requires: [{ kind: 'building', key: 'shipyard', level: 8 }, { kind: 'research', key: 'plasma_tech', level: 7 }], front: 10, ballistics: 2, masking: 2 },
  { id: 49, key: 'small_shield',    tKey: 'smallShield',    name: 'small_shield',    attack: 1,    shield: 2000,  shell: 20000,  cost: { metal: 10000, silicon: 10000, hydrogen: 0     }, requires: [{ kind: 'building', key: 'shipyard', level: 1 }, { kind: 'research', key: 'shield_tech', level: 2 }], front: 16, ballistics: 0, masking: 0 },
  { id: 50, key: 'large_shield',    tKey: 'largeShield',    name: 'large_shield',    attack: 1,    shield: 10000, shell: 100000, cost: { metal: 50000, silicon: 50000, hydrogen: 0     }, requires: [{ kind: 'building', key: 'shipyard', level: 6 }, { kind: 'research', key: 'shield_tech', level: 6 }], front: 18, ballistics: 0, masking: 0 },
];

// Артефакты — только те, что реально реализованы в M5.0.1 (факторы).
// Остальные 300-365 добавятся в M5.1 вместе с one_shot/battle_bonus.
interface ArtefactEntry extends UnitEntry { lifetimeDays: number }
export const ARTEFACTS: ArtefactEntry[] = [
  { id: 300, key: 'merchants_mark',       tKey: 'merchantsMark',       name: 'merchants_mark',       lifetimeDays: 7 },
  { id: 301, key: 'catalyst',             tKey: 'catalyst',            name: 'catalyst',             lifetimeDays: 7 },
  { id: 302, key: 'power_generator',      tKey: 'powerGenerator',      name: 'power_generator',      lifetimeDays: 7 },
  { id: 303, key: 'atomic_densifier',     tKey: 'atomicDensifier',     name: 'atomic_densifier',     lifetimeDays: 7 },
  { id: 305, key: 'supercomputer',        tKey: 'supercomputer',       name: 'supercomputer',        lifetimeDays: 7 },
  { id: 315, key: 'robot_control_system', tKey: 'robotControlSystem',  name: 'robot_control_system', lifetimeDays: 7 },
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
  for (const c of [SHIPS, DEFENSE, RESEARCH, ARTEFACTS, BUILDINGS, MOON_BUILDINGS]) {
    const u = c.find((x) => x.id === id);
    if (u) return imageOf(u.key);
  }
  return '';
}

// nameByKey возвращает отображаемое имя юнита по ключу (fallback на key).
function nameByKey(key: string): string {
  for (const c of [...BUILDINGS, ...MOON_BUILDINGS, ...RESEARCH, ...SHIPS, ...DEFENSE]) {
    if (c.key === key) return c.name;
  }
  return key;
}

// fmtReqs форматирует список требований в читаемую строку.
export function fmtReqs(reqs: Req[]): string {
  return reqs.map((r) => `${nameByKey(r.key)} ур.${r.level}`).join(' + ');
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

// nameOf returns the i18n name for a unit by id.
// Pass a t function from useTranslation('info') for translated output;
// omit for a snake_case fallback (e.g. in non-component contexts).
export function nameOf(id: number, t?: (key: string) => string): string {
  for (const c of [SHIPS, DEFENSE, RESEARCH, ARTEFACTS, BUILDINGS, MOON_BUILDINGS]) {
    const u = c.find((x) => x.id === id);
    if (u) return t ? t(u.tKey) : u.name;
  }
  return `#${id}`;
}

// keyOfId — ключ юнита по числовому id (для wiki-навигации, slug страницы).
export function keyOfId(id: number): string | null {
  for (const c of [SHIPS, DEFENSE, RESEARCH, ARTEFACTS, BUILDINGS, MOON_BUILDINGS]) {
    const u = c.find((x) => x.id === id);
    if (u) return u.key;
  }
  return null;
}

// categoryOfId — wiki-категория для unit_id. ships/defense/buildings/research.
export function categoryOfId(id: number): string | null {
  if (SHIPS.find((x) => x.id === id)) return 'ships';
  if (DEFENSE.find((x) => x.id === id)) return 'defense';
  if (BUILDINGS.find((x) => x.id === id)) return 'buildings';
  if (MOON_BUILDINGS.find((x) => x.id === id)) return 'buildings';
  if (RESEARCH.find((x) => x.id === id)) return 'research';
  return null;
}

export function buildingName(id: number, t?: (key: string) => string): string {
  const b = BUILDINGS.find((x) => x.id === id);
  if (!b) return `#${id}`;
  return t ? t(b.tKey) : b.name;
}

// API functions for resource management
import { api } from './client';
import type { ResourceReport } from './types';

export const resourceAPI = {
  getResourceReport: (planetId: string) =>
    api.get<ResourceReport>(`/api/planets/${planetId}/resource-report`),

  updateResourceFactors: (planetId: string, payload: { factors: Record<string, number> }) =>
    api.post<{ status: string }>(`/api/planets/${planetId}/resource-update`, payload),
};

// Re-export for convenience
export const catalog = {
  getResourceReport: resourceAPI.getResourceReport,
  updateResourceFactors: resourceAPI.updateResourceFactors,
};
