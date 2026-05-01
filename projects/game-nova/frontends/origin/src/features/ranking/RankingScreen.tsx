// S-023 Ranking / Statistics (план 72 Ф.4 Spring 3, расширен 72.1.12).
//
// Pixel-perfect зеркало legacy `templates/standard/playerstats.tpl` +
// statsheader: табы переключения players/alliances/vacation +
// type-select по разным метрикам очков (e/b/r/u/a/dm/max/total).
//
// Endpoints:
//   GET /api/highscore?type=...        → { entries: HighscoreEntry[] }
//   GET /api/highscore/me?type=...     → HighscoreEntry
//   GET /api/highscore/alliances       → { alliances: HighscoreAlliance[] }
//   GET /api/highscore/vacation        → { players: HighscoreVacation[] }
//   GET /api/stats                     → { online_now, online_24h }
//
// Не реализовано (вне scope 72.1.12):
//   - avg-режим (legacy `Ranking::average`).
//   - b_count/r_count/u_count/battles — нет агрегатных колонок.
//   - player_observer / player_old_vacation — нишевые режимы.
//   - Пагинация по 25 (используем топ-100 без пагинации).

import { useQuery } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import {
  fetchHighscore,
  fetchHighscoreAlliances,
  fetchHighscoreMe,
  fetchHighscoreVacation,
  fetchPublicStats,
  type ScoreType,
} from '@/api/highscore';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';
import type { HighscoreEntry } from '@/api/types';

type Mode = 'player' | 'alliance' | 'vacation';

const MODES: Mode[] = ['player', 'alliance', 'vacation'];
const TYPES: ScoreType[] = [
  'total',
  'b',
  'r',
  'u',
  'a',
  'e',
  'dm',
  'max',
];

function isMode(v: string | null): v is Mode {
  return v === 'player' || v === 'alliance' || v === 'vacation';
}
function isType(v: string | null): v is ScoreType {
  return TYPES.includes(v as ScoreType);
}

const MODE_TAB_KEY: Record<Mode, string> = {
  player: 'tabPlayers',
  alliance: 'tabAlliances',
  vacation: 'tabVacation',
};

// План 72.1.12: ключи в i18n существуют как scoreType<Suffix>; мапим сюда.
const TYPE_LABEL_KEY: Record<ScoreType, string> = {
  total: 'scoreTypeTotal',
  b: 'scoreTypeBuildings',
  r: 'scoreTypeResearch',
  u: 'scoreTypeFleet',
  a: 'scoreTypeAchievements',
  e: 'scoreTypeBattle',
  dm: 'scoreTypeDm',
  max: 'scoreTypeMax',
};

function pickScore(e: HighscoreEntry, type: ScoreType): number {
  switch (type) {
    case 'b':
      return e.b_points ?? 0;
    case 'r':
      return e.r_points ?? 0;
    case 'u':
      return e.u_points ?? 0;
    case 'a':
      return e.a_points ?? 0;
    case 'e':
      return e.e_points ?? 0;
    case 'dm':
      return e.dm_points ?? 0;
    case 'max':
      return e.max_points ?? 0;
    default:
      return e.points ?? e.score ?? 0;
  }
}

