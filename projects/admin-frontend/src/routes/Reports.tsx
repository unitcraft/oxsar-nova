// План 56 Ф.6: страница модерации UGC-жалоб (149-ФЗ).
// Источник данных — portal-backend через admin-bff:
//   GET  /api/admin/reports?status=&limit=
//   POST /api/admin/reports/{id}/resolve
// Permission gate: moderation:reports:read (см. план 52 RBAC). До тех
// пор пока identity не кладёт permissions в JWT (план 52 Ф.X), backend
// AdminMiddleware пропускает по role=admin — UI-гейт здесь служит для
// скрытия пункта в навигации, защита defence-in-depth остаётся на
// portal-backend.
import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Lock, Flag } from 'lucide-react';
import {
  reportsApi,
  type Report,
  type ReportStatus,
} from '@/lib/api/reports';
import { ApiError } from '@/lib/api/client';
import { useAuth } from '@/store/auth';
import { PageHeader } from '@/components/layout/PageHeader';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { formatDateTime } from '@/lib/utils';

const PAGE_LIMIT = 100;
const STATUS_OPTIONS: Array<{ value: ReportStatus | ''; label: string }> = [
  { value: 'new', label: 'Новые' },
  { value: 'resolved', label: 'Решённые' },
  { value: 'rejected', label: 'Отклонённые' },
  { value: '', label: 'Все' },
];

