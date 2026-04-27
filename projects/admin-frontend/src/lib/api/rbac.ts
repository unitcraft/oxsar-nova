// RBAC API client — обёртка над admin-bff endpoints (которые проксируют
// на identity-service). DTO повторяют структуру из
// projects/identity/backend/internal/identitysvc/rbac.go.
//
// Все запросы permission-protected на стороне backend; UI скрывает
// кнопки если у юзера нет permission, но это UX, а не security.
import { apiRequest } from './client';

export interface Role {
  id: number;
  name: string;
  description: string;
  is_system: boolean;
  created_at: string;
}

export interface Permission {
  id: number;
  name: string;
  description: string;
  created_at: string;
}

export interface UserRoleAssignment {
  user_id: string;
  role_id: number;
  role_name: string;
  granted_by: string | null;
  granted_at: string;
  expires_at: string | null;
}

export interface AuditRoleChange {
  id: number;
  actor_id: string;
  target_id: string;
  role_name: string;
  action: 'grant' | 'revoke';
  reason: string;
  ip_address: string | null;
  user_agent: string;
  created_at: string;
}

export interface GrantRoleInput {
  role: string;
  reason: string;
  expires_at?: string;
}

export interface AuditQuery {
  actor_id?: string;
  target_id?: string;
  action?: 'grant' | 'revoke';
  since?: string;
  until?: string;
  limit?: number;
  offset?: number;
}

export const rbacApi = {
  listRoles: () =>
    apiRequest<{ roles: Role[] }>('/api/admin/roles').then((r) => r.roles),

  getRolePermissions: (roleId: number) =>
    apiRequest<{ permissions: Permission[] }>(
      `/api/admin/roles/${roleId}/permissions`,
    ).then((r) => r.permissions),

  listUserRoles: (userId: string) =>
    apiRequest<{ assignments: UserRoleAssignment[] }>(
      `/api/admin/users/${encodeURIComponent(userId)}/roles`,
    ).then((r) => r.assignments),

  grantUserRole: (userId: string, input: GrantRoleInput) =>
    apiRequest<{ status: string }>(
      `/api/admin/users/${encodeURIComponent(userId)}/roles`,
      { method: 'POST', body: input },
    ),

  revokeUserRole: (userId: string, role: string, reason: string) => {
    const qs = new URLSearchParams({ reason }).toString();
    return apiRequest<{ status: string }>(
      `/api/admin/users/${encodeURIComponent(userId)}/roles/${encodeURIComponent(role)}?${qs}`,
      { method: 'DELETE' },
    );
  },

  queryAudit: (q: AuditQuery = {}) => {
    const params = new URLSearchParams();
    if (q.actor_id) params.set('actor_id', q.actor_id);
    if (q.target_id) params.set('target_id', q.target_id);
    if (q.action) params.set('action', q.action);
    if (q.since) params.set('since', q.since);
    if (q.until) params.set('until', q.until);
    if (q.limit !== undefined) params.set('limit', String(q.limit));
    if (q.offset !== undefined) params.set('offset', String(q.offset));
    const qs = params.toString();
    const url = `/api/admin/audit/role-changes${qs ? `?${qs}` : ''}`;
    return apiRequest<{ events: AuditRoleChange[] }>(url).then((r) => r.events);
  },
};
