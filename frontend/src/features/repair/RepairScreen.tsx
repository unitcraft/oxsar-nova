import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { SHIPS, DEFENSE, nameOf, imageOf, fmtReqs } from '@/api/catalog';
import type { CombatEntry } from '@/api/catalog';
import type { Inventory, Planet } from '@/api/types';
import { Countdown } from '@/ui/Countdown';
import { ProgressBar } from '@/ui/ProgressBar';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

interface RepairQueueItem {
  id: string;
  planet_id: string;
  user_id: string;
  unit_id: number;
  is_defense: boolean;
  mode: 'disassemble' | 'repair';
  count: number;
  return_metal: number;
  return_silicon: number;
  return_hydrogen: number;
  per_unit_seconds: number;
  start_at: string;
  end_at: string;
  status: string;
}

interface RepairStorage {
  total: number;
  used: number;
  free: number;
}

interface DamagedUnit {
  unit_id: number;
  count: number;
  damaged: number;
  shell_percent: number;
}

export function RepairScreen({ planet }: { planet: Planet }) {
  const { t } = useTranslation('repairUi');
  const qc = useQueryClient();
  const toast = useToast();
  const [tab, setTab] = useState<'repair' | 'disassemble'>('repair');

  const inventory = useQuery({
    queryKey: ['shipyard-inventory', planet.id],
    queryFn: () => api.get<Inventory>(`/api/planets/${planet.id}/shipyard/inventory`),
  });
  const queue = useQuery({
    queryKey: ['repair-queue', planet.id],
    queryFn: () => api.get<{ queue: RepairQueueItem[] | null; storage: RepairStorage }>(`/api/planets/${planet.id}/repair/queue`),
    refetchInterval: 2000,
  });
  const damaged = useQuery({
    queryKey: ['repair-damaged', planet.id],
    queryFn: () => api.get<{ damaged: DamagedUnit[] | null }>(`/api/planets/${planet.id}/repair/damaged`),
    refetchInterval: 5000,
  });

  const repair = useMutation({
    mutationFn: (unitId: number) =>
      api.post<RepairQueueItem>(`/api/planets/${planet.id}/repair/repair`, { unit_id: unitId }),
    onSuccess: (_, unitId) => {
      void qc.invalidateQueries({ queryKey: ['repair-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['repair-damaged', planet.id] });
      void qc.invalidateQueries({ queryKey: ['shipyard-inventory', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', t('toastRepairTitle'), t('toastRepaired', { name: nameOf(unitId) }));
    },
    onError: (err) => { toast.show('danger', t('toastError'), err instanceof Error ? err.message : t('toastRepairErr')); },
  });

  const disassemble = useMutation({
    mutationFn: (p: { unitId: number; count: number }) =>
      api.post<RepairQueueItem>(`/api/planets/${planet.id}/repair/disassemble`, { unit_id: p.unitId, count: p.count }),
    onSuccess: (_, { unitId, count }) => {
      void qc.invalidateQueries({ queryKey: ['repair-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['shipyard-inventory', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', t('toastDisassTitle'), t('toastDisassembled', { name: nameOf(unitId), count: String(count) }));
    },
    onError: (err) => { toast.show('danger', t('toastError'), err instanceof Error ? err.message : t('toastDisassErr')); },
  });

  const cancel = useMutation({
    mutationFn: (queueId: string) =>
      api.delete(`/api/planets/${planet.id}/repair/queue/${queueId}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['repair-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['repair-damaged', planet.id] });
      void qc.invalidateQueries({ queryKey: ['shipyard-inventory', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', t('toastCancelledTitle'), t('toastCancelled'));
    },
    onError: (err) => { toast.show('danger', t('toastError'), err instanceof Error ? err.message : t('toastCancelErr')); },
  });

  const list = (queue.data?.queue ?? []).filter((i) => new Date(i.end_at).getTime() > Date.now());
  const damagedList = damaged.data?.damaged ?? [];
  const storage = queue.data?.storage ?? { total: 0, used: 0, free: 0 };

  const storagePct = storage.total > 0 ? Math.min(100, (storage.used / storage.total) * 100) : 0;
  const storageVariant = storagePct > 80 ? 'danger' : storagePct > 50 ? 'warning' : 'success';

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', gap: 8, flexWrap: 'wrap' }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          {t('title', { planetName: planet.name })}
        </h2>
        {storage.total > 0 && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 3, minWidth: 160 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 13, color: 'var(--ox-fg-muted)' }}>
              <span>{t('storage')}</span>
              <span style={{ fontFamily: 'var(--ox-mono)' }}>{storage.used} / {storage.total}</span>
            </div>
            <ProgressBar pct={storagePct} variant={storageVariant} height={6} />
          </div>
        )}
      </div>

      {list.length > 0 && (
        <div className="ox-panel" style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ fontSize: 13, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 2 }}>
            {t('queueTitle')}
          </div>
          {list.map((q, i) => {
            const total = new Date(q.end_at).getTime() - new Date(q.start_at).getTime();
            const elapsed = Date.now() - new Date(q.start_at).getTime();
            const pct = total > 0 ? Math.min(100, (elapsed / total) * 100) : 100;
            const icon = q.mode === 'repair' ? '🔧' : '♻️';
            return (
              <div key={q.id} style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 15 }}>
                  <span>{i === 0 ? icon : '⏳'}</span>
                  <span style={{ flex: 1, fontWeight: i === 0 ? 600 : 400 }}>
                    {q.mode === 'repair' ? t('modeRepair') : t('modeDisassemble')}: {nameOf(q.unit_id)} × {q.count}
                  </span>
                  {i === 0 ? <Countdown finishAt={q.end_at} /> : (
                    <span style={{ fontSize: 14, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
                      {new Date(q.end_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
                    </span>
                  )}
                  <button
                    type="button"
                    className="btn-ghost btn-sm"
                    style={{ fontSize: 13, padding: '2px 6px', color: 'var(--ox-fg-muted)' }}
                    onClick={() => cancel.mutate(q.id)}
                    title={t('cancelTitle')}
                  >
                    ✕
                  </button>
                </div>
                {i === 0 && <ProgressBar pct={pct} height={4} />}
              </div>
            );
          })}
        </div>
      )}

      <div className="ox-tabs">
        <button type="button" aria-pressed={tab === 'repair'} onClick={() => setTab('repair')}>
          {t('tabRepair', { count: String(damagedList.length) })}
        </button>
        <button type="button" aria-pressed={tab === 'disassemble'} onClick={() => setTab('disassemble')}>
          {t('tabDisassemble')}
        </button>
      </div>

      {tab === 'repair' && (
        damagedList.length === 0 ? (
          <div style={{ color: 'var(--ox-fg-dim)', fontSize: 16, padding: '8px 0' }}>
            {t('noDamaged')}
          </div>
        ) : (
          <div className="ox-cards-grid">
            {damagedList.map((d) => {
              const unitMeta = [...SHIPS, ...DEFENSE].find((u) => u.id === d.unit_id);
              const hasReqs = !!(unitMeta?.requires?.length);
              return (
                <div key={d.unit_id} className="ox-unit-card">
                  <div style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
                    {unitMeta && (
                      <img
                        src={imageOf(unitMeta.key)} alt={nameOf(d.unit_id)} width={64} height={64}
                        style={{ imageRendering: 'pixelated', flexShrink: 0, borderRadius: 6, background: 'rgba(0,0,0,0.3)', padding: 4 }}
                      />
                    )}
                    <div className="ox-unit-card-body" style={{ minWidth: 0, flex: 1, overflow: 'hidden' }}>
                      <div className="ox-unit-card-name">{nameOf(d.unit_id)}</div>
                      <div style={{ fontSize: 14, color: 'var(--ox-fg-dim)' }}>{t('inStock')} {d.count}</div>
                      <div style={{ fontSize: 14, color: 'var(--ox-danger)' }}>{t('damaged')} {d.damaged}</div>
                      <div style={{ marginTop: 4 }}>
                        <ProgressBar pct={d.shell_percent} variant={d.shell_percent < 40 ? 'danger' : 'warning'} height={4} showLabel />
                      </div>
                      {hasReqs && unitMeta?.requires && (
                        <div style={{ fontSize: 10, color: 'var(--ox-fg-muted)', marginTop: 4, fontFamily: 'var(--ox-mono)' }}>
                          🔒 {fmtReqs(unitMeta.requires)}
                        </div>
                      )}
                    </div>
                  </div>
                  <div className="ox-unit-card-footer">
                    <button
                      type="button"
                      className={`btn${hasReqs ? ' btn-danger' : ''} btn-sm`}
                      style={{ width: '100%' }}
                      disabled={repair.isPending || hasReqs}
                      onClick={() => repair.mutate(d.unit_id)}
                    >
                      {t('repairAll')}
                    </button>
                  </div>
                </div>
              );
            })}
          </div>
        )
      )}

      {tab === 'disassemble' && (
        <>
          <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', padding: '4px 0' }}>
            {t('disassembleHint')}
          </div>
          <DisassembleList
            units={[...SHIPS, ...DEFENSE]}
            ships={inventory.data?.ships}
            defense={inventory.data?.defense}
            onGo={(id, n) => disassemble.mutate({ unitId: id, count: n })}
            pending={disassemble.isPending}
          />
        </>
      )}
    </div>
  );
}

function DisassembleList({
  units, ships, defense, onGo, pending,
}: {
  units: CombatEntry[];
  ships: Record<string, number> | undefined;
  defense: Record<string, number> | undefined;
  onGo: (unitId: number, count: number) => void;
  pending: boolean;
}) {
  const { t } = useTranslation('repairUi');
  const [drafts, setDrafts] = useState<Record<number, number>>({});
  const isShip = (id: number) => SHIPS.some((s) => s.id === id);

  return (
    <div className="ox-cards-grid">
      {units.map((u) => {
        const stock = isShip(u.id) ? ships : defense;
        const have = stock?.[u.id.toString()] ?? 0;
        const draft = drafts[u.id] ?? 0;
        if (have === 0) return null;
        const refund = u.cost ? {
          metal:    Math.floor(u.cost.metal    * draft * 0.7),
          silicon:  Math.floor(u.cost.silicon  * draft * 0.7),
          hydrogen: Math.floor(u.cost.hydrogen * draft * 0.7),
        } : null;
        return (
          <div key={u.id} className="ox-unit-card">
            <div style={{ display: 'flex', gap: 10, alignItems: 'flex-start' }}>
              <img
                src={imageOf(u.key)} alt={u.name} width={64} height={64}
                style={{ imageRendering: 'pixelated', flexShrink: 0, borderRadius: 6, background: 'rgba(0,0,0,0.3)', padding: 4 }}
              />
              <div className="ox-unit-card-body" style={{ minWidth: 0, flex: 1, overflow: 'hidden' }}>
                <div className="ox-unit-card-name">{u.name}</div>
                <div style={{ fontSize: 14, color: 'var(--ox-fg-dim)' }}>{t('inStock')} {have}</div>
                {(u.speed != null || (u.fuel != null && u.fuel > 0)) && (
                  <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)', display: 'flex', gap: 8, marginTop: 2 }}>
                    {u.speed != null && <span>🚀 {u.speed.toLocaleString('ru-RU')}</span>}
                    {u.fuel != null && u.fuel > 0 && <span>⛽ {u.fuel}{t('fuelPer')}</span>}
                  </div>
                )}
                {refund && draft > 0 && (
                  <div style={{ fontSize: 13, fontFamily: 'var(--ox-mono)', color: 'var(--ox-success)', marginTop: 2, lineHeight: 1.5 }}>
                    +{refund.metal > 0 && <span style={{ marginRight: 4 }}>🟠{refund.metal.toLocaleString('ru-RU')}</span>}
                    {refund.silicon > 0 && <span style={{ marginRight: 4 }}>💎{refund.silicon.toLocaleString('ru-RU')}</span>}
                    {refund.hydrogen > 0 && <span>💧{refund.hydrogen.toLocaleString('ru-RU')}</span>}
                  </div>
                )}
              </div>
            </div>
            <div className="ox-unit-card-footer" style={{ display: 'flex', gap: 6 }}>
              <input
                type="number" min={0} max={have} value={draft}
                onChange={(e) => setDrafts({ ...drafts, [u.id]: Math.max(0, Math.min(have, Number(e.target.value))) })}
                style={{ width: 64, flexShrink: 0 }}
              />
              <button
                type="button" className="btn btn-sm" style={{ flex: 1 }}
                disabled={pending || draft < 1}
                onClick={() => onGo(u.id, draft)}
              >
                {t('disassembleBtn')}
              </button>
            </div>
          </div>
        );
      })}
    </div>
  );
}
