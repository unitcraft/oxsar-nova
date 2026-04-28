import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { Confirm } from '@/ui/Confirm';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';
import { ReportButton } from '@/components/ReportButton';
import { DescriptionsPanel } from './DescriptionsPanel';
import { RanksPanel } from './RanksPanel';
import { DiplomacyPanel } from './DiplomacyPanel';
import { AuditLogPanel } from './AuditLogPanel';
import { AllianceSearchPanel } from './AllianceSearchPanel';
import { TransferLeadershipDialog } from './TransferLeadershipDialog';
import {
  type AllianceSearchFilters,
  EMPTY_FILTERS,
  buildSearchQuery,
} from './search-filters';

interface Alliance {
  id: string;
  tag: string;
  name: string;
  description: string;
  is_open: boolean;
  owner_id: string;
  owner_name: string;
  member_count: number;
  created_at: string;
}

interface Member {
  user_id: string;
  username: string;
  rank: string;
  rank_name: string;
  joined_at: string;
}

interface Application {
  id: string;
  alliance_id: string;
  user_id: string;
  username: string;
  message: string;
  created_at: string;
}

export function AllianceScreen() {
  const { t } = useTranslation('alliance');
  const qc = useQueryClient();
  const toast = useToast();
  const [view, setView] = useState<'mine' | 'list' | 'create'>('mine');
  const [selectedID, setSelectedID] = useState<string | null>(null);
  const [filters, setFilters] = useState<AllianceSearchFilters>(EMPTY_FILTERS);
  // debouncedQuery — то, что реально идёт в queryKey TanStack Query.
  // Меняется через 300ms после последнего keystroke — иначе кеш
  // разрастается на каждый ввод и мы дёргаем бэк по N раз вместо 1.
  const [debouncedQuery, setDebouncedQuery] = useState<string>('');

  useEffect(() => {
    const next = buildSearchQuery(filters);
    const tid = window.setTimeout(() => setDebouncedQuery(next), 300);
    return () => window.clearTimeout(tid);
  }, [filters]);

  const mine = useQuery({
    queryKey: ['alliances', 'me'],
    queryFn: () => api.get<{ alliance: Alliance | null; members: Member[] | null }>('/api/alliances/me'),
    refetchInterval: 30000,
  });

  const list = useQuery({
    queryKey: ['alliances', 'search', debouncedQuery],
    queryFn: () => {
      const url = debouncedQuery ? `/api/alliances?${debouncedQuery}` : '/api/alliances';
      return api.get<{ alliances: Alliance[] | null }>(url);
    },
    enabled: view === 'list',
  });

  const detail = useQuery({
    queryKey: ['alliances', selectedID],
    queryFn: () => api.get<{ alliance: Alliance; members: Member[] }>(`/api/alliances/${selectedID}`),
    enabled: !!selectedID,
  });

  const join = useMutation({
    mutationFn: ({ id, message }: { id: string; message: string }) =>
      api.post<{ status?: string }>(`/api/alliances/${id}/join`, { message }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      setView('mine');
      toast.show('success', t('title'), t('alreadyApplied'));
    },
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  const leave = useMutation({
    mutationFn: () => api.post<void>('/api/alliances/leave'),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      setView('mine');
      toast.show('info', t('leaveBtn'));
    },
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  const disband = useMutation({
    mutationFn: (id: string) => api.delete<void>(`/api/alliances/${id}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      setView('mine');
      toast.show('info', t('disbandBtn'));
    },
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  const myAlliance = mine.data?.alliance ?? null;
  const myMembers = mine.data?.members ?? [];
  const amOwner = !!myAlliance && myMembers.some((m) => m.user_id === myAlliance.owner_id && m.rank === 'owner');

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
        🤝 {t('title')}
      </h2>

      <div className="ox-tabs">
        <button type="button" aria-pressed={view === 'mine'} onClick={() => setView('mine')}>
          {t('tabMine')}
        </button>
        <button type="button" aria-pressed={view === 'list'} onClick={() => setView('list')}>
          {t('tabList')}
        </button>
        {!myAlliance && (
          <button type="button" aria-pressed={view === 'create'} onClick={() => setView('create')}>
            {t('createTitle')}
          </button>
        )}
      </div>

      {view === 'mine' && (
        mine.isLoading
          ? <div className="ox-skeleton" style={{ height: 120 }} />
          : !myAlliance
            ? (
              <div className="ox-panel" style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-dim)' }}>
                {t('noAlliance')}
              </div>
            )
            : <MyAlliancePanel
                alliance={myAlliance}
                members={myMembers}
                isOwner={amOwner}
                onLeave={() => leave.mutate()}
                onDisband={() => disband.mutate(myAlliance.id)}
                leavePending={leave.isPending}
                disbandPending={disband.isPending}
              />
      )}

      {view === 'list' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <AllianceSearchPanel value={filters} onChange={setFilters} />
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, alignItems: 'start' }}>
          <div className="ox-panel" style={{ overflow: 'hidden' }}>
            {list.isLoading && <div style={{ padding: 16 }}><div className="ox-skeleton" style={{ height: 60 }} /></div>}
            {!list.isLoading && (
              <table className="ox-table" style={{ margin: 0 }}>
                <thead>
                  <tr>
                    <th>{t('labelTag')}</th>
                    <th>{t('sectionInfo')}</th>
                    <th>{t('colMember')}</th>
                    <th>{t('labelOpen')}</th>
                  </tr>
                </thead>
                <tbody>
                  {(list.data?.alliances ?? []).map((al) => (
                    <tr
                      key={al.id}
                      style={{ cursor: 'pointer', background: selectedID === al.id ? 'rgba(99,217,255,0.06)' : undefined }}
                      onClick={() => setSelectedID(al.id)}
                    >
                      <td style={{ fontFamily: 'var(--ox-mono)', fontWeight: 700 }}>[{al.tag}]</td>
                      <td>{al.name}</td>
                      <td className="num">{al.member_count}</td>
                      <td>{al.is_open ? '🔓' : '🔒'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          {selectedID && detail.data && (
            <AllianceDetail
              alliance={detail.data.alliance}
              members={detail.data.members ?? []}
              canJoin={!myAlliance}
              joining={join.isPending}
              onJoin={(msg) => join.mutate({ id: selectedID, message: msg })}
            />
          )}
          </div>
        </div>
      )}

      {view === 'create' && !myAlliance && (
        <CreateForm
          onCreated={() => { void qc.invalidateQueries({ queryKey: ['alliances'] }); setView('mine'); }}
          onCancel={() => setView('mine')}
        />
      )}
    </div>
  );
}

function MyAlliancePanel({
  alliance, members, isOwner, onLeave, onDisband, leavePending, disbandPending,
}: {
  alliance: Alliance;
  members: Member[];
  isOwner: boolean;
  onLeave: () => void;
  onDisband: () => void;
  leavePending: boolean;
  disbandPending: boolean;
}) {
  const { t } = useTranslation('alliance');
  const qc = useQueryClient();
  const toast = useToast();
  const [confirmDisband, setConfirmDisband] = useState(false);
  const [transferOpen, setTransferOpen] = useState(false);

  const setOpen = useMutation({
    mutationFn: ({ id, isOpen }: { id: string; isOpen: boolean }) =>
      api.patch<void>(`/api/alliances/${id}/open`, { is_open: isOpen }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  const approve = useMutation({
    mutationFn: (appID: string) => api.post<void>(`/api/alliances/applications/${appID}/approve`),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
  });

  const reject = useMutation({
    mutationFn: (appID: string) => api.delete<void>(`/api/alliances/applications/${appID}`),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
  });

  const apps = useQuery({
    queryKey: ['alliances', alliance.id, 'applications'],
    queryFn: () => api.get<{ applications: Application[] | null }>(`/api/alliances/${alliance.id}/applications`),
    enabled: isOwner && !alliance.is_open,
    refetchInterval: 15000,
  });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
      {/* Header card */}
      <div className="ox-panel" style={{ padding: '16px 20px' }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
          <span style={{ fontSize: 28, fontFamily: 'var(--ox-mono)', fontWeight: 800, color: 'var(--ox-accent)' }}>
            [{alliance.tag}]
          </span>
          <div>
            <div style={{ fontSize: 16, fontWeight: 700 }}>{alliance.name}</div>
            <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>
              {alliance.owner_name} · {alliance.member_count} · {alliance.is_open ? `🔓 ${t('labelOpen')}` : `🔒 ${t('labelClosed')}`}
            </div>
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
          {isOwner && (
            <button
              type="button"
              className="btn-ghost btn-sm"
              disabled={setOpen.isPending}
              onClick={() => setOpen.mutate({ id: alliance.id, isOpen: !alliance.is_open })}
            >
              {alliance.is_open ? `🔒 ${t('toggleCloseBtn')}` : `🔓 ${t('toggleOpenBtn')}`}
            </button>
          )}
          {!isOwner && (
            <button type="button" className="btn-ghost btn-sm" disabled={leavePending} onClick={onLeave}>
              {t('leaveBtn')}
            </button>
          )}
          {isOwner && members.length > 1 && (
            <button
              type="button"
              className="btn-ghost btn-sm"
              onClick={() => setTransferOpen(true)}
            >
              👑 {t('transferLeadership.openBtn')}
            </button>
          )}
          {isOwner && (
            <button type="button" className="btn-ghost btn-sm" style={{ color: 'var(--ox-danger)' }} onClick={() => setConfirmDisband(true)}>
              {t('disbandBtn')}
            </button>
          )}
        </div>
      </div>

      {/* Descriptions: 3 текста (external/internal/apply) — план 67 D-041, U-015 */}
      <DescriptionsPanel allianceID={alliance.id} canEdit={isOwner} />

      {/* Members */}
      <MembersTable alliance={alliance} members={members} isOwner={isOwner} />

      {/* Custom ranks с гранулярными permissions — план 67 D-014, U-005 */}
      <RanksPanel allianceID={alliance.id} canManage={isOwner} />

      {/* Diplomacy: 5 enum-статусов, accept/reject/break — план 67 D-014 B1 */}
      <DiplomacyPanel allianceID={alliance.id} canManage={isOwner} />

      {/* Audit-log: журнал событий альянса — план 67 U-013 */}
      <AuditLogPanel allianceID={alliance.id} members={members} />

      {/* Applications */}
      {isOwner && !alliance.is_open && (
        <div className="ox-panel" style={{ padding: '12px 16px' }}>
          <div style={{ fontSize: 13, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 8 }}>
            {t('sectionApplications')}
          </div>
          {apps.isLoading && <div className="ox-skeleton" style={{ height: 40 }} />}
          {!apps.isLoading && (apps.data?.applications ?? []).length === 0 && (
            <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)' }}>{t('noApplications')}</div>
          )}
          {(apps.data?.applications ?? []).map((ap) => (
            <div key={ap.id} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '6px 0', borderBottom: '1px solid var(--ox-border)' }}>
              <span style={{ flex: 1, fontSize: 15 }}>
                <b>{ap.username}</b>
                {ap.message && <span style={{ color: 'var(--ox-fg-dim)', marginLeft: 8 }}>{ap.message}</span>}
              </span>
              <button type="button" className="btn btn-sm" disabled={approve.isPending} onClick={() => approve.mutate(ap.id)}>✓ {t('acceptBtn')}</button>
              <button type="button" className="btn-ghost btn-sm" disabled={reject.isPending} onClick={() => reject.mutate(ap.id)}>✕</button>
            </div>
          ))}
        </div>
      )}

      {confirmDisband && (
        <Confirm
          title={t('disbandBtn')}
          message={t('disbandConfirm', { name: alliance.name })}
          confirmLabel={t('confirmBtn')}
          danger
          onConfirm={() => { setConfirmDisband(false); onDisband(); }}
          onCancel={() => setConfirmDisband(false)}
        />
      )}

      {transferOpen && (
        <TransferLeadershipDialog
          allianceID={alliance.id}
          members={members}
          currentOwnerID={alliance.owner_id}
          onClose={() => setTransferOpen(false)}
        />
      )}
    </div>
  );
}

function AllianceDetail({
  alliance, members, canJoin, joining, onJoin,
}: {
  alliance: Alliance;
  members: Member[];
  canJoin: boolean;
  joining: boolean;
  onJoin: (message: string) => void;
}) {
  const { t } = useTranslation('alliance');
  const [message, setMessage] = useState('');
  return (
    <div className="ox-panel" style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div>
        <div style={{ fontSize: 15, fontWeight: 700, display: 'flex', alignItems: 'center', gap: 6 }}>
          <span style={{ color: 'var(--ox-accent)', fontFamily: 'var(--ox-mono)' }}>[{alliance.tag}]</span>{' '}
          {alliance.name}
          <ReportButton targetType="alliance" targetId={alliance.id} compact />
        </div>
        <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)', marginTop: 2 }}>
          {alliance.owner_name} · {alliance.member_count} · {alliance.is_open ? `🔓 ${t('labelOpen')}` : `🔒 ${t('labelClosed')}`}
        </div>
        {alliance.description && (
          <div style={{ fontSize: 15, color: 'var(--ox-fg-dim)', marginTop: 6 }}>{alliance.description}</div>
        )}
      </div>

      <table className="ox-table" style={{ margin: 0, fontSize: 14 }}>
        <thead>
          <tr><th>{t('colMember')}</th><th>{t('colRole')}</th></tr>
        </thead>
        <tbody>
          {members.map((m) => (
            <tr key={m.user_id}>
              <td>{m.username}</td>
              <td style={{ color: 'var(--ox-fg-dim)' }}>{m.rank_name || m.rank}</td>
            </tr>
          ))}
        </tbody>
      </table>

      {canJoin && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {!alliance.is_open && (
            <textarea
              value={message}
              onChange={(e) => setMessage(e.target.value)}
              rows={2}
              style={{ width: '100%', boxSizing: 'border-box', resize: 'vertical' }}
              placeholder={t('createNamePlaceholder')}
            />
          )}
          <button type="button" className="btn btn-sm" disabled={joining} onClick={() => onJoin(message)}>
            {alliance.is_open ? `🤝 ${t('sectionInfo')}` : `📨 ${t('applyBtn')}`}
          </button>
        </div>
      )}
    </div>
  );
}

function CreateForm({ onCreated, onCancel }: { onCreated: () => void; onCancel: () => void }) {
  const { t } = useTranslation('alliance');
  const toast = useToast();
  const [tag, setTag] = useState('');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');

  const create = useMutation({
    mutationFn: () => api.post<{ alliance: Alliance }>('/api/alliances', { tag, name, description }),
    onSuccess: () => { toast.show('success', t('createTitle'), `[${tag}] ${name}`); onCreated(); },
    onError: (e: Error) => toast.show('danger', t('createErr'), e.message),
  });

  return (
    <div className="ox-panel" style={{ padding: 20, maxWidth: 480, display: 'flex', flexDirection: 'column', gap: 14 }}>
      <div style={{ fontSize: 16, fontWeight: 700 }}>{t('createTitle')}</div>

      <div>
        <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>{t('labelTag')} (3–5)</label>
        <input
          value={tag}
          onChange={(e) => setTag(e.target.value.toUpperCase())}
          maxLength={5}
          style={{ width: 100 }}
          placeholder={t('createTagPlaceholder')}
        />
      </div>
      <div>
        <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>{t('sectionInfo')}</label>
        <input value={name} onChange={(e) => setName(e.target.value)} maxLength={64} style={{ width: '100%' }} placeholder={t('createNamePlaceholder')} />
      </div>
      <div>
        <label style={{ fontSize: 14, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>—</label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          style={{ width: '100%', boxSizing: 'border-box', resize: 'vertical' }}
        />
      </div>

      <div style={{ display: 'flex', gap: 8 }}>
        <button
          type="button"
          className="btn btn-sm"
          disabled={create.isPending || tag.length < 3 || name.length < 3}
          onClick={() => create.mutate()}
        >
          {create.isPending ? '…' : `🤝 ${t('createBtn')}`}
        </button>
        <button type="button" className="btn-ghost btn-sm" onClick={onCancel}>{t('cancelBtn')}</button>
      </div>
    </div>
  );
}

function MembersTable({
  alliance, members, isOwner,
}: {
  alliance: Alliance;
  members: Member[];
  isOwner: boolean;
}) {
  const { t } = useTranslation('alliance');
  const qc = useQueryClient();
  const toast = useToast();
  const [editingUID, setEditingUID] = useState<string | null>(null);
  const [rankDraft, setRankDraft] = useState('');

  const setRank = useMutation({
    mutationFn: ({ uid, name }: { uid: string; name: string }) =>
      api.patch<void>(`/api/alliances/${alliance.id}/members/${uid}/rank`, { rank_name: name }),
    onSuccess: () => { void qc.invalidateQueries({ queryKey: ['alliances'] }); setEditingUID(null); },
    onError: (e) => toast.show('danger', t('createErr'), e instanceof Error ? e.message : ''),
  });

  return (
    <div className="ox-panel" style={{ overflow: 'hidden' }}>
      <div style={{ padding: '10px 16px 8px', fontSize: 13, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', borderBottom: '1px solid var(--ox-border)' }}>
        {t('colMember')} ({members.length})
      </div>
      <table className="ox-table" style={{ margin: 0 }}>
        <thead>
          <tr>
            <th>{t('colMember')}</th>
            <th>{t('colRole')}</th>
            <th>{t('labelTag')}</th>
            {isOwner && <th />}
          </tr>
        </thead>
        <tbody>
          {members.map((m) => (
            <tr key={m.user_id}>
              <td style={{ fontWeight: m.rank === 'owner' ? 700 : 400 }}>{m.username}</td>
              <td style={{ color: 'var(--ox-fg-dim)', fontSize: 14 }}>
                {editingUID === m.user_id ? (
                  <span style={{ display: 'flex', gap: 4 }}>
                    <input
                      value={rankDraft}
                      onChange={(e) => setRankDraft(e.target.value)}
                      maxLength={32}
                      style={{ width: 120 }}
                      autoFocus
                    />
                    <button type="button" className="btn btn-sm" disabled={setRank.isPending} onClick={() => setRank.mutate({ uid: m.user_id, name: rankDraft })}>✓</button>
                    <button type="button" className="btn-ghost btn-sm" onClick={() => setEditingUID(null)}>✕</button>
                  </span>
                ) : (
                  m.rank_name || m.rank
                )}
              </td>
              <td style={{ fontSize: 14, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
                {new Date(m.joined_at).toLocaleDateString('ru-RU')}
              </td>
              {isOwner && (
                <td>
                  {m.rank !== 'owner' && (
                    <button
                      type="button"
                      className="btn-ghost btn-sm"
                      onClick={() => { setEditingUID(m.user_id); setRankDraft(m.rank_name); }}
                    >
                      ✎
                    </button>
                  )}
                </td>
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
