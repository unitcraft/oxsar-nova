// API-модуль galaxy origin-фронта (план 72 Ф.2 Spring 1, расширен 72.1.24).
//
// Endpoint (openapi.yaml):
//   GET /api/galaxy/{g}/{s}?from_planet_id=<uuid>
//     — 15 ячеек системы, planet/moon/debris.
//
// План 72.1.24: legacy `Galaxy::subtractHydrogen` списывает 10H при
// просмотре системы, отличной от текущей планеты. Если fromPlanetId
// задан и target-system != src.system → backend списывает 10H.

import { api } from './client';
import type { SystemView } from './types';

export function fetchSystem(
  galaxy: number,
  system: number,
  fromPlanetId?: string,
): Promise<SystemView> {
  const qs = fromPlanetId ? `?from_planet_id=${encodeURIComponent(fromPlanetId)}` : '';
  return api.get<SystemView>(`/api/galaxy/${galaxy}/${system}${qs}`);
}
