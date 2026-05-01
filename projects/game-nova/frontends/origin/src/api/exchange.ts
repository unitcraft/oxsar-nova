// API exchange (биржа артефактов / Stock).
//
// Endpoints (openapi.yaml):
//   GET    /api/exchange/lots                — список лотов (с фильтрами).
//   GET    /api/exchange/lots/{id}           — детали + items.
//   POST   /api/exchange/lots                — создать (план 72.1.8 ч.B).
//   POST   /api/exchange/lots/{id}/buy       — купить (план 72.1.8 ч.A).
//   DELETE /api/exchange/lots/{id}           — отозвать свой (план 72.1.8 ч.A).
//   GET    /api/exchange/stats               — статистика.

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type { ExchangeLot, ExchangeLotsResult } from './types';

export function fetchExchangeLots(params?: {
  artifact_unit_id?: number;
  status?: string;
  limit?: number;
  cursor?: string;
}): Promise<ExchangeLotsResult> {
  const qs = new URLSearchParams();
  if (params?.artifact_unit_id != null)
    qs.set('artifact_unit_id', String(params.artifact_unit_id));
  if (params?.status) qs.set('status', params.status);
  if (params?.limit != null) qs.set('limit', String(params.limit));
  if (params?.cursor) qs.set('cursor', params.cursor);
  const query = qs.toString();
  return api.get<ExchangeLotsResult>(`/api/exchange/lots${query ? `?${query}` : ''}`);
}

// План 72.1.8 ч.A: операции покупки и отзыва лотов.
//
// Backend (internal/exchange/handler.go):
//   Buy: 404 not_found / 409 lot_not_active / 403 cannot_buy_own_lot
//        / 402 insufficient_oxsarits / 409 buyer_has_no_planet.
//   Cancel: 404 not_found / 403 not_a_seller / 409 lot_not_active.
export function buyLot(lotID: string): Promise<{ lot: ExchangeLot }> {
  return api.post<{ lot: ExchangeLot }>(
    `/api/exchange/lots/${encodeURIComponent(lotID)}/buy`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function cancelLot(lotID: string): Promise<void> {
  return api.delete<void>(`/api/exchange/lots/${encodeURIComponent(lotID)}`);
}

// План 72.1.8 ч.B: создание лота (выставить артефакт на продажу).
// Backend (handler.go::Create): требует Idempotency-Key. Возможные
// ошибки: invalid_quantity / invalid_price / invalid_expiry,
// insufficient_artefacts (нет столько артефактов в инвентаре),
// price_cap_exceeded (антифрод), permit_required (нет
// merchant-permit), max_active_lots, max_quantity.
export interface CreateLotPayload {
  artifact_unit_id: number;
  quantity: number;
  price_oxsarit: number;
  expires_in_hours: number;
}

export function createLot(payload: CreateLotPayload): Promise<{ lot: ExchangeLot }> {
  return api.post<{ lot: ExchangeLot }>('/api/exchange/lots', payload, {
    idempotencyKey: newIdempotencyKey(),
  });
}
