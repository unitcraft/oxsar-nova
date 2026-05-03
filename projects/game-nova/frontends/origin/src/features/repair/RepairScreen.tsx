// S-022 Repair — ремонт повреждённых юнитов (план 72.1 ч.20.4).
// Pixel-perfect клон legacy repair.tpl.
//
// Структура:
// 1. Шапка ангара: название «Ремонтный ангар (Уровень N)» + progress bar
// 2. Очередь ремонта (если есть)
// 3. Таблица повреждённых юнитов: иконка / имя / поля / количество input
// 4. Submit «Ремонтировать» внизу формы

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  cancelRepairQueue,
  fetchDamagedShips,
  fetchRepairQueue,
  repairUnits,
  startRepairVIP,
  vipCreditCost,
} from '@/api/repair';
import { fetchResourceReport } from '@/api/resource';
import { QK } from '@/api/query-keys';
import type { ApiError } from '@/api/client';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { findCatalog } from '@/features/common/catalog';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber, formatDuration, secondsUntil } from '@/lib/format';

export function RepairScreen() {
  const { planetId } = useResolvedPlanet();
  const { t } = useTranslation();
  const qc = useQueryClient();

  // План 72.1.56 B6 (legacy 1:1): per-unit input «сколько чинить».
  // Map unit_id → строковое значение из <input>. Пустая строка / 0 → backend
  // починит всех damaged (legacy default).
  const [quantities, setQuantities] = useState<Record<number, string>>({});

  const damagedQ = useQuery({
    queryKey: planetId ? QK.repairDamaged(planetId) : ['noop-rd'],
    queryFn: () =>
      planetId ? fetchDamagedShips(planetId) : Promise.resolve([]),
    enabled: planetId !== null,
  });

  const queueQ = useQuery({
    queryKey: planetId ? QK.repairQueue(planetId) : ['noop-rq'],
    queryFn: () =>
      planetId ? fetchRepairQueue(planetId) : Promise.resolve([]),
    enabled: planetId !== null,
  });

  // Уровень repair_factory (id=100) из resource-report.
  const reportQ = useQuery({
    queryKey: planetId ? QK.resourceReport(planetId) : ['noop-rr-rep'],
    queryFn: () =>
      planetId ? fetchResourceReport(planetId) : Promise.reject(),
    enabled: planetId !== null,
  });

  const [errMsg, setErrMsg] = useState<string | null>(null);

  function invalidateAll() {
    if (!planetId) return;
    void qc.invalidateQueries({ queryKey: QK.repairDamaged(planetId) });
    void qc.invalidateQueries({ queryKey: QK.repairQueue(planetId) });
    void qc.invalidateQueries({ queryKey: QK.shipyardInventory(planetId) });
    void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
    void qc.invalidateQueries({ queryKey: ['me'] });
  }

  const repair = useMutation({
    mutationFn: ({ unitId, quantity }: { unitId: number; quantity?: number }) =>
      repairUnits(planetId!, unitId, quantity),
    onSuccess: (_, vars) => {
      setErrMsg(null);
      // Сбросим input для починенного типа.
      setQuantities((prev) => {
        const copy = { ...prev };
        delete copy[vars.unitId];
        return copy;
      });
      invalidateAll();
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  // План 72.1.25: legacy `abortRepair`.
  const cancel = useMutation({
    mutationFn: (queueId: string) => cancelRepairQueue(planetId!, queueId),
    onSuccess: invalidateAll,
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  // План 72.1.25: legacy `startRepairVIP`.
  const vip = useMutation({
    mutationFn: (queueId: string) => startRepairVIP(planetId!, queueId),
    onSuccess: invalidateAll,
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (!planetId) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  const damaged = damagedQ.data ?? [];
  const queue = (queueQ.data ?? []).filter((q) => q.mode === 'repair');
  const buildings = reportQ.data?.buildings ?? [];
  const repairFactoryLvl =
    buildings.find((b) => b.unit_id === 100)?.level ?? 0;

  function unitName(unitId: number): string {
    const cat = findCatalog(unitId);
    if (!cat) return `#${unitId}`;
    const [g, k] = cat.i18n.split('.') as [string, string];
    return t(g, k);
  }

  function unitIcon(unitId: number): string | null {
    const cat = findCatalog(unitId);
    return cat?.icon ?? null;
  }

  return (
    <>
      {/* Шапка ангара */}
      <table className="ntable">
        <tbody>
          <tr>
            <th colSpan={4}>
              <span style={{ float: 'right' }}>
                Уровень {repairFactoryLvl}
              </span>
              {t('info', 'repairFactory') ?? 'Ремонтный ангар'}
            </th>
          </tr>
          <tr>
            <td colSpan={4}>
              <div style={{ float: 'left', paddingRight: 5 }}>
                <img
                  src="/assets/origin/images/units/repair_factory.gif"
                  alt={t('info', 'repairFactory') ?? ''}
                  onError={(e) => {
                    (e.target as HTMLImageElement).style.display = 'none';
                  }}
                />
              </div>
              <div style={{ display: 'table' }}>
                {t('info', 'repairFactoryDesc') ??
                  'Ремонтный ангар позволяет восстанавливать повреждённые юниты.'}
              </div>
            </td>
          </tr>
          {queue.length > 0 &&
            queue.map((task, idx) => {
              const secLeft = secondsUntil(task.end_at);
              const cost = vipCreditCost(task.count);
              return (
                <tr key={task.id}>
                  <td width="1px">{idx + 1}.</td>
                  <td>
                    {unitName(task.unit_id)}: {formatNumber(task.count)}
                  </td>
                  <td width="120px" className="center">
                    {secLeft > 0 ? formatDuration(secLeft) : '—'}
                  </td>
                  <td width="160px" className="center">
                    <button
                      type="button"
                      className="button"
                      disabled={cancel.isPending}
                      onClick={() => cancel.mutate(task.id)}
                      title={t('repair', 'abortBtn') || 'Отменить'}
                    >
                      ✕
                    </button>
                    {' '}
                    <button
                      type="button"
                      className="button"
                      disabled={vip.isPending}
                      onClick={() => vip.mutate(task.id)}
                      title={t('repair', 'vipBtn', { credits: cost }) || `VIP (${cost} cr)`}
                    >
                      ⚡ {cost}
                    </button>
                  </td>
                </tr>
              );
            })}
        </tbody>
      </table>

      {/* Список повреждённых юнитов */}
      <table className="ntable">
        <tbody>
          <tr>
            <th colSpan={3}>{t('buildings', 'repairNeededUnits') ?? 'Повреждённые юниты'}</th>
          </tr>

          {damaged.length === 0 && (
            <tr>
              <td colSpan={3} align="center">
                {t('buildings', 'repairZeroQuantities') ?? 'Нет повреждённых юнитов'}
              </td>
            </tr>
          )}

          {damaged.map((u) => {
            const icon = unitIcon(u.unit_id);
            return (
              <tr key={u.unit_id}>
                <td width="1px" style={{ verticalAlign: 'top' }}>
                  {icon && (
                    <img
                      src={`/assets/origin/images/units/${icon}.gif`}
                      alt={unitName(u.unit_id)}
                      onError={(e) => {
                        (e.target as HTMLImageElement).style.display = 'none';
                      }}
                    />
                  )}
                </td>
                <td valign="top">
                  <div style={{ width: '100%' }}>
                    <span style={{ float: 'right' }}>
                      {t('buildings', 'repairNeededUnits') ?? 'Повреждено'}: {formatNumber(u.damaged)}
                    </span>
                    {unitName(u.unit_id)}
                  </div>
                  <div style={{ clear: 'both', fontSize: 'smaller' }}>
                    {t('buildings', 'shipsExist', { arg1: formatNumber(u.count) })}
                    {' · '}
                    Целостность: {Math.floor(u.shell_percent)}%
                  </div>
                </td>
                <td width="120px" align="center" valign="top">
                  {/* План 72.1.56 B6: legacy 1:1 — input quantity. */}
                  <input
                    type="text"
                    size={3}
                    value={quantities[u.unit_id] ?? ''}
                    placeholder={String(u.damaged)}
                    onChange={(e) => {
                      const v = e.target.value.replace(/[^0-9]/g, '');
                      setQuantities((p) => ({ ...p, [u.unit_id]: v }));
                    }}
                    style={{ width: 50, marginRight: 4 }}
                  />
                  <span className="true">
                    <input
                      type="button"
                      className="button"
                      value={t('buildings', 'repair') ?? 'Ремонтировать'}
                      onClick={() => {
                        const raw = quantities[u.unit_id];
                        const q = raw ? parseInt(raw, 10) : 0;
                        repair.mutate(
                          q > 0
                            ? { unitId: u.unit_id, quantity: q }
                            : { unitId: u.unit_id },
                        );
                      }}
                      disabled={repair.isPending}
                    />
                  </span>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>

      {errMsg && (
        <div className="idiv">
          <span className="false">{errMsg}</span>
        </div>
      )}
    </>
  );
}
