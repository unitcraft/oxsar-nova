// API-модуль артефактов origin-фронта (план 72 Ф.4 Spring 3, S-013).
//
// Endpoints (openapi.yaml):
//   GET    /api/artefacts                       → { artefacts: Artefact[] }
//   POST   /api/artefacts/{id}/activate         → Artefact
//   POST   /api/artefacts/{id}/deactivate       → 204
//
// Каталог-описание (S-014 ArtefactInfo) — отдельный endpoint
// /api/artefacts/catalog/{type}, см. api/catalog.ts.

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type { Artefact } from './types';

export function fetchArtefacts(): Promise<{ artefacts: Artefact[] | null }> {
  return api.get<{ artefacts: Artefact[] | null }>('/api/artefacts');
}

export function activateArtefact(id: string): Promise<Artefact> {
  return api.post<Artefact>(
    `/api/artefacts/${id}/activate`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function deactivateArtefact(id: string): Promise<void> {
  return api.post<void>(
    `/api/artefacts/${id}/deactivate`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}
