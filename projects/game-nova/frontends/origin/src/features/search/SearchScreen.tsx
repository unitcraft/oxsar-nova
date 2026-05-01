// S-039 Search — глобальный поиск (план 72 Ф.5 Spring 4).
//
// Pixel-perfect зеркало legacy `searchheader.tpl` + `player_search_result.tpl`
// + `ally_search_result.tpl`:
//   Шапка ntable с select (where=players|planets|alliances) и input поиска.
//   Результаты в таблице ниже (тот же ntable стиль, foreach[result]).
//
// Endpoint: GET /api/search?q=&type=

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { search } from '@/api/search';
import type { SearchType } from '@/api/types';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';

// План 72.1.16: legacy выводит «N min ago» / «online» через
// `getTimeTerm(time() - last)`. Воспроизводим компактно.
function formatLastSeen(
  lastSeen: string | undefined,
  t: (g: string, k: string, v?: Record<string, string | number>) => string,
): string {
  if (!lastSeen) return '—';
  const minutes = Math.floor((Date.now() - new Date(lastSeen).getTime()) / 60_000);
  if (minutes < 5) return t('friends', 'statusOnline');
  if (minutes < 60) return t('friends', 'statusMinAgo', { n: minutes });
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return t('friends', 'statusHourAgo', { n: hours });
  const days = Math.floor(hours / 24);
  return t('friends', 'statusDayAgo', { n: days });
}

export function SearchScreen() {
  const { t } = useTranslation();
  const [type, setType] = useState<SearchType>('player');
  const [draft, setDraft] = useState('');
  const [committed, setCommitted] = useState('');

  const q = useQuery({
    queryKey: QK.search(type, committed),
    queryFn: () => search(committed, type),
    enabled: committed.length >= 2,
  });

  function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setCommitted(draft.trim());
  }

  const data = q.data;

  return (
    <>
      <form onSubmit={onSubmit}>
        <table className="ntable">
          <thead>
            <tr>
              <th>{t('statistics', 'browseUniverse')}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>
                <select
                  name="where"
                  value={type}
                  onChange={(e) => setType(e.target.value as SearchType)}
                >
                  <option value="player">{t('statistics', 'players')}</option>
                  <option value="planet">{t('statistics', 'planets')}</option>
                  <option value="alliance">{t('alliance', 'alliances')}</option>
                </select>{' '}
                <input
                  type="text"
                  name="what"
                  maxLength={128}
                  value={draft}
                  placeholder={t('search', 'placeholder')}
                  className="searchInput"
                  onChange={(e) => setDraft(e.target.value)}
                />{' '}
                <input
                  type="submit"
                  value="OK"
                  className="button"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </form>

      {committed.length >= 2 && (
        <ResultsTable type={type} data={data} loading={q.isFetching} />
      )}
      {committed.length > 0 && committed.length < 2 && (
        <div className="idiv">{t('search', 'hint')}</div>
      )}
    </>
  );
}

function ResultsTable({
  type,
  data,
  loading,
}: {
  type: SearchType;
  data: import('@/api/types').SearchResults | undefined;
  loading: boolean;
}) {
  const { t } = useTranslation();
  if (loading && !data) {
    return <div className="idiv">{t('search', 'searching')}</div>;
  }
  if (!data) return null;

  if (type === 'player') {
    return (
      <table className="ntable">
        <thead>
          <tr>
            <th>{t('overview', 'username') || 'Игрок'}</th>
            <th>{t('overview', 'points') || 'Очки'}</th>
            <th>{t('alliance', 'alliance') || 'Альянс'}</th>
            <th>{t('main', 'curHomePlanet') || 'Планета'}</th>
            <th>{t('mission', 'target') || 'Координаты'}</th>
            <th>{t('friends', 'colStatus') || 'Активность'}</th>
          </tr>
        </thead>
        <tbody>
          {data.players.length === 0 ? (
            <tr>
              <td colSpan={6} className="center">
                {t('search', 'notFound')}
              </td>
            </tr>
          ) : (
            data.players.map((p) => (
              <tr
                key={p.user_id}
                style={p.banned ? { textDecoration: 'line-through', opacity: 0.6 } : undefined}
              >
                <td>
                  {p.username}
                  {p.banned && ' '}
                  {p.banned && (
                    <span className="false">{t('search', 'markerBanned')}</span>
                  )}
                </td>
                <td>{Math.round(p.points)}</td>
                <td>{p.alliance_tag ?? '—'}</td>
                <td>{p.home_planet ?? '—'}</td>
                <td>
                  {p.home_galaxy !== undefined ? (
                    <Link to={`/galaxy/${p.home_galaxy}/${p.home_system}`}>
                      [{p.home_galaxy}:{p.home_system}:{p.home_position}]
                    </Link>
                  ) : (
                    '—'
                  )}
                </td>
                <td>{formatLastSeen(p.last_seen, t)}</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    );
  }

  if (type === 'alliance') {
    return (
      <table className="ntable">
        <thead>
          <tr>
            <th>{t('alliance', 'tag') || 'Тег'}</th>
            <th>{t('alliance', 'name') || 'Альянс'}</th>
            <th>{t('alliance', 'totalMembers') || 'Членов'}</th>
            <th>{t('overview', 'points') || 'Очки'}</th>
          </tr>
        </thead>
        <tbody>
          {data.alliances.length === 0 ? (
            <tr>
              <td colSpan={4} className="center">
                {t('search', 'notFound')}
              </td>
            </tr>
          ) : (
            data.alliances.map((a) => (
              <tr key={a.tag}>
                <td>
                  <Link to={`/alliance/${a.tag}`}>{a.tag}</Link>
                </td>
                <td>{a.name}</td>
                <td className="center">{a.members}</td>
                <td>{Math.round(a.points)}</td>
              </tr>
            ))
          )}
        </tbody>
      </table>
    );
  }

  // planet
  return (
    <table className="ntable">
      <thead>
        <tr>
          <th>{t('main', 'curHomePlanet') || 'Планета'}</th>
          <th>{t('mission', 'target') || 'Координаты'}</th>
          <th>{t('overview', 'username') || 'Владелец'}</th>
        </tr>
      </thead>
      <tbody>
        {data.planets.length === 0 ? (
          <tr>
            <td colSpan={3} className="center">
              {t('search', 'notFound')}
            </td>
          </tr>
        ) : (
          data.planets.map((p) => {
            // Legacy: " (HP)" если planet=user.hp, " (MOON)" если ismoon.
            const marker = p.is_home
              ? ' ' + t('search', 'markerHomePlanet')
              : p.is_moon
                ? ' ' + t('search', 'markerMoon')
                : '';
            return (
              <tr key={p.planet_id}>
                <td>
                  {p.name}
                  {marker}
                </td>
                <td>
                  <Link to={`/galaxy/${p.galaxy}/${p.system}`}>
                    [{p.galaxy}:{p.system}:{p.position}]
                  </Link>
                </td>
                <td>{p.owner}</td>
              </tr>
            );
          })
        )}
      </tbody>
    </table>
  );
}
