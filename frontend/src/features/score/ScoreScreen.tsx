import { useState, useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';

type ScoreType = 'total' | 'b' | 'r' | 'u' | 'a' | 'e';
type MainTab = 'players' | 'alliances' | 'vacation' | 'transfers';

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
  home_galaxy?: number | null;
  home_system?: number | null;
  home_position?: number | null;
}

interface AllianceEntry {
  rank: number;
  tag: string;
  name: string;
  points: number;
  count: number;
}

interface VacationEntry {
  rank: number;
  user_id: string;
  username: string;
  alliance_tag?: string | null;
  points: number;
  vacation_since: string;
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

function Skeleton() {
  return (
    <div style={{ padding: 20 }}>
      {Array.from({ length: 5 }).map((_, i) => (
        <div key={i} className="ox-skeleton" style={{ height: 36, marginBottom: 6 }} />
      ))}
    </div>
  );
}

function PlayersTab({ scoreType, setScoreType, initialQuery, onPlanetClick }: {
  scoreType: ScoreType;
  setScoreType: (t: ScoreType) => void;
  initialQuery?: string;
  onPlanetClick?: (g: number, s: number) => void;
}) {
  const [filter, setFilter] = useState(initialQuery ?? '');
  const highlightRef = useRef<HTMLTableRowElement | null>(null);

  useEffect(() => {
    if (initialQuery !== undefined) setFilter(initialQuery);
  }, [initialQuery]);

  useEffect(() => {
    if (highlightRef.current && filter.trim().length >= 2) {
      highlightRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }, [filter]);

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

  const fullList = q.data?.highscore ?? [];
  const filterTrim = filter.trim().toLowerCase();
  const list = filterTrim.length > 0
    ? fullList.filter((e) => e.username.toLowerCase().includes(filterTrim))
    : fullList;
  const typeMeta = SCORE_TYPES.find((s) => s.value === scoreType)!;

  return (
    <>
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

      <input
        type="text"
        placeholder="🔍 Фильтр по нику…"
        value={filter}
        onChange={(e) => setFilter(e.target.value)}
        style={{ padding: '6px 10px', maxWidth: 300 }}
      />

      <div className="ox-panel" style={{ overflow: 'hidden' }}>
        {q.isLoading && <Skeleton />}
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
                  <th style={{ width: 100 }}>Координаты</th>
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
                {list.map((e) => {
                  const isMatch = filterTrim.length > 0 && e.username.toLowerCase().includes(filterTrim);
                  return (
                  <tr
                    key={e.user_id}
                    ref={isMatch && !highlightRef.current ? highlightRef : undefined}
                    style={isMatch ? { background: 'rgba(99,217,255,0.08)' } : undefined}
                  >
                    <td data-label="#" className="num">{MEDAL[e.rank - 1] ?? e.rank}</td>
                    <td data-label="Игрок" style={{ fontWeight: e.rank <= 3 ? 700 : 400 }}>{e.username}</td>
                    <td data-label="Альянс" style={{ fontSize: 11, color: 'var(--ox-fg-dim)', fontFamily: 'var(--ox-mono)' }}>
                      {e.alliance_tag ? `[${e.alliance_tag}]` : '—'}
                    </td>
                    <td data-label="Координаты" style={{ fontFamily: 'var(--ox-mono)', fontSize: 12 }}>
                      {e.home_galaxy != null && e.home_system != null && e.home_position != null ? (
                        <button
                          type="button"
                          onClick={() => onPlanetClick?.(e.home_galaxy!, e.home_system!)}
                          style={{ background: 'transparent', border: 'none', color: 'var(--ox-accent)', cursor: 'pointer', padding: 0, fontFamily: 'inherit', fontSize: 'inherit' }}
                          title="Перейти в галактику"
                        >
                          [{e.home_galaxy}:{e.home_system}:{e.home_position}]
                        </button>
                      ) : <span style={{ color: 'var(--ox-fg-muted)' }}>—</span>}
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
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </>
  );
}

function AlliancesTab() {
  const q = useQuery({
    queryKey: ['highscore', 'alliances'],
    queryFn: () => api.get<{ alliances: AllianceEntry[] | null }>('/api/highscore/alliances'),
    refetchInterval: 60000,
  });

  const list = q.data?.alliances ?? [];

  return (
    <div className="ox-panel" style={{ overflow: 'hidden' }}>
      {q.isLoading && <Skeleton />}
      {!q.isLoading && list.length === 0 && (
        <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
          Нет альянсов с игроками
        </div>
      )}
      {list.length > 0 && (
        <div className="ox-table-responsive">
          <table className="ox-table" style={{ margin: 0 }}>
            <thead>
              <tr>
                <th style={{ width: 48 }}>#</th>
                <th>Альянс</th>
                <th style={{ width: 60 }}>Игроков</th>
                <th>Очки</th>
              </tr>
            </thead>
            <tbody>
              {list.map((e) => (
                <tr key={e.tag}>
                  <td className="num">{MEDAL[e.rank - 1] ?? e.rank}</td>
                  <td>
                    <span style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-accent)', marginRight: 8 }}>[{e.tag}]</span>
                    <span style={{ fontWeight: e.rank <= 3 ? 700 : 400 }}>{e.name}</span>
                  </td>
                  <td className="num" style={{ color: 'var(--ox-fg-dim)' }}>{e.count}</td>
                  <td className="num" style={{ color: 'var(--ox-accent)', fontWeight: 600 }}>
                    {Math.round(e.points).toLocaleString('ru-RU')}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function VacationTab() {
  const q = useQuery({
    queryKey: ['highscore', 'vacation'],
    queryFn: () => api.get<{ players: VacationEntry[] | null }>('/api/highscore/vacation'),
    refetchInterval: 60000,
  });

  const list = q.data?.players ?? [];

  return (
    <div className="ox-panel" style={{ overflow: 'hidden' }}>
      {q.isLoading && <Skeleton />}
      {!q.isLoading && list.length === 0 && (
        <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
          Нет игроков в режиме отпуска
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
                <th>Очки</th>
                <th>В отпуске с</th>
              </tr>
            </thead>
            <tbody>
              {list.map((e) => (
                <tr key={e.user_id}>
                  <td className="num">{e.rank}</td>
                  <td>{e.username}</td>
                  <td style={{ fontSize: 11, color: 'var(--ox-fg-dim)', fontFamily: 'var(--ox-mono)' }}>
                    {e.alliance_tag ? `[${e.alliance_tag}]` : '—'}
                  </td>
                  <td className="num" style={{ color: 'var(--ox-accent)' }}>
                    {Math.round(e.points).toLocaleString('ru-RU')}
                  </td>
                  <td style={{ fontSize: 12, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
                    ✈ {new Date(e.vacation_since).toLocaleDateString('ru-RU')}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

interface TransferRow {
  user_id: string;
  username: string;
  total: number;
  metal: number;
  silicon: number;
  hydrogen: number;
}

function TransfersTab() {
  const [direction, setDirection] = useState<'received' | 'sent'>('received');
  const [period, setPeriod] = useState<'all' | 'week' | 'month'>('all');

  const q = useQuery({
    queryKey: ['transfers', direction, period],
    queryFn: () => api.get<{ transfers: TransferRow[] }>(`/api/stats/resource-transfers?direction=${direction}&period=${period}`),
    refetchInterval: 60000,
  });

  const list = q.data?.transfers ?? [];

  return (
    <>
      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        <div style={{ display: 'flex', gap: 4 }}>
          <button type="button" className={direction === 'received' ? 'btn btn-sm' : 'btn-ghost btn-sm'} onClick={() => setDirection('received')}>📥 Получатели</button>
          <button type="button" className={direction === 'sent' ? 'btn btn-sm' : 'btn-ghost btn-sm'} onClick={() => setDirection('sent')}>📤 Отправители</button>
        </div>
        <div style={{ display: 'flex', gap: 4 }}>
          <button type="button" className={period === 'all' ? 'btn btn-sm' : 'btn-ghost btn-sm'} onClick={() => setPeriod('all')}>Всё время</button>
          <button type="button" className={period === 'month' ? 'btn btn-sm' : 'btn-ghost btn-sm'} onClick={() => setPeriod('month')}>Месяц</button>
          <button type="button" className={period === 'week' ? 'btn btn-sm' : 'btn-ghost btn-sm'} onClick={() => setPeriod('week')}>Неделя</button>
        </div>
      </div>

      <div className="ox-panel" style={{ overflow: 'hidden' }}>
        {q.isLoading && <Skeleton />}
        {!q.isLoading && list.length === 0 && (
          <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
            Нет данных
          </div>
        )}
        {list.length > 0 && (
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th style={{ width: 48 }}>#</th>
                  <th>{direction === 'received' ? 'Получатель' : 'Отправитель'}</th>
                  <th className="num">Металл</th>
                  <th className="num">Кремний</th>
                  <th className="num">Водород</th>
                  <th className="num">Всего (у.е.)</th>
                </tr>
              </thead>
              <tbody>
                {list.map((t, i) => (
                  <tr key={t.user_id}>
                    <td className="num">{MEDAL[i] ?? i + 1}</td>
                    <td style={{ fontWeight: i < 3 ? 700 : 400 }}>{t.username}</td>
                    <td className="num">{Math.round(t.metal).toLocaleString('ru-RU')}</td>
                    <td className="num">{Math.round(t.silicon).toLocaleString('ru-RU')}</td>
                    <td className="num">{Math.round(t.hydrogen).toLocaleString('ru-RU')}</td>
                    <td className="num" style={{ color: 'var(--ox-accent)', fontWeight: 600 }}>
                      {Math.round(t.total).toLocaleString('ru-RU')}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </>
  );
}

export function ScoreScreen({ initialQuery, onPlanetClick }: {
  initialQuery?: string;
  onPlanetClick?: (g: number, s: number) => void;
} = {}) {
  const [mainTab, setMainTab] = useState<MainTab>('players');
  const [scoreType, setScoreType] = useState<ScoreType>('total');

  useEffect(() => {
    if (initialQuery && initialQuery.trim().length > 0) {
      setMainTab('players');
    }
  }, [initialQuery]);

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
        🏆 Рейтинг
      </h2>

      <div className="ox-tabs">
        <button type="button" aria-pressed={mainTab === 'players'} onClick={() => setMainTab('players')}>👤 Игроки</button>
        <button type="button" aria-pressed={mainTab === 'alliances'} onClick={() => setMainTab('alliances')}>🤝 Альянсы</button>
        <button type="button" aria-pressed={mainTab === 'vacation'} onClick={() => setMainTab('vacation')}>✈ В отпуске</button>
        <button type="button" aria-pressed={mainTab === 'transfers'} onClick={() => setMainTab('transfers')}>📦 Торговля</button>
      </div>

      {mainTab === 'players'   && <PlayersTab scoreType={scoreType} setScoreType={setScoreType} initialQuery={initialQuery} onPlanetClick={onPlanetClick} />}
      {mainTab === 'alliances' && <AlliancesTab />}
      {mainTab === 'vacation'  && <VacationTab />}
      {mainTab === 'transfers' && <TransfersTab />}
    </div>
  );
}
