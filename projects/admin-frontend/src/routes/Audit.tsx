// Audit log: фильтры + таблица + пагинация + CSV-export.
// Endpoint: GET /api/admin/audit/role-changes (план 52).
// Permission: audit:read.
//
// Фильтры sync'ятся в URL query (deep-linking, как Sentry):
// ?actor_id=...&target_id=...&action=grant&since=...&until=...
import { useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Download, Lock, RefreshCw, X } from 'lucide-react';
import { rbacApi, type AuditQuery, type AuditRoleChange } from '@/lib/api/rbac';
import { ApiError } from '@/lib/api/client';
import { useAuth } from '@/store/auth';
import { PageHeader } from '@/components/layout/PageHeader';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { formatDateTime } from '@/lib/utils';

const PAGE_SIZE = 50;

export function Audit(): React.ReactElement {
  const hasPermission = useAuth((s) => s.hasPermission);
  const canRead = hasPermission('audit:read');
  const [params, setParams] = useSearchParams();

  const filters = useMemo<AuditQuery>(() => {
    const offset = parseInt(params.get('offset') ?? '0', 10) || 0;
    const action = params.get('action') as AuditQuery['action'];
    const q: AuditQuery = {
      limit: PAGE_SIZE,
      offset,
    };
    if (action === 'grant' || action === 'revoke') q.action = action;
    const actor = params.get('actor_id');
    if (actor) q.actor_id = actor;
    const target = params.get('target_id');
    if (target) q.target_id = target;
    const since = params.get('since');
    if (since) q.since = since;
    const until = params.get('until');
    if (until) q.until = until;
    return q;
  }, [params]);

  const auditQuery = useQuery({
    queryKey: ['rbac', 'audit', filters],
    queryFn: () => rbacApi.queryAudit(filters),
    enabled: canRead,
  });

  if (!canRead) {
    return (
      <>
        <PageHeader title="Audit log" />
        <Card className="max-w-xl">
          <CardContent className="pt-4 flex items-center gap-2 text-sm text-muted-foreground">
            <Lock className="h-4 w-4" aria-hidden="true" />
            Требуется permission <code className="font-mono-sm">audit:read</code>
          </CardContent>
        </Card>
      </>
    );
  }

  function setFilter(key: string, value: string | null): void {
    const next = new URLSearchParams(params);
    if (value === null || value === '') {
      next.delete(key);
    } else {
      next.set(key, value);
    }
    next.delete('offset'); // фильтры → возврат на первую страницу
    setParams(next, { replace: true });
  }

  function setOffset(offset: number): void {
    const next = new URLSearchParams(params);
    if (offset <= 0) next.delete('offset');
    else next.set('offset', String(offset));
    setParams(next, { replace: true });
  }

  function clearAll(): void {
    setParams({}, { replace: true });
  }

  function exportCsv(): void {
    if (!auditQuery.data) return;
    const rows = auditQuery.data;
    const header = [
      'id',
      'created_at',
      'action',
      'role_name',
      'actor_id',
      'target_id',
      'reason',
      'ip_address',
      'user_agent',
    ];
    const lines = [header.join(',')];
    for (const r of rows) {
      lines.push(
        [
          r.id,
          r.created_at,
          r.action,
          csvField(r.role_name),
          r.actor_id,
          r.target_id,
          csvField(r.reason),
          r.ip_address ?? '',
          csvField(r.user_agent),
        ].join(','),
      );
    }
    downloadCsv(lines.join('\n'), `audit-${Date.now()}.csv`);
  }

  const offset = filters.offset ?? 0;
  const data = auditQuery.data ?? [];
  const hasMore = data.length === PAGE_SIZE;
  const hasFilters = Array.from(params.keys()).some((k) => k !== 'offset');

  return (
    <>
      <PageHeader
        title="Audit log"
        description="неизменяемый журнал назначений и отзывов ролей (план 52)"
        action={
          <>
            <Button
              variant="outline"
              size="sm"
              onClick={() => void auditQuery.refetch()}
              disabled={auditQuery.isFetching}
              aria-label="Refresh"
            >
              <RefreshCw
                className={`h-3.5 w-3.5 ${auditQuery.isFetching ? 'animate-spin' : ''}`}
                aria-hidden="true"
              />
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={exportCsv}
              disabled={data.length === 0}
            >
              <Download className="h-3.5 w-3.5" aria-hidden="true" />
              CSV
            </Button>
          </>
        }
      />

      <Card className="mb-3">
        <CardContent className="pt-4">
          <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-5">
            <FilterField
              id="action"
              label="Action"
              value={params.get('action') ?? ''}
              onChange={(v) => setFilter('action', v || null)}
              as="select"
              options={[
                { value: '', label: '— любое —' },
                { value: 'grant', label: 'grant' },
                { value: 'revoke', label: 'revoke' },
              ]}
            />
            <FilterField
              id="actor_id"
              label="Actor (UUID)"
              placeholder="00000000-…"
              value={params.get('actor_id') ?? ''}
              onChange={(v) => setFilter('actor_id', v || null)}
            />
            <FilterField
              id="target_id"
              label="Target (UUID)"
              placeholder="00000000-…"
              value={params.get('target_id') ?? ''}
              onChange={(v) => setFilter('target_id', v || null)}
            />
            <FilterField
              id="since"
              label="Since"
              type="datetime-local"
              value={toLocalInput(params.get('since'))}
              onChange={(v) => setFilter('since', v ? new Date(v).toISOString() : null)}
            />
            <FilterField
              id="until"
              label="Until"
              type="datetime-local"
              value={toLocalInput(params.get('until'))}
              onChange={(v) => setFilter('until', v ? new Date(v).toISOString() : null)}
            />
          </div>
          {hasFilters && (
            <div className="mt-3">
              <Button variant="ghost" size="sm" onClick={clearAll}>
                <X className="h-3.5 w-3.5" aria-hidden="true" />
                Сбросить фильтры
              </Button>
            </div>
          )}
        </CardContent>
      </Card>

      {auditQuery.isLoading && (
        <div className="space-y-2">
          {[0, 1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-9 w-full" />
          ))}
        </div>
      )}
      {auditQuery.error && (
        <Card className="border-destructive/50">
          <CardContent className="pt-4 text-sm text-destructive">
            {auditQuery.error instanceof ApiError && auditQuery.error.status === 403
              ? 'Нет permission audit:read'
              : `Ошибка: ${(auditQuery.error as Error).message}`}
          </CardContent>
        </Card>
      )}
      {auditQuery.data && (
        <>
          <div className="rounded-md border bg-card overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-muted/50 text-2xs uppercase tracking-wide text-muted-foreground">
                <tr className="text-left">
                  <th className="px-3 py-2">when</th>
                  <th className="px-3 py-2 w-20">action</th>
                  <th className="px-3 py-2">role</th>
                  <th className="px-3 py-2">actor</th>
                  <th className="px-3 py-2">target</th>
                  <th className="px-3 py-2">reason</th>
                  <th className="px-3 py-2 w-28">ip</th>
                </tr>
              </thead>
              <tbody>
                {auditQuery.data.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-3 py-6 text-center text-xs text-muted-foreground">
                      — нет записей по текущим фильтрам —
                    </td>
                  </tr>
                )}
                {auditQuery.data.map((row) => (
                  <AuditRow key={row.id} row={row} />
                ))}
              </tbody>
            </table>
          </div>
          <div className="mt-3 flex items-center justify-between text-xs text-muted-foreground">
            <span>
              {data.length === 0
                ? 'нет данных'
                : `показано ${offset + 1}–${offset + data.length}`}
            </span>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={offset === 0}
                onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
              >
                ← Назад
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={!hasMore}
                onClick={() => setOffset(offset + PAGE_SIZE)}
              >
                Вперёд →
              </Button>
            </div>
          </div>
        </>
      )}
    </>
  );
}

