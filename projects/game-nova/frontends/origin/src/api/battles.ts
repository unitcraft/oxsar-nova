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

// Параметры фильтрации /api/users/me/battles (план 72.1.10).
//
// Дефолты сервера: show_drawn=true, show_aliens=false,
// show_no_destroyed=true, new_moon=false, moon_battle=false,
// sort_field=date, sort_order=desc.
export interface BattleListFilters {
  limit?: number;
  cursor?: string;
  date_min?: string; // RFC3339
  date_max?: string;
  user_filter?: string; // uuid оппонента
  alliance_filter?: string;
  show_drawn?: boolean;
  show_aliens?: boolean;
  show_no_destroyed?: boolean;
  new_moon?: boolean;
  moon_battle?: boolean;
  sort_field?: 'date' | 'rounds' | 'debris' | 'loot';
  sort_order?: 'asc' | 'desc';
}

export function fetchMyBattles(params?: BattleListFilters): Promise<BattleListResult> {
  const qs = new URLSearchParams();
  if (params) {
    if (params.limit != null) qs.set('limit', String(params.limit));
    if (params.cursor) qs.set('cursor', params.cursor);
    if (params.date_min) qs.set('date_min', params.date_min);
    if (params.date_max) qs.set('date_max', params.date_max);
    if (params.user_filter) qs.set('user_filter', params.user_filter);
    if (params.alliance_filter) qs.set('alliance_filter', params.alliance_filter);
    if (params.show_drawn != null) qs.set('show_drawn', params.show_drawn ? '1' : '0');
    if (params.show_aliens != null) qs.set('show_aliens', params.show_aliens ? '1' : '0');
    if (params.show_no_destroyed != null) qs.set('show_no_destroyed', params.show_no_destroyed ? '1' : '0');
    if (params.new_moon != null) qs.set('new_moon', params.new_moon ? '1' : '0');
    if (params.moon_battle != null) qs.set('moon_battle', params.moon_battle ? '1' : '0');
    if (params.sort_field) qs.set('sort_field', params.sort_field);
    if (params.sort_order) qs.set('sort_order', params.sort_order);
  }
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
