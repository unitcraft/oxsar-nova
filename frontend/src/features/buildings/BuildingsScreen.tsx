import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { ScreenSkeleton } from '@/ui/Skeleton';

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
import { BUILDINGS, MOON_BUILDINGS, imageOf, costForLevel } from '@/api/catalog';
import type { Planet, QueueItem, UnmetRequirement } from '@/api/types';

import { ProgressBar } from '@/ui/ProgressBar';
import { useToast } from '@/ui/Toast';

function fmtPerHour(v: number): string {
  const h = Math.round(v * 3600);
  if (h >= 1_000_000) return `${(h / 1_000_000).toFixed(1)}M/ч`;
  if (h >= 1_000) return `${Math.round(h / 1_000)}k/ч`;
  return `${h}/ч`;
}

type ProdField = 'metal_per_sec' | 'silicon_per_sec' | 'hydrogen_per_sec';
type EnergyField = 'energy_prod';

interface ProdStat {
  icon: string;
  label: string;
  field: ProdField | EnergyField;
  isEnergy?: boolean;
}

const PROD_STAT: Record<string, ProdStat> = {
  metal_mine:     { icon: '🟠', label: 'Добыча', field: 'metal_per_sec' },
  silicon_lab:    { icon: '💎', label: 'Добыча', field: 'silicon_per_sec' },
  hydrogen_lab:   { icon: '💧', label: 'Добыча', field: 'hydrogen_per_sec' },
  solar_plant:    { icon: '⚡', label: 'Энергия', field: 'energy_prod', isEnergy: true },
  hydrogen_plant: { icon: '⚡', label: 'Энергия', field: 'energy_prod', isEnergy: true },
};

