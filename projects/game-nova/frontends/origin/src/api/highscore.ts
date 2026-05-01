// API-модуль рейтингов / статистики origin-фронта
// (план 72 Ф.4 Spring 3, S-023, S-024, S-032; расширен 72.1.12).
//
// Endpoints (openapi.yaml + backend):
//   GET /api/highscore?type=...       → { entries: HighscoreEntry[] }
//   GET /api/highscore/me?type=...    → HighscoreEntry (поле score = выбранная метрика)
//   GET /api/highscore/alliances      → { alliances: HighscoreAlliance[] }
//   GET /api/highscore/vacation       → { players: HighscoreVacation[] }
//   GET /api/stats                    → { online_now, online_24h }

import { api } from './client';
import type {
  HighscoreEntry,
  HighscoreAlliance,
  HighscoreVacation,
  PublicStats,
} from './types';

// План 72.1.12: legacy `Ranking.class.php` поддерживает много типов; в backend
// `score.columnFor` маппинг: total=points, b=b_points, r=r_points, u=u_points,
// a=a_points, e=e_points, dm=dm_points, max=max_points.
export type ScoreType = 'total' | 'b' | 'r' | 'u' | 'a' | 'e' | 'dm' | 'max';

function withType(path: string, type?: ScoreType): string {
  if (!type || type === 'total') return path;
  const sep = path.includes('?') ? '&' : '?';
  return `${path}${sep}type=${type}`;
}

export function fetchHighscore(type?: ScoreType): Promise<{
  entries: HighscoreEntry[] | null;
}> {
  return api.get<{ entries: HighscoreEntry[] | null }>(
    withType('/api/highscore', type),
  );
}

export function fetchHighscoreMe(type?: ScoreType): Promise<HighscoreEntry> {
  return api.get<HighscoreEntry>(withType('/api/highscore/me', type));
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
