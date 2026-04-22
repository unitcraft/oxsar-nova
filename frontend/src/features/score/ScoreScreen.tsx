import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';

type ScoreType = 'total' | 'b' | 'r' | 'u' | 'a' | 'e';

interface Entry {
  rank: number;
  user_id: string;
  username: string;
  alliance_tag?: string | null;
  points: number;
  b_points: number;
  r_points: number;
  u_points: number;
  a_points: number;
  e_points?: number;
}

interface MyRank {
  rank: number;
  type: string;
}

const SCORE_TYPES: { value: ScoreType; label: string; icon: string }[] = [
  { value: 'total', label: 'Общий',        icon: '🏆' },
  { value: 'b',     label: 'Постройки',    icon: '🏗' },
  { value: 'r',     label: 'Исследования', icon: '🔬' },
  { value: 'u',     label: 'Флот',         icon: '🛸' },
  { value: 'a',     label: 'Достижения',   icon: '🎖' },
  { value: 'e',     label: 'Боевой',       icon: '⚔️' },
];

function getPoints(e: Entry, type: ScoreType): number {
  if (type === 'total') return e.points;
  if (type === 'b') return e.b_points;
  if (type === 'r') return e.r_points;
  if (type === 'u') return e.u_points;
  if (type === 'e') return e.e_points ?? 0;
  return e.a_points;
}

const MEDAL = ['🥇', '🥈', '🥉'];

export function ScoreScreen() {
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
  const typeMeta = SCORE_TYPES.find((s) => s.value === scoreType)!;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
        🏆 Рейтинг
      </h2>

      {/* My rank banner */}
      {myRank.data && (
        <div className="ox-panel" style={{ padding: '12px 20px', display: 'flex', alignItems: 'center', gap: 12 }}>
          <span style={{ fontSize: 28 }}>{MEDAL[myRank.data.rank - 1] ?? '🎯'}</span>
          <div>
            <div style={{ fontSize: 12, color: 'var(--ox-fg-muted)' }}>Ваше место в рейтинге «{typeMeta.label}»</div>
            <div style={{ fontSize: 22, fontWeight: 700, fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)' }}>
              #{myRank.data.rank}
            </div>
          </div>
        </div>
      )}

      {/* Tab switcher */}
      <div className="ox-tabs">
        {SCORE_TYPES.map((s) => (
          <button
            key={s.value}
            type="button"
            aria-pressed={scoreType === s.value}
            onClick={() => setScoreType(s.value)}
          >
            {s.icon} {s.label}
          </button>
        ))}
      </div>

      {/* Leaderboard */}
      <div className="ox-panel" style={{ overflow: 'hidden' }}>
        {q.isLoading && (
          <div style={{ padding: 20 }}>
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="ox-skeleton" style={{ height: 36, marginBottom: 6 }} />
            ))}
          </div>
        )}

        {!q.isLoading && list.length === 0 && (
          <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
            Рейтинг пока пуст — начните строить!
          </div>
        )}

        {list.length > 0 && (
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th style={{ width: 48 }}>#</th>
                  <th>Игрок</th>
                  <th style={{ width: 60 }}>Альянс</th>
                  <th>{typeMeta.label}</th>
                  {scoreType === 'total' && (
                    <>
                      <th>🏗</th>
                      <th>🔬</th>
                      <th>🛸</th>
                    </>
                  )}
                </tr>
              </thead>
              <tbody>
                {list.map((e) => (
                  <tr key={e.user_id}>
                    <td data-label="#" className="num">
                      {MEDAL[e.rank - 1] ?? e.rank}
                    </td>
                    <td data-label="Игрок" style={{ fontWeight: e.rank <= 3 ? 700 : 400 }}>
                      {e.username}
                    </td>
                    <td data-label="Альянс" style={{ fontSize: 11, color: 'var(--ox-fg-dim)', fontFamily: 'var(--ox-mono)' }}>
                      {e.alliance_tag ? `[${e.alliance_tag}]` : '—'}
                    </td>
                    <td data-label={typeMeta.label} className="num" style={{ color: 'var(--ox-accent)', fontWeight: 600 }}>
                      {Math.round(getPoints(e, scoreType)).toLocaleString('ru-RU')}
                    </td>
                    {scoreType === 'total' && (
                      <>
                        <td data-label="Постройки" className="num">{Math.round(e.b_points).toLocaleString('ru-RU')}</td>
                        <td data-label="Исследования" className="num">{Math.round(e.r_points).toLocaleString('ru-RU')}</td>
                        <td data-label="Флот" className="num">{Math.round(e.u_points).toLocaleString('ru-RU')}</td>
                      </>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
