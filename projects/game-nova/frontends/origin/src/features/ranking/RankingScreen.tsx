// S-023 Ranking / Statistics (план 72 Ф.4 Spring 3, расширен 72.1.12 + 72.1.29).
//
// Pixel-perfect зеркало legacy `templates/standard/playerstats.tpl` +
// statsheader: 5 mode'ов (player/observer/old_vacation/alliance/vacation),
// 12 score-types включая b_count/r_count/u_count/battles, avg-режим,
// пагинация по 25 (legacy USER_PER_PAGE).
//
// Endpoints:
//   GET /api/highscore?type=&mode=&avg=&page=  → { entries, total_count, page, per_page }
//   GET /api/highscore/me?type=...             → HighscoreEntry
//   GET /api/highscore/alliances               → { alliances: HighscoreAlliance[] }
//   GET /api/highscore/vacation                → { players: HighscoreVacation[] }
//   GET /api/stats                             → { online_now, online_24h }

import { useQuery } from '@tanstack/react-query';
import { useSearchParams } from 'react-router-dom';
import {
  fetchHighscore,
  fetchHighscoreAlliances,
  fetchHighscoreMe,
  fetchHighscoreVacation,
  fetchPublicStats,
  type HighscoreResult,
  type RankingMode,
  type ScoreType,
} from '@/api/highscore';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';
import type { HighscoreEntry } from '@/api/types';

// План 72.1.29: 5 mode-tab'ов вместо 3 (legacy `Ranking::getRanking`).
type Mode =
  | 'player'
  | 'player_observer'
  | 'player_old_vacation'
  | 'alliance'
  | 'vacation';

const MODES: Mode[] = [
  'player',
  'player_observer',
  'player_old_vacation',
  'alliance',
  'vacation',
];

// 12 score-types (legacy `Ranking::validTypes`).
const TYPES: ScoreType[] = [
  'total',
  'b',
  'r',
  'u',
  'a',
  'e',
  'dm',
  'max',
  'b_count',
  'r_count',
  'u_count',
  'battles',
];

function isMode(v: string | null): v is Mode {
  return MODES.includes(v as Mode);
}
function isType(v: string | null): v is ScoreType {
  return TYPES.includes(v as ScoreType);
}

const MODE_TAB_KEY: Record<Mode, string> = {
  player: 'tabPlayers',
  player_observer: 'tabPlayerObserver',
  player_old_vacation: 'tabPlayerOldVacation',
  alliance: 'tabAlliances',
  vacation: 'tabVacation',
};

const TYPE_LABEL_KEY: Record<ScoreType, string> = {
  total: 'scoreTypeTotal',
  b: 'scoreTypeBuildings',
  r: 'scoreTypeResearch',
  u: 'scoreTypeFleet',
  a: 'scoreTypeAchievements',
  e: 'scoreTypeBattle',
  dm: 'scoreTypeDm',
  max: 'scoreTypeMax',
  b_count: 'scoreTypeBCount',
  r_count: 'scoreTypeRCount',
  u_count: 'scoreTypeUCount',
  battles: 'scoreTypeBattles',
};

function pickScore(e: HighscoreEntry, type: ScoreType, avg: boolean): number {
  if (avg) return e.score_avg ?? 0;
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
    case 'b_count':
      return e.b_count ?? 0;
    case 'r_count':
      return e.r_count ?? 0;
    case 'u_count':
      return e.u_count ?? 0;
    case 'battles':
      return e.battles ?? 0;
    default:
      return e.points ?? e.score ?? 0;
  }
}

// isPlayerMode — все 3 player-вариации.
function isPlayerMode(m: Mode): boolean {
  return m === 'player' || m === 'player_observer' || m === 'player_old_vacation';
}

