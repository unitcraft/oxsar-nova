// S-001 Main — главный экран после логина.
//
// План 72.1 ч.16: pixel-perfect зеркало legacy `main.tpl` —
// все строки таблицы:
//   1. Заголовок: <currentPlanet> [coords] (<username>)
//   2. Главная планета — homeplanet + universe-name
//   3. Профессия — link на /profession
//   4. Серверное время — текущий локальный clock
//   5. (events) — fleet missions
//   6. Параметры планеты:
//      - Диаметр: 18.800 км (застроенная территория: X из Y полей)
//      - Температура: min..max °C
//      - Координаты: [g:s:p]
//
// Военный/накопленный опыт, очки, шахтёр-уровень, межгал.исследования —
// требуют расширения /api/me на бэке (отдельная задача).
//
// Endpoints:
//   GET /api/planets        — список планет (через useResolvedPlanet).
//   GET /api/me             — username, profession, credit (через TopHeader).
//   GET /api/fleet          — fleet missions (events).
//   GET /api/messages/unread-count — счётчик непрочитанных.

import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { fetchFleet } from '@/api/fleet';
import { fetchUnreadCount } from '@/api/messages';
import { fetchMe } from '@/api/me';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { useTranslation } from '@/i18n/i18n';
import {
  formatCoords,
  formatNumber,
  secondsUntil,
  formatDuration,
} from '@/lib/format';

const PROFESSION_LABEL: Record<string, string> = {
  none: '—',
  miner: 'Шахтёр',
  warrior: 'Воин',
  scientist: 'Учёный',
  builder: 'Строитель',
  trader: 'Торговец',
};

export function MainScreen() {
  const { planet, planets, isLoading } = useResolvedPlanet();
  const { t } = useTranslation();

  const meQ = useQuery({
    queryKey: QK.me(),
    queryFn: fetchMe,
    staleTime: 30_000,
  });

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

  const username = meQ.data?.username ?? '';
  const professionKey = meQ.data?.profession ?? 'none';
  const professionLabel =
    PROFESSION_LABEL[professionKey] ?? professionKey;

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th colSpan={3}>
            <b>{home.name}</b>{' '}
            {formatCoords(home.galaxy, home.system, home.position)}
            {username && ` (${username})`}
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
          <td>{t('global', 'menuProfession')}</td>
          <td colSpan={2}>
            <Link to="/profession">{professionLabel}</Link>
          </td>
        </tr>
        <tr>
          <td>{t('main', 'serverTime')}</td>
          <td colSpan={2}>
            <span id="serverwatch">
              {new Date().toLocaleString('ru-RU')}
            </span>
          </td>
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
                    f.state === 'returning'
                      ? (f.return_at ?? f.arrive_at)
                      : f.arrive_at,
                  ),
                )}
              </td>
            </tr>
          ))
        )}

        {/* План 72.1 ч.16: блок параметров планеты из legacy main.tpl. */}
        <tr>
          <td>{t('overview', 'planetDiameter')}</td>
          <td colSpan={2}>
            {formatNumber(home.diameter)} {t('overview', 'diameterKm')} (
            {t('main', 'planetOccupiedFields', {
              used: String(home.used_fields),
              max: String(home.max_fields),
            })}
            )
          </td>
        </tr>
        <tr>
          <td>{t('overview', 'planetTemp')}</td>
          <td colSpan={2}>
            {home.temp_min} °C … {home.temp_max} °C
          </td>
        </tr>
        <tr>
          <td>{t('main', 'position')}</td>
          <td colSpan={2}>
            <Link to={`/galaxy/${home.galaxy}/${home.system}`}>
              {formatCoords(home.galaxy, home.system, home.position)}
            </Link>
          </td>
        </tr>
      </tbody>
    </table>
  );
}
