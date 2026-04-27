// Game events page: dead events + active events с retry/cancel/resurrect.
// Permission gates: game:events:retry / game:events:cancel.
import { useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Lock, RefreshCw, Skull, Activity, RotateCcw, X } from 'lucide-react';
import {
  eventsApi,
  type DeadEvent,
  type EventRow,
  type EventState,
} from '@/lib/api/events';
import { ApiError } from '@/lib/api/client';
import { useAuth } from '@/store/auth';
import { PageHeader } from '@/components/layout/PageHeader';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { formatDateTime } from '@/lib/utils';

const PAGE_SIZE = 100;
type Tab = 'active' | 'dead';

export function GameEvents(): React.ReactElement {
  const hasPermission = useAuth((s) => s.hasPermission);
  const canRetry = hasPermission('game:events:retry');
  const canCancel = hasPermission('game:events:cancel');
  // Доступ к самой странице — наличие любого из game:events:* permissions.
  // Backend всё равно ругнётся 403 если нет.
  const canSee = canRetry || canCancel;
  const [params, setParams] = useSearchParams();

  const tab: Tab = params.get('tab') === 'dead' ? 'dead' : 'active';
  const offset = parseInt(params.get('offset') ?? '0', 10) || 0;
  const stateFilter = (params.get('state') ?? '') as EventState | '';
  const kindFilter = params.get('kind') ?? '';

  function setTab(t: Tab): void {
    const next = new URLSearchParams();
    next.set('tab', t);
    setParams(next, { replace: true });
  }
  function setFilter(key: string, value: string | null): void {
    const next = new URLSearchParams(params);
    if (value === null || value === '') next.delete(key);
    else next.set(key, value);
    next.delete('offset');
    setParams(next, { replace: true });
  }
  function setOffset(n: number): void {
    const next = new URLSearchParams(params);
    if (n <= 0) next.delete('offset');
    else next.set('offset', String(n));
    setParams(next, { replace: true });
  }

  if (!canSee) {
    return (
      <>
        <PageHeader title="Game-ops events" />
        <Card className="max-w-xl">
          <CardContent className="pt-4 flex items-center gap-2 text-sm text-muted-foreground">
            <Lock className="h-4 w-4" aria-hidden="true" />
            Требуется permission <code className="font-mono-sm">game:events:retry</code>
            {' '}или <code className="font-mono-sm">game:events:cancel</code>
          </CardContent>
        </Card>
      </>
    );
  }

  return (
    <>
      <PageHeader
        title="Game-ops events"
        description="управление wait/error events и dead-letter очередью (game-nova)"
      />

      <div className="mb-3 inline-flex rounded-md border bg-card p-0.5">
        <TabButton active={tab === 'active'} onClick={() => setTab('active')}>
          <Activity className="h-3.5 w-3.5" aria-hidden="true" />
          Active
        </TabButton>
        <TabButton active={tab === 'dead'} onClick={() => setTab('dead')}>
          <Skull className="h-3.5 w-3.5" aria-hidden="true" />
          Dead-letter
        </TabButton>
      </div>

      {tab === 'active' ? (
        <ActiveEventsTable
          offset={offset}
          {...(stateFilter ? { stateFilter } : {})}
          kindFilter={kindFilter}
          setFilter={setFilter}
          setOffset={setOffset}
          canRetry={canRetry}
          canCancel={canCancel}
        />
      ) : (
        <DeadEventsTable
          offset={offset}
          kindFilter={kindFilter}
          setFilter={setFilter}
          setOffset={setOffset}
          canRetry={canRetry}
        />
      )}
    </>
  );
}

interface ActiveProps {
  offset: number;
  stateFilter?: EventState;
  kindFilter: string;
  setFilter: (key: string, value: string | null) => void;
  setOffset: (n: number) => void;
  canRetry: boolean;
  canCancel: boolean;
}

