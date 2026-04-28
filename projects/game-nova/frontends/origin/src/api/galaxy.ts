// API-модуль galaxy origin-фронта (план 72 Ф.2 Spring 1).
//
// Endpoint (openapi.yaml):
//   GET /api/galaxy/{g}/{s}  — 15 ячеек системы, planet/moon/debris.

import { api } from './client';
import type { SystemView } from './types';

export function fetchSystem(galaxy: number, system: number): Promise<SystemView> {
  return api.get<SystemView>(`/api/galaxy/${galaxy}/${system}`);
}
