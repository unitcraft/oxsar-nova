// S-R03 Disassemble — утилизация кораблей (план 72.1).
//
// Pixel-perfect клон legacy Repair.class.php (!isRepair mode):
//   - Список кораблей на планете с количеством.
//   - Форма: выбрать юнита + количество → утилизировать (90% ресурсов).
//   - Активная очередь утилизации.
//
// Endpoints:
//   GET  /api/planets/{id}/shipyard/inventory — список кораблей (ships)
//   GET  /api/planets/{id}/repair/queue       — активная очередь
//   POST /api/planets/{id}/repair/disassemble — отправить на утилизацию

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchShipyardInventory } from '@/api/shipyard';
import { fetchRepairQueue, disassembleUnits } from '@/api/repair';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';
import { formatNumber, formatDuration, secondsUntil } from '@/lib/format';

export function DisassembleScreen() {
  const { planetId } = useResolvedPlanet();
  const qc = useQueryClient();

  const [selectedUnitId, setSelectedUnitId] = useState<number | null>(null);
  const [count, setCount] = useState('');

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
    mutationFn: () =>
      disassembleUnits(planetId!, selectedUnitId!, Math.floor(Number(count) || 0)),
    onSuccess: () => {
      setCount('');
      if (planetId) {
        void qc.invalidateQueries({ queryKey: QK.shipyardInventory(planetId) });
        void qc.invalidateQueries({ queryKey: QK.repairQueue(planetId) });
        void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
      }
    },
  });

  if (!planetId) {
    return <div className="idiv">Нет планет</div>;
  }

  const inv: { ships: Record<string, number>; defense: Record<string, number> } =
    invQ.data ?? { ships: {}, defense: {} };
  const queue = (queueQ.data ?? []).filter((q) => q.mode === 'disassemble');
  const allShips = catalogByGroup('ship');

  // Только те корабли, которые есть на планете (кол-во > 0)
  const availableShips = allShips.filter(
    (s) => (inv.ships[String(s.id)] ?? 0) > 0,
  );

  const selectedInStock = selectedUnitId
    ? (inv.ships[String(selectedUnitId)] ?? 0)
    : 0;
  const parsedCount = Math.min(
    selectedInStock,
    Math.max(0, Math.floor(Number(count) || 0)),
  );
  const canSubmit = selectedUnitId !== null && parsedCount > 0 && !disassemble.isPending;

  return (
    <>
      {/* Активная очередь утилизации */}
      {queue.length > 0 && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={3}>Очередь утилизации</th>
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

      {/* Корабли на планете */}
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={3}>Корабли на планете</th>
          </tr>
        </thead>
        <tbody>
          {availableShips.length === 0 && (
            <tr>
              <td colSpan={3} className="center">Нет кораблей для утилизации</td>
            </tr>
          )}
          {availableShips.map((s) => {
            const inStock = inv.ships[String(s.id)] ?? 0;
            const [group, key] = s.i18n.split('.') as [string, string];
            return (
              <tr
                key={s.id}
                style={{ cursor: 'pointer' }}
                onClick={() => setSelectedUnitId(s.id)}
                className={selectedUnitId === s.id ? 'selected' : ''}
              >
                <td width="1px">
                  <input
                    type="radio"
                    name="dis-unit"
                    checked={selectedUnitId === s.id}
                    onChange={() => setSelectedUnitId(s.id)}
                    aria-label={`${group}.${key}`}
                  />
                </td>
                <td>#{s.id} {group}.{key}</td>
                <td className="center">{formatNumber(inStock)} шт.</td>
              </tr>
            );
          })}
        </tbody>
      </table>

      {/* Форма утилизации */}
      {availableShips.length > 0 && (
        <div style={{ marginTop: 8, textAlign: 'center' }}>
          <label>
            Количество:{' '}
            <input
              type="number"
              min={1}
              max={selectedInStock || 1}
              value={count}
              onChange={(e) => setCount(e.target.value)}
              style={{ width: 80 }}
              disabled={selectedUnitId === null}
            />
            {selectedUnitId !== null && (
              <span style={{ marginLeft: 4 }}>
                / {formatNumber(selectedInStock)} шт.
              </span>
            )}
          </label>
          {' '}
          <button
            type="button"
            className="button"
            onClick={() => canSubmit && disassemble.mutate()}
            disabled={!canSubmit}
          >
            {disassemble.isPending ? 'Утилизация…' : 'Утилизировать'}
          </button>
          {disassemble.isSuccess && (
            <span style={{ marginLeft: 8, color: 'green' }}>Отправлено в очередь</span>
          )}
          {disassemble.isError && (
            <span style={{ marginLeft: 8, color: 'red' }}>Ошибка</span>
          )}
          <div style={{ marginTop: 4, fontSize: '0.85em', color: '#aaa' }}>
            Утилизация возвращает 90% стоимости постройки
          </div>
        </div>
      )}
    </>
  );
}
