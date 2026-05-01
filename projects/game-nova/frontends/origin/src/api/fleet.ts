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

// План 72.1.47: load/unload ресурсов для HOLDING-флотов (legacy
// `Mission.class.php::loadResourcesToFleet/unloadResourcesFromFleet`).
export interface LoadUnloadInput {
  current_planet_id: string;
  metal: number;
  silicon: number;
  hydrogen: number;
}

export function loadFleet(fleetId: string, input: LoadUnloadInput): Promise<void> {
  return api.post<void>(`/api/fleet/${fleetId}/load`, input, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function unloadFleet(fleetId: string, input: LoadUnloadInput): Promise<void> {
  return api.post<void>(`/api/fleet/${fleetId}/unload`, input, {
    idempotencyKey: newIdempotencyKey(),
  });
}

// План 72.1.48: formation — конверсия single-атаки в ACS-группу с
// именем, invite по username с проверкой Relation.
export interface ACSGroup {
  id: string;
  name: string;
  leader_user_id: string;
  leader_fleet_id: string;
  created_at: string;
}

export interface ACSInvitation {
  acs_group_id: string;
  group_name: string;
  leader_name: string;
  invited_by: string;
  invited_at: string;
  accepted_at?: string;
}

export function promoteFleetToACS(fleetId: string, name: string): Promise<ACSGroup> {
  return api.post<ACSGroup>(`/api/fleet/${fleetId}/promote-to-acs`, { name }, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function inviteToACS(groupId: string, username: string): Promise<void> {
  return api.post<void>(`/api/acs/${groupId}/invite`, { username }, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function listACSInvitations(): Promise<{ invitations: ACSInvitation[] }> {
  return api.get<{ invitations: ACSInvitation[] }>('/api/acs/invitations');
}

export function acceptACSInvitation(groupId: string): Promise<void> {
  return api.post<void>(`/api/acs/invitations/${groupId}/accept`, undefined, {
    idempotencyKey: newIdempotencyKey(),
  });
}
