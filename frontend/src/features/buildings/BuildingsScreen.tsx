import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

function fmtDuration(secs: number): string {
  if (secs < 60) return `${secs}с`;
  const m = Math.floor(secs / 60) % 60;
  const h = Math.floor(secs / 3600) % 24;
  const d = Math.floor(secs / 86400);
  if (d > 0) return `${d}д ${h}ч ${m}м`;
  if (h > 0) return `${h}ч ${m}м`;
  return `${m}м`;
}
import { api } from '@/api/client';
import { BUILDINGS, imageOf, costForLevel } from '@/api/catalog';
import type { Planet, QueueItem } from '@/api/types';
import { Countdown } from '@/ui/Countdown';
import { ProgressBar } from '@/ui/ProgressBar';
import { useToast } from '@/ui/Toast';

export function BuildingsScreen({ planet }: { planet: Planet }) {
  const qc = useQueryClient();
  const toast = useToast();

  const queue = useQuery({
    queryKey: ['buildings-queue', planet.id],
    queryFn: () => api.get<{ queue: QueueItem[] }>(`/api/planets/${planet.id}/buildings/queue`),
    refetchInterval: 2000,
  });
  const levelsQ = useQuery({
    queryKey: ['buildings-levels', planet.id],
    queryFn: () => api.get<{ levels: Record<string, number>; build_seconds: Record<string, number> }>(`/api/planets/${planet.id}/buildings`),
    refetchInterval: 10000,
  });

  const levels = levelsQ.data?.levels ?? {};
  const buildSeconds = levelsQ.data?.build_seconds ?? {};
  const queueItems = (queue.data?.queue ?? []).filter((i) => new Date(i.end_at).getTime() > Date.now());
  const busyIds = new Set(queueItems.map((q) => q.unit_id));

  const enqueue = useMutation({
    mutationFn: (unitId: number) =>
      api.post<QueueItem>(`/api/planets/${planet.id}/buildings`, { unit_id: unitId }),
    onSuccess: (_, unitId) => {
      void qc.invalidateQueries({ queryKey: ['buildings-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      const name = BUILDINGS.find((b) => b.id === unitId)?.name ?? `#${unitId}`;
      toast.show('success', 'В очередь', `${name} добавлена в очередь строительства`);
    },
    onError: (err) => {
      toast.show('danger', 'Ошибка', err instanceof Error ? err.message : 'Не удалось добавить в очередь');
    },
  });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          Постройки — {planet.name}
        </h2>
        {queueItems.length > 0 && (
          <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>
            В очереди: {queueItems.length}
          </span>
        )}
      </div>

      {/* Active queue */}
      {queueItems.length > 0 && (
        <div className="ox-panel" style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 2 }}>
            Активная очередь
          </div>
          {queueItems.map((item, i) => (
            <QueueRow key={item.id} item={item} isActive={i === 0} />
          ))}
        </div>
      )}

      {/* Building cards */}
      <div className="ox-cards-grid">
        {BUILDINGS.map((b) => {
          const level = levels[b.id] ?? 0;
          const inQueue = busyIds.has(b.id);
          const nextCost = costForLevel(b.costBase, b.costFactor, level + 1);
          const canAfford =
            planet.metal    >= nextCost.metal &&
            planet.silicon  >= nextCost.silicon &&
            planet.hydrogen >= nextCost.hydrogen;
          const secs = buildSeconds[b.id.toString()] ?? 0;
          return (
            <div key={b.id} className="ox-unit-card">
              <div className="ox-unit-card-img">
                <img src={imageOf(b.key)} alt={b.name} width={64} height={64} style={{ imageRendering: 'pixelated' }} />
              </div>
              <div className="ox-unit-card-body">
                <div className="ox-unit-card-name">{b.name}</div>
                <div style={{ fontSize: 12, color: 'var(--ox-fg-dim)', marginBottom: 4 }}>
                  {level > 0 ? `Уровень ${level}` : 'Не построено'}
                </div>
                {!inQueue && (
                  <>
                    <div style={{ fontSize: 11, fontFamily: 'var(--ox-mono)', lineHeight: 1.6 }}>
                      {nextCost.metal > 0 && (
                        <span style={{ marginRight: 6, color: planet.metal >= nextCost.metal ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>
                          ⛏{nextCost.metal.toLocaleString('ru-RU')}
                        </span>
                      )}
                      {nextCost.silicon > 0 && (
                        <span style={{ marginRight: 6, color: planet.silicon >= nextCost.silicon ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>
                          🔷{nextCost.silicon.toLocaleString('ru-RU')}
                        </span>
                      )}
                      {nextCost.hydrogen > 0 && (
                        <span style={{ color: planet.hydrogen >= nextCost.hydrogen ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>
                          💧{nextCost.hydrogen.toLocaleString('ru-RU')}
                        </span>
                      )}
                    </div>
                    {secs > 0 && (
                      <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)', marginTop: 2 }}>
                        ⏱ {fmtDuration(secs)}
                      </div>
                    )}
                  </>
                )}
              </div>
              <div className="ox-unit-card-footer">
                <button
                  type="button"
                  className={`btn${inQueue || !canAfford ? ' btn-ghost' : ''} btn-sm`}
                  style={{ width: '100%' }}
                  disabled={enqueue.isPending || inQueue}
                  onClick={() => enqueue.mutate(b.id)}
                >
                  {inQueue ? '⏳ В очереди' : level === 0 ? 'Построить' : `→ ур. ${level + 1}`}
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function QueueRow({ item, isActive }: { item: QueueItem; isActive: boolean }) {
  const total = new Date(item.end_at).getTime() - new Date(item.start_at).getTime();
  const elapsed = Date.now() - new Date(item.start_at).getTime();
  const pct = total > 0 ? Math.min(100, (elapsed / total) * 100) : 100;
  const name = BUILDINGS.find((b) => b.id === item.unit_id)?.name ?? `#${item.unit_id}`;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
        <span style={{ fontSize: 16 }}>{isActive ? '🏗' : '⏳'}</span>
        <span style={{ flex: 1, fontWeight: isActive ? 600 : 400 }}>
          {name} → ур. {item.target_level}
        </span>
        {isActive
          ? <Countdown finishAt={item.end_at} />
          : <span style={{ fontSize: 12, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
              {new Date(item.end_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
            </span>
        }
      </div>
      {isActive && <ProgressBar pct={pct} height={4} />}
    </div>
  );
}
