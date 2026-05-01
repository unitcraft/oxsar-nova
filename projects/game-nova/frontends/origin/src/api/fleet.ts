// API-модуль fleet origin-фронта (план 72 Ф.2 Spring 1).
//
// В nova-API «отправка миссии» (атака/шпионаж/транспорт/экспедиция)
// унифицирована в /api/fleet POST с полем mission. Отдельного
// /api/missions нет — соответствие legacy-PHP `?go=Missions` обеспечивает
// именно этот endpoint.
//
// Endpoints (openapi.yaml):
//   GET  /api/fleet                  — активные флоты + slots_used/max
//   POST /api/fleet                  — отправить флот (FleetDispatch)
//   POST /api/fleet/{id}/recall      — отозвать флот
//   POST /api/stargate               — прыжок флота между лунами

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type { Fleet, FleetDispatchInput, FleetList } from './types';

export function fetchFleet(): Promise<FleetList> {
  return api.get<FleetList>('/api/fleet');
}

export function dispatchFleet(input: FleetDispatchInput): Promise<Fleet> {
  return api.post<Fleet>('/api/fleet', input, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function recallFleet(id: string): Promise<Fleet> {
  return api.post<Fleet>(
    `/api/fleet/${id}/recall`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}

// План 72.1.47: Stargate Jump (legacy `Mission.class.php::starGateJump`).
// Прыжок между лунами с jump_gate >= 1; cooldown 3600 × 0.7^(level-1) сек.
export interface StargateJumpInput {
  src_planet_id: string;
  dst_planet_id: string;
  ships: Record<string, number>;
}

export interface StargateJumpResult {
  jumped_at: string;
  next_jump_at: string;
  cooldown_sec: number;
  ships: Record<string, number>;
}

export function stargateJump(input: StargateJumpInput): Promise<StargateJumpResult> {
  return api.post<StargateJumpResult>('/api/stargate', input, {
    idempotencyKey: newIdempotencyKey(),
  });
}
