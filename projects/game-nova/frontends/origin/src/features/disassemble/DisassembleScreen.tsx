// S-R03 Disassemble — утилизация (план 72.1 ч.19).
// Pixel-perfect клон legacy disassemble.tpl.

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchShipyardInventory } from '@/api/shipyard';
import { fetchRepairQueue, disassembleUnits } from '@/api/repair';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber, formatDuration, secondsUntil } from '@/lib/format';

export function DisassembleScreen() {
  const { planetId } = useResolvedPlanet();
  const { t } = useTranslation();
  const qc = useQueryClient();

  // counts per unit_id
  const [counts, setCounts] = useState<Record<number, string>>({});

  const invQ = useQuery({
    queryKey: planetId ? QK.shipyardInventory(planetId) : ['noop-dis-inv'],
    queryFn: () =>
      planetId
        ? fetchShipyardInventory(planetId)
        : Promise.resolve({ ships: {}, defense: {} }),
    enabled: planetId !== null,
  });

  const queueQ = useQuery({
    queryKey: planetId ? QK.repairQueue(planetId) : ['noop-rq'],
    queryFn: () => (planetId ? fetchRepairQueue(planetId) : Promise.resolve([])),
    enabled: planetId !== null,
  });

  const disassemble = useMutation({
    mutationFn: ({ unitId, count }: { unitId: number; count: number }) =>
      disassembleUnits(planetId!, unitId, count),
    onSuccess: () => {
      setCounts({});
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.shipyardInventory(planetId) });
        void qc.invalidateQueries({ queryKey: QK.repairQueue(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
    },
  });

  if (!planetId) return <div className="idiv">{t('overview', 'noPlanets')}</div>;

  const inv: { ships: Record<string, number>; defense: Record<string, number> } =
    invQ.data ?? { ships: {}, defense: {} };

  const disassembleQueue = (queueQ.data ?? []).filter(
    (q) => q.mode === 'disassemble',
  );

  const allShips = catalogByGroup('ship');
  const allDefense = catalogByGroup('defense');

  // Только юниты, которые реально есть на планете
  const availableShips = allShips.filter(
    (s) => (inv.ships[String(s.id)] ?? 0) > 0,
  );
  const availableDefense = allDefense.filter(
    (s) => (inv.defense[String(s.id)] ?? 0) > 0,
  );

  function inStockShip(id: number): number {
    return inv.ships[String(id)] ?? 0;
  }
  function inStockDefense(id: number): number {
    return inv.defense[String(id)] ?? 0;
  }

  return (
    <>
      {/* Очередь */}
      {disassembleQueue.length > 0 && (
        <table className="ntable">
          <tbody>
            <tr>
              <th colSpan={3}>Очередь утилизации</th>
            </tr>
            {disassembleQueue.map((task, idx) => (
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

      {/* Основная таблица */}
      <table className="ntable">
        <tbody>
          <tr>
            <th colSpan={3}>Утилизация</th>
          </tr>
          <tr>
            <th colSpan={2}>&nbsp;</th>
            <th style={{ textAlign: 'center' }}>Количество</th>
          </tr>

          {(availableShips.length > 0 || availableDefense.length > 0) && (
            <>
              {availableShips.length > 0 && (
                <tr>
                  <th colSpan={3}>Флот</th>
                </tr>
              )}
              {availableShips.map((s) => {
                const stock = inStockShip(s.id);
                const [group, key] = s.i18n.split('.') as [string, string];
                const cnt = Math.min(
                  stock,
                  Math.max(0, Math.floor(Number(counts[s.id] ?? '') || 0)),
                );
                return (
                  <tr key={s.id}>
                    <td width="1px">
                      <img
                        src={`/assets/origin/images/units/${s.id}.gif`}
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
                      {formatNumber(stock)}
                      <br /><br />
                      <input
                        type="number"
                        min={0}
                        max={stock}
                        value={counts[s.id] ?? ''}
                        onChange={(e) =>
                          setCounts((prev) => ({ ...prev, [s.id]: e.target.value }))
                        }
                        style={{ width: 60, textAlign: 'center' }}
                        aria-label={t(group, key)}
                      />
                      <br />
                      <input
                        type="button"
                        className="button"
                        value="Утилизировать"
                        disabled={disassemble.isPending || cnt <= 0}
                        onClick={() =>
                          cnt > 0 && disassemble.mutate({ unitId: s.id, count: cnt })
                        }
                      />
                    </td>
                  </tr>
                );
              })}

              {availableDefense.length > 0 && (
                <tr>
                  <th colSpan={3}>Оборона</th>
                </tr>
              )}
              {availableDefense.map((s) => {
                const stock = inStockDefense(s.id);
                const [group, key] = s.i18n.split('.') as [string, string];
                const cnt = Math.min(
                  stock,
                  Math.max(0, Math.floor(Number(counts[s.id] ?? '') || 0)),
                );
                return (
                  <tr key={s.id}>
                    <td width="1px">
                      <img
                        src={`/assets/origin/images/units/${s.id}.gif`}
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
                      {formatNumber(stock)}
                      <br /><br />
                      <input
                        type="number"
                        min={0}
                        max={stock}
                        value={counts[s.id] ?? ''}
                        onChange={(e) =>
                          setCounts((prev) => ({ ...prev, [s.id]: e.target.value }))
                        }
                        style={{ width: 60, textAlign: 'center' }}
                        aria-label={t(group, key)}
                      />
                      <br />
                      <input
                        type="button"
                        className="button"
                        value="Утилизировать"
                        disabled={disassemble.isPending || cnt <= 0}
                        onClick={() =>
                          cnt > 0 && disassemble.mutate({ unitId: s.id, count: cnt })
                        }
                      />
                    </td>
                  </tr>
                );
              })}
            </>
          )}

          {availableShips.length === 0 && availableDefense.length === 0 && (
            <tr>
              <td colSpan={3} align="center">Нет юнитов для утилизации</td>
            </tr>
          )}
        </tbody>
      </table>
    </>
  );
}
