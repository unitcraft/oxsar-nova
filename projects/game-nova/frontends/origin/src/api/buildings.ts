// API-модуль buildings origin-фронта (план 72 Ф.2 Spring 1).
//
// Endpoints (openapi.yaml):
//   POST   /api/planets/{id}/buildings            — поставить апгрейд
//   GET    /api/planets/{id}/buildings/queue      — текущая очередь
//   DELETE /api/planets/{id}/buildings/queue/{taskId} — отмена
//
// Список доступных зданий с уровнями (информационная таблица для
// рендера экрана) в openapi.yaml сейчас отсутствует — есть только
// очередь. На MVP экран рендерит список юнитов из локального справочника
// (см. configs/units там, где это уместно), либо из mock'а с TODO.

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type { QueueItem } from './types';

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
