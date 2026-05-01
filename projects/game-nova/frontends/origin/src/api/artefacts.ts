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

export interface ArtefactsResponse {
  artefacts: Artefact[] | null;
  // План 72.1.45: legacy `Artefacts.class.php` показывает в шапке
  // tech_level (research UNIT_ARTEFACTS_TECH=111), storage_slots (=tech_level),
  // used_slots (count active artefacts).
  tech_level: number;
  storage_slots: number;
  used_slots: number;
}

export function fetchArtefacts(): Promise<ArtefactsResponse> {
  return api.get<ArtefactsResponse>('/api/artefacts');
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

// План 72.1.45 §2: история приобретений артефакта (legacy ArtefactInfo).
export interface ArtefactHistoryEntry {
  acquired_at: string;
  user_id: string;
  username: string;
  battle_report_id?: string;
  opponent_name?: string;
  source: 'battle' | 'expedition' | 'quest' | 'admin' | 'market';
}

export function fetchArtefactHistory(
  unitId: number,
): Promise<{ entries: ArtefactHistoryEntry[] }> {
  return api.get<{ entries: ArtefactHistoryEntry[] }>(
    `/api/artefacts/info/${unitId}/history`,
  );
}
