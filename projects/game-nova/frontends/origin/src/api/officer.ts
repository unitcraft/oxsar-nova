// API-модуль officers origin-фронта (план 72 Ф.5 Spring 4 ч.2 — S-040).
//
// Endpoints (openapi.yaml + backend internal/officer/):
//   GET  /api/officers                       → { officers: Officer[] }
//   POST /api/officers/{key}/activate        → Officer  body: { auto_renew? }
//
// Officer DTO от backend содержит полный набор полей:
// title/description/duration_days/cost_credit/effect/activated_at/expires_at.
// R9 Idempotency-Key — на activate (мутация, списывает credit).

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type {
  Officer,
  OfficerActivateRequest,
  OfficersList,
} from './types';

export function fetchOfficers(): Promise<OfficersList> {
  return api.get<OfficersList>('/api/officers');
}

export function activateOfficer(
  key: string,
  body: OfficerActivateRequest = {},
): Promise<Officer> {
  return api.post<Officer>(`/api/officers/${encodeURIComponent(key)}/activate`, body, {
    idempotencyKey: newIdempotencyKey(),
  });
}
