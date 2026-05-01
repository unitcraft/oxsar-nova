// API-модуль /p2p-exchange (план 72.1 §20.12 P2P-биржа task).
//
// Endpoint: GET /api/exchange/broker-stats — статистика лотов
// текущего юзера за период (legacy `Exchange.class.php::showStatistics`).

import { api } from './client';

export interface BrokerStatsRow {
  lot_id: string;
  unit_id: number;
  quantity: number;
  price: number;
  status: 'sold' | 'cancelled' | 'expired';
  sold_at: string;
  profit: number;
}

export interface BrokerStatsSummary {
  total: number;
  sold: number;
  turnover: number;
  profit: number;
}

export interface BrokerStatsResponse {
  rows: BrokerStatsRow[];
  summary: BrokerStatsSummary;
  pages: number;
  page: number;
  fee: number;
  // План 72.1.45 §9: title из exchange_settings (легаси показывает в шапке).
  title?: string;
}

export interface BrokerStatsFilters {
  date_min?: string; // YYYY-MM-DD
  date_max?: string;
  sort_field?: 'date' | 'lot' | 'lot_price' | 'lot_amount' | 'lot_profit';
  sort_order?: 'asc' | 'desc';
  page?: number;
}

export function fetchBrokerStats(
  filters: BrokerStatsFilters = {},
): Promise<BrokerStatsResponse> {
  const qs = new URLSearchParams();
  if (filters.date_min) qs.set('date_min', filters.date_min);
  if (filters.date_max) qs.set('date_max', filters.date_max);
  if (filters.sort_field) qs.set('sort_field', filters.sort_field);
  if (filters.sort_order) qs.set('sort_order', filters.sort_order);
  if (filters.page != null) qs.set('page', String(filters.page));
  const url = `/api/exchange/broker-stats${qs.toString() ? `?${qs}` : ''}`;
  return api.get<BrokerStatsResponse>(url);
}
