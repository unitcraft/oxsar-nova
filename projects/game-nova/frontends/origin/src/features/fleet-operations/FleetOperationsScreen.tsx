// S-024 Fleet operations (план 72 Ф.3 Spring 2 ч.2).
//
// Pixel-perfect зеркало legacy `templates/standard/missions.tpl` /
// missions2.tpl. Управление активными миссиями: список с типом, целью,
// временем прилёта, кнопкой recall.
//
// Endpoints (openapi.yaml):
//   GET  /api/fleet                  → активные флоты + slots_used/max
//   POST /api/fleet/{id}/recall      → отзыв (Idempotency-Key R9)
//
// Замечание о MissionCode → label маппинге:
// nova-API возвращает `mission` как integer (6/7/8/9/10/11/12/15).
// MISSION_LABEL_KEY мапит на ключи bundle fleet:* (missionAttack/...).
// Если код не известен — fallback на `fleet.missionFallback`.

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchFleet, loadFleet, recallFleet, unloadFleet } from '@/api/fleet';
import type { ApiError } from '@/api/client';
import { QK } from '@/api/query-keys';
import type { Fleet } from '@/api/types';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { formatCoords, formatDuration, secondsUntil } from '@/lib/format';
import { useTranslation } from '@/i18n/i18n';
import { useState } from 'react';

// План 72.1.47: исправлен mapping mission→i18n key (был перепутан, см.
// event/kinds.go: 6=Position, 7=Transport, 8=Colonize, 9=Recycling,
// 10=AttackSingle, 11=Spy, 12=AttackAlliance, 15=Expedition, 17=Holding).
const MISSION_LABEL_KEY: Record<number, string> = {
  6: 'missionRebase',
  7: 'missionTransport',
  8: 'missionColonize',
  9: 'missionRecycle',
  10: 'missionAttack',
  11: 'missionSpy',
  12: 'missionAttack', // ACS отображаем как Атака
  15: 'missionExpedition',
  17: 'missionHolding',
};

