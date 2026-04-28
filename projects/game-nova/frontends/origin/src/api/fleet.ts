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
