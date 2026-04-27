// Список ролей с lazy-fetch permissions при раскрытии row.
// Endpoint: GET /api/admin/roles, GET /api/admin/roles/{id}/permissions.
// Permission: roles:read.
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { ChevronDown, ChevronRight, ShieldCheck, Lock } from 'lucide-react';
import { rbacApi, type Role, type Permission } from '@/lib/api/rbac';
import { useAuth } from '@/store/auth';
import { PageHeader } from '@/components/layout/PageHeader';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';

export function Roles(): React.ReactElement {
  const hasPermission = useAuth((s) => s.hasPermission);
  const canRead = hasPermission('roles:read');
  const [expandedId, setExpandedId] = useState<number | null>(null);

  const rolesQuery = useQuery({
    queryKey: ['rbac', 'roles'],
    queryFn: rbacApi.listRoles,
    enabled: canRead,
  });

  if (!canRead) {
    return (
      <>
        <PageHeader title="Roles" />
        <Card className="max-w-xl">
          <CardContent className="pt-4 flex items-center gap-2 text-sm text-muted-foreground">
            <Lock className="h-4 w-4" aria-hidden="true" />
            Требуется permission <code className="font-mono-sm">roles:read</code>
          </CardContent>
        </Card>
      </>
    );
  }

  return (
    <>
      <PageHeader
        title="Roles"
        description="справочник ролей и привязанных permissions"
      />
      {rolesQuery.isLoading && (
        <div className="space-y-2">
          {[0, 1, 2].map((i) => (
            <Skeleton key={i} className="h-10 w-full max-w-2xl" />
          ))}
        </div>
      )}
      {rolesQuery.error && (
        <Card className="max-w-xl border-destructive/50">
          <CardContent className="pt-4 text-sm text-destructive">
            Ошибка загрузки: {String((rolesQuery.error as Error).message)}
          </CardContent>
        </Card>
      )}
      {rolesQuery.data && (
        <div className="rounded-md border bg-card overflow-hidden max-w-3xl">
          <table className="w-full text-sm">
            <thead className="bg-muted/50">
              <tr className="text-left text-2xs uppercase tracking-wide text-muted-foreground">
                <th className="px-3 py-2 w-8" />
                <th className="px-3 py-2">name</th>
                <th className="px-3 py-2">description</th>
                <th className="px-3 py-2 w-20">type</th>
              </tr>
            </thead>
            <tbody>
              {rolesQuery.data.map((role) => (
                <RoleRow
                  key={role.id}
                  role={role}
                  expanded={expandedId === role.id}
                  onToggle={() => setExpandedId(expandedId === role.id ? null : role.id)}
                />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
}

interface RoleRowProps {
  role: Role;
  expanded: boolean;
  onToggle: () => void;
}

function RoleRow({ role, expanded, onToggle }: RoleRowProps): React.ReactElement {
  const permsQuery = useQuery({
    queryKey: ['rbac', 'roles', role.id, 'permissions'],
    queryFn: () => rbacApi.getRolePermissions(role.id),
    enabled: expanded,
    staleTime: 5 * 60_000,
  });

  return (
    <>
      <tr className="border-t hover:bg-accent/30 cursor-pointer" onClick={onToggle}>
        <td className="px-3 py-2 align-top">
          {expanded ? (
            <ChevronDown className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          ) : (
            <ChevronRight className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          )}
        </td>
        <td className="px-3 py-2 align-top">
          <div className="flex items-center gap-2 font-mono-sm">
            <ShieldCheck className="h-3.5 w-3.5 text-primary" aria-hidden="true" />
            {role.name}
          </div>
        </td>
        <td className="px-3 py-2 align-top text-xs text-muted-foreground">
          {role.description || '—'}
        </td>
        <td className="px-3 py-2 align-top">
          {role.is_system ? (
            <Badge variant="secondary">system</Badge>
          ) : (
            <Badge variant="outline">custom</Badge>
          )}
        </td>
      </tr>
      {expanded && (
        <tr className="border-t bg-muted/20">
          <td />
          <td colSpan={3} className="px-3 py-3">
            {permsQuery.isLoading && <Skeleton className="h-16 w-64" />}
            {permsQuery.error && (
              <p className="text-xs text-destructive">
                Ошибка: {String((permsQuery.error as Error).message)}
              </p>
            )}
            {permsQuery.data && (
              <PermissionsList permissions={permsQuery.data} />
            )}
          </td>
        </tr>
      )}
    </>
  );
}

function PermissionsList({ permissions }: { permissions: Permission[] }): React.ReactElement {
  if (permissions.length === 0) {
    return <p className="text-xs text-muted-foreground">— нет permissions —</p>
  }
  return (
    <div className="flex flex-wrap gap-1">
      {permissions.map((p) => (
        <Badge key={p.id} variant="outline" className="font-mono-sm">
          {p.name}
        </Badge>
      ))}
    </div>
  );
}
