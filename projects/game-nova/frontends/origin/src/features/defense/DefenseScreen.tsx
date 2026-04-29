// S-R02 Defense — оборона планеты (план 72.1).
//
// Pixel-perfect клон legacy Defense.class.php:
//   - Очередь верфи (элементы с is_defense=true) если есть.
//   - Таблица оборонительных юнитов с полем ввода количества + кнопка.
//
// Использует тот же endpoint верфи, что и ShipyardScreen:
//   GET  /api/planets/{id}/shipyard/queue
//   GET  /api/planets/{id}/shipyard/inventory
//   POST /api/planets/{id}/shipyard

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  buildShipyard,
  fetchShipyardInventory,
  fetchShipyardQueue,
} from '@/api/shipyard';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup, type CatalogEntry } from '@/features/common/catalog';
import { formatNumber, formatDuration, secondsUntil } from '@/lib/format';

interface BuildRowProps {
  entry: CatalogEntry;
  inStock: number;
  onBuild: (unitId: number, count: number) => void;
  isPending: boolean;
}

function BuildRow({ entry, inStock, onBuild, isPending }: BuildRowProps) {
  const [count, setCount] = useState('');
  const parsed = Math.max(0, Math.floor(Number(count) || 0));
  const [group, key] = entry.i18n.split('.') as [string, string];
  // Используем ключ i18n напрямую как label (строковый fallback)
  const label = `${group}.${key}`;
  return (
    <tr>
      <td width="1px">#{entry.id}</td>
      <td>
        {label}{' '}
        <span className="normal">({formatNumber(inStock)} шт.)</span>
      </td>
      <td width="120px" style={{ textAlign: 'center' }}>
        <input
          type="number"
          className="center"
          min={0}
          value={count}
          onChange={(e) => setCount(e.target.value)}
          aria-label={label}
          style={{ width: '70px' }}
        />
      </td>
      <td width="100px">
        <button
          type="button"
          className="button"
          onClick={() => parsed > 0 && onBuild(entry.id, parsed)}
          disabled={isPending || parsed <= 0}
        >
          Построить
        </button>
      </td>
    </tr>
  );
}

export function DefenseScreen() {
  const { planetId } = useResolvedPlanet();
  const qc = useQueryClient();

  const queueQ = useQuery({
    queryKey: planetId ? QK.shipyardQueue(planetId) : ['noop-dq'],
    queryFn: () => (planetId ? fetchShipyardQueue(planetId) : Promise.resolve([])),
    enabled: planetId !== null,
  });

  const invQ = useQuery({
    queryKey: planetId ? QK.shipyardInventory(planetId) : ['noop-di'],
    queryFn: () =>
      planetId
        ? fetchShipyardInventory(planetId)
        : Promise.resolve({ ships: {}, defense: {} }),
    enabled: planetId !== null,
  });

  const build = useMutation({
    mutationFn: ({ unitId, count }: { unitId: number; count: number }) =>
      buildShipyard(planetId!, unitId, count),
    onSuccess: () => {
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.shipyardQueue(planetId) });
        void qc.invalidateQueries({ queryKey: QK.shipyardInventory(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
    },
  });

  if (!planetId) {
    return <div className="idiv">Нет планет</div>;
  }

  const queue = (queueQ.data ?? []);
  const defense = catalogByGroup('defense');
  const inv: { ships: Record<string, number>; defense: Record<string, number> } =
    invQ.data ?? { ships: {}, defense: {} };

  function inStock(id: number): number {
    return inv.defense[String(id)] ?? 0;
  }

  return (
    <>
      {queue.length > 0 && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={4}>Очередь строительства</th>
            </tr>
          </thead>
          <tbody>
            {queue.map((task, idx) => (
              <tr key={task.id}>
                <td width="1px">{idx + 1}.</td>
                <td>#{task.unit_id} × {formatNumber(task.count)}</td>
                <td>{formatDuration(secondsUntil(task.end_at))}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={4}>Оборона</th>
          </tr>
        </thead>
        <tbody>
          {defense.length === 0 && (
            <tr>
              <td colSpan={4} className="center">Нет доступных оборонительных единиц</td>
            </tr>
          )}
          {defense.map((entry) => (
            <BuildRow
              key={entry.id}
              entry={entry}
              inStock={inStock(entry.id)}
              onBuild={(unitId, count) => build.mutate({ unitId, count })}
              isPending={build.isPending}
            />
          ))}
        </tbody>
      </table>
    </>
  );
}
