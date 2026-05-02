// План 72.1.20 — MonitorPlanet: legacy паритет (range + hydrogen + events).
//
// URL: /monitor-planet?id=<planet_id>
//
// Pixel-perfect клон legacy `monitor_planet.tpl` + `MonitorPlanet.class.php`:
//   - заголовок «Мониторинг планеты <name> [g:s:p] игрока <username>»
//   - таблица event'ов (timer + сообщение).
//
// КАЖДОЕ сканирование списывает 5000H с источника (легаси
// STAR_SURVEILLANCE_CONSUMPTION). Поэтому НЕ делаем auto-refetch — игрок
// нажимает «Сканировать» каждый раз вручную.

import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import { fetchMonitorPlanet, type MonitorResult } from '@/api/monitor';
import type { ApiError } from '@/api/client';
import type { PhalanxScan } from '@/api/types';
import { ConfirmDialog, useConfirm } from '@/features/common/ConfirmDialog';
import { useTranslation } from '@/i18n/i18n';
import { formatDuration, secondsUntil } from '@/lib/format';

export function MonitorPlanetScreen() {
  const { t } = useTranslation();
  const [params] = useSearchParams();
  const planetId = params.get('id') ?? '';

  const [data, setData] = useState<MonitorResult | null>(null);
  const [errMsg, setErrMsg] = useState<string | null>(null);
  // План 72.1.53: in-game confirm-диалог вместо window.confirm.
  const { confirm, dialogProps } = useConfirm();

  const scanMut = useMutation({
    mutationFn: () => fetchMonitorPlanet(planetId),
    onSuccess: (r) => {
      setData(r);
      setErrMsg(null);
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (!planetId) {
    return (
      <div className="idiv">
        {t('monitorPlanet', 'noPlanetSelected') ||
          'Не выбрана планета. Откройте /monitor-planet?id=<planet_id>.'}
      </div>
    );
  }

  return (
    <>
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('monitorPlanet', 'title') || 'Мониторинг'}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td colSpan={2}>
              <button
                type="button"
                className="button"
                disabled={scanMut.isPending}
                onClick={async () => {
                  // План 72.1.47 + 72.1.53: pre-scan confirmation
                  // для 5000H (legacy STAR_SURVEILLANCE_CONSUMPTION
                  // списывается безусловно при каждом нажатии).
                  // Стилизованный in-game dialog заменил window.confirm.
                  const ok = await confirm({
                    title: (t('monitorPlanet', 'scanBtn') as string) ||
                      'Сканировать',
                    message:
                      (t('monitorPlanet', 'scanConfirm') as string) ||
                      'Сканирование стоит 5000H. Продолжить?',
                    confirmLabel: 'OK',
                  });
                  if (!ok) return;
                  scanMut.mutate();
                }}
              >
                {scanMut.isPending
                  ? '…'
                  : t('monitorPlanet', 'scanBtn') || 'Сканировать (5000H)'}
              </button>
              {errMsg && (
                <>
                  {' '}
                  <span className="false">{errMsg}</span>
                </>
              )}
            </td>
          </tr>
        </tbody>
      </table>

      {data && (
        <>
          <table className="ntable">
            <thead>
              <tr>
                <th colSpan={2}>
                  {data.target_planet.name} [{data.target_planet.galaxy}:
                  {data.target_planet.system}:{data.target_planet.position}]
                  {data.target_planet.username
                    ? ` — ${data.target_planet.username}`
                    : ''}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td>{t('monitorPlanet', 'scannerLabel') || 'Источник'}</td>
                <td>
                  {data.scanner.name} [{data.scanner.galaxy}:
                  {data.scanner.system}:{data.scanner.position}]
                </td>
              </tr>
              {data.detected && (
                <tr>
                  <td colSpan={2} className="false2">
                    {t('monitorPlanet', 'detectedNote') ||
                      'Цель обнаружила сканирование (отправлено уведомление).'}
                  </td>
                </tr>
              )}
            </tbody>
          </table>

          <table className="ntable">
            <thead>
              <tr>
                <th colSpan={2}>
                  {t('monitorPlanet', 'eventsTitle') || 'События флота'}
                </th>
              </tr>
            </thead>
            <tbody>
              {data.events.length === 0 ? (
                <tr>
                  <td colSpan={2} className="center">
                    {t('mission', 'noMatchesFound') || 'Нет событий'}
                  </td>
                </tr>
              ) : (
                data.events.map((e) => <EventRow key={e.fleet_id} e={e} />)
              )}
            </tbody>
          </table>
        </>
      )}
      {/* План 72.1.53: in-game confirm-dialog (см. ConfirmDialog.tsx). */}
      <ConfirmDialog {...dialogProps} />
    </>
  );
}

function EventRow({ e }: { e: PhalanxScan }) {
  const arriveSec = secondsUntil(e.arrive_at);
  const isReturning = e.state === 'returning';
  const targetSec = isReturning ? secondsUntil(e.return_at) : arriveSec;
  return (
    <tr>
      <td width="100px" align="center">
        <span>{formatDuration(targetSec)}</span>
      </td>
      <td>
        <span className={isReturning ? 'true' : 'false'}>
          #{e.mission} {e.owner_name || '—'}: [{e.src_galaxy}:{e.src_system}:
          {e.src_position}] → [{e.dst_galaxy}:{e.dst_system}:{e.dst_position}
          {e.dst_is_moon ? '🌙' : ''}] — {e.state}
        </span>
      </td>
    </tr>
  );
}
