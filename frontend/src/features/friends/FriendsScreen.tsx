import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';

interface Friend {
  user_id: string;
  username: string;
  points: number;
  last_seen?: string | null;
  alliance_tag?: string | null;
}

function formatActivity(lastSeen?: string | null): { label: string; color: string } {
  if (!lastSeen) return { label: '—', color: 'var(--ox-fg-muted)' };
  const mins = Math.floor((Date.now() - new Date(lastSeen).getTime()) / 60000);
  if (mins < 5) return { label: 'онлайн', color: 'var(--ox-success)' };
  if (mins < 60) return { label: `${mins} мин назад`, color: 'var(--ox-fg-dim)' };
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return { label: `${hrs} ч назад`, color: 'var(--ox-fg-muted)' };
  const days = Math.floor(hrs / 24);
  return { label: `${days} дн назад`, color: 'var(--ox-fg-muted)' };
}

export function FriendsScreen() {
  const qc = useQueryClient();
  const q = useQuery({
    queryKey: ['friends'],
    queryFn: () => api.get<{ friends: Friend[] }>('/api/friends'),
    refetchInterval: 60000,
  });

  const remove = useMutation({
    mutationFn: (userID: string) => api.delete<void>(`/api/friends/${userID}`),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['friends'] }),
  });

  const list = q.data?.friends ?? [];

  return (
    <div style={{ maxWidth: 900, margin: '0 auto', padding: 24, display: 'flex', flexDirection: 'column', gap: 16 }}>
      <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700 }}>⭐ Друзья</h2>
      <p style={{ margin: 0, fontSize: 12, color: 'var(--ox-fg-dim)' }}>
        Друзья подсвечены в галактике звёздочкой. Добавить в друзья можно из результатов поиска (Ctrl+K).
      </p>

      <div className="ox-panel" style={{ overflow: 'hidden' }}>
        {q.isLoading && <div style={{ padding: 20 }}><div className="ox-skeleton" style={{ height: 120 }} /></div>}
        {!q.isLoading && list.length === 0 && (
          <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
            Список друзей пуст
          </div>
        )}
        {list.length > 0 && (
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th>Игрок</th>
                  <th style={{ width: 80 }}>Альянс</th>
                  <th className="num">Очки</th>
                  <th>Активность</th>
                  <th style={{ width: 100 }}></th>
                </tr>
              </thead>
              <tbody>
                {list.map((f) => {
                  const act = formatActivity(f.last_seen);
                  return (
                    <tr key={f.user_id}>
                      <td style={{ fontWeight: 600 }}>⭐ {f.username}</td>
                      <td style={{ fontSize: 11, color: 'var(--ox-fg-dim)', fontFamily: 'var(--ox-mono)' }}>
                        {f.alliance_tag ? `[${f.alliance_tag}]` : '—'}
                      </td>
                      <td className="num" style={{ color: 'var(--ox-accent)' }}>
                        {Math.round(f.points).toLocaleString('ru-RU')}
                      </td>
                      <td style={{ fontSize: 12, color: act.color }}>{act.label}</td>
                      <td>
                        <button
                          type="button"
                          className="btn-ghost btn-sm"
                          style={{ color: 'var(--ox-danger)' }}
                          disabled={remove.isPending}
                          onClick={() => remove.mutate(f.user_id)}
                        >
                          Удалить
                        </button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
