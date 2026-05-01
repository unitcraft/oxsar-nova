// API-модуль empire origin-фронта (план 72.1.37).
//
// Endpoint: GET /api/empire — все планеты игрока с агрегатами
// (buildings/ships/defense per-planet) + общие исследования (research per-user).
//
// Legacy `Empire.class.php` показывает 5 вкладок:
// constructions/shipyard/defense/moon/research. Origin изначально
// показывал только верхнюю таблицу (план 72 Spring 1); 72.1.37
// добавляет вкладки с агрегатами.

import { api } from './client';

export interface EmpirePlanet {
  id: string;
  name: string;
  galaxy: number;
  system: number;
  position: number;
  is_moon: boolean;
  diameter: number;
  used_fields: number;
  temp_min: number;
  temp_max: number;
  metal: number;
  silicon: number;
  hydrogen: number;
  buildings: Record<string, number>;
  ships: Record<string, number>;
  defense: Record<string, number>;
}

export interface EmpireResponse {
  planets: EmpirePlanet[];
  research: Record<string, number>;
}

export function fetchEmpire(): Promise<EmpireResponse> {
  return api.get<EmpireResponse>('/api/empire');
}