// Backend mode mapping. Frontend `Mode` includes alliance/vacation (которые
// идут через отдельные endpoint'ы); player-mode'ы передаются как есть.
function backendMode(m: Mode): RankingMode {
  if (m === 'player_observer') return 'player_observer';
  if (m === 'player_old_vacation') return 'player_old_vacation';
  return 'player';
}

export function RankingScreen() {
  const { t } = useTranslation();
  const [params, setParams] = useSearchParams();

  const mode: Mode = isMode(params.get('mode')) ? (params.get('mode') as Mode) : 'player';
  const type: ScoreType = isType(params.get('type'))
    ? (params.get('type') as ScoreType)
    : 'total';
  const avg = params.get('avg') === 'true';
  const page = Math.max(1, Number(params.get('page')) || 1);

  const playerEnabled = isPlayerMode(mode);
  const queryOpts = { type, mode: backendMode(mode), avg, page };

  const playerQ = useQuery({
    queryKey: ['highscore', queryOpts.type, queryOpts.mode, avg, page],
    queryFn: () => fetchHighscore(queryOpts),
    staleTime: 30_000,
    enabled: playerEnabled,
  });
  const meQ = useQuery({
    queryKey: QK.highscoreMe(type),
    queryFn: () => fetchHighscoreMe(type),
    staleTime: 30_000,
    enabled: playerEnabled,
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
    p.delete('page'); // reset pagination
    if (!isPlayerMode(next)) {
      p.delete('type');
      p.delete('avg');
    }
    setParams(p);
  }
  function setType(next: ScoreType) {
    const p = new URLSearchParams(params);
    p.set('type', next);
    p.delete('page');
    setParams(p);
  }
  function toggleAvg() {
    const p = new URLSearchParams(params);
    if (avg) p.delete('avg');
    else p.set('avg', 'true');
    p.delete('page');
    setParams(p);
  }
  function setPage(next: number) {
    const p = new URLSearchParams(params);
    if (next > 1) p.set('page', String(next));
    else p.delete('page');
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
          {playerEnabled && meQ.data && (
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
        {playerEnabled && (
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
            {' | '}
            <label>
              <input
                type="checkbox"
                checked={avg}
                onChange={toggleAvg}
              />{' '}
              {t('score', 'avgCheckbox')}
            </label>
          </>
        )}
      </div>

      {playerEnabled && (
        <PlayerTable
          type={type}
          avg={avg}
          page={page}
          q={playerQ}
          onPage={setPage}
        />
      )}
      {mode === 'alliance' && <AllianceTable q={allyQ} />}
      {mode === 'vacation' && <VacationTable q={vacQ} />}
    </>
  );
}

function PlayerTable({
  type,
  avg,
  page,
  q,
  onPage,
}: {
  type: ScoreType;
  avg: boolean;
  page: number;
  q: ReturnType<typeof useQuery<HighscoreResult, Error>>;
  onPage: (n: number) => void;
}) {
  const { t } = useTranslation();
  const entries = q.data?.entries ?? [];
  const totalCount = q.data?.total_count ?? 0;
  const perPage = q.data?.per_page ?? 25;
  const totalPages = Math.max(1, Math.ceil(totalCount / perPage));

  return (
    <>
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
            <th>
              {avg
                ? t('score', 'colScoreAvg')
                : t('score', TYPE_LABEL_KEY[type])}
            </th>
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
              <td>{formatNumber(pickScore(e, type, avg))}</td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* План 72.1.29: пагинация по 25 (legacy USER_PER_PAGE). */}
      {totalPages > 1 && (
        <div className="idiv center">
          <button
            type="button"
            className="button"
            disabled={page <= 1}
            onClick={() => onPage(page - 1)}
          >
            ◀
          </button>{' '}
          <span>
            {t('score', 'pageLabel', { page, total: totalPages })}
          </span>{' '}
          <button
            type="button"
            className="button"
            disabled={page >= totalPages}
            onClick={() => onPage(page + 1)}
          >
            ▶
          </button>
        </div>
      )}
    </>
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
