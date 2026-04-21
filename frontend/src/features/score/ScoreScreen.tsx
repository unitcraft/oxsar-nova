import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

type ScoreType = 'total' | 'b' | 'r' | 'u' | 'a';

interface Entry {
  rank: number;
  user_id: string;
  username: string;
  points: number;
  b_points: number;
  r_points: number;
  u_points: number;
  a_points: number;
}

interface MyRank {
  rank: number;
  type: string;
}

const SCORE_TYPES: { value: ScoreType; label: string }[] = [
  { value: 'total', label: 'Общий' },
  { value: 'b', label: 'Постройки' },
  { value: 'r', label: 'Исследования' },
  { value: 'u', label: 'Флот' },
  { value: 'a', label: 'Достижения' },
];

export function ScoreScreen() {
  const { t } = useTranslation();
  const [scoreType, setScoreType] = useState<ScoreType>('total');

  const q = useQuery({
    queryKey: ['highscore', scoreType],
    queryFn: () => api.get<{ highscore: Entry[] | null }>(`/api/highscore?type=${scoreType}`),
    refetchInterval: 60000,
  });

  const myRank = useQuery({
    queryKey: ['highscore', 'me', scoreType],
    queryFn: () => api.get<MyRank>(`/api/highscore/me?type=${scoreType}`),
    refetchInterval: 60000,
  });

  const list = q.data?.highscore ?? [];

  const colLabel = (type: ScoreType): string => {
    const found = SCORE_TYPES.find((s) => s.value === type);
    return found?.label ?? 'Очки';
  };

  return (
    <section>
      <h2>{t('global', 'MENU_HIGHSCORE') || 'Рейтинг'}</h2>

      <div style={{ marginBottom: 12 }}>
        {SCORE_TYPES.map((s) => (
          <button
            key={s.value}
            type="button"
            aria-pressed={scoreType === s.value}
            style={{ marginRight: 6 }}
            onClick={() => setScoreType(s.value)}
          >
            {s.label}
          </button>
        ))}
      </div>

      {myRank.data && (
        <p>
          Ваше место в рейтинге «{colLabel(scoreType)}»:{' '}
          <b>#{myRank.data.rank}</b>
        </p>
      )}

      {q.isLoading && <p>…</p>}
      {q.error && (
        <p className="ox-error">
          {t('global', 'ERROR')}:{' '}
          {q.error instanceof Error ? q.error.message : ''}
        </p>
      )}

      {list.length > 0 && (
        <table className="ox-table">
          <thead>
            <tr>
              <th>#</th>
              <th>Игрок</th>
              <th>{colLabel(scoreType)}</th>
              {scoreType === 'total' && (
                <>
                  <th>Постройки</th>
                  <th>Исследования</th>
                  <th>Флот</th>
                </>
              )}
            </tr>
          </thead>
          <tbody>
            {list.map((e) => (
              <tr key={e.user_id}>
                <td className="num">{e.rank}</td>
                <td>{e.username}</td>
                <td className="num">
                  {scoreType === 'total'
                    ? e.points.toFixed(0)
                    : scoreType === 'b'
                    ? e.b_points.toFixed(0)
                    : scoreType === 'r'
                    ? e.r_points.toFixed(0)
                    : scoreType === 'u'
                    ? e.u_points.toFixed(0)
                    : e.a_points.toFixed(0)}
                </td>
                {scoreType === 'total' && (
                  <>
                    <td className="num">{e.b_points.toFixed(0)}</td>
                    <td className="num">{e.r_points.toFixed(0)}</td>
                    <td className="num">{e.u_points.toFixed(0)}</td>
                  </>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {!q.isLoading && list.length === 0 && (
        <p>Рейтинг пока пуст — начните строить!</p>
      )}
    </section>
  );
}
