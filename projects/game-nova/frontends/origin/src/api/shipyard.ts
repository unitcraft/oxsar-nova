// API-модуль shipyard origin-фронта (план 72 Ф.2 Spring 1).
//
// Endpoints (openapi.yaml):
//   GET  /api/planets/{id}/shipyard/queue      — очередь верфи
//   GET  /api/planets/{id}/shipyard/inventory  — флот + оборона на планете
//   POST /api/planets/{id}/shipyard            — поставить производство

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type { ShipyardQueueItem, ShipyardInventory } from './types';

export async function fetchShipyardQueue(
  planetId: string,
): Promise<ShipyardQueueItem[]> {
  const res = await api.get<{ queue: ShipyardQueueItem[] }>(
    `/api/planets/${planetId}/shipyard/queue`,
  );
  return res.queue ?? [];
}

export function fetchShipyardInventory(
  planetId: string,
): Promise<ShipyardInventory> {
  return api.get<ShipyardInventory>(`/api/planets/${planetId}/shipyard/inventory`);
}

export function buildShipyard(
  planetId: string,
  unitId: number,
  count: number,
): Promise<ShipyardQueueItem> {
  return api.post<ShipyardQueueItem>(
    `/api/planets/${planetId}/shipyard`,
    { unit_id: unitId, count },
    { idempotencyKey: newIdempotencyKey() },
  );
}
