// Game-ops events API: dead events + active events с фильтрами.
// Endpoints (game-nova через admin-bff /api/admin/events*):
//   GET  /api/admin/events?state=&kind=&limit=&offset=
//   GET  /api/admin/events/dead?kind=&limit=&offset=
//   POST /api/admin/events/{id}/retry
//   POST /api/admin/events/{id}/cancel
//   POST /api/admin/events/dead/{id}/resurrect
import { apiRequest } from './client';

export type EventState = 'wait' | 'error' | 'ok';

export interface EventRow {
  id: string;
  user_id: string | null;
  planet_id: string | null;
  kind: number;
  state: EventState;
  fire_at: string;
  created_at: string;
  processed_at: string | null;
  attempt: number;
  last_error: string;
}

export interface DeadEvent {
  id: string;
  user_id?: string;
  planet_id?: string;
  kind: number;
  fire_at: string;
  payload: unknown;
  created_at: string;
  processed_at?: string;
  attempt: number;
  last_error: string;
  failed_at: string;
}

export interface EventsQuery {
  state?: EventState;
  kind?: number;
  limit?: number;
  offset?: number;
}

export const eventsApi = {
  listEvents: (q: EventsQuery = {}) => {
    const p = new URLSearchParams();
    if (q.state) p.set('state', q.state);
    if (q.kind !== undefined) p.set('kind', String(q.kind));
    if (q.limit !== undefined) p.set('limit', String(q.limit));
    if (q.offset !== undefined) p.set('offset', String(q.offset));
    const qs = p.toString();
    return apiRequest<{ events: EventRow[] }>(
      `/api/admin/events${qs ? `?${qs}` : ''}`,
    ).then((r) => r.events);
  },

  listDead: (q: { kind?: number; limit?: number; offset?: number } = {}) => {
    const p = new URLSearchParams();
    if (q.kind !== undefined) p.set('kind', String(q.kind));
    if (q.limit !== undefined) p.set('limit', String(q.limit));
    if (q.offset !== undefined) p.set('offset', String(q.offset));
    const qs = p.toString();
    return apiRequest<{ events: DeadEvent[] }>(
      `/api/admin/events/dead${qs ? `?${qs}` : ''}`,
    ).then((r) => r.events);
  },

  retry: (id: string) =>
    apiRequest<{ status: string }>(
      `/api/admin/events/${encodeURIComponent(id)}/retry`,
      { method: 'POST' },
    ),

  cancel: (id: string) =>
    apiRequest<{ status: string }>(
      `/api/admin/events/${encodeURIComponent(id)}/cancel`,
      { method: 'POST' },
    ),

  resurrect: (id: string) =>
    apiRequest<{ status: string }>(
      `/api/admin/events/dead/${encodeURIComponent(id)}/resurrect`,
      { method: 'POST' },
    ),
};
