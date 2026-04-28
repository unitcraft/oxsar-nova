// S-023 Battle stats (план 72 Ф.3 Spring 2 ч.2).
//
// Pixel-perfect зеркало legacy `templates/standard/battlestats.tpl`:
// фильтры (даты, фильтр по нику/альянсу, чекбоксы) + таблица боёв.
//
// Замечание (P72.S2.G — simplifications.md):
// nova-API на 2026-04-28 НЕ предоставляет endpoint'а
// `/api/users/me/battles` или агрегированный `/api/battlestats`.
// Прокси `GET /api/highscore/me` отдаёт только итоговые score/rank, без
// детализации боёв.
//
// В первой итерации S-023:
//   - Рендерим pixel-perfect шапку фильтров и таблицу как ntable.
//   - Список боёв пустой (пока backend не появится).
//   - Показываем общий ранг игрока из /api/highscore/me как top-row.
//
// При появлении endpoint'а — заменяем `[]` на реальный fetch с
// query-параметрами из формы.

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

interface HighscoreEntry {
  user_id: string;
  username: string;
  score: number;
  rank: number;
}

export function BattleStatsScreen() {
  const { t } = useTranslation();
  const [dateFirst, setDateFirst] = useState('');
  const [dateLast, setDateLast] = useState('');
  const [userFilter, setUserFilter] = useState('');
  const [allianceFilter, setAllianceFilter] = useState('');
  const [showDrawn, setShowDrawn] = useState(false);
  const [showAliens, setShowAliens] = useState(false);

  const meQ = useQuery({
    queryKey: ['highscore', 'me'],
    queryFn: () => api.get<HighscoreEntry>('/api/highscore/me'),
    staleTime: 60_000,
  });

  // P72.S2.G: реальные battles — пустой массив, пока endpoint не появится.
  const battles: never[] = [];

  return (
    <form
      method="post"
      onSubmit={(ev) => ev.preventDefault()}
    >
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
                    <td>
                      <label htmlFor="date_last">
                        {t('statistics', 'dateLast')}
                      </label>
                    </td>
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
                    <td>
                      <label htmlFor="date_first">
                        {t('statistics', 'dateFirst')}
                      </label>
                    </td>
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
                    <td>
                      <label htmlFor="user_filter">
                        {t('statistics', 'bsUserFilter')}
                      </label>
                    </td>
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
                    <td>
                      <label htmlFor="alliance_filter">
                        {t('statistics', 'bsAllianceFilter')}
                      </label>
                    </td>
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
                        name="show_drawn"
                        type="checkbox"
                        checked={showDrawn}
                        onChange={(e) => setShowDrawn(e.target.checked)}
                      />{' '}
                      <label htmlFor="show_drawn">
                        {t('statistics', 'bsShowDrawn')}
                      </label>
                    </td>
                  </tr>
                  <tr>
                    <td colSpan={2}>
                      <input
                        id="show_aliens"
                        name="show_aliens"
                        type="checkbox"
                        checked={showAliens}
                        onChange={(e) => setShowAliens(e.target.checked)}
                      />{' '}
                      <label htmlFor="show_aliens">
                        {t('statistics', 'bsShowUfoBattles')}
                      </label>
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
                value={t('alliance', 'commit') || 'OK'}
                disabled
                title={t('battlestats', 'empty')}
              />
            </td>
          </tr>
        </tbody>
      </table>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={4} className="center">
              {meQ.data ? (
                <>
                  {t('battlestats', 'statTotal')}: <b>{meQ.data.score}</b> ·{' '}
                  {t('alliance', 'rank')}: <b>{meQ.data.rank}</b>
                </>
              ) : (
                '…'
              )}
            </th>
          </tr>
          <tr>
            <th>{t('battlestats', 'colDate')}</th>
            <th>{t('battlestats', 'colOpponent')}</th>
            <th>{t('battlestats', 'colResult')}</th>
            <th>{t('battlestats', 'colLoot')}</th>
          </tr>
        </thead>
        <tbody>
          {battles.length === 0 ? (
            <tr>
              <td colSpan={4} className="center">
                {t('battlestats', 'empty')}
              </td>
            </tr>
          ) : null}
        </tbody>
      </table>
    </form>
  );
}
