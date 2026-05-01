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

// План 72.1.28: multi-resource Credit_ex (legacy `Market::Credit_ex`).
// Любой комбинацию M/Si/H одной транзакцией. Пользователь указывает
// сколько ресурсов хочет купить, backend считает суммарную стоимость.
export interface CreditExchangeMultiResult {
  direction: string;
  metal: number;
  silicon: number;
  hydrogen: number;
  credits: number;
}

export function exchangeCreditMulti(input: {
  planetId: string;
  metal: number;
  silicon: number;
  hydrogen: number;
}): Promise<CreditExchangeMultiResult> {
  return api.post<CreditExchangeMultiResult>(
    `/api/planets/${input.planetId}/market/credit-multi`,
    {
      metal: input.metal,
      silicon: input.silicon,
      hydrogen: input.hydrogen,
    },
    { idempotencyKey: newIdempotencyKey() },
  );
}

// Клиент-side preview total cost (legacy `Market::Credit_ex`).
// Курсы: 100 metal = 1 cr, 50 silicon = 1 cr, 25 hydrogen = 1 cr.
// Каждый по отдельности ceil, потом sum (не суммарный ceil).
export function multiCreditCost(
  metal: number,
  silicon: number,
  hydrogen: number,
): number {
  const m = metal > 0 ? Math.ceil(metal / 100) : 0;
  const s = silicon > 0 ? Math.ceil(silicon / 50) : 0;
  const h = hydrogen > 0 ? Math.ceil(hydrogen / 25) : 0;
  return m + s + h;
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
