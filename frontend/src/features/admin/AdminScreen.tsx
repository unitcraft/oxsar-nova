import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';
import { Confirm } from '@/ui/Confirm';
import { AdminUserProfilePanel } from './AdminUserProfilePanel';

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

type AdminTab = 'users' | 'events' | 'audit';

export function AdminScreen() {
  const { t } = useTranslation('adminUi');
  const qc = useQueryClient();
  const [tab, setTab] = useState<AdminTab>('users');
  const [creditUserID, setCreditUserID] = useState('');
  const [creditAmount, setCreditAmount] = useState(0);
  const [roleUserID, setRoleUserID] = useState('');
  const [roleValue, setRoleValue] = useState('');
  const [profileUserID, setProfileUserID] = useState<string | null>(null);
  const [pendingConfirm, setPendingConfirm] = useState<{
    message: string;
    danger: boolean;
    action: () => void;
  } | null>(null);
  const ask = (message: string, action: () => void, danger = true) =>
    setPendingConfirm({ message, action, danger });

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
      <h2>{t('title')}</h2>

      <div className="ox-tabs" style={{ marginBottom: 16 }}>
        <button type="button" aria-pressed={tab === 'users'} onClick={() => setTab('users')}>
          {t('tabUsers')}
        </button>
        <button type="button" aria-pressed={tab === 'events'} onClick={() => setTab('events')}>
          {t('tabEvents')}
        </button>
        <button type="button" aria-pressed={tab === 'audit'} onClick={() => setTab('audit')}>
          {t('tabAudit')}
        </button>
      </div>

      {tab === 'audit' && <AdminAuditTab />}
      {tab === 'events' && <AdminEventsTab />}
      {tab !== 'users' ? null : (<>

      {stats.data && (
        <div style={{ display: 'flex', gap: 24, marginBottom: 16, flexWrap: 'wrap' }}>
          <StatCard label={t('statUsers')} value={stats.data.users} />
          <StatCard label={t('statPlanets')} value={stats.data.planets} />
          <StatCard label={t('statFleets')} value={stats.data.fleets_active} />
          <StatCard label={t('statEvents')} value={stats.data.events_pending} />
        </div>
      )}

      <h3>{t('sectionActions')}</h3>
      <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap', marginBottom: 16 }}>
        <div style={{ border: '1px solid #444', padding: 12, borderRadius: 4 }}>
          <b>{t('creditTitle')}</b>
          <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
            <input
              placeholder="user_id"
              value={creditUserID}
              onChange={(e) => setCreditUserID(e.target.value)}
              style={{ width: 280 }}
            />
            <input
              type="number"
              placeholder={t('creditAmountPlaceholder')}
              value={creditAmount}
              onChange={(e) => setCreditAmount(Number(e.target.value))}
              style={{ width: 80 }}
            />
            <button
              type="button"
              disabled={!creditUserID || credit.isPending}
              onClick={() =>
                ask(
                  t('creditConfirm', { amount: String(creditAmount), id: creditUserID.slice(0, 8) }),
                  () => credit.mutate({ id: creditUserID, amount: creditAmount }),
                  creditAmount < 0,
                )
              }
            >
              {t('okBtn')}
            </button>
          </div>
        </div>

        <div style={{ border: '1px solid #444', padding: 12, borderRadius: 4 }}>
          <b>{t('roleTitle')}</b>
          <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
            <input
              placeholder="user_id"
              value={roleUserID}
              onChange={(e) => setRoleUserID(e.target.value)}
              style={{ width: 280 }}
            />
            <select value={roleValue} onChange={(e) => setRoleValue(e.target.value)}>
              <option value="">user</option>
              <option value="support">support</option>
              <option value="admin">admin</option>
              <option value="superadmin">superadmin</option>
            </select>
            <button
              type="button"
              disabled={!roleUserID || setRole.isPending}
              onClick={() =>
                ask(
                  t('roleConfirm', { role: roleValue || 'user', id: roleUserID.slice(0, 8) }),
                  () => setRole.mutate({ id: roleUserID, role: roleValue }),
                )
              }
            >
              {t('okBtn')}
            </button>
          </div>
        </div>
      </div>

      <AutomsgsPanel />

      <h3>{t('sectionUsers')}</h3>
      {users.isLoading && <p>…</p>}
      {users.data && (
        <table className="ox-table">
          <thead>
            <tr>
              <th>{t('colPlayer')}</th>
              <th>{t('colRole')}</th>
              <th>{t('colCredits')}</th>
              <th>{t('colScore')}</th>
              <th>{t('colCreated')}</th>
              <th>{t('colActions')}</th>
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
                  <button
                    type="button"
                    className="btn-ghost btn-sm"
                    onClick={() => setProfileUserID(u.id)}
                    title={t('profileBtn')}
                  >
                    {t('profileBtn')}
                  </button>
                  {u.banned_at ? (
                    <button
                      type="button"
                      disabled={unban.isPending}
                      onClick={() => ask(t('unbanConfirm', { name: u.username }), () => unban.mutate(u.id), false)}
                    >
                      {t('unbanBtn')}
                    </button>
                  ) : (
                    <button
                      type="button"
                      disabled={ban.isPending}
                      onClick={() => ask(t('banConfirm', { name: u.username }), () => ban.mutate(u.id))}
                    >
                      {t('banBtn')}
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      </>)}

      {profileUserID && (
        <AdminUserProfilePanel
          userID={profileUserID}
          onClose={() => setProfileUserID(null)}
        />
      )}

      {pendingConfirm && (
        <Confirm
          title={t('confirmTitle')}
          message={pendingConfirm.message}
          confirmLabel={t('confirmBtn')}
          cancelLabel={t('cancelBtn')}
          danger={pendingConfirm.danger}
          onConfirm={() => {
            pendingConfirm.action();
            setPendingConfirm(null);
          }}
          onCancel={() => setPendingConfirm(null)}
        />
      )}
    </section>
  );
}

function AutomsgsPanel() {
  const { t } = useTranslation('adminUi');
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
      <h3>{t('sectionAutomsgs')}</h3>
      {defs.isLoading && <p>…</p>}
      {list.length > 0 && (
        <table className="ox-table" style={{ marginBottom: 8 }}>
          <thead>
            <tr>
              <th>{t('automsgColKey')}</th>
              <th>{t('automsgColTitle')}</th>
              <th>{t('automsgColFolder')}</th>
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
                    {t('automsgEditBtn')}
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
              {t('automsgTitleLabel')}{' '}
              <input
                value={editing.title}
                onChange={(e) => setEditing({ ...editing, title: e.target.value })}
                style={{ width: 340 }}
              />
            </label>
          </div>
          <div style={{ marginTop: 8 }}>
            <label>
              {t('automsgFolderLabel')}{' '}
              <input
                type="number"
                value={editing.folder}
                onChange={(e) => setEditing({ ...editing, folder: Number(e.target.value) })}
                style={{ width: 60 }}
              />
            </label>
          </div>
          <div style={{ marginTop: 8 }}>
            <div style={{ marginBottom: 4 }}>{t('automsgBodyLabel')}</div>
            <textarea
              value={editing.body_template}
              onChange={(e) => setEditing({ ...editing, body_template: e.target.value })}
              rows={6}
              style={{ width: '100%', boxSizing: 'border-box', fontFamily: 'monospace', fontSize: '0.85em' }}
            />
          </div>
          <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
            <button type="button" disabled={save.isPending} onClick={() => save.mutate(editing)}>
              {t('automsgSaveBtn')}
            </button>
            <button type="button" onClick={() => setEditing(null)}>
              {t('automsgCancelBtn')}
            </button>
          </div>
          {save.isError && (
            <p className="ox-error">
              {save.error instanceof Error ? save.error.message : t('automsgSaveErr')}
            </p>
          )}
        </div>
      )}
    </div>
  );
}

interface EventRow {
  id: string;
  user_id?: string;
  planet_id?: string;
  kind: number;
  state: string;
  fire_at: string;
  created_at: string;
  processed_at?: string;
  attempt: number;
  last_error?: string;
}

interface EventsStats {
  by_state: { state: string; count: number }[];
  top_errors_24h: { kind: number; count: number }[];
  oldest_wait_lag_s: number | null;
}

function AdminEventsMonitor() {
  const { t } = useTranslation('adminUi');
  const qc = useQueryClient();
  const [stateFilter, setStateFilter] = useState<'all' | 'wait' | 'error' | 'ok'>('error');

  const stats = useQuery({
    queryKey: ['admin-events-stats'],
    queryFn: () => api.get<EventsStats>('/api/admin/events/stats'),
    refetchInterval: 15000,
  });

  const events = useQuery({
    queryKey: ['admin-events', stateFilter],
    queryFn: () => {
      const q = stateFilter === 'all' ? '' : `?state=${stateFilter}&limit=50`;
      return api.get<{ events: EventRow[] }>(`/api/admin/events${q}`);
    },
    refetchInterval: 15000,
  });

  const retry = useMutation({
    mutationFn: (id: string) => api.post(`/api/admin/events/${id}/retry`, {}),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['admin-events'] });
      void qc.invalidateQueries({ queryKey: ['admin-events-stats'] });
    },
  });
  const cancel = useMutation({
    mutationFn: (id: string) => api.post(`/api/admin/events/${id}/cancel`, {}),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['admin-events'] }),
  });

  return (
    <section style={{ marginBottom: 16 }}>
      <h3>{t('eventsMonitorTitle')}</h3>

      {stats.data && (
        <div style={{ display: 'flex', gap: 12, marginBottom: 8, flexWrap: 'wrap' }}>
          {stats.data.by_state?.map((s) => (
            <StatCard key={s.state} label={`${s.state}`} value={s.count} />
          ))}
          <StatCard
            label="lag wait (sec)"
            value={Math.round(stats.data.oldest_wait_lag_s ?? 0)}
          />
        </div>
      )}

      {stats.data && stats.data.top_errors_24h?.length > 0 && (
        <div style={{ marginBottom: 8, fontSize: 14, color: 'var(--ox-fg-dim)' }}>
          {t('eventsTopErrors')}
          {' '}
          {stats.data.top_errors_24h.map((e) => (
            <span key={e.kind} style={{ marginRight: 8, fontFamily: 'var(--ox-mono)' }}>
              kind={e.kind}:{e.count}
            </span>
          ))}
        </div>
      )}

      <div style={{ display: 'flex', gap: 6, marginBottom: 8 }}>
        {(['error', 'wait', 'ok', 'all'] as const).map((s) => (
          <button
            key={s}
            type="button"
            className={stateFilter === s ? 'btn btn-sm' : 'btn-ghost btn-sm'}
            onClick={() => setStateFilter(s)}
          >
            {s}
          </button>
        ))}
      </div>

      {events.isLoading && <div>{t('eventsLoading')}</div>}
      {events.data && events.data.events.length === 0 && (
        <div style={{ color: 'var(--ox-fg-muted)', fontSize: 14 }}>{t('eventsEmpty')}</div>
      )}
      {events.data && events.data.events.length > 0 && (
        <div style={{ overflowX: 'auto' }}>
          <table className="ox-table" style={{ fontSize: 14 }}>
            <thead>
              <tr>
                <th>{t('eventsColId')}</th>
                <th>{t('eventsColKind')}</th>
                <th>{t('eventsColState')}</th>
                <th>{t('eventsColAtt')}</th>
                <th>{t('eventsColFire')}</th>
                <th>{t('eventsColLastErr')}</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {events.data.events.map((e) => (
                <tr key={e.id}>
                  <td style={{ fontFamily: 'var(--ox-mono)', fontSize: 10 }}>{e.id.slice(0, 8)}</td>
                  <td className="num">{e.kind}</td>
                  <td>
                    <span style={{
                      color: e.state === 'error' ? 'var(--ox-danger)'
                           : e.state === 'wait' ? 'var(--ox-warning, #f59e0b)'
                           : 'var(--ox-fg-dim)',
                      fontFamily: 'var(--ox-mono)',
                    }}>{e.state}</span>
                  </td>
                  <td className="num">{e.attempt}</td>
                  <td style={{ fontSize: 10, fontFamily: 'var(--ox-mono)' }}>
                    {new Date(e.fire_at).toLocaleString('ru-RU', { dateStyle: 'short', timeStyle: 'short' })}
                  </td>
                  <td style={{ fontSize: 13, color: 'var(--ox-fg-dim)', maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }} title={e.last_error ?? ''}>
                    {e.last_error ?? '—'}
                  </td>
                  <td>
                    <button type="button" className="btn-ghost btn-sm" disabled={retry.isPending} onClick={() => retry.mutate(e.id)}>↻</button>
                    <button type="button" className="btn-ghost btn-sm" disabled={cancel.isPending} onClick={() => cancel.mutate(e.id)} style={{ color: 'var(--ox-danger)' }}>✕</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </section>
  );
}

function StatCard({ label, value }: { label: string; value: number }) {
  return (
    <div style={{ border: '1px solid #444', padding: '8px 16px', borderRadius: 4, minWidth: 120 }}>
      <div style={{ fontSize: 13, color: 'var(--ox-muted, #888)' }}>{label}</div>
      <div style={{ fontSize: 24, fontWeight: 700 }}>{value}</div>
    </div>
  );
}

interface DeadEvent {
  id: string;
  user_id?: string | null;
  planet_id?: string | null;
  kind: number;
  fire_at: string;
  payload: Record<string, unknown>;
  attempt: number;
  last_error: string;
  failed_at: string;
}

function AdminEventsTab() {
  const { t } = useTranslation('adminUi');
  const qc = useQueryClient();
  const dead = useQuery({
    queryKey: ['admin-events-dead'],
    queryFn: () => api.get<{ events: DeadEvent[] }>('/api/admin/events/dead?limit=200'),
    refetchInterval: 30000,
  });

  const resurrect = useMutation({
    mutationFn: (id: string) => api.post(`/api/admin/events/dead/${id}/resurrect`, {}),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['admin-events-dead'] });
      void qc.invalidateQueries({ queryKey: ['admin-events'] });
    },
  });

  const list = dead.data?.events ?? [];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      <AdminEventsMonitor />

      <div>
        <h3 style={{ margin: '0 0 8px 0' }}>{t('deadTitle')} ({list.length})</h3>
        {dead.isLoading && <p>{t('deadLoading')}</p>}
        {dead.isError && <p style={{ color: 'var(--ox-danger)' }}>{t('deadLoadErr')}</p>}
        {list.length === 0 && !dead.isLoading && (
          <p style={{ color: 'var(--ox-fg-muted)' }}>{t('deadEmpty')}</p>
        )}
        {list.length > 0 && (
          <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 14 }}>
            <thead>
              <tr style={{ textAlign: 'left', borderBottom: '1px solid #444' }}>
                <th style={{ padding: '6px 8px' }}>{t('deadColFailed')}</th>
                <th style={{ padding: '6px 8px' }}>{t('deadColKind')}</th>
                <th style={{ padding: '6px 8px' }}>{t('deadColAttempt')}</th>
                <th style={{ padding: '6px 8px' }}>{t('deadColError')}</th>
                <th style={{ padding: '6px 8px' }}>{t('deadColTarget')}</th>
                <th style={{ padding: '6px 8px' }} />
              </tr>
            </thead>
            <tbody>
              {list.map((e) => (
                <tr key={e.id} style={{ borderBottom: '1px solid #2a2a2a' }}>
                  <td style={{ padding: '6px 8px', whiteSpace: 'nowrap', fontFamily: 'var(--ox-mono)', fontSize: 13 }}>
                    {new Date(e.failed_at).toLocaleString('ru-RU')}
                  </td>
                  <td style={{ padding: '6px 8px' }}>{e.kind}</td>
                  <td style={{ padding: '6px 8px' }}>{e.attempt}</td>
                  <td style={{ padding: '6px 8px', color: 'var(--ox-danger)', maxWidth: 320, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                      title={e.last_error}>
                    {e.last_error}
                  </td>
                  <td style={{ padding: '6px 8px', fontFamily: 'var(--ox-mono)', fontSize: 13 }}>
                    {e.user_id ? e.user_id.slice(0, 8) : '—'} / {e.planet_id ? e.planet_id.slice(0, 8) : '—'}
                  </td>
                  <td style={{ padding: '6px 8px' }}>
                    <button
                      type="button"
                      className="btn-sm"
                      disabled={resurrect.isPending}
                      onClick={() => {
                        if (confirm(t('deadRetryConfirm', { id: e.id.slice(0, 8) }))) {
                          resurrect.mutate(e.id);
                        }
                      }}
                    >
                      ♻ Retry
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

interface AuditEntry {
  id: string;
  admin_id: string;
  admin_name: string;
  action: string;
  target_kind: string;
  target_id: string;
  payload: Record<string, unknown>;
  status: number;
  ip?: string;
  user_agent?: string;
  created_at: string;
}

interface AuditResponse {
  entries: AuditEntry[];
  limit: number;
  offset: number;
}

function AdminAuditTab() {
  const { t } = useTranslation('adminUi');
  const [actionFilter, setActionFilter] = useState('');
  const [targetFilter, setTargetFilter] = useState('');

  const q = useQuery({
    queryKey: ['admin-audit', actionFilter, targetFilter],
    queryFn: () => {
      const qs = new URLSearchParams();
      if (actionFilter) qs.set('action', actionFilter);
      if (targetFilter) qs.set('target_id', targetFilter);
      qs.set('limit', '100');
      return api.get<AuditResponse>(`/api/admin/audit?${qs.toString()}`);
    },
    refetchInterval: 15000,
  });

  const entries = q.data?.entries ?? [];

  return (
    <div>
      <div style={{ display: 'flex', gap: 12, marginBottom: 12, flexWrap: 'wrap' }}>
        <input
          placeholder={t('auditActionPlaceholder')}
          value={actionFilter}
          onChange={(e) => setActionFilter(e.target.value)}
          style={{ minWidth: 220 }}
        />
        <input
          placeholder={t('auditTargetPlaceholder')}
          value={targetFilter}
          onChange={(e) => setTargetFilter(e.target.value)}
          style={{ minWidth: 220 }}
        />
        <button type="button" onClick={() => q.refetch()}>{t('auditRefreshBtn')}</button>
      </div>

      {q.isLoading && <p>{t('auditLoading')}</p>}
      {q.isError && <p style={{ color: 'var(--ox-danger)' }}>{t('auditLoadErr')}</p>}
      {q.data && entries.length === 0 && (
        <p style={{ color: 'var(--ox-fg-muted)' }}>{t('auditEmpty')}</p>
      )}

      {entries.length > 0 && (
        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 14 }}>
          <thead>
            <tr style={{ textAlign: 'left', borderBottom: '1px solid #444' }}>
              <th style={{ padding: '6px 8px' }}>{t('auditColDate')}</th>
              <th style={{ padding: '6px 8px' }}>{t('auditColAdmin')}</th>
              <th style={{ padding: '6px 8px' }}>{t('auditColAction')}</th>
              <th style={{ padding: '6px 8px' }}>{t('auditColTarget')}</th>
              <th style={{ padding: '6px 8px' }}>{t('auditColPayload')}</th>
              <th style={{ padding: '6px 8px' }}>{t('auditColIp')}</th>
            </tr>
          </thead>
          <tbody>
            {entries.map((e) => (
              <tr key={e.id} style={{ borderBottom: '1px solid #2a2a2a' }}>
                <td style={{ padding: '6px 8px', fontFamily: 'var(--ox-mono)', fontSize: 13, whiteSpace: 'nowrap' }}>
                  {new Date(e.created_at).toLocaleString('ru-RU')}
                </td>
                <td style={{ padding: '6px 8px' }}>{e.admin_name || e.admin_id.slice(0, 8)}</td>
                <td style={{ padding: '6px 8px', fontFamily: 'var(--ox-mono)' }}>{e.action}</td>
                <td style={{ padding: '6px 8px' }}>
                  {e.target_kind && <span style={{ color: 'var(--ox-fg-muted)' }}>{e.target_kind}:</span>}
                  {e.target_id && <span style={{ fontFamily: 'var(--ox-mono)', fontSize: 13, marginLeft: 4 }}>{e.target_id.slice(0, 12)}</span>}
                </td>
                <td style={{ padding: '6px 8px', fontFamily: 'var(--ox-mono)', fontSize: 12, maxWidth: 320, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}
                    title={JSON.stringify(e.payload)}>
                  {JSON.stringify(e.payload)}
                </td>
                <td style={{ padding: '6px 8px', fontFamily: 'var(--ox-mono)', fontSize: 13 }}>
                  {e.ip || '—'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
