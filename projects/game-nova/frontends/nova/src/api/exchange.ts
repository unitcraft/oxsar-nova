// API-клиент биржи артефактов (план 76, использует backend плана 68).
// Все мутации передают Idempotency-Key (R9). TanStack Query-keys
// строятся в features/exchange/keys.ts, чтобы не размазывать
// серилизацию фильтров по компонентам.

import { api, type MutationOpts } from './client';

export type LotStatus = 'active' | 'sold' | 'cancelled' | 'expired';

export interface ExchangeLot {
  id: string;
  seller_user_id: string;
  seller_username?: string | null;
  artifact_unit_id: number;
  quantity: number;
  // price_oxsarit — за весь лот (целые оксариты).
  price_oxsarit: number;
  // unit_price_oxsarit — backend возвращает price/quantity, удобно для
  // сортировки списка по «выгодности» в UI.
  unit_price_oxsarit?: number;
  status: LotStatus;
  created_at: string;
  expires_at: string;
  buyer_user_id?: string | null;
  sold_at?: string | null;
}

export interface ListLotsResponse {
  lots: ExchangeLot[];
  next_cursor: string | null;
}

export interface LotDetailResponse {
  lot: ExchangeLot;
  items: Array<{ artefact_id: string }>;
}

export interface CreateLotRequest {
  artifact_unit_id: number;
  quantity: number;
  price_oxsarit: number;
  expires_in_hours: number;
}

export interface ExchangeStatsItem {
  artifact_unit_id: number;
  active_lots: number;
  // avg_unit_price — rolling-30d AVG(price/quantity) среди bought-лотов.
  // null когда продаж не было — UI должен показывать «—».
  avg_unit_price: number | null;
  last_30d_volume: number;
}

export const exchangeApi = {
  listLots: (queryString: string) =>
    api.get<ListLotsResponse>(
      queryString ? `/api/exchange/lots?${queryString}` : '/api/exchange/lots',
    ),

  getLot: (id: string) =>
    api.get<LotDetailResponse>(`/api/exchange/lots/${id}`),

  createLot: (body: CreateLotRequest, opts: MutationOpts) =>
    api.post<{ lot: ExchangeLot }>('/api/exchange/lots', body, opts),

  buyLot: (id: string, opts: MutationOpts) =>
    api.post<{ lot: ExchangeLot }>(`/api/exchange/lots/${id}/buy`, undefined, opts),

  cancelLot: (id: string) =>
    api.delete<void>(`/api/exchange/lots/${id}`),

  stats: () =>
    api.get<{ items: ExchangeStatsItem[] }>('/api/exchange/stats'),
};
