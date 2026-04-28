// S-023 Ranking / Statistics (план 72 Ф.4 Spring 3).
//
// Pixel-perfect зеркало legacy `templates/standard/playerstats.tpl`
// (с шапкой `statsheader.tpl`, в первой итерации не воспроизводится —
// табы переключения players/alliances/vacation/transfers).
//
// Endpoints (openapi.yaml):
//   GET /api/highscore        → { entries: HighscoreEntry[] }
//   GET /api/highscore/me     → HighscoreEntry  (мой ранг)
//   GET /api/stats            → { online_now, online_24h }
//
// Это совмещённый экран «Рейтинг + публичная статистика игры»,
// решающий задачу промпта S-032 «агрегированная статистика». Табы
// player/alliance/vacation/transfers (как в legacy) — отдельный
// план, см. simplifications P72.S3.F.

import { useQuery } from '@tanstack/react-query';
import {
  fetchHighscore,
  fetchHighscoreMe,
  fetchPublicStats,
} from '@/api/highscore';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { formatNumber } from '@/lib/format';

export function RankingScreen() {
  const { t } = useTranslation();

  const hsQ = useQuery({
    queryKey: QK.highscore(),
    queryFn: fetchHighscore,
    staleTime: 30_000,
  });
  const meQ = useQuery({
    queryKey: QK.highscoreMe(),
    queryFn: fetchHighscoreMe,
    staleTime: 30_000,
  });
  const statsQ = useQuery({
    queryKey: QK.publicStats(),
    queryFn: fetchPublicStats,
    staleTime: 60_000,
  });

  const entries = hsQ.data?.entries ?? [];

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
            <td>{t('common', 'online') || 'Онлайн сейчас'}</td>
            <td>
              <b>{formatNumber(statsQ.data?.online_now ?? 0)}</b>
            </td>
            <td>{t('common', 'online24h') || 'За 24 часа'}</td>
            <td>
              <b>{formatNumber(statsQ.data?.online_24h ?? 0)}</b>
            </td>
          </tr>
          {meQ.data && (
            <tr>
              <td colSpan={4}>
                {t('score', 'myRankLabel', {
                  typeLabel: t('score', 'scoreTypeTotal'),
                })}
                : <b>{meQ.data.rank}</b> ({formatNumber(meQ.data.score)})
              </td>
            </tr>
          )}
        </tbody>
      </table>
      <table className="ntable">
        <colgroup>
          <col width="1" />
          <col width="*" />
          <col width="*" />
        </colgroup>
        <thead>
          <tr>
            <th>#</th>
            <th>{t('score', 'colPlayer')}</th>
            <th>{t('score', 'colPoints')}</th>
          </tr>
        </thead>
        <tbody>
          {entries.length === 0 && (
            <tr>
              <td colSpan={3} className="center">
                {t('score', 'emptyPlayers')}
              </td>
            </tr>
          )}
          {entries.map((e) => (
            <tr key={e.user_id}>
              <td>{e.rank}</td>
              <td>{e.username}</td>
              <td>{formatNumber(e.score)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </>
  );
}
