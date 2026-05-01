// S-R03 Disassemble — утилизация (план 72.1 ч.19, расширен 72.1.25).
// Pixel-perfect клон legacy disassemble.tpl:
//   - Очередь утилизации с abort и VIP-кнопкой.
//   - Таблица юнитов с required/earn ресурсами.
//   - umode/observer/under-attack блокировки приходят с backend как 400.

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchShipyardInventory } from '@/api/shipyard';
import type { ShipyardInventory } from '@/api/types';
import {
  cancelRepairQueue,
  disassembleUnits,
  fetchRepairQueue,
  startRepairVIP,
  vipCreditCost,
} from '@/api/repair';
import { QK } from '@/api/query-keys';
import type { ApiError } from '@/api/client';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { catalogByGroup } from '@/features/common/catalog';
import { useTranslation } from '@/i18n/i18n';
import {
  formatNumber,
  formatDuration,
  secondsUntil,
} from '@/lib/format';

// План 72.1.25: legacy formula `setDisassembleUnitRequirements`:
//   required = ceil(base * 0.2 / 10) * 10  (списывается)
//   return   = ceil(base * 0.9 / 10) * 10  (зачисляется)
//   earn     = return - required
function ceil10(v: number): number {
  if (v <= 0) return 0;
  return Math.ceil(v / 10) * 10;
}

function disassembleEcon(
  baseCost: { metal: number; silicon: number; hydrogen: number } | undefined,
  count: number,
) {
  if (!baseCost || count <= 0) {
    return {
      reqMetal: 0, reqSilicon: 0, reqHydrogen: 0,
      retMetal: 0, retSilicon: 0, retHydrogen: 0,
      earnMetal: 0, earnSilicon: 0, earnHydrogen: 0,
    };
  }
  const reqMetal = ceil10(baseCost.metal * 0.2) * count;
  const reqSilicon = ceil10(baseCost.silicon * 0.2) * count;
  const reqHydrogen = ceil10(baseCost.hydrogen * 0.2) * count;
  const retMetal = ceil10(baseCost.metal * 0.9) * count;
  const retSilicon = ceil10(baseCost.silicon * 0.9) * count;
  const retHydrogen = 0; // legacy: returnHydrogen = 0
  return {
    reqMetal, reqSilicon, reqHydrogen,
    retMetal, retSilicon, retHydrogen,
    earnMetal: retMetal - reqMetal,
    earnSilicon: retSilicon - reqSilicon,
    earnHydrogen: retHydrogen - reqHydrogen,
  };
}

