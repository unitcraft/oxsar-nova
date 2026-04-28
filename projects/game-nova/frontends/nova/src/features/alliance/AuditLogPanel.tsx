// План 67 Ф.5 ч.2 — alliance audit-log UI (U-013).
//
// Backend: GET /api/alliances/{id}/audit?action=&actor_id=&limit=&offset=
// Доступ — любой член альянса (бэкенд проверяет).
//
// Фильтры: dropdown по action (18 known + "all"), input по actor (имя)
// — выбор из members. Пагинация offset/limit, default 50/page.

import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';
import {
  AUDIT_ACTIONS,
  auditActionLabelKey,
  auditTargetKindLabelKey,
  formatRelativeTime,
} from './audit';

interface AuditEntry {
  id: string;
  alliance_id: string;
  actor_id: string;
  actor_name: string;
  action: string;
  target_kind: string;
  target_id: string;
  payload: Record<string, unknown> | null;
  created_at: string;
}

interface Member {
  user_id: string;
  username: string;
}

const PAGE_SIZE = 50;

export function AuditLogPanel({
  allianceID,
  members,
}: {
  allianceID: string;
  members: Member[];
}) {
  const { t } = useTranslation('alliance');
  const [action, setAction] = useState<string>('');
  const [actorID, setActorID] = useState<string>('');
  const [offset, setOffset] = useState(0);

  // Сброс пагинации при смене фильтров — иначе можем оказаться на
  // несуществующей странице.
  const setActionAndReset = (a: string) => { setAction(a); setOffset(0); };
  const setActorAndReset = (uid: string) => { setActorID(uid); setOffset(0); };

  const queryStr = useMemo(() => {
    const params = new URLSearchParams();
    params.set('limit', String(PAGE_SIZE));
    params.set('offset', String(offset));
    if (action) params.set('action', action);
    if (actorID) params.set('actor_id', actorID);
    return params.toString();
  }, [action, actorID, offset]);

  const audit = useQuery({
    queryKey: ['alliances', allianceID, 'audit', queryStr],
    queryFn: () =>
      api.get<{ entries: AuditEntry[] | null; limit: number; offset: number }>(
        `/api/alliances/${allianceID}/audit?${queryStr}`,
      ),
  });

  const entries = audit.data?.entries ?? [];
  const now = new Date();

  return (
    <div className="ox-panel" style={{ padding: '12px 16px', display: 'flex', flexDirection: 'column', gap: 10 }}>
      <div
        style={{
          fontSize: 13,
          fontWeight: 700,
          letterSpacing: '0.08em',
          textTransform: 'uppercase',
          color: 'var(--ox-fg-muted)',
        }}
      >
        {t('audit.title')}
      </div>

      <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap', alignItems: 'center' }}>
        <label style={{ display: 'flex', flexDirection: 'column', gap: 2, flex: '1 1 200px' }}>
          <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>{t('audit.filterAction')}</span>
          <select value={action} onChange={(e) => setActionAndReset(e.target.value)}>
            <option value="">{t('diplomacy.filterAll')}</option>
            {AUDIT_ACTIONS.map((a) => (
              <option key={a} value={a}>
                {t(auditActionLabelKey(a))}
              </option>
            ))}
          </select>
        </label>
        <label style={{ display: 'flex', flexDirection: 'column', gap: 2, flex: '1 1 200px' }}>
          <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>{t('audit.filterActor')}</span>
          <select value={actorID} onChange={(e) => setActorAndReset(e.target.value)}>
            <option value="">{t('diplomacy.filterAll')}</option>
            {members.map((m) => (
              <option key={m.user_id} value={m.user_id}>{m.username}</option>
            ))}
          </select>
        </label>
      </div>

      {audit.isLoading && <div className="ox-skeleton" style={{ height: 80 }} />}

      {!audit.isLoading && entries.length === 0 && (
        <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', fontStyle: 'italic' }}>
          {t('audit.empty')}
        </div>
      )}

      {entries.length > 0 && (
        <table className="ox-table" style={{ margin: 0, fontSize: 14 }}>
          <thead>
            <tr>
              <th>{t('audit.colWhen')}</th>
              <th>{t('audit.colActor')}</th>
              <th>{t('audit.colAction')}</th>
              <th>{t('audit.colTarget')}</th>
            </tr>
          </thead>
          <tbody>
            {entries.map((e) => {
              const when = new Date(e.created_at);
              const rel = formatRelativeTime(now, when);
              const actorLabel = e.actor_name || (e.actor_id ? '—' : t('audit.actorSystem'));
              const actionLabel = t(auditActionLabelKey(e.action), { name: e.action });
              const targetLabel =
                e.target_kind
                  ? `${t(auditTargetKindLabelKey(e.target_kind), { kind: e.target_kind })}${e.target_id ? ' · ' + e.target_id.slice(0, 8) : ''}`
                  : '—';
              return (
                <tr key={e.id}>
                  <td
                    style={{ color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)', fontSize: 13 }}
                    title={when.toLocaleString('ru-RU')}
                  >
                    {t(rel.key, rel.vars)}
                  </td>
                  <td>{actorLabel}</td>
                  <td style={{ fontWeight: 700 }}>{actionLabel}</td>
                  <td style={{ color: 'var(--ox-fg-dim)' }}>{targetLabel}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}

      <div style={{ display: 'flex', gap: 8, alignItems: 'center', marginTop: 4 }}>
        <button
          type="button"
          className="btn-ghost btn-sm"
          disabled={offset === 0}
          onClick={() => setOffset(Math.max(0, offset - PAGE_SIZE))}
        >
          ← {t('audit.prevPage')}
        </button>
        <button
          type="button"
          className="btn-ghost btn-sm"
          disabled={entries.length < PAGE_SIZE}
          onClick={() => setOffset(offset + PAGE_SIZE)}
        >
          {t('audit.nextPage')} →
        </button>
        <span style={{ marginLeft: 'auto', fontSize: 13, color: 'var(--ox-fg-muted)' }}>
          {t('audit.pageInfo', { from: String(offset + 1), to: String(offset + entries.length) })}
        </span>
      </div>
    </div>
  );
}
