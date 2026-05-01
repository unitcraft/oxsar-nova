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
