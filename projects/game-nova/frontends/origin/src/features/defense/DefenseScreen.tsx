// S-R02 Defense — оборона (план 72.1 ч.19).
// Pixel-perfect клон legacy shipyard.tpl (раздел оборона).

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
import { useTranslation } from '@/i18n/i18n';
import { formatNumber, formatDuration, secondsUntil } from '@/lib/format';

interface ConstructCellProps {
  entry: CatalogEntry;
  inStock: number;
  onBuild: (unitId: number, count: number) => void;
  isPending: boolean;
}

function ConstructCell({ entry, inStock, onBuild, isPending }: ConstructCellProps) {
  const [count, setCount] = useState('');
  const { t } = useTranslation();
  const [group, key] = entry.i18n.split('.') as [string, string];
  const parsed = Math.max(0, Math.floor(Number(count) || 0));
  return (
    <>
      {inStock > 0 && (
        <>
          {formatNumber(inStock)}
          <br />
        </>
      )}
      <br />
      <input
        type="number"
        name={`unit_${entry.id}`}
        value={count}
        min={0}
        onChange={(e) => setCount(e.target.value)}
        aria-label={t(group, key)}
        style={{ width: 60, textAlign: 'center' }}
      />
      <br />
      <input
        type="button"
        className="button"
        value={t('shipyard', 'build')}
        onClick={() => parsed > 0 && onBuild(entry.id, parsed)}
        disabled={isPending || parsed <= 0}
      />
    </>
  );
}

export function DefenseScreen() {
  const { planetId } = useResolvedPlanet();
  const { t } = useTranslation();
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

  if (!planetId) return <div className="idiv">{t('overview', 'noPlanets')}</div>;

  const queue = queueQ.data ?? [];
  const inv: { ships: Record<string, number>; defense: Record<string, number> } =
    invQ.data ?? { ships: {}, defense: {} };
  const defense = catalogByGroup('defense');

  function inStock(id: number): number {
    return inv.defense[String(id)] ?? 0;
  }

  return (
    <>
      {queue.length > 0 && (
        <table className="ntable">
          <tbody>
            <tr>
              <th colSpan={4}>{t('shipyard', 'queue')}</th>
            </tr>
            {queue.map((task, idx) => (
              <tr key={task.id}>
                <td width="1px">{idx + 1}.</td>
                <td>
                  #{task.unit_id}&nbsp;{formatNumber(task.count)}
                </td>
                <td>{formatDuration(secondsUntil(task.end_at))}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <table className="ntable">
        <tbody>
          <tr>
            <th colSpan={3}>{t('shipyard', 'tabDefense')}</th>
          </tr>
          <tr>
            <th colSpan={2}>&nbsp;</th>
            <th style={{ textAlign: 'center' }}>{t('shipyard', 'quantity') ?? 'Количество'}</th>
          </tr>

          {defense.map((entry) => {
            const [group, key] = entry.i18n.split('.') as [string, string];
            return (
              <tr key={entry.id}>
                <td width="1px">
                  <img
                    src={`/assets/origin/images/units/${entry.id}.gif`}
                    alt=""
                    onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
                  />
                </td>
                <td valign="top">
                  <div style={{ width: '100%' }}>
                    {t(group, key)}
                  </div>
                </td>
                <td width="100px" align="center" valign="top">
                  <ConstructCell
                    entry={entry}
                    inStock={inStock(entry.id)}
                    onBuild={(unitId, count) => build.mutate({ unitId, count })}
                    isPending={build.isPending}
                  />
                </td>
              </tr>
            );
          })}

          {defense.length === 0 && (
            <tr>
              <td colSpan={3} className="center">—</td>
            </tr>
          )}
        </tbody>
      </table>
    </>
  );
}
