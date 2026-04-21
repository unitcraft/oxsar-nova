import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

// AchievementsScreen — lazy-trigger: backend CheckAll выполняется
// при каждом GET, поэтому список всегда актуален без ws/refresh.

interface Entry {
  key: string;
  title: string;
  description: string;
  points: number;
  unlocked_at?: string | null;
  progress?: number | undefined;
  progress_max?: number | undefined;
}

export function AchievementsScreen() {
  const { t, tf } = useTranslation();
  const q = useQuery({
    queryKey: ['achievements'],
    queryFn: () => api.get<{ achievements: Entry[] | null }>('/api/achievements'),
    refetchInterval: 15000,
  });

  const list = q.data?.achievements ?? [];
  const unlockedCount = list.filter((e) => e.unlocked_at).length;
  const totalPoints = list
    .filter((e) => e.unlocked_at)
    .reduce((acc, e) => acc + e.points, 0);

  return (
    <section>
      <h2>{tf('global', 'MENU_ACHIEVEMENTS', 'Достижения')}</h2>
      <p>
        <b>{tf('Main', 'ACH_PROGRESS', 'Открыто')}:</b> {unlockedCount} / {list.length}
        {' · '}
        <b>{tf('Main', 'ACH_POINTS', 'Очки')}:</b> {totalPoints}
      </p>

      {q.isLoading && <p>…</p>}
      {q.error && (
        <p className="ox-error">
          {t('global', 'ERROR')}: {q.error instanceof Error ? q.error.message : ''}
        </p>
      )}

      {list.length > 0 && (
        <table className="ox-table">
          <thead>
            <tr>
              <th>{tf('Main', 'ACH_TITLE', 'Достижение')}</th>
              <th>{tf('Main', 'ACH_DESC', 'Описание')}</th>
              <th>{tf('Main', 'ACH_POINTS', 'Очки')}</th>
              <th>{tf('Main', 'ACH_UNLOCKED_AT', 'Получено')}</th>
            </tr>
          </thead>
          <tbody>
            {list.map((e) => (
              <tr
                key={e.key}
                style={{
                  opacity: e.unlocked_at ? 1 : 0.5,
                  fontWeight: e.unlocked_at ? 600 : 400,
                }}
              >
                <td>
                  {e.unlocked_at ? '✓ ' : '○ '}
                  {e.title}
                </td>
                <td>
                  {e.description}
                  {e.progress_max != null && !e.unlocked_at && (
                    <span style={{ marginLeft: 8, color: '#888', fontSize: '0.85em' }}>
                      {e.progress ?? 0} / {e.progress_max}
                    </span>
                  )}
                </td>
                <td className="num">{e.points}</td>
                <td>{e.unlocked_at ? new Date(e.unlocked_at).toLocaleString('ru-RU') : '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}
