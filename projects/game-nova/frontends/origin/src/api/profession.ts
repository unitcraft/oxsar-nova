// API-модуль professions origin-фронта (план 72 Ф.5 Spring 4 ч.2 — S-041).
//
// Endpoints:
//   GET  /api/professions       → { professions: Profession[] }
//   GET  /api/professions/me    → ProfessionInfo (текущая + cooldown)
//   POST /api/professions/me    → 204  body: { profession }
//
// Backend (internal/profession/service.go): смена 1000 кр, интервал
// 14 дней, ключ профессии валидируется по configs/professions.yml.
// R9 Idempotency-Key — на смену.

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type {
  ProfessionInfo,
  ProfessionsList,
} from './types';

export function fetchProfessions(): Promise<ProfessionsList> {
  return api.get<ProfessionsList>('/api/professions');
}

export function fetchProfessionMe(): Promise<ProfessionInfo> {
  return api.get<ProfessionInfo>('/api/professions/me');
}

export function changeProfession(profession: string): Promise<void> {
  return api.post<void>(
    '/api/professions/me',
    { profession },
    { idempotencyKey: newIdempotencyKey() },
  );
}
