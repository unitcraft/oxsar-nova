import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';
import { BUILDINGS, MOON_BUILDINGS, RESEARCH, imageOf, costForLevel, fmtReqs } from '@/api/catalog';
import type { Planet, QueueItem, ResearchState } from '@/api/types';
import { ProgressBar } from '@/ui/ProgressBar';
import { useToast } from '@/ui/Toast';

function fmtDuration(secs: number): string {
  if (secs < 60) return `${secs}с`;
  const m = Math.floor(secs / 60) % 60;
  const h = Math.floor(secs / 3600) % 24;
  const d = Math.floor(secs / 86400);
  if (d > 0) return `${d}д ${h}ч ${m}м`;
  if (h > 0) return `${h}ч ${m}м`;
  return `${m}м`;
}

function fmtSecs(sec: number): string {
  if (sec <= 0) return '00:00:00';
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = sec % 60;
  if (h > 0) return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

function useLiveProgress(startAt: string, endAt: string): { pct: number; secsLeft: number } {
  const calc = () => {
    const now = Date.now();
    const total = new Date(endAt).getTime() - new Date(startAt).getTime();
    const elapsed = now - new Date(startAt).getTime();
    const secsLeft = Math.max(0, Math.round((new Date(endAt).getTime() - now) / 1000));
    const pct = secsLeft === 0 ? 100 : total > 0 ? Math.min(99, (elapsed / total) * 100) : 100;
    return { pct, secsLeft };
  };
  const [state, setState] = useState(calc);
  useEffect(() => {
    const t = setInterval(() => setState(calc), 1000);
    return () => clearInterval(t);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [startAt, endAt]);
  return state;
}

export function ResearchScreen({ planet, onOpenInfo }: { planet: Planet; onOpenInfo: (id: number, level: number) => void }) {
  const { t } = useTranslation('researchUi');
  const { t: tg } = useTranslation('global');
  const levelAbbr = t('levelAbbr');
  const qc = useQueryClient();
  const toast = useToast();

  const state = useQuery({
    queryKey: ['research'],
    queryFn: () => api.get<ResearchState & { research_seconds: Record<string, number> }>('/api/research'),
    refetchInterval: 2000,
  });

  const buildingsLevelsQ = useQuery({
    queryKey: ['buildings-levels', planet.id],
    queryFn: () => api.get<{ levels: Record<string, number> }>(`/api/planets/${planet.id}/buildings`),
    staleTime: 10_000,
  });

  const enqueue = useMutation({
    mutationFn: (unitId: number) =>
      api.post<QueueItem>(`/api/planets/${planet.id}/research`, { unit_id: unitId }),
    onSuccess: (_, unitId) => {
      void qc.invalidateQueries({ queryKey: ['research'] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      const name = RESEARCH.find((r) => r.id === unitId)?.name ?? `#${unitId}`;
      toast.show('success', t('started'), t('startedBody', { name }));
    },
    onError: (err) => {
      const msg = err instanceof Error ? err.message : '';
      const text = msg.includes('queue busy') ? t('errQueueBusy')
        : msg.includes('not enough') ? t('errNotEnough')
        : msg || t('errDefault');
      toast.show('danger', tg('error'), text);
    },
  });

  const levels = state.data?.levels ?? {};
  const resSeconds = state.data?.research_seconds ?? {};
  const buildingLevels = buildingsLevelsQ.data?.levels ?? {};
  const queue = (state.data?.queue ?? []).filter((i) => new Date(i.end_at).getTime() > Date.now());
  const active = queue[0];
  const isBusy = !!active;

  const allBuildings = [...BUILDINGS, ...MOON_BUILDINGS];
  const isReqMet = (req: { kind: 'building' | 'research'; key: string; level: number }) => {
    if (req.kind === 'building') {
      const b = allBuildings.find((x) => x.key === req.key);
      if (!b) return true;
      return (buildingLevels[b.id.toString()] ?? 0) >= req.level;
    }
    const r = RESEARCH.find((x) => x.key === req.key);
    if (!r) return true;
    return (levels[r.id.toString()] ?? 0) >= req.level;
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          {t('title')}
        </h2>
        {planet.research_factor != null && planet.research_factor > 1 && (
          <span style={{ fontSize: 14, color: 'var(--ox-success)', fontFamily: 'var(--ox-mono)' }}>
            🔬 +{Math.round((planet.research_factor - 1) * 100)}% {t('researchBonus')}
          </span>
        )}
      </div>

      {active ? (
        <ActiveResearchBanner item={active} />
      ) : (
        <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', padding: '8px 0' }}>
          🔬 {t('labFree')}
        </div>
      )}

      <div className="ox-cards-grid">
        {RESEARCH.map((r) => {
          const level = levels[r.id.toString()] ?? 0;
          const isActive = active?.unit_id === r.id;
          const nextCost = costForLevel(r.costBase, r.costFactor, level + 1);
          const secs = resSeconds[r.id.toString()] ?? 0;
          const canAfford =
            planet.metal    >= nextCost.metal &&
            planet.silicon  >= nextCost.silicon &&
            planet.hydrogen >= nextCost.hydrogen;
          const isLocked = level === 0 && (r.requires ?? []).some((req) => !isReqMet(req));
          return (
            <div key={r.id} className="ox-unit-card" style={isActive ? { borderColor: 'var(--ox-accent)', boxShadow: '0 0 0 1px var(--ox-accent)' } : undefined}>
              <div style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
                <img
                  src={imageOf(r.key)} alt={r.name} width={64} height={64}
                  style={{ imageRendering: 'pixelated', flexShrink: 0, borderRadius: 6, background: 'rgba(0,0,0,0.3)', padding: 4, cursor: 'pointer' }}
                  onClick={() => onOpenInfo(r.id, level)}
                  title={t('details')}
                />
                <div style={{ minWidth: 0, flex: 1, overflow: 'hidden' }}>
                  <div className="ox-unit-card-name" style={{ cursor: 'pointer' }} onClick={() => onOpenInfo(r.id, level)}>{r.name}</div>
                  <div style={{ fontSize: 14, color: level > 0 ? 'var(--ox-fg-dim)' : 'var(--ox-fg-muted)', marginBottom: 2 }}>
                    {level > 0 ? t('level', { n: String(level) }) : t('notStudied')}
                  </div>
                  <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)', marginBottom: 2, fontStyle: 'italic' }}>
                    {r.benefit}
                  </div>
                  {level === 0 && r.requires && r.requires.length > 0 && (
                    <div style={{ fontSize: 10, color: 'var(--ox-fg-muted)', marginBottom: 2, fontFamily: 'var(--ox-mono)' }}>
                      🔒 {fmtReqs(r.requires)}
                    </div>
                  )}
                  {!isActive && (
                    <>
                      <div style={{ fontSize: 13, fontFamily: 'var(--ox-mono)', lineHeight: 1.6 }}>
                        {nextCost.metal > 0 && (
                          <span style={{ marginRight: 6, color: planet.metal >= nextCost.metal ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>
                            🟠{nextCost.metal.toLocaleString('ru-RU')}
                          </span>
                        )}
                        {nextCost.silicon > 0 && (
                          <span style={{ marginRight: 6, color: planet.silicon >= nextCost.silicon ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>
                            💎{nextCost.silicon.toLocaleString('ru-RU')}
                          </span>
                        )}
                        {nextCost.hydrogen > 0 && (
                          <span style={{ color: planet.hydrogen >= nextCost.hydrogen ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>
                            💧{nextCost.hydrogen.toLocaleString('ru-RU')}
                          </span>
                        )}
                      </div>
                      {!canAfford && (
                        <div style={{ fontSize: 10, color: 'var(--ox-danger)', marginTop: 2, fontFamily: 'var(--ox-mono)' }}>
                          {[
                            nextCost.metal    > planet.metal    && `🟠−${(nextCost.metal    - planet.metal   ).toLocaleString('ru-RU')}`,
                            nextCost.silicon  > planet.silicon  && `💎−${(nextCost.silicon  - planet.silicon ).toLocaleString('ru-RU')}`,
                            nextCost.hydrogen > planet.hydrogen && `💧−${(nextCost.hydrogen - planet.hydrogen).toLocaleString('ru-RU')}`,
                          ].filter(Boolean).join(' ')}
                        </div>
                      )}
                      {secs > 0 && (
                        <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)', marginTop: 2 }}>
                          ⏱ {fmtDuration(secs)}
                        </div>
                      )}
                    </>
                  )}
                </div>
              </div>
              <div className="ox-unit-card-footer">
                <button
                  type="button"
                  className={`btn${isActive || isBusy ? ' btn-ghost' : (!canAfford || isLocked) ? ' btn-danger' : ''} btn-sm`}
                  style={{ width: '100%' }}
                  disabled={enqueue.isPending || isBusy || isActive || !canAfford || isLocked}
                  onClick={() => enqueue.mutate(r.id)}
                >
                  {isActive ? `🔬 ${t('inProgress')}` : isBusy ? `⏳ ${t('busy')}` : level === 0 ? t('study') : `→ ${levelAbbr} ${level + 1}`}
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
  const { t } = useTranslation('researchUi');
  const { pct, secsLeft } = useLiveProgress(item.start_at, item.end_at);
  const name = RESEARCH.find((r) => r.id === item.unit_id)?.name ?? `#${item.unit_id}`;

  return (
    <div className="ox-panel" style={{ padding: '14px 16px', display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 15 }}>
        <span style={{ fontSize: 20 }}>🔬</span>
        <span style={{ flex: 1, fontWeight: 600 }}>{name} → {t('levelAbbr')} {item.target_level}</span>
        <span className={`ox-timer${secsLeft < 60 ? ' urgent' : ''}`}>{fmtSecs(secsLeft)}</span>
      </div>
      <ProgressBar pct={pct} height={5} variant="default" />
    </div>
  );
}
