// API-модуль simulator origin-фронта (план 72.1 ч.20.7).
//
// Endpoint:
//   POST /api/simulator/run — чистый бой (battle.Calculate)
//
// Семантика: симулятор не знает о юзере/планетах — только статы юнитов
// и tech-уровни. Все cost/quantity передаются явно.

import { api } from './client';

export interface SimUnit {
  unit_id: number;
  mode?: number;
  quantity: number;
  damaged?: number;
  shell_percent?: number;
  front?: number;
  attack: number;
  shield?: number;
  shell: number;
  name?: string;
  cost?: { metal: number; silicon: number; hydrogen: number };
}

export interface SimTech {
  gun?: number;
  shield?: number;
  shell?: number;
  laser?: number;
  ion?: number;
  plasma?: number;
  ballistics?: number;
  masking?: number;
}

export interface SimSide {
  user_id: string;
  username?: string;
  is_aliens?: boolean;
  tech?: SimTech;
  units: SimUnit[];
  primary_target?: number;
  // План 72.1.34 ч.C: legacy 6 боевых артефактов с
  // effect_type=ARTEFACT_EFFECT_TYPE_BATTLE (id 316-318, 359-361).
  // Backend агрегирует через ComputeBattleModifier и применяет
  // multiplier к attack/shield/shell.
  battle_artefact_ids?: number[];
}

export interface SimInput {
  seed?: number;
  rounds?: number;
  num_sim?: number;
  attackers: SimSide[];
  defenders: SimSide[];
  is_moon?: boolean;
  // План 72.1.34 ч.B: цель-здание (legacy target_buildingid).
  target_building_id?: number;
  target_building_level?: number;
}

// Расширенная статистика раунда (план 72.1 ч.20.11.4) — порт
// oxsar2-java/Assault.java rendering.
export interface SimRoundTrace {
  index: number;
  attackers_alive: number;
  defenders_alive: number;
  attacker_side: SimRoundSide;
  defender_side: SimRoundSide;
}

export interface SimRoundSide {
  username?: string;
  galaxy?: number;
  system?: number;
  position?: number;
  is_moon?: boolean;
  gun_power_pct: number;
  shield_power_pct: number;
  armoring_pct: number;
  ballistics_lvl: number;
  masking_lvl: number;
  shots: number;
  power: number;
  shield_absorbed: number;
  shell_destroyed: number;
  units_destroyed: number;
  units: SimRoundUnit[];
}

export interface SimRoundUnit {
  unit_id: number;
  name?: string;
  start_turn_quantity: number;
  start_turn_quantity_diff: number;
  start_turn_damaged: number;
  damaged_shell_percent: number;
  attack: number;
  shield: number;
  shell: number;
  front: number;
  ballistics_level?: number;
  masking_level?: number;
  start_battle_quantity: number;
  alive_percent: number;
}

export interface SimUnitResult {
  unit_id: number;
  quantity_start: number;
  quantity_end: number;
  damaged_end?: number;
  shell_percent_end?: number;
}

export interface SimSideResult {
  user_id: string;
  username?: string;
  lost_metal: number;
  lost_silicon: number;
  lost_hydrogen: number;
  // План 72.1.34: legacy simulator.tpl показывает потери в очках
  // (Σ qty_lost × (cost) / 1000 × 2) — backend уже считает.
  lost_points?: number;
  lost_units?: number;
  units: SimUnitResult[];
}

export interface SimReport {
  seed: number;
  rounds: number;
  winner: string;
  rounds_trace?: SimRoundTrace[];
  attackers?: SimSideResult[];
  defenders?: SimSideResult[];
  debris_metal?: number;
  debris_silicon?: number;
  moon_chance?: number;
  moon_created?: boolean;
  attacker_exp?: number;
  defender_exp?: number;
  haul_metal?: number;
  haul_silicon?: number;
  haul_hydrogen?: number;
  // План 72.1.34 ч.B: target-building destroy chance.
  building_destroy_chance?: number;
  target_destroyed?: boolean;
}

// Сводка по num_sim итераций (план 72.1 ч.20.11.7) — pixel-perfect клон
// блока «Результаты» legacy simulator.tpl.
export interface SimStats {
  num_sim: number;
  attacker_win_pct: number;
  defender_win_pct: number;
  draw_pct: number;
  avg_rounds: number;
  avg_moon_chance: number;
  attacker_lost_metal: number;
  attacker_lost_silicon: number;
  attacker_lost_hydrogen: number;
  defender_lost_metal: number;
  defender_lost_silicon: number;
  defender_lost_hydrogen: number;
  // План 72.1.34: legacy simulator.tpl показывает потери в очках/юнитах.
  attacker_lost_points?: number;
  defender_lost_points?: number;
  attacker_lost_units?: number;
  defender_lost_units?: number;
  debris_metal: number;
  debris_silicon: number;
  attacker_exp: number;
  defender_exp: number;
  gen_time_all: number;
  gen_time: number;
}

export interface SimRunResponse {
  id: string;
  stats: SimStats;
  report: SimReport;
}

export function runSimulation(input: SimInput): Promise<SimRunResponse> {
  return api.post<SimRunResponse>('/api/simulator/run', input);
}