export function BuildingsScreen({ planet, onOpenInfo }: { planet: Planet; onOpenInfo: (id: number, level: number) => void }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [showLocked, setShowLocked] = useState<boolean>(
    () => localStorage.getItem('buildings-show-locked') === 'true'
  );

  const queue = useQuery({
    queryKey: ['buildings-queue', planet.id],
    queryFn: () => api.get<{ queue: QueueItem[] }>(`/api/planets/${planet.id}/buildings/queue`),
    refetchInterval: 2000,
  });
  const levelsQ = useQuery({
    queryKey: ['buildings-levels', planet.id],
    queryFn: () => api.get<{
      levels: Record<string, number>;
      build_seconds: Record<string, number>;
      requirements_unmet: Record<string, UnmetRequirement[]>;
    }>(`/api/planets/${planet.id}/buildings`),
    refetchInterval: 10000,
  });

  const enqueue = useMutation({
    mutationFn: (unitId: number) =>
      api.post<QueueItem>(`/api/planets/${planet.id}/buildings`, { unit_id: unitId }),
    onSuccess: (_, unitId) => {
      void qc.invalidateQueries({ queryKey: ['buildings-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      const name = ([...BUILDINGS, ...MOON_BUILDINGS]).find((b) => b.id === unitId)?.name ?? `#${unitId}`;
      toast.show('success', 'В очередь', `${name} добавлена в очередь строительства`);
    },
    onError: (err) => {
      const msg = err instanceof Error ? err.message : '';
      const text = msg.includes('queue busy') ? 'Очередь занята, дождитесь завершения постройки'
        : msg.includes('not enough') ? 'Недостаточно ресурсов'
        : msg.includes('moon-only') ? 'Это здание доступно только на луне'
        : msg.includes('not available on moon') ? 'Это здание недоступно на луне'
        : msg || 'Не удалось добавить в очередь';
      toast.show('danger', 'Ошибка', text);
    },
  });

  const cancel = useMutation({
    mutationFn: (taskId: string) =>
      api.delete<void>(`/api/planets/${planet.id}/buildings/queue/${taskId}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['buildings-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('info', 'Отменено', 'Строительство отменено, ресурсы возвращены');
    },
    onError: (err) => {
      toast.show('danger', 'Ошибка', err instanceof Error ? err.message : 'Не удалось отменить');
    },
  });

  const levels = levelsQ.data?.levels ?? {};
  const buildSeconds = levelsQ.data?.build_seconds ?? {};
  const requirementsUnmet = levelsQ.data?.requirements_unmet ?? {};
  const queueItems = (queue.data?.queue ?? []).filter((i) => new Date(i.end_at).getTime() > Date.now());
  const busyIds = new Set(queueItems.map((q) => q.unit_id));
  const buildingList = planet.is_moon ? MOON_BUILDINGS : BUILDINGS;

  if (levelsQ.isLoading) {
    return <ScreenSkeleton />;
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          Постройки — {planet.name}
        </h2>
        <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
          {planet.build_factor != null && planet.build_factor > 1 && (
            <span style={{ fontSize: 12, color: 'var(--ox-success)', fontFamily: 'var(--ox-mono)' }}>
              🏗 +{Math.round((planet.build_factor - 1) * 100)}% строительство
            </span>
          )}
          {planet.produce_factor != null && planet.produce_factor > 1 && (
            <span style={{ fontSize: 12, color: 'var(--ox-success)', fontFamily: 'var(--ox-mono)' }}>
              🟠 +{Math.round((planet.produce_factor - 1) * 100)}% добыча
            </span>
          )}
          {queueItems.length > 0 && (
            <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>
              В очереди: {queueItems.length}
            </span>
          )}
          <button
            type="button"
            className="btn-ghost btn-sm"
            style={{ fontSize: 12 }}
            onClick={() => {
              const next = !showLocked;
              setShowLocked(next);
              localStorage.setItem('buildings-show-locked', String(next));
            }}
          >
            {showLocked ? '👁 Все здания' : '🔒 Только доступные'}
          </button>
        </div>
      </div>

      {/* Active queue */}
      {queueItems.length > 0 && (
        <div className="ox-panel" style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 2 }}>
            Активная очередь
          </div>
          {queueItems.map((item, i) => (
            <QueueRow key={item.id} item={item} isActive={i === 0} onCancel={() => cancel.mutate(item.id)} cancelPending={cancel.isPending} />
          ))}
        </div>
      )}

      {/* Building cards */}
      <div className="ox-cards-grid">
        {buildingList.filter((b) => {
          if (showLocked) return true;
          const lvl = levels[b.id] ?? 0;
          const unmetArr = requirementsUnmet[b.key] ?? [];
          return lvl > 0 || unmetArr.length === 0;
        }).map((b) => {
          const level = levels[b.id] ?? 0;
          const maxLevel = b.maxLevel ?? 50;
          const isMax = level >= maxLevel;
          const inQueue = busyIds.has(b.id);
          const nextCost = costForLevel(b.costBase, b.costFactor, level + 1);
          const canAfford =
            planet.metal    >= nextCost.metal &&
            planet.silicon  >= nextCost.silicon &&
            planet.hydrogen >= nextCost.hydrogen;
          const secs = buildSeconds[b.id.toString()] ?? 0;
          const unmet = requirementsUnmet[b.key] ?? [];
          const isLocked = unmet.length > 0;
          return (
            <div key={b.id} className="ox-unit-card">
              <div style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
                <img
                  src={imageOf(b.key)} alt={b.name} width={64} height={64}
                  style={{ imageRendering: 'pixelated', flexShrink: 0, borderRadius: 6, background: 'rgba(0,0,0,0.3)', padding: 4, cursor: 'pointer' }}
                  onClick={() => onOpenInfo(b.id, level)}
                  title="Подробнее"
                />
              <div className="ox-unit-card-body" style={{ minWidth: 0, flex: 1, overflow: 'hidden' }}>
                <div className="ox-unit-card-name" style={{ cursor: 'pointer' }} onClick={() => onOpenInfo(b.id, level)}>{b.name}</div>
                {b.description && (
                  <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)', marginBottom: 2, fontStyle: 'italic' }}>
                    {b.description}
                  </div>
                )}
                <div style={{ fontSize: 12, color: 'var(--ox-fg-dim)', marginBottom: 2 }}>
                  {level > 0 ? `Уровень ${level}` : 'Не построено'}
                </div>
                {isLocked && (
                  <div style={{ fontSize: 11, color: 'var(--ox-danger)', marginBottom: 4 }}>
                    {unmet.map((r) => (
                      <div key={`${r.kind}-${r.key}`}>
                        🔒 {r.key} ур.{r.required} (у вас: {r.current})
                      </div>
                    ))}
                  </div>
                )}
                {(() => {
                  const stat = PROD_STAT[b.key];
                  if (!stat || level === 0) return null;
                  const raw = planet[stat.field];
                  const display = stat.isEnergy
                    ? `${Math.round(raw as number)}`
                    : fmtPerHour(raw as number);
                  return (
                    <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)', marginBottom: 2, fontFamily: 'var(--ox-mono)' }}>
                      {stat.icon} {display}
                    </div>
                  );
                })()}
                {!inQueue && (
                  <>
                    <div style={{ fontSize: 11, fontFamily: 'var(--ox-mono)', lineHeight: 1.6 }}>
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
                      <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)', marginTop: 2 }}>
                        ⏱ {fmtDuration(secs)}
                      </div>
                    )}
                  </>
                )}
              </div>
              </div>
              <div className="ox-unit-card-footer">
                {isMax ? (
                  <div style={{ textAlign: 'center', fontSize: 12, color: 'var(--ox-fg-muted)', fontWeight: 700, padding: '4px 0' }}>MAX</div>
                ) : (
                  <button
                    type="button"
                    className={`btn${inQueue || !canAfford || isLocked ? ' btn-ghost' : ''} btn-sm`}
                    style={{ width: '100%' }}
                    disabled={enqueue.isPending || inQueue || isLocked}
                    onClick={() => enqueue.mutate(b.id)}
                  >
                    {inQueue ? '⏳ В очереди' : isLocked ? '🔒 Заблокировано' : level === 0 ? 'Построить' : `→ ур. ${level + 1}`}
                  </button>
                )}
              </div>
            </div>
          );
        })}
      </div>

    </div>
  );
}

function useBuildProgress(startAt: string, endAt: string): { pct: number; secsLeft: number } {
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

function fmtSecs(sec: number): string {
  if (sec <= 0) return '00:00:00';
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = sec % 60;
  if (h > 0) return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
  return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')}`;
}

function QueueRow({ item, isActive, onCancel, cancelPending }: { item: QueueItem; isActive: boolean; onCancel: () => void; cancelPending: boolean }) {
  const { pct, secsLeft } = useBuildProgress(item.start_at, item.end_at);
  const [confirming, setConfirming] = useState(false);
  const name = ([...BUILDINGS, ...MOON_BUILDINGS]).find((b) => b.id === item.unit_id)?.name ?? `#${item.unit_id}`;

  function handleCancelClick() {
    if (secsLeft === 0) return; // уже завершилось пока думал
    setConfirming(true);
  }

  function handleConfirm() {
    setConfirming(false);
    if (secsLeft === 0) return; // завершилось пока думал
    onCancel();
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
        <span style={{ fontSize: 16 }}>{isActive ? '🏗' : '⏳'}</span>
        <span style={{ flex: 1, fontWeight: isActive ? 600 : 400 }}>
          {name} → ур. {item.target_level}
        </span>
        {isActive
          ? <span className={`ox-timer${secsLeft < 60 ? ' urgent' : ''}`}>{fmtSecs(secsLeft)}</span>
          : <span style={{ fontSize: 12, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
              {new Date(item.end_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
            </span>
        }
        {confirming ? (
          <>
            <span style={{ fontSize: 11, color: 'var(--ox-danger)', flexShrink: 0 }}>Отменить?</span>
            <button
              type="button"
              className="btn-sm"
              style={{ fontSize: 11, padding: '2px 8px', flexShrink: 0, background: 'var(--ox-danger)', color: '#fff', border: 'none', borderRadius: 4 }}
              disabled={cancelPending}
              onClick={handleConfirm}
            >
              Да
            </button>
            <button
              type="button"
              className="btn-ghost btn-sm"
              style={{ fontSize: 11, padding: '2px 8px', flexShrink: 0 }}
              onClick={() => setConfirming(false)}
            >
              Нет
            </button>
          </>
        ) : (
          <button
            type="button"
            className="btn-ghost btn-sm"
            disabled={cancelPending || secsLeft === 0}
            onClick={handleCancelClick}
            title="Отменить (ресурсы вернутся)"
            style={{ fontSize: 11, padding: '2px 8px', flexShrink: 0 }}
          >
            ✕
          </button>
        )}
      </div>
      {isActive && <ProgressBar pct={pct} height={4} />}
    </div>
  );
}