interface FilterFieldProps {
  id: string;
  label: string;
  value: string;
  onChange: (v: string) => void;
  type?: string;
  placeholder?: string;
  as?: 'input' | 'select';
  options?: { value: string; label: string }[];
}

function FilterField({
  id,
  label,
  value,
  onChange,
  type = 'text',
  placeholder,
  as = 'input',
  options,
}: FilterFieldProps): React.ReactElement {
  return (
    <div className="space-y-1">
      <label htmlFor={id} className="text-2xs uppercase tracking-wide text-muted-foreground">
        {label}
      </label>
      {as === 'select' ? (
        <select
          id={id}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="flex h-8 w-full rounded-md border border-input bg-background px-2 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
        >
          {options?.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
      ) : (
        <Input
          id={id}
          type={type}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          {...(placeholder ? { placeholder } : {})}
          className="font-mono-sm"
        />
      )}
    </div>
  );
}

function AuditRow({ row }: { row: AuditRoleChange }): React.ReactElement {
  return (
    <tr className="border-t hover:bg-accent/30">
      <td className="px-3 py-2 font-mono-sm text-xs whitespace-nowrap">
        {formatDateTime(row.created_at)}
      </td>
      <td className="px-3 py-2">
        <Badge variant={row.action === 'grant' ? 'success' : 'destructive'}>
          {row.action}
        </Badge>
      </td>
      <td className="px-3 py-2 font-mono-sm">
        {row.role_name}
      </td>
      <td className="px-3 py-2 font-mono-sm text-2xs text-muted-foreground">
        {shortUuid(row.actor_id)}
      </td>
      <td className="px-3 py-2 font-mono-sm text-2xs text-muted-foreground">
        {shortUuid(row.target_id)}
      </td>
      <td className="px-3 py-2 text-xs">
        {row.reason}
      </td>
      <td className="px-3 py-2 font-mono-sm text-2xs text-muted-foreground">
        {row.ip_address ?? '—'}
      </td>
    </tr>
  );
}

// === helpers ===

// Возьмём последний блок UUIDv7 для коротких меток (по memory-правилу:
// первые 8 байт — timestamp, не уникальны в одной ms).
function shortUuid(id: string): string {
  const last = id.split('-').pop() ?? id;
  return last.slice(0, 8);
}

function csvField(s: string): string {
  if (s.includes('"') || s.includes(',') || s.includes('\n')) {
    return `"${s.replace(/"/g, '""')}"`;
  }
  return s;
}

function downloadCsv(content: string, filename: string): void {
  const blob = new Blob([content], { type: 'text/csv;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

function toLocalInput(iso: string | null): string {
  if (!iso) return '';
  // ISO → datetime-local (YYYY-MM-DDTHH:MM).
  const d = new Date(iso);
  if (isNaN(d.getTime())) return '';
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}
