import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';

interface AdminStats {
  users: number;
  planets: number;
  fleets_active: number;
  events_pending: number;
}

interface AdminUser {
  id: string;
  username: string;
  email: string;
  role: string | undefined;
  banned_at: string | null | undefined;
  credit: number;
  score: number;
  created_at: string;
}

export function AdminScreen() {
  const qc = useQueryClient();
  const [creditUserID, setCreditUserID] = useState('');
  const [creditAmount, setCreditAmount] = useState(0);
  const [roleUserID, setRoleUserID] = useState('');
  const [roleValue, setRoleValue] = useState('');

  const stats = useQuery({
    queryKey: ['admin-stats'],
    queryFn: () => api.get<AdminStats>('/api/admin/stats'),
    refetchInterval: 30000,
  });

  const users = useQuery({
    queryKey: ['admin-users'],
    queryFn: () => api.get<{ users: AdminUser[] }>('/api/admin/users'),
    staleTime: 10000,
  });

  const ban = useMutation({
    mutationFn: (id: string) => api.post(`/api/admin/users/${id}/ban`, {}),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['admin-users'] }),
  });

  const unban = useMutation({
    mutationFn: (id: string) => api.post(`/api/admin/users/${id}/unban`, {}),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['admin-users'] }),
  });

  const credit = useMutation({
    mutationFn: ({ id, amount }: { id: string; amount: number }) =>
      api.post(`/api/admin/users/${id}/credit`, { amount }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['admin-users'] });
      setCreditUserID('');
      setCreditAmount(0);
    },
  });

  const setRole = useMutation({
    mutationFn: ({ id, role }: { id: string; role: string }) =>
      api.post(`/api/admin/users/${id}/role`, { role }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['admin-users'] });
      setRoleUserID('');
      setRoleValue('');
    },
  });

  return (
    <section>
      <h2>Admin Panel</h2>

      {stats.data && (
        <div style={{ display: 'flex', gap: 24, marginBottom: 16, flexWrap: 'wrap' }}>
          <StatCard label="Пользователей" value={stats.data.users} />
          <StatCard label="Планет" value={stats.data.planets} />
          <StatCard label="Флотов в пути" value={stats.data.fleets_active} />
          <StatCard label="Событий в очереди" value={stats.data.events_pending} />
        </div>
      )}

      <h3>Действия</h3>
      <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap', marginBottom: 16 }}>
        <div style={{ border: '1px solid #444', padding: 12, borderRadius: 4 }}>
          <b>Начислить кредиты</b>
          <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
            <input
              placeholder="user_id"
              value={creditUserID}
              onChange={(e) => setCreditUserID(e.target.value)}
              style={{ width: 280 }}
            />
            <input
              type="number"
              placeholder="сумма"
              value={creditAmount}
              onChange={(e) => setCreditAmount(Number(e.target.value))}
              style={{ width: 80 }}
            />
            <button
              type="button"
              disabled={!creditUserID || credit.isPending}
              onClick={() => credit.mutate({ id: creditUserID, amount: creditAmount })}
            >
              ОК
            </button>
          </div>
        </div>

        <div style={{ border: '1px solid #444', padding: 12, borderRadius: 4 }}>
          <b>Установить роль</b>
          <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
            <input
              placeholder="user_id"
              value={roleUserID}
              onChange={(e) => setRoleUserID(e.target.value)}
              style={{ width: 280 }}
            />
            <select value={roleValue} onChange={(e) => setRoleValue(e.target.value)}>
              <option value="">user</option>
              <option value="admin">admin</option>
              <option value="superadmin">superadmin</option>
            </select>
            <button
              type="button"
              disabled={!roleUserID || setRole.isPending}
              onClick={() => setRole.mutate({ id: roleUserID, role: roleValue })}
            >
              ОК
            </button>
          </div>
        </div>
      </div>

      <h3>Пользователи</h3>
      {users.isLoading && <p>…</p>}
      {users.data && (
        <table className="ox-table">
          <thead>
            <tr>
              <th>Username</th>
              <th>Role</th>
              <th>Credit</th>
              <th>Score</th>
              <th>Создан</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {users.data.users.map((u) => (
              <tr key={u.id} style={{ opacity: u.banned_at ? 0.5 : 1 }}>
                <td>
                  {u.username}
                  {u.banned_at ? ' 🚫' : ''}
                </td>
                <td>{u.role ?? 'user'}</td>
                <td>{u.credit}</td>
                <td>{u.score}</td>
                <td>{new Date(u.created_at).toLocaleDateString('ru-RU')}</td>
                <td style={{ display: 'flex', gap: 4 }}>
                  {u.banned_at ? (
                    <button
                      type="button"
                      disabled={unban.isPending}
                      onClick={() => unban.mutate(u.id)}
                    >
                      Разбан
                    </button>
                  ) : (
                    <button
                      type="button"
                      disabled={ban.isPending}
                      onClick={() => ban.mutate(u.id)}
                    >
                      Бан
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div style={{ border: '1px solid #444', padding: '8px 16px', borderRadius: 4, minWidth: 120 }}>
      <div style={{ fontSize: 11, color: 'var(--ox-muted, #888)' }}>{label}</div>
      <div style={{ fontSize: 24, fontWeight: 700 }}>{value}</div>
    </div>
  );
}