function ActiveEventsTable({
  offset,
  stateFilter,
  kindFilter,
  setFilter,
  setOffset,
  canRetry,
  canCancel,
}: ActiveProps): React.ReactElement {
  const queryClient = useQueryClient();
  const filters = useMemo(() => {
    const f: { state?: EventState; kind?: number; limit: number; offset: number } = {
      limit: PAGE_SIZE,
      offset,
    };
    if (stateFilter) f.state = stateFilter;
    const k = parseInt(kindFilter, 10);
    if (!Number.isNaN(k)) f.kind = k;
    return f;
  }, [offset, stateFilter, kindFilter]);

  const q = useQuery({
    queryKey: ['game', 'events', 'active', filters],
    queryFn: () => eventsApi.listEvents(filters),
  });

  const retryMut = useMutation({
    mutationFn: (id: string) => eventsApi.retry(id),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['game', 'events'] }),
  });
  const cancelMut = useMutation({
    mutationFn: (id: string) => eventsApi.cancel(id),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['game', 'events'] }),
  });

  return (
    <>
      <Filters
        {...(stateFilter ? { stateFilter } : {})}
        kindFilter={kindFilter}
        setFilter={setFilter}
        showState
      />
      <ActionStatus retry={retryMut} cancel={cancelMut} />

      <div className="rounded-md border bg-card overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-muted/50 text-2xs uppercase tracking-wide text-muted-foreground">
            <tr className="text-left">
              <th className="px-3 py-2">id</th>
              <th className="px-3 py-2 w-16">kind</th>
              <th className="px-3 py-2 w-20">state</th>
              <th className="px-3 py-2">fire_at</th>
              <th className="px-3 py-2 w-12">try</th>
              <th className="px-3 py-2">last_error</th>
              {(canRetry || canCancel) && <th className="px-3 py-2 w-32" />}
            </tr>
          </thead>
          <tbody>
            {q.isLoading && <SkeletonRows cols={canRetry || canCancel ? 7 : 6} />}
            {q.error && <ErrorRow err={q.error} colSpan={7} />}
            {q.data?.length === 0 && <EmptyRow colSpan={7} />}
            {q.data?.map((row) => (
              <ActiveRow
                key={row.id}
                row={row}
                canRetry={canRetry}
                canCancel={canCancel}
                onRetry={() => retryMut.mutate(row.id)}
                onCancel={() => cancelMut.mutate(row.id)}
                pending={retryMut.isPending || cancelMut.isPending}
              />
            ))}
          </tbody>
        </table>
      </div>
      <Pagination
        offset={offset}
        count={q.data?.length ?? 0}
        size={PAGE_SIZE}
        setOffset={setOffset}
      />
    </>
  );
}

interface DeadProps {
  offset: number;
  kindFilter: string;
  setFilter: (key: string, value: string | null) => void;
  setOffset: (n: number) => void;
  canRetry: boolean;
}

function DeadEventsTable({
  offset,
  kindFilter,
  setFilter,
  setOffset,
  canRetry,
}: DeadProps): React.ReactElement {
  const queryClient = useQueryClient();
  const filters = useMemo(() => {
    const f: { kind?: number; limit: number; offset: number } = {
      limit: PAGE_SIZE,
      offset,
    };
    const k = parseInt(kindFilter, 10);
    if (!Number.isNaN(k)) f.kind = k;
    return f;
  }, [offset, kindFilter]);

  const q = useQuery({
    queryKey: ['game', 'events', 'dead', filters],
    queryFn: () => eventsApi.listDead(filters),
  });

  const resurrectMut = useMutation({
    mutationFn: (id: string) => eventsApi.resurrect(id),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ['game', 'events'] }),
  });

  return (
    <>
      <Filters kindFilter={kindFilter} setFilter={setFilter} />
      <ActionStatus resurrect={resurrectMut} />

      <div className="rounded-md border bg-card overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-muted/50 text-2xs uppercase tracking-wide text-muted-foreground">
            <tr className="text-left">
              <th className="px-3 py-2">id</th>
              <th className="px-3 py-2 w-16">kind</th>
              <th className="px-3 py-2">failed_at</th>
              <th className="px-3 py-2 w-12">try</th>
              <th className="px-3 py-2">last_error</th>
              {canRetry && <th className="px-3 py-2 w-28" />}
            </tr>
          </thead>
          <tbody>
            {q.isLoading && <SkeletonRows cols={canRetry ? 6 : 5} />}
            {q.error && <ErrorRow err={q.error} colSpan={6} />}
            {q.data?.length === 0 && <EmptyRow colSpan={6} />}
            {q.data?.map((row) => (
              <DeadRow
                key={row.id}
                row={row}
                canRetry={canRetry}
                onResurrect={() => resurrectMut.mutate(row.id)}
                pending={resurrectMut.isPending}
              />
            ))}
          </tbody>
        </table>
      </div>
      <Pagination
        offset={offset}
        count={q.data?.length ?? 0}
        size={PAGE_SIZE}
        setOffset={setOffset}
      />
    </>
  );
}

