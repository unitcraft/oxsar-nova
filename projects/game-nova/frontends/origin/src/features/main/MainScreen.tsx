// S-001 Main — главный экран после логина (план 72 Ф.2 Spring 1).
//
// Pixel-perfect зеркало legacy `main.tpl`:
//   1) ntable шапка: Главная планета (homeplanet) + universe-name.
//   2) Серверное время.
//   3) Блок «События» — текущие миссии в пути (из /api/fleet).
//   4) Блок «Непрочитанные сообщения» — счётчик из
//      /api/messages/unread-count.
//
// Endpoints:
//   GET /api/planets                    — список планет (используется
//                                         через useResolvedPlanet).
//   GET /api/fleet                      — активные флоты.
//   GET /api/messages/unread-count      — счётчик.
//
// Pixel-perfect: имена CSS-классов идентичны legacy (.ntable, .center,
// шапка <th colSpan>). Декоративные части legacy (BBCODE-юзербар,
// invite_friend, mini_games) исключены из Spring 1 и Ф.X плана 72.

import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { fetchFleet } from '@/api/fleet';
import { fetchUnreadCount } from '@/api/messages';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { useTranslation } from '@/i18n/i18n';
import { formatCoords, formatNumber, secondsUntil, formatDuration } from '@/lib/format';

export function MainScreen() {
  const { planet, planets, isLoading } = useResolvedPlanet();
  const { t } = useTranslation();

  const fleetQ = useQuery({
    queryKey: QK.fleet(),
    queryFn: fetchFleet,
    refetchInterval: 10_000,
  });

  const unreadQ = useQuery({
    queryKey: QK.unreadCount(),
    queryFn: fetchUnreadCount,
    refetchInterval: 30_000,
  });

  if (isLoading) {
    return <div className="idiv">…</div>;
  }
  if (planets.length === 0) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }
  const home = planet ?? planets[0];
  if (!home) {
    return <div className="idiv">{t('overview', 'noPlanets')}</div>;
  }

  const fleets = fleetQ.data?.fleets ?? [];
  const slotsUsed = fleetQ.data?.slots_used ?? 0;
  const slotsMax = fleetQ.data?.slots_max ?? 0;
  const unread = unreadQ.data?.count ?? 0;

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={3}>
            <b>{home.name}</b>{' '}
            {formatCoords(home.galaxy, home.system, home.position)}
          </th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td>{t('main', 'curHomePlanet')}</td>
          <td colSpan={2}>
            {home.name}{' '}
            <span style={{ float: 'right' }}>Oxsar Classic</span>
          </td>
        </tr>
        <tr>
          <td>{t('main', 'serverTime')}</td>
          <td colSpan={2}>
            <span id="serverwatch">{new Date().toLocaleString('ru-RU')}</span>
          </td>
        </tr>
        <tr>
          <td>{t('overview', 'resMetalLabel')}</td>
          <td colSpan={2}>{formatNumber(home.metal)}</td>
        </tr>
        <tr>
          <td>{t('overview', 'resSiliconLabel')}</td>
          <td colSpan={2}>{formatNumber(home.silicon)}</td>
        </tr>
        <tr>
          <td>{t('overview', 'resHydrogenLabel')}</td>
          <td colSpan={2}>{formatNumber(home.hydrogen)}</td>
        </tr>

        {unread > 0 && (
          <tr>
            <td colSpan={3} className="center">
              <Link to="/empire">
                {t('overview', 'unreadPrefix')} {unread}{' '}
                {unread === 1
                  ? t('overview', 'unreadSingle')
                  : t('overview', 'unreadPlural')}
              </Link>
            </td>
          </tr>
        )}

        <tr>
          <th colSpan={3}>
            {t('main', 'events')}{' '}
            <span style={{ float: 'right' }}>
              {t('fleet', 'slots')} {slotsUsed}/{slotsMax}
            </span>
          </th>
        </tr>
        {fleets.length === 0 ? (
          <tr>
            <td colSpan={3} className="center">
              —
            </td>
          </tr>
        ) : (
          fleets.map((f) => (
            <tr key={f.id}>
              <td className="center">
                {f.state === 'returning'
                  ? t('overview', 'fleetReturning')
                  : t('overview', 'fleetOutbound')}
              </td>
              <td colSpan={2}>
                {formatCoords(f.dst_galaxy, f.dst_system, f.dst_position)}{' '}
                {formatDuration(
                  secondsUntil(
                    f.state === 'returning' ? (f.return_at ?? f.arrive_at) : f.arrive_at,
                  ),
                )}
              </td>
            </tr>
          ))
        )}
      </tbody>
    </table>
  );
}