export function DisassembleScreen() {
  const { planetId } = useResolvedPlanet();
  const { t } = useTranslation();
  const qc = useQueryClient();

  const [counts, setCounts] = useState<Record<number, string>>({});
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const invQ = useQuery({
    queryKey: planetId ? QK.shipyardInventory(planetId) : ['noop-dis-inv'],
    queryFn: () =>
      planetId
        ? fetchShipyardInventory(planetId)
        : Promise.resolve<ShipyardInventory>({
            ships: {},
            defense: {},
            ship_costs: {},
            defense_costs: {},
          }),
    enabled: planetId !== null,
  });

  const queueQ = useQuery({
    queryKey: planetId ? QK.repairQueue(planetId) : ['noop-rq'],
    queryFn: () => (planetId ? fetchRepairQueue(planetId) : Promise.resolve([])),
    enabled: planetId !== null,
  });

  function invalidateAll() {
    if (!planetId) return;
    void qc.invalidateQueries({ queryKey: QK.shipyardInventory(planetId) });
    void qc.invalidateQueries({ queryKey: QK.repairQueue(planetId) });
    void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
    void qc.invalidateQueries({ queryKey: ['me'] });
  }

  const disassemble = useMutation({
    mutationFn: ({ unitId, count }: { unitId: number; count: number }) =>
      disassembleUnits(planetId!, unitId, count),
    onSuccess: () => {
      setCounts({});
      setErrMsg(null);
      invalidateAll();
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const cancel = useMutation({
    mutationFn: (queueId: string) => cancelRepairQueue(planetId!, queueId),
    onSuccess: invalidateAll,
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const vip = useMutation({
    mutationFn: (queueId: string) => startRepairVIP(planetId!, queueId),
    onSuccess: invalidateAll,
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (!planetId) return <div className="idiv">{t('overview', 'noPlanets')}</div>;

  const inv = invQ.data ?? { ships: {}, defense: {}, ship_costs: {}, defense_costs: {} };
  const shipCosts = inv.ship_costs ?? {};
  const defCosts = inv.defense_costs ?? {};

  const disassembleQueue = (queueQ.data ?? []).filter((q) => q.mode === 'disassemble');

  const allShips = catalogByGroup('ship');
  const allDefense = catalogByGroup('defense');
  const availableShips = allShips.filter((s) => (inv.ships[String(s.id)] ?? 0) > 0);
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
      {/* Очередь утилизации */}
      {disassembleQueue.length > 0 && (
        <table className="ntable">
          <tbody>
            <tr>
              <th colSpan={4}>{t('repair', 'disassembleQueue') || 'Очередь утилизации'}</th>
            </tr>
            {disassembleQueue.map((task, idx) => {
              const secLeft = secondsUntil(task.end_at);
              const cost = vipCreditCost(task.count);
              return (
                <tr key={task.id}>
                  <td width="1px">{idx + 1}.</td>
                  <td>
                    #{task.unit_id}&nbsp;{formatNumber(task.count)}
                  </td>
                  <td>{secLeft > 0 ? formatDuration(secLeft) : '—'}</td>
                  <td className="center">
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
      )}

      {/* Основная таблица */}
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={5}>{t('repair', 'disassembleTitle') || 'Утилизация'}</th>
          </tr>
          <tr>
            <th colSpan={2}>&nbsp;</th>
            <th className="center">{t('repair', 'colRequired') || 'Стоимость'}</th>
            <th className="center">{t('repair', 'colEarn') || 'Получите'}</th>
            <th className="center">{t('repair', 'colCount') || 'Количество'}</th>
          </tr>
        </thead>
        <tbody>
          {availableShips.length > 0 && (
            <tr>
              <th colSpan={5}>{t('shipyard', 'tabFleet') || 'Флот'}</th>
            </tr>
          )}
          {availableShips.map((s) => {
            const stock = inStockShip(s.id);
            const [group, key] = s.i18n.split('.') as [string, string];
            const cnt = Math.min(
              stock,
              Math.max(0, Math.floor(Number(counts[s.id] ?? '') || 0)),
            );
            const econ = disassembleEcon(shipCosts[String(s.id)], cnt);
            return (
              <UnitRow
                key={s.id}
                imgId={s.id}
                name={t(group, key)}
                stock={stock}
                count={counts[s.id] ?? ''}
                onChange={(v) => setCounts((p) => ({ ...p, [s.id]: v }))}
                onSubmit={() => cnt > 0 && disassemble.mutate({ unitId: s.id, count: cnt })}
                disabled={disassemble.isPending}
                econ={econ}
                actionLabel={t('repair', 'disassembleBtn') || 'Утилизировать'}
              />
            );
          })}

          {availableDefense.length > 0 && (
            <tr>
              <th colSpan={5}>{t('shipyard', 'tabDefense') || 'Оборона'}</th>
            </tr>
          )}
          {availableDefense.map((s) => {
            const stock = inStockDefense(s.id);
            const [group, key] = s.i18n.split('.') as [string, string];
            const cnt = Math.min(
              stock,
              Math.max(0, Math.floor(Number(counts[s.id] ?? '') || 0)),
            );
            const econ = disassembleEcon(defCosts[String(s.id)], cnt);
            return (
              <UnitRow
                key={s.id}
                imgId={s.id}
                name={t(group, key)}
                stock={stock}
                count={counts[s.id] ?? ''}
                onChange={(v) => setCounts((p) => ({ ...p, [s.id]: v }))}
                onSubmit={() => cnt > 0 && disassemble.mutate({ unitId: s.id, count: cnt })}
                disabled={disassemble.isPending}
                econ={econ}
                actionLabel={t('repair', 'disassembleBtn') || 'Утилизировать'}
              />
            );
          })}

          {availableShips.length === 0 && availableDefense.length === 0 && (
            <tr>
              <td colSpan={5} className="center">
                {t('repair', 'emptyDisassemble') || 'Нет юнитов для утилизации'}
              </td>
            </tr>
          )}
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

interface EconView {
  reqMetal: number;
  reqSilicon: number;
  reqHydrogen: number;
  earnMetal: number;
  earnSilicon: number;
  earnHydrogen: number;
}

function UnitRow({
  imgId,
  name,
  stock,
  count,
  onChange,
  onSubmit,
  disabled,
  econ,
  actionLabel,
}: {
  imgId: number;
  name: string;
  stock: number;
  count: string;
  onChange: (v: string) => void;
  onSubmit: () => void;
  disabled: boolean;
  econ: EconView;
  actionLabel: string;
}) {
  const cnt = Math.min(
    stock,
    Math.max(0, Math.floor(Number(count) || 0)),
  );
  return (
    <tr>
      <td width="1px">
        <img
          src={`/assets/origin/images/units/${imgId}.gif`}
          alt=""
          onError={(e) => { (e.target as HTMLImageElement).style.display = 'none'; }}
        />
      </td>
      <td valign="top">
        <div style={{ width: '100%' }}>{name}</div>
      </td>
      <td className="center">
        {cnt > 0 ? (
          <span>
            {econ.reqMetal > 0 && <>М: <b>{formatNumber(econ.reqMetal)}</b><br /></>}
            {econ.reqSilicon > 0 && <>К: <b>{formatNumber(econ.reqSilicon)}</b><br /></>}
            {econ.reqHydrogen > 0 && <>В: <b>{formatNumber(econ.reqHydrogen)}</b></>}
          </span>
        ) : '—'}
      </td>
      <td className="center">
        {cnt > 0 ? (
          <span className="true">
            {econ.earnMetal > 0 && <>+{formatNumber(econ.earnMetal)}М<br /></>}
            {econ.earnSilicon > 0 && <>+{formatNumber(econ.earnSilicon)}К<br /></>}
            {econ.earnHydrogen > 0 && <>+{formatNumber(econ.earnHydrogen)}В</>}
          </span>
        ) : '—'}
      </td>
      <td width="120px" className="center" valign="top">
        {formatNumber(stock)}
        <br /><br />
        <input
          type="number"
          min={0}
          max={stock}
          value={count}
          onChange={(e) => onChange(e.target.value)}
          style={{ width: 60, textAlign: 'center' }}
          aria-label={name}
        />
        <br />
        <input
          type="button"
          className="button"
          value={actionLabel}
          disabled={disabled || cnt <= 0}
          onClick={onSubmit}
        />
      </td>
    </tr>
  );
}
