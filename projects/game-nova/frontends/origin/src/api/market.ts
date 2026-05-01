// API-модули resource exchange + artefact market origin-фронта
// (план 72 Ф.3 Spring 2 ч.2).
//
// Endpoints (openapi.yaml):
//   GET  /api/market/rates                          → MarketRates
//   POST /api/planets/{id}/market/exchange          → ExchangeResult
//   GET  /api/artefact-market/offers                → { offers: [] }
//   GET  /api/artefact-market/credit                → { credit: number }
//   POST /api/artefact-market/offers/{id}/buy
//   DELETE /api/artefact-market/offers/{id}
//   POST /api/artefacts/{id}/sell                    → создать offer

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type {
  ArtMarketCredit,
  ArtMarketOffer,
  ExchangeResult,
  MarketRates,
  ResourceKind,
} from './types';

// ---- Resource exchange ----

export function fetchMarketRates(): Promise<MarketRates> {
  return api.get<MarketRates>('/api/market/rates');
}

export function exchangeResource(input: {
  planetId: string;
  from: ResourceKind;
  to: ResourceKind;
  amount: number;
}): Promise<ExchangeResult> {
  return api.post<ExchangeResult>(
    `/api/planets/${input.planetId}/market/exchange`,
    { from: input.from, to: input.to, amount: input.amount },
    { idempotencyKey: newIdempotencyKey() },
  );
}

// План 72.1.21: покупка ресурса за кредиты (legacy `Credit_ex`).
// amount — сколько кредитов потратить; ответ содержит фактический
// resource_delta (полученное количество ресурса).
export interface CreditExchangeResult {
  direction: string; // всегда "from_credit"
  resource: string;
  resource_delta: number;
  credit_delta: number; // отрицательное число — списание
}

export function exchangeCredit(input: {
  planetId: string;
  resource: ResourceKind;
  amount: number; // кредиты
}): Promise<CreditExchangeResult> {
  return api.post<CreditExchangeResult>(
    `/api/planets/${input.planetId}/market/credit`,
    {
      direction: 'from_credit',
      resource: input.resource,
      amount: input.amount,
    },
    { idempotencyKey: newIdempotencyKey() },
  );
}

// ---- Artefact market ----

export function fetchArtMarketOffers(): Promise<{
  offers: ArtMarketOffer[] | null;
}> {
  return api.get<{ offers: ArtMarketOffer[] | null }>(
    '/api/artefact-market/offers',
  );
}

export function fetchArtMarketCredit(): Promise<ArtMarketCredit> {
  return api.get<ArtMarketCredit>('/api/artefact-market/credit');
}

export function buyArtMarketOffer(offerID: string): Promise<void> {
  return api.post<void>(
    `/api/artefact-market/offers/${offerID}/buy`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function cancelArtMarketOffer(offerID: string): Promise<void> {
  return api.delete<void>(`/api/artefact-market/offers/${offerID}`, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function sellArtefact(
  artefactID: string,
  price: number,
): Promise<void> {
  return api.post<void>(
    `/api/artefacts/${artefactID}/sell`,
    { price },
    { idempotencyKey: newIdempotencyKey() },
  );
}
