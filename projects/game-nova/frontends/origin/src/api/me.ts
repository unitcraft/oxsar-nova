// API-модуль /api/me для origin-фронта (план 72.1 ч.16).
//
// Endpoint:
//   GET /api/me → { user_id, username, profession, credit, roles[] }
//
// Используется в TopHeader (credit для блока «Кредиты») и MainScreen
// (profession для строки «Профессия» в legacy-эталоне).

import { api } from './client';

export interface MeInfo {
  user_id: string;
  username: string;
  profession: string;
  credit: number;
  roles: string[];
}

export function fetchMe(): Promise<MeInfo> {
  return api.get<MeInfo>('/api/me');
}
