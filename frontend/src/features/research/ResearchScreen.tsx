import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { RESEARCH, imageOf, costForLevel } from '@/api/catalog';
import type { Planet, QueueItem, ResearchState } from '@/api/types';
import { Countdown } from '@/ui/Countdown';
import { ProgressBar } from '@/ui/ProgressBar';
import { useToast } from '@/ui/Toast';

export function ResearchScreen({ planet }: { planet: Planet }) {
  const qc = useQueryClient();
  const toast = useToast();

  const state = useQuery({
    queryKey: ['research'],
    queryFn: () => api.get<ResearchState>('/api/research'),
    refetchInterval: 2000,
  });

  const enqueue = useMutation({
    mutationFn: (unitId: number) =>
      api.post<QueueItem>(`/api/planets/${planet.id}/research`, { unit_id: unitId }),
    onSuccess: (_, unitId) => {
      void qc.invalidateQueries({ queryKey: ['research'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      const name = RESEARCH.find((r) => r.id === unitId)?.name ?? `#${unitId}`;
      toast.show('success', 'Исследование', `${name} запущено`);
    },
    onError: (err) => {
      toast.show('danger', 'Ошибка', err instanceof Error ? err.message : 'Не удалось запустить');
    },
  });

  const levels = state.data?.levels ?? {};
  const queue = (state.data?.queue ?? []).filter((i) => new Date(i.end_at).getTime() > Date.now());
  const active = queue[0];
  const isBusy = !!active;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          Исследования
        </h2>
      </div>

      {/* Active research */}
      {active ? (
        <ActiveResearchBanner item={active} />
      ) : (
        <div style={{ fontSize: 13, color: 'var(--ox-fg-dim)', padding: '8px 0' }}>
          🔬 Лаборатория свободна — выберите технологию для исследования
        </div>
      )}

      <div className="ox-cards-grid">
        {RESEARCH.map((r) => {
          const level = levels[r.id.toString()] ?? 0;
          const isActive = active?.unit_id === r.id;
          const nextCost = costForLevel(r.costBase, r.costFactor, level + 1);
          const canAfford =
            planet.metal    >= nextCost.metal &&
            planet.silicon  >= nextCost.silicon &&
            planet.hydrogen >= nextCost.hydrogen;
          return (
            <div key={r.id} className="ox-unit-card" style={isActive ? { borderColor: 'var(--ox-accent)', boxShadow: '0 0 0 1px var(--ox-accent)' } : undefined}>
              <div className="ox-unit-card-img">
                <img src={imageOf(r.key)} alt={r.name} width={64} height={64} style={{ imageRendering: 'pixelated' }} />
              </div>
              <div className="ox-unit-card-body">
                <div className="ox-unit-card-name">{r.name}</div>
                <div style={{ fontSize: 12, color: level > 0 ? 'var(--ox-fg-dim)' : 'var(--ox-fg-muted)', marginBottom: 4 }}>
                  {level > 0 ? `Уровень ${level}` : 'Не изучено'}
                </div>
                {!isActive && (
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
                )}
              </div>
              <div className="ox-unit-card-footer">
                <button
                  type="button"
                  className={`btn${isBusy || isActive || !canAfford ? ' btn-ghost' : ''} btn-sm`}
                  style={{ width: '100%' }}
                  disabled={enqueue.isPending || isBusy}
                  onClick={() => enqueue.mutate(r.id)}
                >
                  {isActive ? '🔬 Изучается' : isBusy ? '⏳ Занято' : level === 0 ? 'Изучить' : `→ ур. ${level + 1}`}
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

function ActiveResearchBanner({ item }: { item: QueueItem }) {
  const total = new Date(item.end_at).getTime() - new Date(item.start_at).getTime();
  const elapsed = Date.now() - new Date(item.start_at).getTime();
  const pct = total > 0 ? Math.min(100, (elapsed / total) * 100) : 100;
  const name = RESEARCH.find((r) => r.id === item.unit_id)?.name ?? `#${item.unit_id}`;

  return (
    <div className="ox-panel" style={{ padding: '14px 16px', display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 13 }}>
        <span style={{ fontSize: 20 }}>🔬</span>
        <span style={{ flex: 1, fontWeight: 600 }}>{name} → ур. {item.target_level}</span>
        <Countdown finishAt={item.end_at} />
      </div>
      <ProgressBar pct={pct} height={5} variant="default" />
    </div>
  );
}
