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
  { id: 1, key: 'metal_mine', name: 'Metal Mine' },
  { id: 2, key: 'silicon_lab', name: 'Silicon Lab' },
  { id: 3, key: 'hydrogen_lab', name: 'Hydrogen Lab' },
  { id: 4, key: 'solar_plant', name: 'Solar Plant' },
  { id: 6, key: 'robotic_factory', name: 'Robotic Factory' },
  { id: 8, key: 'shipyard', name: 'Shipyard' },
  { id: 9, key: 'metal_storage', name: 'Metal Storage' },
  { id: 10, key: 'silicon_storage', name: 'Silicon Storage' },
  { id: 11, key: 'hydrogen_storage', name: 'Hydrogen Storage' },
  { id: 12, key: 'research_lab', name: 'Research Lab' },
];

export const RESEARCH: UnitEntry[] = [
  { id: 13, key: 'spyware', name: 'Espionage' },
  { id: 14, key: 'computer_tech', name: 'Computer Tech' },
  { id: 15, key: 'gun_tech', name: 'Weapons' },
  { id: 16, key: 'shield_tech', name: 'Shielding' },
  { id: 17, key: 'shell_tech', name: 'Armor' },
  { id: 18, key: 'energy_tech', name: 'Energy' },
  { id: 19, key: 'hyperspace_tech', name: 'Hyperspace' },
  { id: 20, key: 'combustion_engine', name: 'Combustion Drive' },
  { id: 21, key: 'impulse_engine', name: 'Impulse Drive' },
  { id: 22, key: 'hyperspace_engine', name: 'Hyperspace Drive' },
  { id: 23, key: 'laser_tech', name: 'Laser' },
  { id: 24, key: 'ion_tech', name: 'Ion' },
  { id: 25, key: 'plasma_tech', name: 'Plasma' },
  { id: 27, key: 'expo_tech', name: 'Expedition' },
  { id: 103, key: 'ballistics_tech', name: 'Ballistics' },
  { id: 104, key: 'masking_tech', name: 'Masking' },
];

export const SHIPS: CombatEntry[] = [
  { id: 29, key: 'small_transporter', name: 'Small Transporter', attack: 5,      shield: 10,    shell: 4000 },
  { id: 30, key: 'large_transporter', name: 'Large Transporter', attack: 5,      shield: 25,    shell: 12000 },
  { id: 31, key: 'light_fighter',     name: 'Light Fighter',     attack: 50,     shield: 10,    shell: 4000 },
  { id: 32, key: 'strong_fighter',    name: 'Heavy Fighter',     attack: 150,    shield: 25,    shell: 10000 },
  { id: 33, key: 'cruiser',           name: 'Cruiser',           attack: 400,    shield: 50,    shell: 27000 },
  { id: 34, key: 'battle_ship',       name: 'Battleship',        attack: 1000,   shield: 200,   shell: 60000 },
  { id: 36, key: 'colony_ship',       name: 'Colony Ship',       attack: 50,     shield: 100,   shell: 30000 },
  { id: 37, key: 'recycler',          name: 'Recycler',          attack: 1,      shield: 10,    shell: 16000 },
  { id: 38, key: 'espionage_sensor',  name: 'Espionage Probe',   attack: 0,      shield: 0,     shell: 1000 },
  { id: 39, key: 'solar_satellite',   name: 'Solar Satellite',   attack: 1,      shield: 1,     shell: 2000 },
  { id: 40, key: 'bomber',            name: 'Bomber',            attack: 1000,   shield: 500,   shell: 75000 },
  { id: 42, key: 'death_star',        name: 'Deathstar',         attack: 200000, shield: 50000, shell: 9000000 },
];

export const DEFENSE: CombatEntry[] = [
  { id: 43, key: 'rocket_launcher', name: 'Rocket Launcher',    attack: 80,   shield: 20,    shell: 2000 },
  { id: 44, key: 'light_laser',     name: 'Light Laser',        attack: 100,  shield: 25,    shell: 2000 },
  { id: 45, key: 'strong_laser',    name: 'Heavy Laser',        attack: 250,  shield: 100,   shell: 8000 },
  { id: 47, key: 'gauss_gun',       name: 'Gauss Cannon',       attack: 1100, shield: 200,   shell: 35000 },
  { id: 48, key: 'plasma_gun',      name: 'Plasma Turret',      attack: 3000, shield: 300,   shell: 100000 },
  { id: 49, key: 'small_shield',    name: 'Small Shield Dome',  attack: 1,    shield: 2000,  shell: 20000 },
  { id: 50, key: 'large_shield',    name: 'Large Shield Dome',  attack: 1,    shield: 10000, shell: 100000 },
];

// Артефакты — только те, что реально реализованы в M5.0.1 (факторы).
// Остальные 300-365 добавятся в M5.1 вместе с one_shot/battle_bonus.
export const ARTEFACTS: UnitEntry[] = [
  { id: 300, key: 'merchants_mark', name: "Merchant's Mark" },
  { id: 301, key: 'catalyst', name: 'Catalyst' },
  { id: 302, key: 'power_generator', name: 'Power Generator' },
  { id: 303, key: 'atomic_densifier', name: 'Atomic Densifier' },
  { id: 305, key: 'supercomputer', name: 'Supercomputer' },
  { id: 315, key: 'robot_control_system', name: 'Robot Control System' },
];

export function nameOf(id: number): string {
  for (const c of [BUILDINGS, RESEARCH, SHIPS, DEFENSE, ARTEFACTS]) {
    const u = c.find((x) => x.id === id);
    if (u) return u.name;
  }
  return `#${id}`;
}
