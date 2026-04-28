// API-модуль planets origin-фронта (план 72 Ф.2 Spring 1).
//
// Endpoints — существующие в openapi.yaml:
//   GET /api/planets             — список планет игрока
//   GET /api/planets/{id}        — детали одной планеты
//
// Запрос empire-overview (агрегированный «первый экран») в openapi.yaml
// отсутствует — для S-001 Main мы собираем обзор клиентски из /planets
// + /fleet + /messages/unread-count. Запись об отсутствующем endpoint —
// в docs/simplifications.md.

import { api } from './client';
import type { Planet } from './types';

export async function fetchPlanets(): Promise<Planet[]> {
  const res = await api.get<{ planets: Planet[] }>('/api/planets');
  return res.planets ?? [];
}

export function fetchPlanet(id: string): Promise<Planet> {
  return api.get<Planet>(`/api/planets/${id}`);
}
