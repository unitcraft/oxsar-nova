// API-модуль рейтингов / статистики origin-фронта
// (план 72 Ф.4 Spring 3, S-023, S-024, S-032; расширен 72.1.12 + 72.1.29).
//
// Endpoints (openapi.yaml + backend):
//   GET /api/highscore?type=&mode=&avg=&page=  → { entries, total_count, page, per_page }
//   GET /api/highscore/me?type=...             → HighscoreEntry
//   GET /api/highscore/alliances               → { alliances: HighscoreAlliance[] }
//   GET /api/highscore/vacation                → { players: HighscoreVacation[] }
//   GET /api/stats                             → { online_now, online_24h }

import { api } from './client';
import type {
  HighscoreEntry,
  HighscoreAlliance,
  HighscoreVacation,
  PublicStats,
} from './types';

// План 72.1.12 + 72.1.29: 12 score-types (legacy `Ranking::validTypes`).
//   total → points (b+r+u)
//   b/r/u/a/e/dm/max → users.<X>_points
//   b_count → COUNT buildings, r_count → COUNT research,
//   u_count → SUM ships+defense, battles → users.battles
export type ScoreType =
  | 'total'
  | 'b'
  | 'r'
  | 'u'
  | 'a'
  | 'e'
  | 'dm'
  | 'max'
  | 'b_count'
  | 'r_count'
  | 'u_count'
  | 'battles';

// План 72.1.29: legacy 4 mode'а. alliance — отдельный endpoint.
export type RankingMode =
  | 'player'
  | 'player_observer'
  | 'player_old_vacation';

export interface HighscoreOptions {
  type?: ScoreType;
  mode?: RankingMode;
  avg?: boolean;
  page?: number;
}

function buildQuery(opts: HighscoreOptions): string {
  const qs = new URLSearchParams();
  if (opts.type && opts.type !== 'total') qs.set('type', opts.type);
  if (opts.mode && opts.mode !== 'player') qs.set('mode', opts.mode);
  if (opts.avg) qs.set('avg', 'true');
  if (opts.page && opts.page > 1) qs.set('page', String(opts.page));
  const q = qs.toString();
  return q ? `?${q}` : '';
}

export interface HighscoreResult {
  entries: HighscoreEntry[] | null;
  total_count: number;
  page: number;
  per_page: number;
}

export function fetchHighscore(opts: HighscoreOptions = {}): Promise<HighscoreResult> {
  return api.get<HighscoreResult>(`/api/highscore${buildQuery(opts)}`);
}

export function fetchHighscoreMe(type?: ScoreType): Promise<HighscoreEntry> {
  const path = type && type !== 'total'
    ? `/api/highscore/me?type=${type}`
    : '/api/highscore/me';
  return api.get<HighscoreEntry>(path);
}

export function fetchHighscoreAlliances(): Promise<{
  alliances: HighscoreAlliance[] | null;
}> {
  return api.get<{ alliances: HighscoreAlliance[] | null }>(
    '/api/highscore/alliances',
  );
}

export function fetchHighscoreVacation(): Promise<{
  players: HighscoreVacation[] | null;
}> {
  return api.get<{ players: HighscoreVacation[] | null }>(
    '/api/highscore/vacation',
  );
}

export function fetchPublicStats(): Promise<PublicStats> {
  return api.get<PublicStats>('/api/stats');
}
