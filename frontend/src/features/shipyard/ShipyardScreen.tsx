import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { SHIPS, DEFENSE, nameOf, imageOf } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';
import type { Inventory, Planet, ShipyardQueueItem } from '@/api/types';

export function ShipyardScreen({ planet }: { planet: Planet }) {
  const { t, tf } = useTranslation();
  const qc = useQueryClient();
  const queue = useQuery({
    queryKey: ['shipyard-queue', planet.id],
    queryFn: () => api.get<{ queue: ShipyardQueueItem[] }>(`/api/planets/${planet.id}/shipyard/queue`),
    refetchInterval: 2000,
  });
  const inventory = useQuery({
    queryKey: ['shipyard-inventory', planet.id],
    queryFn: () => api.get<Inventory>(`/api/planets/${planet.id}/shipyard/inventory`),
  });

  const enqueue = useMutation({
    mutationFn: (p: { unitId: number; count: number }) =>
      api.post<ShipyardQueueItem>(`/api/planets/${planet.id}/shipyard`, {
        unit_id: p.unitId,
        count: p.count,
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['shipyard-queue', planet.id] });
      void qc.invalidateQueries({ queryKey: ['shipyard-inventory', planet.id] });
      void qc.invalidateQueries({ queryKey: ['planets'] });
    },
  });

  return (
    <section>
      <h2>
        {t('global', 'MENU_SHIPYARD')} — {planet.name}
      </h2>

      <UnitList
        title={tf('Main', 'UNITS_SHIPS', 'Корабли')}
        units={SHIPS}
        stock={inventory.data?.ships}
        onBuild={(id, n) => enqueue.mutate({ unitId: id, count: n })}
        pending={enqueue.isPending}
      />
      <UnitList
        title={tf('Main', 'UNITS_DEFENSE', 'Оборона')}
        units={DEFENSE}
        stock={inventory.data?.defense}
        onBuild={(id, n) => enqueue.mutate({ unitId: id, count: n })}
        pending={enqueue.isPending}
      />

      {enqueue.isError && (
        <div className="ox-error">
          {enqueue.error instanceof Error ? enqueue.error.message : t('global', 'ERROR')}
        </div>
      )}

      <h3>{tf('Main', 'SHIPYARD_QUEUE', 'Очередь верфи')}</h3>
      {queue.data && (queue.data.queue ?? []).length > 0 ? (
        <ul>
          {(queue.data.queue ?? []).map((q) => (
            <li key={q.id}>
              {nameOf(q.unit_id)} × {q.count}, {tf('Main', 'UNTIL', 'до')}{' '}
              {new Date(q.end_at).toLocaleTimeString('ru-RU')}
            </li>
          ))}
        </ul>
      ) : (
        <p>{tf('Main', 'QUEUE_EMPTY', 'Очередь пуста.')}</p>
      )}
    </section>
  );
}

function UnitList({
  title,
  units,
  stock,
  onBuild,
  pending,
}: {
  title: string;
  units: { id: number; key: string; name: string }[];
  stock: Record<string, number> | undefined;
  onBuild: (unitId: number, count: number) => void;
  pending: boolean;
}) {
  const { tf } = useTranslation();
  const [drafts, setDrafts] = useState<Record<number, number>>({});
  return (
    <>
      <h3>{title}</h3>
      <table className="ox-table">
        <thead>
          <tr>
            <th>{tf('Main', 'UNIT', 'Юнит')}</th>
            <th>{tf('Main', 'IN_STOCK', 'В наличии')}</th>
            <th>{tf('Main', 'COUNT', 'Количество')}</th>
            <th>{tf('Main', 'ACTION', 'Действие')}</th>
          </tr>
        </thead>
        <tbody>
          {units.map((u) => (
            <tr key={u.id}>
              <td style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <img src={imageOf(u.key)} alt="" width={40} height={40} style={{ imageRendering: 'pixelated' }} />
                {u.name}
              </td>
              <td className="num">{stock?.[u.id.toString()] ?? 0}</td>
              <td>
                <input
                  type="number"
                  min={1}
                  value={drafts[u.id] ?? 1}
                  onChange={(e) =>
                    setDrafts({ ...drafts, [u.id]: Math.max(1, Number(e.target.value)) })
                  }
                  style={{ width: 80 }}
                />
              </td>
              <td>
                <button
                  type="button"
                  disabled={pending}
                  onClick={() => onBuild(u.id, drafts[u.id] ?? 1)}
                >
                  {tf('Main', 'BUILD_BUTTON', 'Построить')}
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  );
}
