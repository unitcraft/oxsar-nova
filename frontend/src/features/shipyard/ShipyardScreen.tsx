import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { SHIPS, DEFENSE, nameOf, imageOf, fmtReqs } from '@/api/catalog';
import type { CombatEntry } from '@/api/catalog';
import type { Inventory, Planet, ShipyardQueueItem } from '@/api/types';
import { Countdown } from '@/ui/Countdown';
import { ProgressBar } from '@/ui/ProgressBar';
import { useToast } from '@/ui/Toast';

export function ShipyardScreen({ planet }: { planet: Planet }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [tab, setTab] = useState<'ships' | 'defense'>('ships');
  const [showLocked, setShowLocked] = useState<boolean>(
    () => localStorage.getItem('shipyard-show-locked') === 'true'
  );

  const queue = useQuery({
    queryKey: ['shipyard-queue', planet.id],
    queryFn: () => api.get<{ queue: ShipyardQueueItem[] }>(`/api/planets/${planet.id}/shipyard/queue`),
    refetchInterval: 2000,
  });
  const inventory = useQuery({
    queryKey: ['shipyard-inventory', planet.id],
    queryFn: () => api.get<Inventory>(`/api/planets/${planet.id}/shipyard/inventory`),
    refetchInterval: 15000,
  });

  const cancel = useMutation({
    mutationFn: (queueId: string) =>
      api.delete(`/api/planets/${planet.id}/shipyard/${queueId}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['shipyard-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', 'Отменено', 'Задание отменено, ресурсы возвращены');
    },
    onError: (err) => {
      toast.show('danger', 'Ошибка', err instanceof Error ? err.message : 'Не удалось отменить');
    },
  });

  const enqueue = useMutation({
    mutationFn: (p: { unitId: number; count: number }) =>
      api.post<ShipyardQueueItem>(`/api/planets/${planet.id}/shipyard`, {
        unit_id: p.unitId,
        count: p.count,
      }),
    onSuccess: (_, { unitId, count }) => {
      void qc.invalidateQueries({ queryKey: ['shipyard-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['shipyard-inventory', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', 'В очередь', `${nameOf(unitId)} × ${count} добавлено в верфь`);
    },
    onError: (err) => {
      toast.show('danger', 'Ошибка', err instanceof Error ? err.message : 'Не удалось добавить');
    },
  });

  const queueItems = (queue.data?.queue ?? []).filter((i) => new Date(i.end_at).getTime() > Date.now());
  const ships = inventory.data?.ships ?? {};
  const defense = inventory.data?.defense ?? {};

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          Верфь — {planet.name}
        </h2>
      </div>

      {/* Queue */}
      {queueItems.length > 0 && (
        <div className="ox-panel" style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 2 }}>
            Очередь верфи
          </div>
          {queueItems.map((item, i) => (
            <ShipQueueRow key={item.id} item={item} isActive={i === 0} onCancel={() => cancel.mutate(item.id)} />
          ))}
        </div>
      )}

      {/* Tab switcher + filter */}
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
        <div className="ox-tabs" style={{ flex: 1 }}>
          <button type="button" aria-pressed={tab === 'ships'} onClick={() => setTab('ships')}>
            🛸 Корабли
          </button>
          <button type="button" aria-pressed={tab === 'defense'} onClick={() => setTab('defense')}>
            🛡 Оборона
          </button>
        </div>
        <button
          type="button"
          className="btn-ghost btn-sm"
          style={{ fontSize: 12, whiteSpace: 'nowrap' }}
          onClick={() => {
            const next = !showLocked;
            setShowLocked(next);
            localStorage.setItem('shipyard-show-locked', String(next));
          }}
        >
          {showLocked ? '✅ Все' : '🔒 Скрыть недоступные'}
        </button>
      </div>

      {/* Unit cards */}
      <UnitCards
        units={tab === 'ships' ? SHIPS : DEFENSE}
        stock={tab === 'ships' ? ships : defense}
        planet={planet}
        onBuild={(unitId, count) => enqueue.mutate({ unitId, count })}
        pending={enqueue.isPending}
        showLocked={showLocked}
      />
    </div>
  );
}

function UnitCards({
  units, stock, planet, onBuild, pending, showLocked,
}: {
  units: CombatEntry[];
  stock: Record<string, number>;
  planet: Planet;
  onBuild: (unitId: number, count: number) => void;
  pending: boolean;
  showLocked: boolean;
}) {
  const [drafts, setDrafts] = useState<Record<number, number>>({});

  const visibleUnits = showLocked ? units : units.filter((u) => !u.requires?.length);

  return (
    <div className="ox-cards-grid">
      {visibleUnits.map((u) => {
        const inStock = stock[u.id.toString()] ?? 0;
        const count = drafts[u.id] ?? 1;
        const c = u.cost;
        const totalCost = c ? {
          metal:    c.metal    * count,
          silicon:  c.silicon  * count,
          hydrogen: c.hydrogen * count,
        } : null;
        const canAfford = !totalCost || (
          planet.metal    >= totalCost.metal &&
          planet.silicon  >= totalCost.silicon &&
          planet.hydrogen >= totalCost.hydrogen
        );
        return (
          <div key={u.id} className="ox-unit-card">
            <div style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
              <img
                src={imageOf(u.key)} alt={u.name} width={64} height={64}
                style={{ imageRendering: 'pixelated', flexShrink: 0, borderRadius: 6, background: 'rgba(0,0,0,0.3)', padding: 4 }}
              />
              <div className="ox-unit-card-body" style={{ minWidth: 0, flex: 1, overflow: 'hidden' }}>
                <div className="ox-unit-card-name">{u.name}</div>
                {u.description && (
                  <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)', fontStyle: 'italic', marginTop: 2 }}>
                    {u.description}
                  </div>
                )}
                <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)', display: 'flex', gap: 8, flexWrap: 'wrap', marginTop: 4 }}>
                  <span>⚔ {u.attack.toLocaleString('ru-RU')}</span>
                  <span>🛡 {u.shield.toLocaleString('ru-RU')}</span>
                  <span>❤ {u.shell.toLocaleString('ru-RU')}</span>
                  {u.cargo != null && u.cargo > 0 && <span title="Грузоподъёмность">📦 {u.cargo.toLocaleString('ru-RU')}</span>}
                </div>
                {(u.speed != null || u.fuel != null) && (
                  <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)', display: 'flex', gap: 8, flexWrap: 'wrap', marginTop: 2 }}>
                    {u.speed != null && <span>🚀 {u.speed.toLocaleString('ru-RU')}</span>}
                    {u.fuel != null && u.fuel > 0 && <span>⛽ {u.fuel}/ед.</span>}
                  </div>
                )}
                {u.requires && u.requires.length > 0 && (
                  <div style={{ fontSize: 10, color: 'var(--ox-fg-muted)', marginTop: 4, fontFamily: 'var(--ox-mono)' }}>
                    🔒 {fmtReqs(u.requires)}
                  </div>
                )}
                {inStock > 0 && (
                  <div style={{ fontSize: 11, color: 'var(--ox-fg-dim)', marginTop: 4 }}>
                    В наличии: {inStock.toLocaleString('ru-RU')}
                  </div>
                )}
                {c && (
                  <>
                    <div style={{ fontSize: 11, fontFamily: 'var(--ox-mono)', lineHeight: 1.6, marginTop: 4 }}>
                      {c.metal > 0 && <span style={{ marginRight: 6, color: planet.metal >= c.metal * count ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>🟠{(c.metal * count).toLocaleString('ru-RU')}</span>}
                      {c.silicon > 0 && <span style={{ marginRight: 6, color: planet.silicon >= c.silicon * count ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>💎{(c.silicon * count).toLocaleString('ru-RU')}</span>}
                      {c.hydrogen > 0 && <span style={{ color: planet.hydrogen >= c.hydrogen * count ? 'var(--ox-fg-dim)' : 'var(--ox-danger)' }}>💧{(c.hydrogen * count).toLocaleString('ru-RU')}</span>}
                    </div>
                    {!canAfford && (
                      <div style={{ fontSize: 10, color: 'var(--ox-danger)', marginTop: 2, fontFamily: 'var(--ox-mono)' }}>
                        {[
                          c.metal    * count > planet.metal    && `🟠−${(c.metal    * count - planet.metal   ).toLocaleString('ru-RU')}`,
                          c.silicon  * count > planet.silicon  && `💎−${(c.silicon  * count - planet.silicon ).toLocaleString('ru-RU')}`,
                          c.hydrogen * count > planet.hydrogen && `💧−${(c.hydrogen * count - planet.hydrogen).toLocaleString('ru-RU')}`,
                        ].filter(Boolean).join(' ')}
                      </div>
                    )}
                  </>
                )}
              </div>
            </div>
            <div className="ox-unit-card-footer" style={{ display: 'flex', gap: 6 }}>
              <input
                type="number"
                min={1}
                value={count}
                onChange={(e) => setDrafts({ ...drafts, [u.id]: Math.max(1, Number(e.target.value)) })}
                style={{ width: 64, flexShrink: 0 }}
              />
              <button
                type="button"
                className={`btn${canAfford ? '' : ' btn-ghost'} btn-sm`}
                style={{ flex: 1 }}
                disabled={pending}
                onClick={() => onBuild(u.id, count)}
              >
                Строить
              </button>
            </div>
          </div>
        );
      })}
    </div>
  );
}

function ShipQueueRow({ item, isActive, onCancel }: { item: ShipyardQueueItem; isActive: boolean; onCancel: () => void }) {
  const total = new Date(item.end_at).getTime() - new Date(item.start_at).getTime();
  const elapsed = Date.now() - new Date(item.start_at).getTime();
  const pct = total > 0 ? Math.min(100, (elapsed / total) * 100) : 100;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
        <span style={{ fontSize: 16 }}>{isActive ? '🚀' : '⏳'}</span>
        <span style={{ flex: 1, fontWeight: isActive ? 600 : 400 }}>
          {nameOf(item.unit_id)} × {item.count}
        </span>
        {isActive
          ? <Countdown finishAt={item.end_at} />
          : <span style={{ fontSize: 12, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
              {new Date(item.end_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
            </span>
        }
        <button
          type="button"
          className="btn-ghost btn-sm"
          style={{ fontSize: 11, padding: '2px 6px', color: 'var(--ox-fg-muted)' }}
          onClick={onCancel}
          title="Отменить"
        >
          ✕
        </button>
      </div>
      {isActive && <ProgressBar pct={pct} height={4} />}
    </div>
  );
}
