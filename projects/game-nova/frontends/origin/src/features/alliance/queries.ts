// Pure-функции построения query-параметров alliance-API
// (план 72 Ф.3 Spring 2 ч.1).
//
// Вынесено из api/alliance.ts отдельно, чтобы тестировать без импорта
// auth-store + http-клиента (vitest без jsdom не имеет localStorage).

import type { AllianceListFilters } from '@/api/types';

export function buildSearchQuery(f: AllianceListFilters): string {
  const usp = new URLSearchParams();
  if (f.q && f.q.trim() !== '') usp.set('q', f.q.trim());
  if (f.is_open !== undefined) usp.set('is_open', String(f.is_open));
  if (f.min_members !== undefined) usp.set('min_members', String(f.min_members));
  if (f.max_members !== undefined) usp.set('max_members', String(f.max_members));
  if (f.limit !== undefined) usp.set('limit', String(f.limit));
  if (f.offset !== undefined) usp.set('offset', String(f.offset));
  return usp.toString();
}

export interface AuditQuery {
  action?: string | undefined;
  actor_id?: string | undefined;
  limit?: number | undefined;
  offset?: number | undefined;
}

export function buildAuditQuery(q: AuditQuery): string {
  const usp = new URLSearchParams();
  if (q.action) usp.set('action', q.action);
  if (q.actor_id) usp.set('actor_id', q.actor_id);
  if (q.limit !== undefined) usp.set('limit', String(q.limit));
  if (q.offset !== undefined) usp.set('offset', String(q.offset));
  return usp.toString();
}
