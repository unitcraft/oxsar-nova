// API-модуль /api/monitor-planet (план 72.1.20).
//
// Endpoint:
//   GET /api/monitor-planet?id=<planet_id>
// Ответ — { target_planet, scanner, events, detected }.

import { api } from './client';
import type { PhalanxScan } from './types';

export interface MonitorPlanetInfo {
  planet_id: string;
  name: string;
  user_id: string;
  username?: string;
  galaxy: number;
  system: number;
  position: number;
}

export interface MonitorResult {
  target_planet: MonitorPlanetInfo;
  scanner: MonitorPlanetInfo;
  events: PhalanxScan[];
  detected: boolean;
}

export function fetchMonitorPlanet(planetId: string): Promise<MonitorResult> {
  return api.get<MonitorResult>(
    `/api/monitor-planet?id=${encodeURIComponent(planetId)}`,
  );
}
