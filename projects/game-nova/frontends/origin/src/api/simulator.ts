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
}

export interface SimInput {
  seed?: number;
  rounds?: number;
  num_sim?: number;
  attackers: SimSide[];
  defenders: SimSide[];
  is_moon?: boolean;
}

export interface SimRoundTrace {
  index: number;
  attackers_alive: number;
  defenders_alive: number;
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
}

export function runSimulation(input: SimInput): Promise<SimReport> {
  return api.post<SimReport>('/api/simulator/run', input);
}
