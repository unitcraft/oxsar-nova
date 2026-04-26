import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

interface Friend {
  user_id: string;
  username: string;
  points: number;
  last_seen?: string | null;
  alliance_tag?: string | null;
}

function formatActivity(lastSeen: string | null | undefined, t: (k: string, v?: Record<string, string>) => string): { label: string; color: string } {
  if (!lastSeen) return { label: '—', color: 'var(--ox-fg-muted)' };
  const mins = Math.floor((Date.now() - new Date(lastSeen).getTime()) / 60000);
  if (mins < 5) return { label: t('statusOnline'), color: 'var(--ox-success)' };
  if (mins < 60) return { label: t('statusMinAgo', { n: String(mins) }), color: 'var(--ox-fg-dim)' };
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return { label: t('statusHourAgo', { n: String(hrs) }), color: 'var(--ox-fg-muted)' };
  const days = Math.floor(hrs / 24);
  return { label: t('statusDayAgo', { n: String(days) }), color: 'var(--ox-fg-muted)' };
}

export function FriendsScreen() {
  const { t } = useTranslation('friends');
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
      <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700 }}>⭐ {t('title')}</h2>
      <p style={{ margin: 0, fontSize: 14, color: 'var(--ox-fg-dim)' }}>
        {t('addPlaceholder')}
      </p>

      <div className="ox-panel" style={{ overflow: 'hidden' }}>
        {q.isLoading && <div style={{ padding: 20 }}><div className="ox-skeleton" style={{ height: 120 }} /></div>}
        {!q.isLoading && list.length === 0 && (
          <div style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-muted)' }}>
            {t('empty')}
          </div>
        )}
        {list.length > 0 && (
          <div className="ox-table-responsive">
            <table className="ox-table" style={{ margin: 0 }}>
              <thead>
                <tr>
                  <th>{t('colPlayer')}</th>
                  <th style={{ width: 80 }}>{t('colStatus')}</th>
                  <th className="num">{t('colActions')}</th>
                  <th>{t('statusOnline')}</th>
                  <th style={{ width: 100 }}></th>
                </tr>
              </thead>
              <tbody>
                {list.map((f) => {
                  const act = formatActivity(f.last_seen, t);
                  return (
                    <tr key={f.user_id}>
                      <td style={{ fontWeight: 600 }}>⭐ {f.username}</td>
                      <td style={{ fontSize: 13, color: 'var(--ox-fg-dim)', fontFamily: 'var(--ox-mono)' }}>
                        {f.alliance_tag ? `[${f.alliance_tag}]` : '—'}
                      </td>
                      <td className="num" style={{ color: 'var(--ox-accent)' }}>
                        {Math.round(f.points).toLocaleString('ru-RU')}
                      </td>
                      <td style={{ fontSize: 14, color: act.color }}>{act.label}</td>
                      <td>
                        <button
                          type="button"
                          className="btn-ghost btn-sm"
                          style={{ color: 'var(--ox-danger)' }}
                          disabled={remove.isPending}
                          onClick={() => remove.mutate(f.user_id)}
                        >
                          {t('removeBtn')}
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
