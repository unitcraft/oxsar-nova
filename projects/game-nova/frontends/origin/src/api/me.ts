// API-модуль /api/me для origin-фронта.
//
// План 72.1 ч.17: pixel-perfect MainScreen, расширен набор полей.
// Контракт описан в openapi.yaml#/components/schemas/MeInfo (там
// authoritative). Здесь — ручной TS-тип для совместимости с принятым
// в origin-фронте стилем (ad-hoc DTO в `api/`, см. `types.ts`).
// Перейдём на `import type { components } from './schema'` в Ф.7
// финализации плана 72 (когда вся origin-API будет сгенерирована).
//
// Endpoint:
//   GET /api/me → MeInfo

import { api } from './client';

export interface MeInfo {
  user_id: string;
  username: string;
  roles: string[];
  credit: number;
  profession: 'none' | 'miner' | 'attacker' | 'defender' | 'tank';
  // Pixel-perfect MainScreen (план 72.1 ч.17):
  points: number;
  rank: number;
  total_users: number;
  max_points: number;
  combat_experience: number;
  accumulated_experience: number;
  miner_level: number;
  miner_points: number;
  miner_need_points: number;
  dm_points: number;
  intergalactic_research_level: number;
  battles: number;
  // Vacation-режим (опциональные):
  vacation_since?: string;
  vacation_unlock_at?: string;
  vacation_last_end?: string;
}

export function fetchMe(): Promise<MeInfo> {
  return api.get<MeInfo>('/api/me');
}
