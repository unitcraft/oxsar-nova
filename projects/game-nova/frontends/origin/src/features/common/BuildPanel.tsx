// BuildPanel — общая таблица для shipyard/defense (план 72.1 ч.20.3).
// Pixel-perfect клон legacy shipyard.tpl.
//
// Структура: <table.ntable> очередь + <table.ntable> заголовок + ряды юнитов.
// Каждый ряд: иконка / описание+cost / quantity input + count в наличии.
// Внизу таблицы — общая кнопка submit на multiple-build (как legacy form).

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  buildShipyard,
  cancelShipyardTask,
  fetchShipyardCapacity,
  fetchShipyardInventory,
  fetchShipyardQueue,
} from '@/api/shipyard';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup, type CatalogEntry } from '@/features/common/catalog';
import { RequiredResTable } from '@/features/common/RequiredResTable';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber, formatDuration, secondsUntil } from '@/lib/format';
import type { ShipyardInventory } from '@/api/types';

interface BuildPanelProps {
  /** Группа юнитов: 'ship' для верфи, 'defense' для обороны. */
  group: 'ship' | 'defense';
  /** Заголовок таблицы (i18n) */
  title: string;
}

export function BuildPanel({ group, title }: BuildPanelProps) {
  const { planetId, planet } = useResolvedPlanet();
  const { t } = useTranslation();
  const qc = useQueryClient();
  const [counts, setCounts] = useState<Record<number, string>>({});

  const queueQ = useQuery({
    queryKey: planetId ? QK.shipyardQueue(planetId) : ['noop-sq'],
    queryFn: () => (planetId ? fetchShipyardQueue(planetId) : Promise.resolve([])),
    enabled: planetId !== null,
  });

  const invQ = useQuery<ShipyardInventory>({
    queryKey: planetId ? QK.shipyardInventory(planetId) : ['noop-si'],
    queryFn: () =>
      planetId
        ? fetchShipyardInventory(planetId)
        : Promise.resolve<ShipyardInventory>({ ships: {}, defense: {} }),
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

  // План 72.1.41: cancel-задача shipyard (legacy `Shipyard::abort`).
  const cancel = useMutation({
    mutationFn: (queueId: string) => cancelShipyardTask(planetId!, queueId),
    onSuccess: () => {
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.shipyardQueue(planetId) });
        void qc.invalidateQueries({ queryKey: QK.shipyardInventory(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
    },
  });

  // План 72.1.41: capacity-индикатор для DefenseScreen.
  const capQ = useQuery({
    queryKey: planetId ? ['shipyard-capacity', planetId] : ['noop-cap'],
    queryFn: () =>
      planetId
        ? fetchShipyardCapacity(planetId)
        : Promise.resolve({
            free_shield_fields: 0,
            max_shield_fields: 0,
            free_rocket_fields: 0,
            max_rocket_fields: 0,
          }),
    enabled: planetId !== null && group === 'defense',
  });

  if (!planetId) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  const queue = queueQ.data ?? [];
  const inv = invQ.data ?? { ships: {}, defense: {} };
  const stockMap: Record<string, number> =
    group === 'ship' ? inv.ships ?? {} : inv.defense ?? {};
  const costsMap = group === 'ship' ? inv.ship_costs ?? {} : inv.defense_costs ?? {};
  const secsMap = group === 'ship' ? inv.ship_seconds ?? {} : inv.defense_seconds ?? {};
  const catalog = catalogByGroup(group);

  const available = planet
    ? {
        metal: Math.floor(planet.metal),
        silicon: Math.floor(planet.silicon),
        hydrogen: Math.floor(planet.hydrogen),
      }
    : { metal: 0, silicon: 0, hydrogen: 0 };

  function inStock(id: number): number {
    return Number(stockMap[String(id)] ?? 0);
  }

  function canBuildOne(unitId: number): boolean {
    const c = costsMap[String(unitId)];
    if (!c) return false;
    return (
      available.metal >= c.metal &&
      available.silicon >= c.silicon &&
      available.hydrogen >= c.hydrogen
    );
  }

  function onSubmitAll(e: React.FormEvent) {
    e.preventDefault();
    for (const entry of catalog) {
      const cnt = Math.max(0, Math.floor(Number(counts[entry.id] ?? '') || 0));
      if (cnt > 0) {
        build.mutate({ unitId: entry.id, count: cnt });
      }
    }
    setCounts({});
  }

  function entryName(entry: CatalogEntry): string {
    const [g, k] = entry.i18n.split('.') as [string, string];
    return t(g, k);
  }

  return (
    <>
      {queue.length > 0 && (
        <table className="ntable">
          <tbody>
            <tr>
              <th colSpan={5}>{t('buildings', 'outstandingMissions')}</th>
            </tr>
            {queue.map((task, idx) => {
              const cat = catalog.find((c) => c.id === task.unit_id);
              const name = cat ? entryName(cat) : `#${task.unit_id}`;
              return (
                <tr key={task.id}>
                  <td width="1px">{idx + 1}.</td>
                  <td colSpan={2}>
                    {name}: {formatNumber(task.count)}
                  </td>
                  <td width="100px">
                    {formatDuration(secondsUntil(task.end_at))}
                  </td>
                  {/* План 72.1.41: cancel-кнопка (legacy Shipyard::abort). */}
                  <td width="60px" align="center">
                    <button
                      type="button"
                      className="button"
                      disabled={cancel.isPending}
                      title={t('buildings', 'cancelTask') || 'Отменить'}
                      onClick={() => {
                        if (
                          window.confirm(
                            (t('buildings', 'cancelConfirm') as string) ||
                              'Отменить задачу? Возврат ресурсов в зависимости от прогресса.',
                          )
                        ) {
                          cancel.mutate(task.id);
                        }
                      }}
                    >
                      ✕
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}

      {/* План 72.1.41: capacity-индикатор для DefenseScreen. */}
      {group === 'defense' && capQ.data && (
        <table className="ntable">
          <tbody>
            <tr>
              <td className="center">
                🛡 {t('buildings', 'shieldFields') || 'Поля щитов'}:{' '}
                <b className={capQ.data.free_shield_fields === 0 ? 'false' : 'true'}>
                  {capQ.data.free_shield_fields} / {capQ.data.max_shield_fields}
                </b>
                {' · '}
                🚀 {t('buildings', 'rocketFields') || 'Поля ракет'}:{' '}
                <b className={capQ.data.free_rocket_fields === 0 ? 'false' : 'true'}>
                  {capQ.data.free_rocket_fields} / {capQ.data.max_rocket_fields}
                </b>
              </td>
            </tr>
          </tbody>
        </table>
      )}

      <form onSubmit={onSubmitAll} style={{ padding: 0, margin: 0 }}>
        <table className="ntable">
          <tbody>
            <tr>
              <th colSpan={3}>{title}</th>
            </tr>
            <tr>
              <th colSpan={2}>&nbsp;</th>
              <th style={{ textAlign: 'center' }}>
                {t('buildings', 'quantity') ?? 'Количество'}
              </th>
            </tr>

            {catalog.map((entry) => {
              const [g, k] = entry.i18n.split('.') as [string, string];
              const stock = inStock(entry.id);
              const cost = costsMap[String(entry.id)] ?? {
                metal: 0,
                silicon: 0,
                hydrogen: 0,
              };
              const secs = secsMap[String(entry.id)] ?? 0;
              const descKey = `${k}Desc`;
              const desc = t(g, descKey);
              const hasDesc = !desc.startsWith('[');
              const enough = canBuildOne(entry.id);
              const cnt = counts[entry.id] ?? '';

              return (
                <tr key={entry.id}>
                  <td width="1px" style={{ verticalAlign: 'top' }}>
                    <img
                      src={`/assets/origin/images/units/${entry.icon}.gif`}
                      alt={entryName(entry)}
                      onError={(e) => {
                        (e.target as HTMLImageElement).style.display = 'none';
                      }}
                    />
                  </td>
                  <td style={{ verticalAlign: 'top' }}>
                    <div style={{ width: '100%' }}>{entryName(entry)}</div>
                    {hasDesc && (
                      <div style={{ clear: 'both', fontSize: 'smaller' }}>
                        {desc}
                      </div>
                    )}
                    <div style={{ marginTop: 6 }}>
                      <RequiredResTable
                        metal={cost.metal}
                        silicon={cost.silicon}
                        hydrogen={cost.hydrogen}
                        available={available}
                        seconds={secs}
                      />
                    </div>
                  </td>
                  <td
                    width="100px"
                    align="center"
                    style={{ verticalAlign: 'top' }}
                  >
                    {stock > 0 && (
                      <>
                        {formatNumber(stock)}
                        <br />
                      </>
                    )}
                    <br />
                    <input
                      type="number"
                      name={String(entry.id)}
                      value={cnt}
                      min={0}
                      onChange={(e) =>
                        setCounts((prev) => ({
                          ...prev,
                          [entry.id]: e.target.value,
                        }))
                      }
                      aria-label={entryName(entry)}
                      className={enough ? '' : 'notavailable'}
                      style={{ width: 60, textAlign: 'center' }}
                    />
                  </td>
                </tr>
              );
            })}

            <tr>
              <td colSpan={3} align="center">
                <input
                  type="submit"
                  name="sendmission"
                  value={t('buildings', 'build') ?? 'Построить'}
                  className="button"
                  disabled={build.isPending}
                />
              </td>
            </tr>
          </tbody>
        </table>
      </form>
    </>
  );
}
