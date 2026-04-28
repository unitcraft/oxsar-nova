// API-модуль search origin-фронта (план 72 Ф.5 Spring 4 — S-039).
//
// Endpoint (openapi.yaml):
//   GET /api/search?q=...&type=player|alliance|planet (type опционально)

import { api } from './client';
import type { SearchResults, SearchType } from './types';

export function search(q: string, type?: SearchType): Promise<SearchResults> {
  const params = new URLSearchParams();
  params.set('q', q);
  if (type) params.set('type', type);
  return api.get<SearchResults>(`/api/search?${params.toString()}`);
}
