// S-023 Battle stats (план 72.1 ч.20.8 — battle viewer).
//
// Pixel-perfect клон legacy battlestats.tpl + список реальных боёв
// + кнопка «Просмотр» открывает /battle-reports/:id.

import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { fetchMyBattles } from '@/api/battles';
import { QK } from '@/api/query-keys';
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
  const [dateFirst, setDateFirst] = useState('');
  const [dateLast, setDateLast] = useState('');
  const [userFilter, setUserFilter] = useState('');
  const [allianceFilter, setAllianceFilter] = useState('');
  const [showDrawn, setShowDrawn] = useState(false);
  const [showAliens, setShowAliens] = useState(false);
  const [cursor, setCursor] = useState<string | undefined>(undefined);

  const meQ = useQuery({
    queryKey: ['highscore', 'me'],
    queryFn: () => api.get<HighscoreEntry>('/api/highscore/me'),
    staleTime: 60_000,
  });

  const battlesQ = useQuery({
    queryKey: QK.myBattles(cursor),
    queryFn: () => fetchMyBattles(cursor !== undefined ? { cursor, limit: 20 } : { limit: 20 }),
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
