// User detail: показывает текущие role assignments + grant/revoke формы.
// Endpoints: GET /api/admin/users/{id}/roles, POST .../roles,
// DELETE .../roles/{role}?reason=...
//
// Permission gates на UI:
// - roles:read   → вообще видеть страницу.
// - roles:grant  → видна кнопка "Grant role".
// - roles:revoke → виден inline "Revoke" возле каждой роли.
import { useState } from 'react';
import { Navigate, useParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Trash2, Plus, Lock, ScrollText } from 'lucide-react';
import { rbacApi, type UserRoleAssignment } from '@/lib/api/rbac';
import { useAuth } from '@/store/auth';
import { PageHeader } from '@/components/layout/PageHeader';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { ApiError } from '@/lib/api/client';
import { GrantRoleDialog } from '@/components/rbac/GrantRoleDialog';
import { formatDateTime } from '@/lib/utils';

const UUID_RE = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export function UserDetail(): React.ReactElement {
  const { id = '' } = useParams<{ id: string }>();
  const hasPermission = useAuth((s) => s.hasPermission);
  const canRead = hasPermission('roles:read');
  const canGrant = hasPermission('roles:grant');
  const canRevoke = hasPermission('roles:revoke');
  const [grantOpen, setGrantOpen] = useState(false);

  const queryClient = useQueryClient();

  if (!UUID_RE.test(id)) {
    return <Navigate to="/users/lookup" replace />;
  }
  if (!canRead) {
    return <PermissionDenied perm="roles:read" />;
  }

  const assignmentsQuery = useQuery({
    queryKey: ['rbac', 'users', id, 'roles'],
    queryFn: () => rbacApi.listUserRoles(id),
  });

  const revokeMutation = useMutation({
    mutationFn: ({ role, reason }: { role: string; reason: string }) =>
      rbacApi.revokeUserRole(id, role, reason),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['rbac', 'users', id, 'roles'] });
      void queryClient.invalidateQueries({ queryKey: ['rbac', 'audit'] });
    },
  });

  function onRevoke(assignment: UserRoleAssignment): void {
    const reason = window.prompt(
      `Причина отзыва роли "${assignment.role_name}"? (обязательное поле для audit)`,
    );
    if (!reason || reason.trim() === '') return;
    revokeMutation.mutate({ role: assignment.role_name, reason: reason.trim() });
  }

  return (
    <>
      <PageHeader
        title="User detail"
        description={<span className="font-mono-sm text-foreground">{id}</span>}
        action={
          canGrant ? (
            <Button size="sm" onClick={() => setGrantOpen(true)}>
              <Plus className="h-3.5 w-3.5" aria-hidden="true" />
              Grant role
            </Button>
          ) : null
        }
      />

      {assignmentsQuery.isLoading && (
        <div className="space-y-2 max-w-2xl">
          <Skeleton className="h-8 w-full" />
          <Skeleton className="h-8 w-full" />
        </div>
      )}
      {assignmentsQuery.error && (
        <Card className="max-w-xl border-destructive/50">
          <CardContent className="pt-4 text-sm text-destructive">
            {assignmentsQuery.error instanceof ApiError && assignmentsQuery.error.status === 404
              ? 'Юзер не найден'
              : `Ошибка: ${(assignmentsQuery.error as Error).message}`}
          </CardContent>
        </Card>
      )}
      {assignmentsQuery.data && (
        <Card className="max-w-3xl">
          <CardContent className="pt-4">
            {assignmentsQuery.data.length === 0 ? (
              <p className="text-sm text-muted-foreground">— нет назначенных ролей —</p>
            ) : (
              <table className="w-full text-sm">
                <thead className="text-2xs uppercase tracking-wide text-muted-foreground">
                  <tr className="text-left">
                    <th className="px-2 py-2">role</th>
                    <th className="px-2 py-2">granted</th>
                    <th className="px-2 py-2">expires</th>
                    {canRevoke && <th className="px-2 py-2 w-10" />}
                  </tr>
                </thead>
                <tbody>
                  {assignmentsQuery.data.map((a) => (
                    <tr key={`${a.role_id}-${a.granted_at}`} className="border-t">
                      <td className="px-2 py-2">
                        <Badge>{a.role_name}</Badge>
                      </td>
                      <td className="px-2 py-2 font-mono-sm text-xs text-muted-foreground">
                        {formatDateTime(a.granted_at)}
                      </td>
                      <td className="px-2 py-2 font-mono-sm text-xs text-muted-foreground">
                        {a.expires_at ? formatDateTime(a.expires_at) : '—'}
                      </td>
                      {canRevoke && (
                        <td className="px-2 py-2">
                          <Button
                            size="icon"
                            variant="ghost"
                            aria-label={`Revoke ${a.role_name}`}
                            disabled={revokeMutation.isPending}
                            onClick={() => onRevoke(a)}
                          >
                            <Trash2 className="h-3.5 w-3.5 text-destructive" aria-hidden="true" />
                          </Button>
                        </td>
                      )}
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </CardContent>
        </Card>
      )}

      <Card className="mt-3 max-w-3xl">
        <CardContent className="pt-4 flex items-center gap-2 text-2xs text-muted-foreground">
          <ScrollText className="h-3 w-3" aria-hidden="true" />
          Все назначения и отзывы пишутся в audit log (план 52 §audit).
        </CardContent>
      </Card>

      <GrantRoleDialog
        open={grantOpen}
        onOpenChange={setGrantOpen}
        userId={id}
        existingRoles={assignmentsQuery.data?.map((a) => a.role_name) ?? []}
      />
    </>
  );
}

function PermissionDenied({ perm }: { perm: string }): React.ReactElement {
  return (
    <>
      <PageHeader title="User detail" />
      <Card className="max-w-xl">
        <CardContent className="pt-4 flex items-center gap-2 text-sm text-muted-foreground">
          <Lock className="h-4 w-4" aria-hidden="true" />
          Требуется permission <code className="font-mono-sm">{perm}</code>
        </CardContent>
      </Card>
    </>
  );
}
