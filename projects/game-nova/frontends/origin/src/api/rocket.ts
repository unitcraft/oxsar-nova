// API-модуль rockets origin-фронта (план 72.1.35).
//
// Endpoints:
//   GET  /api/planets/{id}/rockets         — текущий stock IPM на планете.
//   POST /api/planets/{id}/rockets/launch  — запуск ракетной атаки.
//
// Legacy `RocketAttack.class.php::sendRockets`: проверка target в
// missile range, NS::isFirstRun дедуп, под-атакой, бан/umode/observer
// у обоих сторон, isPlanetUnderAttack у атакующего, бэшинг,
// protection_time. Добавляется event KindRocketAttack с fire_at =
// now + getRocketFlightDuration(diff).

import { api } from './client';
import { newIdempotencyKey } from './idempotency';

export interface RocketStock {
  count: number;
}

export function fetchRocketStock(planetId: string): Promise<RocketStock> {
  return api.get<RocketStock>(`/api/planets/${planetId}/rockets`);
}

export interface LaunchRequest {
  dst: {
    galaxy: number;
    system: number;
    position: number;
    is_moon?: boolean;
  };
  count: number;
  target_unit_id?: number; // 0 = без приоритета (= "all" в legacy).
}

export interface LaunchResult {
  fleet_id?: string;
  arrival_at?: string;
  count: number;
}

export function launchRocket(
  planetId: string,
  req: LaunchRequest,
): Promise<LaunchResult> {
  return api.post<LaunchResult>(
    `/api/planets/${planetId}/rockets/launch`,
    req,
    { idempotencyKey: newIdempotencyKey() },
  );
}