export function RankingScreen() {
  const { t } = useTranslation();
  const [params, setParams] = useSearchParams();

  const mode: Mode = isMode(params.get('mode')) ? (params.get('mode') as Mode) : 'player';
  const type: ScoreType = isType(params.get('type'))
    ? (params.get('type') as ScoreType)
    : 'total';

  const playerQ = useQuery({
    queryKey: QK.highscore(type),
    queryFn: () => fetchHighscore(type),
    staleTime: 30_000,
    enabled: mode === 'player',
  });
  const meQ = useQuery({
    queryKey: QK.highscoreMe(type),
    queryFn: () => fetchHighscoreMe(type),
    staleTime: 30_000,
    enabled: mode === 'player',
  });
  const allyQ = useQuery({
    queryKey: QK.highscoreAlliances(),
    queryFn: fetchHighscoreAlliances,
    staleTime: 30_000,
    enabled: mode === 'alliance',
  });
  const vacQ = useQuery({
    queryKey: QK.highscoreVacation(),
    queryFn: fetchHighscoreVacation,
    staleTime: 30_000,
    enabled: mode === 'vacation',
  });
  const statsQ = useQuery({
    queryKey: QK.publicStats(),
    queryFn: fetchPublicStats,
    staleTime: 60_000,
  });

  function setMode(next: Mode) {
    const p = new URLSearchParams(params);
    p.set('mode', next);
    if (next !== 'player') p.delete('type');
    setParams(p);
  }
  function setType(next: ScoreType) {
    const p = new URLSearchParams(params);
    p.set('type', next);
    setParams(p);
  }

  return (
    <>
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={4}>{t('score', 'title')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>{t('score', 'online')}</td>
            <td>
              <b>{formatNumber(statsQ.data?.online_now ?? 0)}</b>
            </td>
            <td>{t('score', 'online24h')}</td>
            <td>
              <b>{formatNumber(statsQ.data?.online_24h ?? 0)}</b>
            </td>
          </tr>
          {mode === 'player' && meQ.data && (
            <tr>
              <td colSpan={4}>
                {t('score', 'myRankLabel', {
                  typeLabel: t('score', TYPE_LABEL_KEY[type]),
                })}
                : <b>{meQ.data.rank}</b> ({formatNumber(meQ.data.score ?? 0)})
              </td>
            </tr>
          )}
        </tbody>
      </table>

      <div className="idiv">
        {MODES.map((m, i) => (
          <span key={m}>
            {i > 0 && ' | '}
            <button
              type="button"
              className={'tab-link' + (mode === m ? ' true' : '')}
              onClick={() => setMode(m)}
            >
              {t('score', MODE_TAB_KEY[m])}
            </button>
          </span>
        ))}
        {mode === 'player' && (
          <>
            {' | '}
            <label>
              {t('score', 'sortByLabel')}{' '}
              <select
                value={type}
                onChange={(e) => setType(e.target.value as ScoreType)}
              >
                {TYPES.map((tp) => (
                  <option key={tp} value={tp}>
                    {t('score', TYPE_LABEL_KEY[tp])}
                  </option>
                ))}
              </select>
            </label>
          </>
        )}
      </div>

      {mode === 'player' && <PlayerTable type={type} q={playerQ} />}
      {mode === 'alliance' && <AllianceTable q={allyQ} />}
      {mode === 'vacation' && <VacationTable q={vacQ} />}
    </>
  );
}

function PlayerTable({
  type,
  q,
}: {
  type: ScoreType;
  q: ReturnType<typeof useQuery<{ entries: HighscoreEntry[] | null }, Error>>;
}) {
  const { t } = useTranslation();
  const entries = q.data?.entries ?? [];
  return (
    <table className="ntable">
      <colgroup>
        <col width="1" />
        <col width="*" />
        <col width="*" />
        <col width="*" />
      </colgroup>
      <thead>
        <tr>
          <th>#</th>
          <th>{t('score', 'colPlayer')}</th>
          <th>{t('alliance', 'tag')}</th>
          <th>{t('score', 'colPoints')}</th>
        </tr>
      </thead>
      <tbody>
        {entries.length === 0 && (
          <tr>
            <td colSpan={4} className="center">
              {t('score', 'emptyPlayers')}
            </td>
          </tr>
        )}
        {entries.map((e) => (
          <tr key={e.user_id}>
            <td>{e.rank}</td>
            <td>{e.username}</td>
            <td>{e.alliance_tag ?? ''}</td>
            <td>{formatNumber(pickScore(e, type))}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function AllianceTable({
  q,
}: {
  q: ReturnType<
    typeof useQuery<
      { alliances: import('@/api/types').HighscoreAlliance[] | null },
      Error
    >
  >;
}) {
  const { t } = useTranslation();
  const list = q.data?.alliances ?? [];
  return (
    <table className="ntable">
      <thead>
        <tr>
          <th>#</th>
          <th>{t('alliance', 'tag')}</th>
          <th>{t('alliance', 'name')}</th>
          <th>{t('score', 'colMembers')}</th>
          <th>{t('score', 'colPoints')}</th>
        </tr>
      </thead>
      <tbody>
        {list.length === 0 && (
          <tr>
            <td colSpan={5} className="center">
              {t('score', 'emptyAlliances')}
            </td>
          </tr>
        )}
        {list.map((a) => (
          <tr key={a.tag}>
            <td>{a.rank}</td>
            <td>{a.tag}</td>
            <td>{a.name}</td>
            <td>{a.count}</td>
            <td>{formatNumber(a.points)}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

function VacationTable({
  q,
}: {
  q: ReturnType<
    typeof useQuery<
      { players: import('@/api/types').HighscoreVacation[] | null },
      Error
    >
  >;
}) {
  const { t } = useTranslation();
  const list = q.data?.players ?? [];
  return (
    <table className="ntable">
      <thead>
        <tr>
          <th>#</th>
          <th>{t('score', 'colPlayer')}</th>
          <th>{t('alliance', 'tag')}</th>
          <th>{t('score', 'colPoints')}</th>
          <th>{t('score', 'colVacationSince')}</th>
        </tr>
      </thead>
      <tbody>
        {list.length === 0 && (
          <tr>
            <td colSpan={5} className="center">
              {t('score', 'emptyVacation')}
            </td>
          </tr>
        )}
        {list.map((p) => (
          <tr key={p.user_id}>
            <td>{p.rank}</td>
            <td>{p.username}</td>
            <td>{p.alliance_tag ?? ''}</td>
            <td>{formatNumber(p.points)}</td>
            <td>{new Date(p.vacation_since).toLocaleDateString('ru-RU')}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