export function FleetOperationsScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const [errMsg, setErrMsg] = useState<string | null>(null);
  // План 72.1.47: load/unload требуют current_planet_id — берём текущую
  // планету пользователя.
  const { planetId } = useResolvedPlanet();

  const fleetQ = useQuery({
    queryKey: QK.fleet(),
    queryFn: fetchFleet,
    refetchInterval: 5_000,
  });

  const recall = useMutation({
    mutationFn: (id: string) => recallFleet(id),
    onSuccess: () => void qc.invalidateQueries({ queryKey: QK.fleet() }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const load = useMutation({
    mutationFn: ({ id, m, s, h }: { id: string; m: number; s: number; h: number }) =>
      loadFleet(id, {
        current_planet_id: planetId ?? '',
        metal: m, silicon: s, hydrogen: h,
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.fleet() });
      if (planetId) void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const unload = useMutation({
    mutationFn: ({ id, m, s, h }: { id: string; m: number; s: number; h: number }) =>
      unloadFleet(id, {
        current_planet_id: planetId ?? '',
        metal: m, silicon: s, hydrogen: h,
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: QK.fleet() });
      if (planetId) void qc.invalidateQueries({ queryKey: QK.planet(planetId) });
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (fleetQ.isLoading) return <div className="idiv">…</div>;

  const fleets = fleetQ.data?.fleets ?? [];
  const slotsUsed = fleetQ.data?.slots_used ?? 0;
  const slotsMax = fleetQ.data?.slots_max ?? 0;

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={5}>
            {t('fleet', 'activeFleets', { count: fleets.length })}{' '}
            <span style={{ float: 'right' }}>
              {t('fleet', 'slots')} {slotsUsed}/{slotsMax}
            </span>
          </th>
        </tr>
        <tr>
          <th>{t('fleet', 'colMission')}</th>
          <th>{t('fleet', 'colDestination')}</th>
          <th>{t('fleet', 'colComposition')}</th>
          <th>{t('fleet', 'colArrival')}</th>
          <th>{t('alliance', 'operations')}</th>
        </tr>
      </thead>
      <tbody>
        {fleets.length === 0 && (
          <tr>
            <td colSpan={5} className="center">
              —
            </td>
          </tr>
        )}
        {fleets.map((f) => (
          <FleetRow
            key={f.id}
            fleet={f}
            onRecall={() => recall.mutate(f.id)}
            onLoad={(m, s, h) => load.mutate({ id: f.id, m, s, h })}
            onUnload={(m, s, h) => unload.mutate({ id: f.id, m, s, h })}
            disabled={recall.isPending || load.isPending || unload.isPending}
          />
        ))}
        {errMsg && (
          <tr>
            <td colSpan={5} className="center">
              <span className="false">{errMsg}</span>
            </td>
          </tr>
        )}
      </tbody>
    </table>
  );
}

function FleetRow({
  fleet,
  onRecall,
  onLoad,
  onUnload,
  disabled,
}: {
  fleet: Fleet;
  onRecall: () => void;
  onLoad: (m: number, s: number, h: number) => void;
  onUnload: (m: number, s: number, h: number) => void;
  disabled: boolean;
}) {
  const { t } = useTranslation();
  const missionKey = MISSION_LABEL_KEY[fleet.mission] ?? 'missionFallback';
  const total = Object.values(fleet.ships).reduce((s, n) => s + (n || 0), 0);
  const arrival =
    fleet.state === 'returning'
      ? fleet.return_at ?? fleet.arrive_at
      : fleet.arrive_at;
  // План 72.1.47: state='hold' — флот в HOLDING на цели, можно load/unload.
  const isHold = fleet.state === 'hold';
  const stateLabel =
    fleet.state === 'returning'
      ? t('fleet', 'stateReturning')
      : fleet.state === 'outbound'
      ? t('fleet', 'stateOutbound')
      : isHold
        ? t('fleet', 'stateHold') || '🛡 На цели'
        : t('fleet', 'stateArrived');

  // План 72.1.47: load/unload форма для HOLDING-флота.
  const [showLoadForm, setShowLoadForm] = useState(false);
  const [loadMode, setLoadMode] = useState<'load' | 'unload'>('unload');
  const [m, setM] = useState('0');
  const [s, setS] = useState('0');
  const [h, setH] = useState('0');

  function submit() {
    const mm = Math.max(0, Math.floor(Number(m) || 0));
    const ss = Math.max(0, Math.floor(Number(s) || 0));
    const hh = Math.max(0, Math.floor(Number(h) || 0));
    if (mm + ss + hh === 0) return;
    if (loadMode === 'load') onLoad(mm, ss, hh);
    else onUnload(mm, ss, hh);
    setShowLoadForm(false);
    setM('0');
    setS('0');
    setH('0');
  }

  return (
    <>
      <tr>
        <td>{t('fleet', missionKey)}</td>
        <td>
          {formatCoords(fleet.dst_galaxy, fleet.dst_system, fleet.dst_position)}
          {fleet.dst_is_moon ? ' 🌑' : ''}
        </td>
        <td className="center">{total}</td>
        <td>
          {stateLabel} · {formatDuration(secondsUntil(arrival))}
        </td>
        <td className="center">
          {fleet.state === 'outbound' && (
            <input
              type="button"
              className="button"
              value={t('fleet', 'recall')}
              disabled={disabled}
              onClick={() => {
                if (window.confirm(t('fleet', 'recall') + '?')) onRecall();
              }}
            />
          )}
          {isHold && (
            <input
              type="button"
              className="button"
              value={t('fleet', 'controlBtn') || '⚙ Управление'}
              disabled={disabled}
              onClick={() => setShowLoadForm((v) => !v)}
            />
          )}
          {!isHold && fleet.state !== 'outbound' && '—'}
        </td>
      </tr>
      {isHold && showLoadForm && (
        <tr>
          <td colSpan={5}>
            <div style={{ display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap' }}>
              <select
                value={loadMode}
                onChange={(e) => setLoadMode(e.target.value as 'load' | 'unload')}
              >
                <option value="unload">{t('fleet', 'unloadOpt') || 'Выгрузить с флота'}</option>
                <option value="load">{t('fleet', 'loadOpt') || 'Загрузить с планеты'}</option>
              </select>
              <span>
                {t('overview', 'metal') || 'Металл'}:{' '}
                <input
                  type="number"
                  min={0}
                  value={m}
                  onChange={(e) => setM(e.target.value)}
                  style={{ width: 100 }}
                />
              </span>
              <span>
                {t('overview', 'silicon') || 'Кремний'}:{' '}
                <input
                  type="number"
                  min={0}
                  value={s}
                  onChange={(e) => setS(e.target.value)}
                  style={{ width: 100 }}
                />
              </span>
              <span>
                {t('overview', 'hydrogen') || 'Водород'}:{' '}
                <input
                  type="number"
                  min={0}
                  value={h}
                  onChange={(e) => setH(e.target.value)}
                  style={{ width: 100 }}
                />
              </span>
              <input
                type="button"
                className="button"
                value={t('fleet', 'sendButton') || 'OK'}
                disabled={disabled}
                onClick={submit}
              />
            </div>
          </td>
        </tr>
      )}
    </>
  );
}
