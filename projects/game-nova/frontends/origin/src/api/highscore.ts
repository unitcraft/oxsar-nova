// API-модуль рейтингов / статистики origin-фронта
// (план 72 Ф.4 Spring 3, S-023, S-024, S-032).
//
// Endpoints (openapi.yaml):
//   GET /api/highscore        → { entries: HighscoreEntry[] }
//   GET /api/highscore/me     → HighscoreEntry
//   GET /api/stats            → { online_now, online_24h }
//
// `/api/records` (отдельный от highscore) и `/api/highscore?type=...`
// в openapi.yaml на 2026-04-28 отсутствуют — Records-экран использует
// тот же /api/highscore, см. simplifications.md P72.S3.D.

import { api } from './client';
import type { HighscoreEntry, PublicStats } from './types';

export function fetchHighscore(): Promise<{
  entries: HighscoreEntry[] | null;
}> {
  return api.get<{ entries: HighscoreEntry[] | null }>('/api/highscore');
}

export function fetchHighscoreMe(): Promise<HighscoreEntry> {
  return api.get<HighscoreEntry>('/api/highscore/me');
}

export function fetchPublicStats(): Promise<PublicStats> {
  return api.get<PublicStats>('/api/stats');
}