export function Reports(): React.ReactElement {
  const hasPermission = useAuth((s) => s.hasPermission);
  // План 52 RBAC: при добавлении permissions в JWT — поменять gate на
  // moderation:reports:read. Сейчас identity не кладёт permissions, и
  // hasPermission всегда вернёт false для всех permission-gates → пока
  // оставляем role-based fallback.
  const canRead =
    hasPermission('moderation:reports:read') || useAuth.getState().claims?.roles.includes('admin') === true;
  const canResolve =
    hasPermission('moderation:reports:resolve') ||
    useAuth.getState().claims?.roles.includes('admin') === true;

  const [status, setStatus] = useState<ReportStatus | ''>('new');
  const [active, setActive] = useState<Report | null>(null);

  const list = useQuery({
    queryKey: ['reports', status],
    queryFn: () => reportsApi.list({ status, limit: PAGE_LIMIT }),
    enabled: canRead,
    refetchInterval: 30_000,
  });

  if (!canRead) {
    return (
      <>
        <PageHeader title="Жалобы" />
        <Card className="max-w-xl">
          <CardContent className="pt-4 flex items-center gap-2 text-sm text-muted-foreground">
            <Lock className="h-4 w-4" aria-hidden="true" />
            Требуется permission <code className="font-mono-sm">moderation:reports:read</code>
          </CardContent>
        </Card>
      </>
    );
  }

  return (
    <>
      <PageHeader
        title="Жалобы"
        description="UGC-модерация: жалобы игроков на пользователей, альянсы, чат, планеты (149-ФЗ)"
      />

      <div className="mb-3 flex flex-wrap items-center gap-2">
        <span className="text-2xs uppercase tracking-wide text-muted-foreground">
          Статус
        </span>
        {STATUS_OPTIONS.map((o) => (
          <Button
            key={o.value || 'all'}
            type="button"
            size="sm"
            variant={status === o.value ? 'default' : 'outline'}
            onClick={() => setStatus(o.value)}
          >
            {o.label}
          </Button>
        ))}
        <span className="ml-auto text-xs text-muted-foreground">
          {list.data?.length ?? 0} запис.
        </span>
      </div>

      <div className="rounded-md border bg-card overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-muted/50 text-2xs uppercase tracking-wide text-muted-foreground">
            <tr className="text-left">
              <th className="px-3 py-2">Дата</th>
              <th className="px-3 py-2">От</th>
              <th className="px-3 py-2 w-24">Тип</th>
              <th className="px-3 py-2">Цель</th>
              <th className="px-3 py-2">Причина</th>
              <th className="px-3 py-2 w-24">Статус</th>
              <th className="px-3 py-2 w-20" />
            </tr>
          </thead>
          <tbody>
            {list.isLoading && <SkeletonRows />}
            {list.error && <ErrorRow err={list.error} />}
            {!list.isLoading && list.data && list.data.length === 0 && (
              <tr>
                <td colSpan={7} className="px-3 py-6 text-center text-xs text-muted-foreground">
                  — нет жалоб —
                </td>
              </tr>
            )}
            {(list.data ?? []).map((r) => (
              <tr key={r.id} className="border-t hover:bg-accent/30">
                <td className="px-3 py-2 font-mono-sm text-xs whitespace-nowrap">
                  {formatDateTime(r.created_at)}
                </td>
                <td className="px-3 py-2">{r.reporter_name || r.reporter_id.slice(0, 8)}</td>
                <td className="px-3 py-2">
                  <code className="font-mono-sm text-xs">{r.target_type}</code>
                </td>
                <td className="px-3 py-2 font-mono-sm text-2xs">
                  {r.target_id.length > 14 ? `${r.target_id.slice(0, 14)}…` : r.target_id}
                </td>
                <td className="px-3 py-2 text-xs">{r.reason}</td>
                <td className="px-3 py-2">
                  <Badge variant={statusBadge(r.status)}>{r.status}</Badge>
                </td>
                <td className="px-3 py-2">
                  <Button size="sm" variant="ghost" onClick={() => setActive(r)}>
                    <Flag className="h-3 w-3" aria-hidden="true" />
                    Открыть
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {active && (
        <ReportDetail
          report={active}
          canResolve={canResolve}
          onClose={() => setActive(null)}
        />
      )}
    </>
  );
}

interface DetailProps {
  report: Report;
  canResolve: boolean;
  onClose: () => void;
}

function ReportDetail({ report, canResolve, onClose }: DetailProps): React.ReactElement {
  const queryClient = useQueryClient();
  const [note, setNote] = useState('');

  const resolveMut = useMutation({
    mutationFn: (status: 'resolved' | 'rejected') =>
      reportsApi.resolve(report.id, { status, note }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['reports'] });
      onClose();
    },
  });

  const isPending = resolveMut.isPending;
  const error = resolveMut.error;

  return (
    <Dialog open onOpenChange={(o) => { if (!o) onClose(); }}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <DialogTitle>Жалоба от {report.reporter_name || report.reporter_id.slice(0, 8)}</DialogTitle>
        </DialogHeader>

        <dl className="grid grid-cols-[max-content_1fr] gap-x-3 gap-y-1 text-xs">
          <dt className="text-muted-foreground">Дата:</dt>
          <dd className="font-mono-sm">{formatDateTime(report.created_at)}</dd>
          <dt className="text-muted-foreground">Тип цели:</dt>
          <dd><code className="font-mono-sm">{report.target_type}</code></dd>
          <dt className="text-muted-foreground">ID цели:</dt>
          <dd className="font-mono-sm break-all">{report.target_id}</dd>
          <dt className="text-muted-foreground">Причина:</dt>
          <dd>{report.reason}</dd>
          <dt className="text-muted-foreground">Статус:</dt>
          <dd><Badge variant={statusBadge(report.status)}>{report.status}</Badge></dd>
        </dl>

        {report.comment && (
          <div className="mt-3 rounded border bg-muted/30 p-2 text-xs whitespace-pre-wrap">
            {report.comment}
          </div>
        )}

        {report.status === 'new' ? (
          <>
            <div className="mt-4 space-y-1">
              <label
                htmlFor="resolution-note"
                className="text-2xs uppercase tracking-wide text-muted-foreground"
              >
                Пометка о решении
              </label>
              <textarea
                id="resolution-note"
                value={note}
                onChange={(e) => setNote(e.target.value)}
                rows={3}
                maxLength={1000}
                placeholder="Что сделано (warn/mute/ban/rename) или почему отклонено"
                className="w-full rounded-md border border-input bg-background p-2 text-sm shadow-sm"
                disabled={!canResolve || isPending}
              />
            </div>

            {error && (
              <div role="alert" className="mt-2 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
                {formatError(error)}
              </div>
            )}

            <DialogFooter>
              <Button type="button" variant="outline" onClick={onClose}>
                Закрыть
              </Button>
              <Button
                type="button"
                variant="ghost"
                disabled={!canResolve || isPending}
                onClick={() => resolveMut.mutate('rejected')}
              >
                Отклонить
              </Button>
              <Button
                type="button"
                disabled={!canResolve || isPending}
                onClick={() => resolveMut.mutate('resolved')}
              >
                Принять
              </Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <p className="mt-3 text-xs text-muted-foreground">
              Решено {report.resolver_name || '?'} —{' '}
              {report.resolved_at ? formatDateTime(report.resolved_at) : ''}
            </p>
            {report.resolution_note && (
              <div className="mt-2 rounded border bg-muted/30 p-2 text-xs whitespace-pre-wrap">
                {report.resolution_note}
              </div>
            )}
            <DialogFooter>
              <Button type="button" onClick={onClose}>
                Закрыть
              </Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}

function statusBadge(s: ReportStatus): 'destructive' | 'success' | 'secondary' {
  if (s === 'new') return 'destructive';
  if (s === 'resolved') return 'success';
  return 'secondary';
}

function SkeletonRows(): React.ReactElement {
  return (
    <>
      {[0, 1, 2].map((i) => (
        <tr key={i} className="border-t">
          {[0, 1, 2, 3, 4, 5, 6].map((j) => (
            <td key={j} className="px-3 py-2">
              <Skeleton className="h-4 w-full" />
            </td>
          ))}
        </tr>
      ))}
    </>
  );
}

function ErrorRow({ err }: { err: unknown }): React.ReactElement {
  return (
    <tr>
      <td colSpan={7} className="px-3 py-6 text-center text-xs text-destructive">
        {formatError(err)}
      </td>
    </tr>
  );
}

function formatError(err: unknown): string {
  if (err instanceof ApiError) {
    if (err.status === 403) return 'Нет permission';
    return err.body?.message || err.body?.error || `HTTP ${err.status}`;
  }
  if (err instanceof Error) return err.message;
  return 'Ошибка';
}
