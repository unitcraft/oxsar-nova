// S-001 Main — главный экран после логина.
//
// План 72.1 ч.17: pixel-perfect зеркало legacy `main.tpl`. По сравнению
// с ч.16 добавлены:
//   - 3-колонный блок Луна / Планет-картинка / planetMainSelection (sidebar
//     других планет)
//   - Военный/накопленный опыт, очки + ранг, max_points, dm_points
//   - Шахтёр-уровень + прогресс к следующему
//   - Уровень межгалактических исследований
//
// Источники данных:
//   GET /api/me            — MeInfo (расширен в плане 72.1 ч.17)
//   GET /api/planets       — useResolvedPlanet
//   GET /api/fleet         — fleet missions (events)
//   GET /api/messages/unread-count — счётчик непрочитанных
//   GET /api/professions/me — текущая профессия (label из конфига)

import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { fetchFleet } from '@/api/fleet';
import { fetchUnreadCount } from '@/api/messages';
import { fetchMe } from '@/api/me';
import { fetchProfessionMe } from '@/api/profession';
import { QK } from '@/api/query-keys';
import { useResolvedPlanet } from '@/features/common/useResolvedPlanet';
import { useTranslation } from '@/i18n/i18n';
import {
  formatCoords,
  formatNumber,
  secondsUntil,
  formatDuration,
} from '@/lib/format';
import {
  planetImageUrl,
  planetImageSmallUrl,
  moonImageUrl,
  moonImageSmallUrl,
} from '@/lib/planet-image';

export function MainScreen() {
  const { planet, planets, isLoading } = useResolvedPlanet();
  const { t } = useTranslation();

  const meQ = useQuery({
    queryKey: QK.me(),
    queryFn: fetchMe,
    staleTime: 30_000,
  });

  // Профессия — отдельный endpoint (label из конфига, локализован на бэке).
  const profQ = useQuery({
    queryKey: QK.professionMe(),
    queryFn: fetchProfessionMe,
    staleTime: 60_000,
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
  const me = meQ.data;

  const professionLabel =
    profQ.data?.label && profQ.data.label !== ''
      ? profQ.data.label
      : t('profession', 'title');

  // Планет-картинка — детерминированная по home.id.
  const planetImg = planetImageUrl(home.planet_type, home.id);
  // Луна — пока не отдаём в /api/planets, ставим placeholder (если будет
  // moon-планета в общем списке планет — отрисуем sidebar-иконкой).
  const moonExists = planets.some(
    (p) => p.is_moon && p.galaxy === home.galaxy && p.system === home.system,
  );

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

        {/*
          3-колонный блок (legacy main.tpl §107-117):
          Луна | Планет-картинка | sidebar других планет.
        */}
        <tr>
          <td className="center" style={{ width: '33%' }}>
            {moonExists ? (
              <>
                {t('global', 'moon')}
                <br />
                <img
                  src={moonImageUrl()}
                  alt={t('global', 'moon')}
                  style={{ maxWidth: 200 }}
                />
              </>
            ) : (
              <small style={{ color: '#888' }}>—</small>
            )}
          </td>
          <td className="center" style={{ width: '33%' }}>
            <img
              src={planetImg}
              alt={home.name}
              style={{ maxWidth: 200 }}
            />
          </td>
          <td className="center" style={{ width: '34%' }}>
            <PlanetSidebar
              currentPlanetId={home.id}
              planets={planets}
            />
          </td>
        </tr>

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

        {/*
          Боевой опыт / очки / шахтёр / межгал-исследования
          (legacy main.tpl §132-165). Если /api/me ещё не отдал —
          показываем placeholder, не таблицу-с-нулями.
        */}
        {me && (
          <>
            <tr>
              <td>{t('main', 'battleExperience')}</td>
              <td colSpan={2}>{formatNumber(me.combat_experience)}</td>
            </tr>
            <tr>
              <td>{t('main', 'battleActiveExperience')}</td>
              <td colSpan={2}>{formatNumber(me.accumulated_experience)}</td>
            </tr>
            <tr>
              <td>{t('main', 'points')}</td>
              <td colSpan={2}>
                {formatNumber(me.points)} ({t('main', 'rankLabel')}: {me.rank})
              </td>
            </tr>
            {me.max_points > me.points && (
              <tr>
                <td>{t('main', 'maxPoints')}</td>
                <td colSpan={2}>{formatNumber(me.max_points)}</td>
              </tr>
            )}
            {me.dm_points > 0 && (
              <tr>
                <td>{t('main', 'dmPoints')}</td>
                <td colSpan={2}>{formatNumber(me.dm_points)}</td>
              </tr>
            )}
            <tr>
              <td>
                {t('main', 'minerLevel', { level: String(me.miner_level) })}
              </td>
              <td colSpan={2}>
                {formatNumber(me.miner_points)} /{' '}
                {formatNumber(me.miner_need_points)}
              </td>
            </tr>
            {me.intergalactic_research_level > 0 && (
              <tr>
                <td>{t('main', 'intergalacticResearchLevel')}</td>
                <td colSpan={2}>{me.intergalactic_research_level}</td>
              </tr>
            )}
          </>
        )}
      </tbody>
    </table>
  );
}

// Sidebar других планет (legacy `planetMainSelection`-include).
// Список малых иконок-планет с подписью имени + координат, кликабельных
// для смены текущей планеты через ?planet_id=...
function PlanetSidebar({
  currentPlanetId,
  planets,
}: {
  currentPlanetId: string;
  planets: Array<{
    id: string;
    name: string;
    galaxy: number;
    system: number;
    position: number;
    planet_type?: string | null;
    is_moon?: boolean;
  }>;
}) {
  const others = planets.filter(
    (p) => p.id !== currentPlanetId && !p.is_moon,
  );
  if (others.length === 0) {
    return <small style={{ color: '#888' }}>—</small>;
  }
  return (
    <ul
      style={{
        listStyle: 'none',
        padding: 0,
        margin: 0,
        fontSize: '0.85em',
      }}
    >
      {others.map((p) => (
        <li key={p.id} style={{ marginBottom: 4 }}>
          <Link
            to={`?planet_id=${encodeURIComponent(p.id)}`}
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: 4,
            }}
          >
            <img
              src={
                p.is_moon
                  ? moonImageSmallUrl()
                  : planetImageSmallUrl(p.planet_type ?? null, p.id)
              }
              alt={p.name}
              width={24}
              height={24}
            />
            <span>
              {p.name} [{p.galaxy}:{p.system}:{p.position}]
            </span>
          </Link>
        </li>
      ))}
    </ul>
  );
}