// === sub-components ===

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}): React.ReactElement {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`flex items-center gap-1.5 rounded px-3 py-1 text-xs ${
        active ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:text-foreground'
      }`}
    >
      {children}
    </button>
  );
}

interface FiltersProps {
  stateFilter?: EventState;
  kindFilter: string;
  setFilter: (key: string, value: string | null) => void;
  showState?: boolean;
}

function Filters({
  stateFilter,
  kindFilter,
  setFilter,
  showState,
}: FiltersProps): React.ReactElement {
  return (
    <Card className="mb-3">
      <CardContent className="pt-4">
        <div className="flex flex-wrap gap-3">
          {showState && (
            <div className="space-y-1">
              <label className="text-2xs uppercase tracking-wide text-muted-foreground">
                state
              </label>
              <select
                value={stateFilter ?? ''}
                onChange={(e) => setFilter('state', e.target.value || null)}
                className="flex h-8 w-32 rounded-md border border-input bg-background px-2 text-sm shadow-sm"
              >
                <option value="">— любое —</option>
                <option value="wait">wait</option>
                <option value="error">error</option>
                <option value="ok">ok</option>
              </select>
            </div>
          )}
          <div className="space-y-1">
            <label className="text-2xs uppercase tracking-wide text-muted-foreground">
              kind (int)
            </label>
            <Input
              value={kindFilter}
              onChange={(e) => setFilter('kind', e.target.value || null)}
              placeholder="напр. 12"
              className="w-32 font-mono-sm"
              inputMode="numeric"
            />
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

function ActionStatus({
  retry,
  cancel,
  resurrect,
}: {
  retry?: { error: unknown };
  cancel?: { error: unknown };
  resurrect?: { error: unknown };
}): React.ReactElement | null {
  const errs = [retry?.error, cancel?.error, resurrect?.error].filter(Boolean);
  if (errs.length === 0) return null;
  const first = errs[0];
  let msg = 'Ошибка';
  if (first instanceof ApiError) {
    msg = first.body?.message || first.body?.error || `HTTP ${first.status}`;
  } else if (first instanceof Error) {
    msg = first.message;
  }
  return (
    <div
      role="alert"
      className="mb-3 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive"
    >
      Ошибка операции: {msg}
    </div>
  );
}

interface ActiveRowProps {
  row: EventRow;
  canRetry: boolean;
  canCancel: boolean;
  onRetry: () => void;
  onCancel: () => void;
  pending: boolean;
}

function ActiveRow({
  row,
  canRetry,
  canCancel,
  onRetry,
  onCancel,
  pending,
}: ActiveRowProps): React.ReactElement {
  return (
    <tr className="border-t hover:bg-accent/30">
      <td className="px-3 py-2 font-mono-sm text-2xs">{shortId(row.id)}</td>
      <td className="px-3 py-2 font-mono-sm">{row.kind}</td>
      <td className="px-3 py-2">
        <Badge variant={stateBadge(row.state)}>{row.state}</Badge>
      </td>
      <td className="px-3 py-2 font-mono-sm text-xs whitespace-nowrap">
        {formatDateTime(row.fire_at)}
      </td>
      <td className="px-3 py-2 font-mono-sm text-xs">{row.attempt}</td>
      <td className="px-3 py-2 text-xs text-muted-foreground truncate max-w-md">
        {row.last_error || '—'}
      </td>
      {(canRetry || canCancel) && (
        <td className="px-3 py-2">
          <div className="flex gap-1">
            {canRetry && row.state === 'error' && (
              <Button size="sm" variant="outline" disabled={pending} onClick={onRetry}>
                <RotateCcw className="h-3 w-3" aria-hidden="true" />
                Retry
              </Button>
            )}
            {canCancel && row.state !== 'ok' && (
              <Button size="sm" variant="ghost" disabled={pending} onClick={onCancel}>
                <X className="h-3 w-3" aria-hidden="true" />
                Cancel
              </Button>
            )}
          </div>
        </td>
      )}
    </tr>
  );
}

interface DeadRowProps {
  row: DeadEvent;
  canRetry: boolean;
  onResurrect: () => void;
  pending: boolean;
}

function DeadRow({ row, canRetry, onResurrect, pending }: DeadRowProps): React.ReactElement {
  return (
    <tr className="border-t hover:bg-accent/30">
      <td className="px-3 py-2 font-mono-sm text-2xs">{shortId(row.id)}</td>
      <td className="px-3 py-2 font-mono-sm">{row.kind}</td>
      <td className="px-3 py-2 font-mono-sm text-xs whitespace-nowrap">
        {formatDateTime(row.failed_at)}
      </td>
      <td className="px-3 py-2 font-mono-sm text-xs">{row.attempt}</td>
      <td className="px-3 py-2 text-xs text-muted-foreground truncate max-w-md">
        {row.last_error || '—'}
      </td>
      {canRetry && (
        <td className="px-3 py-2">
          <Button size="sm" variant="outline" disabled={pending} onClick={onResurrect}>
            <RefreshCw className="h-3 w-3" aria-hidden="true" />
            Resurrect
          </Button>
        </td>
      )}
    </tr>
  );
}

function SkeletonRows({ cols }: { cols: number }): React.ReactElement {
  return (
    <>
      {[0, 1, 2].map((i) => (
        <tr key={i} className="border-t">
          {Array.from({ length: cols }).map((_, j) => (
            <td key={j} className="px-3 py-2">
              <Skeleton className="h-4 w-full" />
            </td>
          ))}
        </tr>
      ))}
    </>
  );
}

function ErrorRow({ err, colSpan }: { err: unknown; colSpan: number }): React.ReactElement {
  let msg = 'Ошибка';
  if (err instanceof ApiError) {
    if (err.status === 403) msg = 'Нет permission';
    else msg = err.body?.message || err.body?.error || `HTTP ${err.status}`;
  } else if (err instanceof Error) msg = err.message;
  return (
    <tr>
      <td colSpan={colSpan} className="px-3 py-6 text-center text-xs text-destructive">
        {msg}
      </td>
    </tr>
  );
}

function EmptyRow({ colSpan }: { colSpan: number }): React.ReactElement {
  return (
    <tr>
      <td colSpan={colSpan} className="px-3 py-6 text-center text-xs text-muted-foreground">
        — нет данных —
      </td>
    </tr>
  );
}

function Pagination({
  offset,
  count,
  size,
  setOffset,
}: {
  offset: number;
  count: number;
  size: number;
  setOffset: (n: number) => void;
}): React.ReactElement {
  const hasMore = count === size;
  return (
    <div className="mt-3 flex items-center justify-between text-xs text-muted-foreground">
      <span>{count === 0 ? 'нет данных' : `показано ${offset + 1}–${offset + count}`}</span>
      <div className="flex gap-2">
        <Button
          variant="outline"
          size="sm"
          disabled={offset === 0}
          onClick={() => setOffset(Math.max(0, offset - size))}
        >
          ← Назад
        </Button>
        <Button variant="outline" size="sm" disabled={!hasMore} onClick={() => setOffset(offset + size)}>
          Вперёд →
        </Button>
      </div>
    </div>
  );
}

function shortId(id: string): string {
  const last = id.split('-').pop() ?? id;
  return last.slice(0, 8);
}

function stateBadge(s: EventState): 'success' | 'destructive' | 'secondary' {
  if (s === 'ok') return 'success';
  if (s === 'error') return 'destructive';
  return 'secondary';
}

