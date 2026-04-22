import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { SHIPS, DEFENSE, nameOf, imageOf } from '@/api/catalog';
import type { Inventory, Planet } from '@/api/types';
import { Countdown } from '@/ui/Countdown';
import { ProgressBar } from '@/ui/ProgressBar';
import { useToast } from '@/ui/Toast';

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

interface DamagedUnit {
  unit_id: number;
  count: number;
  damaged: number;
  shell_percent: number;
}

export function RepairScreen({ planet }: { planet: Planet }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [tab, setTab] = useState<'repair' | 'disassemble'>('repair');

  const inventory = useQuery({
    queryKey: ['shipyard-inventory', planet.id],
    queryFn: () => api.get<Inventory>(`/api/planets/${planet.id}/shipyard/inventory`),
  });
  const queue = useQuery({
    queryKey: ['repair-queue', planet.id],
    queryFn: () => api.get<{ queue: RepairQueueItem[] | null }>(`/api/planets/${planet.id}/repair/queue`),
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
      toast.show('success', 'Ремонт', `${nameOf(unitId)} отправлен на ремонт`);
    },
    onError: (err) => { toast.show('danger', 'Ошибка', err instanceof Error ? err.message : 'Ошибка ремонта'); },
  });

  const disassemble = useMutation({
    mutationFn: (p: { unitId: number; count: number }) =>
      api.post<RepairQueueItem>(`/api/planets/${planet.id}/repair/disassemble`, { unit_id: p.unitId, count: p.count }),
    onSuccess: (_, { unitId, count }) => {
      void qc.invalidateQueries({ queryKey: ['repair-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['shipyard-inventory', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
      toast.show('success', 'Разбор', `${nameOf(unitId)} × ${count} отправлено на разбор`);
    },
    onError: (err) => { toast.show('danger', 'Ошибка', err instanceof Error ? err.message : 'Ошибка разбора'); },
  });

  const list = (queue.data?.queue ?? []).filter((i) => new Date(i.end_at).getTime() > Date.now());
  const damagedList = damaged.data?.damaged ?? [];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
        Ремонтный ангар — {planet.name}
      </h2>

      {/* Queue */}
      {list.length > 0 && (
        <div className="ox-panel" style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 10 }}>
          <div style={{ fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 2 }}>
            Очередь ангара
          </div>
          {list.map((q, i) => {
            const total = new Date(q.end_at).getTime() - new Date(q.start_at).getTime();
            const elapsed = Date.now() - new Date(q.start_at).getTime();
            const pct = total > 0 ? Math.min(100, (elapsed / total) * 100) : 100;
            const icon = q.mode === 'repair' ? '🔧' : '♻️';
            return (
              <div key={q.id} style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 13 }}>
                  <span>{i === 0 ? icon : '⏳'}</span>
                  <span style={{ flex: 1, fontWeight: i === 0 ? 600 : 400 }}>
                    {q.mode === 'repair' ? 'Ремонт' : 'Разбор'}: {nameOf(q.unit_id)} × {q.count}
                  </span>
                  {i === 0 ? <Countdown finishAt={q.end_at} /> : (
                    <span style={{ fontSize: 12, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
                      {new Date(q.end_at).toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })}
                    </span>
                  )}
                </div>
                {i === 0 && <ProgressBar pct={pct} height={4} />}
              </div>
            );
          })}
        </div>
      )}

      {/* Tabs */}
      <div className="ox-tabs">
        <button type="button" aria-pressed={tab === 'repair'} onClick={() => setTab('repair')}>
          🔧 Ремонт ({damagedList.length})
        </button>
        <button type="button" aria-pressed={tab === 'disassemble'} onClick={() => setTab('disassemble')}>
          ♻️ Разбор
        </button>
      </div>

      {tab === 'repair' && (
        damagedList.length === 0 ? (
          <div style={{ color: 'var(--ox-fg-dim)', fontSize: 14, padding: '8px 0' }}>
            🔧 Нет повреждённых кораблей
          </div>
        ) : (
          <div className="ox-cards-grid">
            {damagedList.map((d) => {
              const unitMeta = [...SHIPS, ...DEFENSE].find((u) => u.id === d.unit_id);
              return (
                <div key={d.unit_id} className="ox-unit-card">
                  <div className="ox-unit-card-img">
                    {unitMeta && <img src={imageOf(unitMeta.key)} alt={nameOf(d.unit_id)} width={64} height={64} style={{ imageRendering: 'pixelated' }} />}
                  </div>
                  <div className="ox-unit-card-body">
                    <div className="ox-unit-card-name">{nameOf(d.unit_id)}</div>
                    <div style={{ fontSize: 12, color: 'var(--ox-fg-dim)' }}>В наличии: {d.count}</div>
                    <div style={{ fontSize: 12, color: 'var(--ox-danger)' }}>Повреждено: {d.damaged}</div>
                    <div style={{ marginTop: 4 }}>
                      <ProgressBar pct={d.shell_percent} variant={d.shell_percent < 40 ? 'danger' : 'warning'} height={4} showLabel />
                    </div>
                  </div>
                  <div className="ox-unit-card-footer">
                    <button
                      type="button"
                      className="btn btn-sm"
                      style={{ width: '100%' }}
                      disabled={repair.isPending}
                      onClick={() => repair.mutate(d.unit_id)}
                    >
                      Починить все
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
          <div style={{ fontSize: 13, color: 'var(--ox-fg-dim)', padding: '4px 0' }}>
            Разбор здоровых юнитов возвращает ~70% стоимости.
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
  units: { id: number; key: string; name: string }[];
  ships: Record<string, number> | undefined;
  defense: Record<string, number> | undefined;
  onGo: (unitId: number, count: number) => void;
  pending: boolean;
}) {
  const [drafts, setDrafts] = useState<Record<number, number>>({});
  const isShip = (id: number) => SHIPS.some((s) => s.id === id);

  return (
    <div className="ox-cards-grid">
      {units.map((u) => {
        const stock = isShip(u.id) ? ships : defense;
        const have = stock?.[u.id.toString()] ?? 0;
        const draft = drafts[u.id] ?? 0;
        if (have === 0) return null;
        return (
          <div key={u.id} className="ox-unit-card">
            <div className="ox-unit-card-img">
              <img src={imageOf(u.key)} alt={u.name} width={64} height={64} style={{ imageRendering: 'pixelated' }} />
            </div>
            <div className="ox-unit-card-body">
              <div className="ox-unit-card-name">{u.name}</div>
              <div style={{ fontSize: 12, color: 'var(--ox-fg-dim)' }}>В наличии: {have}</div>
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
                Разобрать
              </button>
            </div>
          </div>
        );
      })}
    </div>
  );
}
