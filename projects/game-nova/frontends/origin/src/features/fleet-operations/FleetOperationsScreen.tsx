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
import { fetchFleet, recallFleet } from '@/api/fleet';
import type { ApiError } from '@/api/client';
import { QK } from '@/api/query-keys';
import type { Fleet } from '@/api/types';
import { formatCoords, formatDuration, secondsUntil } from '@/lib/format';
import { useTranslation } from '@/i18n/i18n';
import { useState } from 'react';

const MISSION_LABEL_KEY: Record<number, string> = {
  6: 'missionAttack',
  7: 'missionExpedition',
  8: 'missionTransport',
  9: 'missionRebase',
  10: 'missionColonize',
  11: 'missionRecycle',
  12: 'missionSpy',
  15: 'missionAttack', // legacy alliance attack — отображаем как Атака
};

export function FleetOperationsScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const [errMsg, setErrMsg] = useState<string | null>(null);

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
            disabled={recall.isPending}
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
  disabled,
}: {
  fleet: Fleet;
  onRecall: () => void;
  disabled: boolean;
}) {
  const { t } = useTranslation();
  const missionKey = MISSION_LABEL_KEY[fleet.mission] ?? 'missionFallback';
  const total = Object.values(fleet.ships).reduce((s, n) => s + (n || 0), 0);
  const arrival =
    fleet.state === 'returning'
      ? fleet.return_at ?? fleet.arrive_at
      : fleet.arrive_at;
  const stateLabel =
    fleet.state === 'returning'
      ? t('fleet', 'stateReturning')
      : fleet.state === 'outbound'
      ? t('fleet', 'stateOutbound')
      : t('fleet', 'stateArrived');

  return (
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
        {fleet.state === 'outbound' ? (
          <input
            type="button"
            className="button"
            value={t('fleet', 'recall')}
            disabled={disabled}
            onClick={() => {
              if (window.confirm(t('fleet', 'recall') + '?')) onRecall();
            }}
          />
        ) : (
          '—'
        )}
      </td>
    </tr>
  );
}
