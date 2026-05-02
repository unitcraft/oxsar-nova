// S-023 Battle stats (план 72.1 ч.20.8 — battle viewer).
//
// Pixel-perfect клон legacy battlestats.tpl + список реальных боёв
// + кнопка «Просмотр» открывает /battle-reports/:id.

import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { fetchMyBattles } from '@/api/battles';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';

interface HighscoreEntry {
  user_id: string;
  username: string;
  score: number;
  rank: number;
}

function fmtDate(iso: string): string {
  try {
    const d = new Date(iso);
    return d.toLocaleString('ru-RU', {
      day: '2-digit',
      month: '2-digit',
      year: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    });
  } catch {
    return iso;
  }
}

export function BattleStatsScreen() {
  const { t } = useTranslation();
  // Все 11 фильтров (план 72.1.10 — порт legacy showBattles).
  // Дефолты — соответствуют legacy index() (показываем всё кроме
  // пришельцев и moon-only).
  const [dateFirst, setDateFirst] = useState('');
  const [dateLast, setDateLast] = useState('');
  const [userFilter, setUserFilter] = useState('');
  const [allianceFilter, setAllianceFilter] = useState('');
  const [showDrawn, setShowDrawn] = useState(true);
  const [showAliens, setShowAliens] = useState(false);
  const [showNoDestroyed, setShowNoDestroyed] = useState(true);
  const [newMoon, setNewMoon] = useState(false);
  const [moonBattle, setMoonBattle] = useState(false);
  // План 72.1.10 wave 2: добавлены legacy sort-поля outcome (winner)
  // и moon (is_moon).
  // План 72.1.10 wave 3: добавлен planet_name (через JOIN planets).
  const [sortField, setSortField] = useState<
    'date' | 'rounds' | 'debris' | 'loot' | 'outcome' | 'moon' | 'planet_name'
  >('date');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('desc');
  const [cursor, setCursor] = useState<string | undefined>(undefined);

  const meQ = useQuery({
    queryKey: ['highscore', 'me'],
    queryFn: () => api.get<HighscoreEntry>('/api/highscore/me'),
    staleTime: 60_000,
  });

  // Сборка набора фильтров для useQuery: пустые поля не попадают в
  // params, чтобы дефолты сервера оставались дефолтами.
  const filterParams: import('@/api/battles').BattleListFilters = {
    limit: 20,
    show_drawn: showDrawn,
    show_aliens: showAliens,
    show_no_destroyed: showNoDestroyed,
    new_moon: newMoon,
    moon_battle: moonBattle,
    sort_field: sortField,
    sort_order: sortOrder,
  };
  if (cursor !== undefined) filterParams.cursor = cursor;
  if (dateFirst) filterParams.date_min = new Date(dateFirst).toISOString();
  if (dateLast) filterParams.date_max = new Date(dateLast).toISOString();
  if (userFilter) filterParams.user_filter = userFilter;
  if (allianceFilter) filterParams.alliance_filter = allianceFilter;

  // queryKey включает все фильтры — при их изменении query
  // автоматически перезапускается.
  const queryKey = [
    'my-battles',
    cursor ?? '',
    dateFirst, dateLast, userFilter, allianceFilter,
    showDrawn, showAliens, showNoDestroyed, newMoon, moonBattle,
    sortField, sortOrder,
  ];
  const battlesQ = useQuery({
    queryKey,
    queryFn: () => fetchMyBattles(filterParams),
  });

  const battles = battlesQ.data?.battles ?? [];
  const nextCursor = battlesQ.data?.next_cursor;

  return (
    <form method="post" onSubmit={(ev) => ev.preventDefault()}>
      <table className="ntable">
        <thead>
          <tr>
            <th>{t('battlestats', 'title')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <table>
                <tbody>
                  <tr>
                    <td><label htmlFor="date_last">{t('statistics', 'dateLast')}</label></td>
                    <td>
                      <input
                        id="date_last"
                        name="date_last"
                        type="text"
                        value={dateLast}
                        onChange={(e) => setDateLast(e.target.value)}
                      />
                    </td>
                  </tr>
                  <tr>
                    <td><label htmlFor="date_first">{t('statistics', 'dateFirst')}</label></td>
                    <td>
                      <input
                        id="date_first"
                        name="date_first"
                        type="text"
                        value={dateFirst}
                        onChange={(e) => setDateFirst(e.target.value)}
                      />
                    </td>
                  </tr>
                  <tr>
                    <td><label htmlFor="user_filter">{t('statistics', 'bsUserFilter')}</label></td>
                    <td>
                      <input
                        id="user_filter"
                        name="user_filter"
                        type="text"
                        value={userFilter}
                        onChange={(e) => setUserFilter(e.target.value)}
                      />
                    </td>
                  </tr>
                  <tr>
                    <td><label htmlFor="alliance_filter">{t('statistics', 'bsAllianceFilter')}</label></td>
                    <td>
                      <input
                        id="alliance_filter"
                        name="alliance_filter"
                        type="text"
                        value={allianceFilter}
                        onChange={(e) => setAllianceFilter(e.target.value)}
                      />
                    </td>
                  </tr>
                  <tr>
                    <td colSpan={2}>
                      <input
                        id="show_drawn"
                        type="checkbox"
                        checked={showDrawn}
                        onChange={(e) => setShowDrawn(e.target.checked)}
                      />{' '}
                      <label htmlFor="show_drawn">{t('statistics', 'bsShowDrawn')}</label>
                    </td>
                  </tr>
                  <tr>
                    <td colSpan={2}>
                      <input
                        id="show_aliens"
                        type="checkbox"
                        checked={showAliens}
                        onChange={(e) => setShowAliens(e.target.checked)}
                      />{' '}
                      <label htmlFor="show_aliens">{t('statistics', 'bsShowUfoBattles')}</label>
                    </td>
                  </tr>
                  <tr>
                    <td colSpan={2}>
                      <input
                        id="show_no_destroyed"
                        type="checkbox"
                        checked={showNoDestroyed}
                        onChange={(e) => setShowNoDestroyed(e.target.checked)}
                      />{' '}
                      <label htmlFor="show_no_destroyed">{t('statistics', 'bsShowNoDestroyed')}</label>
                    </td>
                  </tr>
                  <tr>
                    <td colSpan={2}>
                      <input
                        id="new_moon"
                        type="checkbox"
                        checked={newMoon}
                        onChange={(e) => setNewMoon(e.target.checked)}
                      />{' '}
                      <label htmlFor="new_moon">{t('statistics', 'bsNewMoon')}</label>
                    </td>
                  </tr>
                  <tr>
                    <td colSpan={2}>
                      <input
                        id="moon_battle"
                        type="checkbox"
                        checked={moonBattle}
                        onChange={(e) => setMoonBattle(e.target.checked)}
                      />{' '}
                      <label htmlFor="moon_battle">{t('statistics', 'bsMoonBattle')}</label>
                    </td>
                  </tr>
                  <tr>
                    <td><label htmlFor="sort_field">{t('statistics', 'bsSortBy')}</label></td>
                    <td>
                      <select
                        id="sort_field"
                        value={sortField}
                        onChange={(e) => setSortField(e.target.value as typeof sortField)}
                      >
                        <option value="date">{t('statistics', 'bsSortDate')}</option>
                        <option value="rounds">{t('statistics', 'bsSortRounds')}</option>
                        <option value="debris">{t('statistics', 'bsSortDebris')}</option>
                        <option value="loot">{t('statistics', 'bsSortLoot')}</option>
                        <option value="outcome">{t('statistics', 'bsSortOutcome')}</option>
                        <option value="moon">{t('statistics', 'bsSortMoon')}</option>
                        <option value="planet_name">{t('statistics', 'bsSortPlanetName')}</option>
                      </select>{' '}
                      <select
                        value={sortOrder}
                        onChange={(e) => setSortOrder(e.target.value as typeof sortOrder)}
                      >
                        <option value="desc">{t('statistics', 'bsSortDesc')}</option>
                        <option value="asc">{t('statistics', 'bsSortAsc')}</option>
                      </select>
                    </td>
                  </tr>
                </tbody>
              </table>
            </td>
          </tr>
          <tr>
            <td className="center">
              <input
                type="submit"
                name="go"
                className="button"
                value="OK"
              />
            </td>
          </tr>
        </tbody>
      </table>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={6} className="center">
              {meQ.data ? (
                <>
                  {t('battlestats', 'statTotal')}: <b>{meQ.data.score}</b> ·{' '}
                  {t('alliance', 'rank')}: <b>{meQ.data.rank}</b>
                </>
              ) : '…'}
            </th>
          </tr>
          <tr>
            <th>{t('battlestats', 'colDate')}</th>
            <th>{t('battlestats', 'colOpponent') ?? 'Противник'}</th>
            <th>{t('battlestats', 'colResult') ?? 'Результат'}</th>
            <th>{t('battlestats', 'colLoot') ?? 'Трофеи'}</th>
            <th>{t('mission', 'debris') ?? 'Обломки'}</th>
            <th>&nbsp;</th>
          </tr>
        </thead>
        <tbody>
          {battlesQ.isLoading && (
            <tr><td colSpan={6} className="center">…</td></tr>
          )}
          {!battlesQ.isLoading && battles.length === 0 && (
            <tr>
              <td colSpan={6} className="center">
                {t('battlestats', 'empty')}
              </td>
            </tr>
          )}
          {battles.map((b) => {
            const myWin =
              (b.is_attacker && b.winner === 'attackers') ||
              (!b.is_attacker && b.winner === 'defenders');
            const isDraw = b.winner === 'draw';
            const resultLabel = isDraw
              ? (t('mission', 'draw') ?? 'Ничья')
              : myWin
                ? (t('mission', 'attackerWins') ?? 'Победа')
                : (t('mission', 'defenderWins') ?? 'Поражение');
            const resultClass = isDraw ? '' : myWin ? 'true' : 'false';
            return (
              <tr key={b.id}>
                <td className="center">{fmtDate(b.at)}</td>
                <td className="center">
                  {b.is_attacker
                    ? (b.defender_user_id ?? '—')
                    : (b.attacker_user_id ?? '—')}
                  {/* План 72.1.50 ч.5 (72.1.10 wave 3): legacy
                      `Battlestats.class.php` рендерит название планеты
                      и координаты под именем оппонента. */}
                  {b.planet_name && (
                    <>
                      <br />
                      <small>
                        {b.is_moon_target
                          ? `🌙 ${b.planet_name}`
                          : b.planet_name}
                        {b.galaxy != null && b.system != null && b.position != null && (
                          <> [{b.galaxy}:{b.system}:{b.position}]</>
                        )}
                      </small>
                    </>
                  )}
                </td>
                <td className="center">
                  <span className={resultClass}>{resultLabel}</span>
                  <br />
                  <small>R: {b.rounds}</small>
                </td>
                <td className="center">
                  <small>
                    М: {formatNumber(b.loot_metal)}<br />
                    К: {formatNumber(b.loot_silicon)}<br />
                    В: {formatNumber(b.loot_hydrogen)}
                  </small>
                </td>
                <td className="center">
                  <small>
                    М: {formatNumber(b.debris_metal)}<br />
                    К: {formatNumber(b.debris_silicon)}
                  </small>
                </td>
                <td className="center">
                  <Link to={`/battle-report/${b.id}`} className="button">
                    {t('alliance', 'detailsBtn') ?? 'Просмотр'}
                  </Link>
                </td>
              </tr>
            );
          })}
        </tbody>
        {(cursor || nextCursor) && (
          <tfoot>
            <tr>
              <td colSpan={6} className="center">
                {cursor && (
                  <button
                    type="button"
                    className="button"
                    style={{ marginRight: 8 }}
                    onClick={() => setCursor(undefined)}
                  >
                    ◀ {t('mission', 'toBegin') ?? 'В начало'}
                  </button>
                )}
                {nextCursor && (
                  <button
                    type="button"
                    className="button"
                    onClick={() => setCursor(nextCursor)}
                    disabled={battlesQ.isFetching}
                  >
                    {t('mission', 'next') ?? 'Дальше'} ▶
                  </button>
                )}
              </td>
            </tr>
          </tfoot>
        )}
      </table>
    </form>
  );
}
