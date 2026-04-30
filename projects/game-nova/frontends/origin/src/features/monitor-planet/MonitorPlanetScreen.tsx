// S-MP MonitorPlanet — мониторинг активностей флота (план 72.1 ч.20.5).
// Pixel-perfect клон legacy monitor_planet.tpl.
//
// Структура: одна таблица «Активности флота» с timer + сообщением.

import { useQuery } from '@tanstack/react-query';
import { fetchFleet } from '@/api/fleet';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { formatDuration, secondsUntil } from '@/lib/format';

export function MonitorPlanetScreen() {
  const { t } = useTranslation();
  const q = useQuery({
    queryKey: QK.fleet(),
    queryFn: fetchFleet,
    refetchInterval: 5000,
  });

  const fleets = q.data?.fleets ?? [];

  return (
    <table className="ntable">
      <tbody>
        <tr>
          <th colSpan={2}>Активности флота</th>
        </tr>
        {fleets.length === 0 && (
          <tr>
            <td>&nbsp;</td>
            <td>{t('mission', 'noMatchesFound')}</td>
          </tr>
        )}
        {fleets.map((f) => (
          <tr key={f.id}>
            <td width="100px" align="center">
              <span>{formatDuration(secondsUntil(f.arrive_at))}</span>
            </td>
            <td>
              <span className={f.state === 'outbound' ? 'false' : 'true'}>
                #{f.mission}: [{f.dst_galaxy}:{f.dst_system}:{f.dst_position}]
                {' '}— {f.state}
              </span>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
