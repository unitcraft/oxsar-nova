// S-004 Shipyard — верфь (план 72 Ф.2 Spring 1).
//
// Pixel-perfect зеркало legacy `shipyard.tpl`:
//   1) Очередь верфи (если есть).
//   2) Таблица кораблей с input количества + кнопка «Построить».
//   3) Таблица обороны — то же самое.
//
// Endpoints:
//   GET  /api/planets/{id}/shipyard/queue
//   GET  /api/planets/{id}/shipyard/inventory
//   POST /api/planets/{id}/shipyard

import { useState } from 'react';
import { useParams } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  buildShipyard,
  fetchShipyardInventory,
  fetchShipyardQueue,
} from '@/api/shipyard';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup, type CatalogEntry } from '@/features/common/catalog';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber, formatDuration, secondsUntil } from '@/lib/format';

interface BuildRowProps {
  entry: CatalogEntry;
  inStock: number;
  onBuild: (unitId: number, count: number) => void;
  isPending: boolean;
}

function BuildRow({ entry, inStock, onBuild, isPending }: BuildRowProps) {
  const [count, setCount] = useState('');
  const { t } = useTranslation();
  const [group, key] = entry.i18n.split('.') as [string, string];
  const parsed = Math.max(0, Math.floor(Number(count) || 0));
  return (
    <tr>
      <td width="1px">#{entry.id}</td>
      <td>
        {t(group, key)}{' '}
        <span className="normal">
          {t('shipyard', 'inStock', { count: formatNumber(inStock) })}
        </span>
      </td>
      <td width="120px" style={{ textAlign: 'center' }}>
        <input
          type="number"
          className="center"
          min={0}
          value={count}
          onChange={(e) => setCount(e.target.value)}
          aria-label={`${t(group, key)} ${t('mission', 'capicity')}`}
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
          {t('shipyard', 'build')}
        </button>
      </td>
    </tr>
  );
}

export function ShipyardScreen() {
  const { planetId: urlId } = useParams();
  const { planetId } = useResolvedPlanet(urlId);
  const { t } = useTranslation();
  const qc = useQueryClient();

  const queueQ = useQuery({
    queryKey: planetId ? QK.shipyardQueue(planetId) : ['noop-queue'],
    queryFn: () => (planetId ? fetchShipyardQueue(planetId) : Promise.resolve([])),
    enabled: planetId !== null,
  });

  const invQ = useQuery({
    queryKey: planetId ? QK.shipyardInventory(planetId) : ['noop-inv'],
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
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  const queue = queueQ.data ?? [];
  const inv: { ships: Record<string, number>; defense: Record<string, number> } =
    invQ.data ?? { ships: {}, defense: {} };
  const ships = catalogByGroup('ship');
  const defense = catalogByGroup('defense');

  function inStock(group: 'ships' | 'defense', id: number): number {
    return inv[group][String(id)] ?? 0;
  }

  return (
    <>
      {queue.length > 0 && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={4}>{t('shipyard', 'queue')}</th>
            </tr>
          </thead>
          <tbody>
            {queue.map((task, idx) => (
              <tr key={task.id}>
                <td width="1px">{idx + 1}.</td>
                <td>
                  #{task.unit_id} × {formatNumber(task.count)}
                </td>
                <td>{formatDuration(secondsUntil(task.end_at))}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={4}>{t('shipyard', 'tabShips')}</th>
          </tr>
        </thead>
        <tbody>
          {ships.map((entry) => (
            <BuildRow
              key={entry.id}
              entry={entry}
              inStock={inStock('ships', entry.id)}
              onBuild={(unitId, count) => build.mutate({ unitId, count })}
              isPending={build.isPending}
            />
          ))}
        </tbody>
      </table>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={4}>{t('shipyard', 'tabDefense')}</th>
          </tr>
        </thead>
        <tbody>
          {defense.map((entry) => (
            <BuildRow
              key={entry.id}
              entry={entry}
              inStock={inStock('defense', entry.id)}
              onBuild={(unitId, count) => build.mutate({ unitId, count })}
              isPending={build.isPending}
            />
          ))}
        </tbody>
      </table>
    </>
  );
}
