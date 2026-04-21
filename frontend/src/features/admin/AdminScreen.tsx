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

interface AutomsgDef {
  key: string;
  title: string;
  body_template: string;
  folder: number;
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
      <h2>Панель администратора</h2>

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

      <AutomsgsPanel />

      <h3>Пользователи</h3>
      {users.isLoading && <p>…</p>}
      {users.data && (
        <table className="ox-table">
          <thead>
            <tr>
              <th>Игрок</th>
              <th>Роль</th>
              <th>Кредиты</th>
              <th>Очки</th>
              <th>Создан</th>
              <th>Действия</th>
            </tr>
          </thead>
          <tbody>
            {(users.data.users ?? []).map((u) => (
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

function AutomsgsPanel() {
  const qc = useQueryClient();
  const [editing, setEditing] = useState<AutomsgDef | null>(null);

  const defs = useQuery({
    queryKey: ['admin-automsgs'],
    queryFn: () => api.get<{ defs: AutomsgDef[] | null }>('/api/admin/automsgs'),
  });

  const save = useMutation({
    mutationFn: (d: AutomsgDef) =>
      api.put<void>(`/api/admin/automsgs/${d.key}`, {
        title: d.title,
        body_template: d.body_template,
        folder: d.folder,
      }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['admin-automsgs'] });
      setEditing(null);
    },
  });

  const list = defs.data?.defs ?? [];

  return (
    <div style={{ marginBottom: 16 }}>
      <h3>Шаблоны сообщений</h3>
      {defs.isLoading && <p>…</p>}
      {list.length > 0 && (
        <table className="ox-table" style={{ marginBottom: 8 }}>
          <thead>
            <tr>
              <th>Ключ</th>
              <th>Заголовок</th>
              <th>Папка</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {list.map((d) => (
              <tr key={d.key}>
                <td style={{ fontFamily: 'monospace', fontSize: '0.85em' }}>{d.key}</td>
                <td>{d.title}</td>
                <td>{d.folder}</td>
                <td>
                  <button type="button" onClick={() => setEditing({ ...d })}>
                    Правка
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {editing && (
        <div style={{ border: '1px solid #444', padding: 12, borderRadius: 4, maxWidth: 600 }}>
          <b style={{ fontFamily: 'monospace' }}>{editing.key}</b>
          <div style={{ marginTop: 8 }}>
            <label>
              Заголовок:{' '}
              <input
                value={editing.title}
                onChange={(e) => setEditing({ ...editing, title: e.target.value })}
                style={{ width: 340 }}
              />
            </label>
          </div>
          <div style={{ marginTop: 8 }}>
            <label>
              Папка:{' '}
              <input
                type="number"
                value={editing.folder}
                onChange={(e) => setEditing({ ...editing, folder: Number(e.target.value) })}
                style={{ width: 60 }}
              />
            </label>
          </div>
          <div style={{ marginTop: 8 }}>
            <div style={{ marginBottom: 4 }}>Шаблон тела (поддерживает {'{{variable}}'})</div>
            <textarea
              value={editing.body_template}
              onChange={(e) => setEditing({ ...editing, body_template: e.target.value })}
              rows={6}
              style={{ width: '100%', boxSizing: 'border-box', fontFamily: 'monospace', fontSize: '0.85em' }}
            />
          </div>
          <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
            <button type="button" disabled={save.isPending} onClick={() => save.mutate(editing)}>
              Сохранить
            </button>
            <button type="button" onClick={() => setEditing(null)}>
              Отмена
            </button>
          </div>
          {save.isError && (
            <p className="ox-error">
              {save.error instanceof Error ? save.error.message : 'ошибка'}
            </p>
          )}
        </div>
      )}
    </div>
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
