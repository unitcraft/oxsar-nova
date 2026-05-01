// API-модуль buildings origin-фронта (план 72.1 ч.20).
//
// Endpoints (openapi.yaml):
//   GET    /api/planets/{id}/buildings            — levels + cost + time + unmet
//   POST   /api/planets/{id}/buildings            — поставить апгрейд
//   GET    /api/planets/{id}/buildings/queue      — текущая очередь
//   DELETE /api/planets/{id}/buildings/queue/{taskId} — отмена

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type { QueueItem } from './types';

export interface BuildingCost {
  metal: number;
  silicon: number;
  hydrogen: number;
}

export interface RequirementUnmet {
  unit_id: number;
  required_level: number;
  actual_level: number;
}

export interface BuildingsOverview {
  levels: Record<string, number>;
  build_seconds: Record<string, number>;
  build_costs: Record<string, BuildingCost>;
  requirements_unmet: Record<string, RequirementUnmet[]>;
}

export function fetchBuildingsOverview(planetId: string): Promise<BuildingsOverview> {
  return api.get<BuildingsOverview>(`/api/planets/${planetId}/buildings`);
}

export async function fetchBuildingQueue(planetId: string): Promise<QueueItem[]> {
  const res = await api.get<{ queue: QueueItem[] }>(
    `/api/planets/${planetId}/buildings/queue`,
  );
  return res.queue ?? [];
}

export function enqueueBuilding(
  planetId: string,
  unitId: number,
): Promise<QueueItem> {
  return api.post<QueueItem>(
    `/api/planets/${planetId}/buildings`,
    { unit_id: unitId },
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function cancelBuildingTask(planetId: string, taskId: string): Promise<void> {
  return api.delete<void>(
    `/api/planets/${planetId}/buildings/queue/${taskId}`,
    { idempotencyKey: newIdempotencyKey() },
  );
}

// План 72.1.33: legacy `BuildingInfo::DEMOLISH_NOW` — снос здания
// на 1 уровень. Cost = (1 / spec.demolish) × cost_at_current_level.
// Длительность = build duration × 0.5.
export function demolishBuilding(
  planetId: string,
  unitId: number,
): Promise<QueueItem> {
  return api.post<QueueItem>(
    `/api/planets/${planetId}/buildings/${unitId}/demolish`,
    {},
    { idempotencyKey: newIdempotencyKey() },
  );
}

// План 72.1.33 ч.2: legacy `BuildingInfo::packCurrentConstruction` —
// упаковка здания в packed-building артефакт (требует held packing-
// building артефакта на этой планете, иначе backend вернёт 409).
export interface PackedArtefact {
  id: string;
  user_id: string;
  unit_id: number;
  state: string;
}

export function packBuilding(
  planetId: string,
  unitId: number,
): Promise<PackedArtefact> {
  return api.post<PackedArtefact>(
    `/api/planets/${planetId}/buildings/${unitId}/pack`,
    {},
    { idempotencyKey: newIdempotencyKey() },
  );
}

// План 72.1.33 ч.2: legacy `BuildingInfo::packCurrentResearch` —
// упаковка исследования. planetId — текущая планета (контекст).
export function packResearch(
  planetId: string,
  unitId: number,
): Promise<PackedArtefact> {
  return api.post<PackedArtefact>(
    `/api/planets/${planetId}/research/${unitId}/pack`,
    {},
    { idempotencyKey: newIdempotencyKey() },
  );
}

// План 72.1.44: VIP-instant старт стройки за credits (legacy
// `EventHandler::startConstructionEventVIP` для UNIT_TYPE_CONSTRUCTION).
export function startBuildingVIP(
  planetId: string,
  taskId: string,
): Promise<QueueItem> {
  return api.post<QueueItem>(
    `/api/planets/${planetId}/buildings/queue/${taskId}/vip`,
    {},
    { idempotencyKey: newIdempotencyKey() },
  );
}
