// UGC reports API: список + резолюция жалоб (149-ФЗ).
// Endpoints проксируются admin-bff в portal-backend (план 56 Ф.6):
//   GET  /api/admin/reports?status=&limit=
//   POST /api/admin/reports/{id}/resolve
import { apiRequest } from './client';

export type ReportStatus = 'new' | 'resolved' | 'rejected';
export type ReportTargetType = 'user' | 'alliance' | 'chat_msg' | 'planet';

export interface Report {
  id: string;
  reporter_id: string;
  reporter_name: string;
  target_type: ReportTargetType;
  target_id: string;
  reason: string;
  comment: string;
  status: ReportStatus;
  resolved_by?: string;
  resolver_name?: string;
  resolution_note?: string;
  created_at: string;
  resolved_at?: string;
}

export interface ReportsQuery {
  status?: ReportStatus | '';
  limit?: number;
}

export interface ResolveBody {
  status: 'resolved' | 'rejected';
  note: string;
}

export const reportsApi = {
  list: (q: ReportsQuery = {}) => {
    const p = new URLSearchParams();
    if (q.status) p.set('status', q.status);
    if (q.limit !== undefined) p.set('limit', String(q.limit));
    const qs = p.toString();
    return apiRequest<{ reports: Report[] }>(
      `/api/admin/reports${qs ? `?${qs}` : ''}`,
    ).then((r) => r.reports ?? []);
  },

  resolve: (id: string, body: ResolveBody) =>
    apiRequest<{ status: string }>(
      `/api/admin/reports/${encodeURIComponent(id)}/resolve`,
      { method: 'POST', body },
    ),
};
