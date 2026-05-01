// API-модуль battles origin-фронта (план 72.1 ч.20.8 — battle viewer).
//
// Endpoints:
//   GET /api/users/me/battles         — список моих боёв (cursor)
//   GET /api/battle-reports/{id}      — детальный отчёт

import { api } from './client';
import type { SimReport } from './simulator';

export interface BattleListItem {
  id: string;
  attacker_user_id?: string | null;
  defender_user_id?: string | null;
  winner: 'attackers' | 'defenders' | 'draw';
  rounds: number;
  debris_metal: number;
  debris_silicon: number;
  loot_metal: number;
  loot_silicon: number;
  loot_hydrogen: number;
  is_attacker: boolean;
  at: string;
}

export interface BattleListResult {
  battles: BattleListItem[];
  next_cursor?: string | null;
}

export function fetchMyBattles(params?: {
  limit?: number;
  cursor?: string;
}): Promise<BattleListResult> {
  const qs = new URLSearchParams();
  if (params?.limit) qs.set('limit', String(params.limit));
  if (params?.cursor) qs.set('cursor', params.cursor);
  const url = `/api/users/me/battles${qs.toString() ? `?${qs}` : ''}`;
  return api.get<BattleListResult>(url);
}

export interface BattleReportFull {
  report: SimReport;
  started_at: string; // RFC3339, время записи отчёта (Java assaultTime).
}

export function fetchBattleReport(id: string): Promise<BattleReportFull> {
  return api.get<BattleReportFull>(`/api/battle-reports/${id}`);
}
