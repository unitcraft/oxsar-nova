// API-модуль research origin-фронта (план 72 Ф.2 Spring 1).
//
// Endpoints (openapi.yaml):
//   GET  /api/research                       — очередь + уровни (по всем
//                                              планетам, агрегировано)
//   POST /api/planets/{id}/research          — поставить исследование

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type { ResearchOverview, QueueItem } from './types';

export function fetchResearch(): Promise<ResearchOverview> {
  return api.get<ResearchOverview>('/api/research');
}

export function startResearch(planetId: string, unitId: number): Promise<QueueItem> {
  return api.post<QueueItem>(
    `/api/planets/${planetId}/research`,
    { unit_id: unitId },
    { idempotencyKey: newIdempotencyKey() },
  );
}
