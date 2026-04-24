import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { ProgressBar } from '@/ui/ProgressBar';

interface Entry {
  key: string;
  title: string;
  description: string;
  points: number;
  category: string;
  unlocked_at?: string | null;
  progress?: number | undefined;
  progress_max?: number | undefined;
}

type CategoryFilter = 'all' | 'starter' | 'passive';

const CATEGORY_LABELS: Record<string, string> = {
  starter: '🎓 Старт',
  passive: '🥇 Достижения',
};

function AchievementCard({ e }: { e: Entry }) {
  const done = !!e.unlocked_at;
  const hasProg = e.progress_max != null && !done;
  const pct = hasProg ? Math.min(100, ((e.progress ?? 0) / e.progress_max!) * 100) : 0;
  return (
    <div
      className="ox-panel"
      style={{
        padding: '12px 16px',
        opacity: done ? 1 : 0.65,
        borderColor: done ? 'var(--ox-success)' : undefined,
        display: 'flex', gap: 14, alignItems: 'flex-start',
      }}
    >
      <div style={{ fontSize: 28, flexShrink: 0 }}>
        {done ? '✅' : '🔒'}
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontWeight: 700, fontSize: 16, marginBottom: 2 }}>{e.title}</div>
        <div style={{ fontSize: 14, color: 'var(--ox-fg-dim)', marginBottom: hasProg ? 6 : 0 }}>
          {e.description}
          {hasProg && (
            <span style={{ marginLeft: 8, fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg-muted)' }}>
              {e.progress ?? 0}/{e.progress_max}
            </span>
          )}
        </div>
        {hasProg && <ProgressBar pct={pct} height={3} />}
        {done && e.unlocked_at && (
          <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)', marginTop: 4 }}>
            Получено {new Date(e.unlocked_at).toLocaleString('ru-RU')}
          </div>
        )}
      </div>
      <div style={{ fontSize: 15, fontWeight: 700, color: 'var(--ox-success)', fontFamily: 'var(--ox-mono)', flexShrink: 0 }}>
        +{e.points}
      </div>
    </div>
  );
}

export function AchievementsScreen() {
  const [filter, setFilter] = useState<CategoryFilter>('all');

  const q = useQuery({
    queryKey: ['achievements'],
    queryFn: () => api.get<{ achievements: Entry[] | null }>('/api/achievements'),
    refetchInterval: 15000,
  });

  const list = q.data?.achievements ?? [];
  const unlocked = list.filter((e) => e.unlocked_at);
  const totalPoints = unlocked.reduce((acc, e) => acc + e.points, 0);

  const visible = filter === 'all' ? list : list.filter((e) => e.category === filter);

  const categories: CategoryFilter[] = ['all', 'starter', 'passive'];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
        <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
          🎖 Достижения
        </h2>
        <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)' }}>
          Открыто:{' '}
          <span style={{ fontWeight: 700, color: 'var(--ox-accent)', fontFamily: 'var(--ox-mono)' }}>
            {unlocked.length}/{list.length}
          </span>
          {' · '}
          Очки:{' '}
          <span style={{ fontWeight: 700, fontFamily: 'var(--ox-mono)', color: 'var(--ox-success)' }}>
            {totalPoints}
          </span>
        </div>
      </div>

      {list.length > 0 && (
        <div style={{ marginBottom: 4 }}>
          <ProgressBar pct={list.length > 0 ? (unlocked.length / list.length) * 100 : 0} variant="success" height={6} showLabel />
        </div>
      )}

      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        {categories.map((cat) => (
          <button
            key={cat}
            type="button"
            onClick={() => setFilter(cat)}
            style={{
              padding: '4px 12px', borderRadius: 16, border: '1px solid',
              borderColor: filter === cat ? 'var(--ox-accent)' : 'var(--ox-border)',
              background: filter === cat ? 'var(--ox-bg-active)' : 'transparent',
              color: filter === cat ? 'var(--ox-accent)' : 'var(--ox-fg-dim)',
              fontSize: 15, fontWeight: 600, cursor: 'pointer',
            }}
          >
            {cat === 'all' ? '📋 Все' : (CATEGORY_LABELS[cat] ?? cat)}
          </button>
        ))}
      </div>

      {q.isLoading && (
        <div>
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="ox-skeleton" style={{ height: 64, marginBottom: 8 }} />
          ))}
        </div>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
        {visible.map((e) => <AchievementCard key={e.key} e={e} />)}
      </div>
    </div>
  );
}
