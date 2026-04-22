import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { Confirm } from '@/ui/Confirm';
import { useToast } from '@/ui/Toast';

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

interface Relationship {
  target_alliance_id: string;
  target_tag: string;
  target_name: string;
  relation: string;
  status: string;
  initiator: boolean;
  set_at: string;
}

const REL_LABEL: Record<string, string> = { nap: 'НЕН', war: 'ВОЙНА', ally: 'СОЮЗ' };
const REL_COLOR: Record<string, string> = { nap: 'var(--ox-fg-dim)', war: 'var(--ox-danger)', ally: 'var(--ox-success)' };

export function AllianceScreen() {
  const qc = useQueryClient();
  const toast = useToast();
  const [view, setView] = useState<'mine' | 'list' | 'create'>('mine');
  const [selectedID, setSelectedID] = useState<string | null>(null);

  const mine = useQuery({
    queryKey: ['alliances', 'me'],
    queryFn: () => api.get<{ alliance: Alliance | null; members: Member[] | null }>('/api/alliances/me'),
    refetchInterval: 30000,
  });

  const list = useQuery({
    queryKey: ['alliances'],
    queryFn: () => api.get<{ alliances: Alliance[] | null }>('/api/alliances'),
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
      toast.show('success', 'Альянс', 'Заявка отправлена');
    },
    onError: (e) => toast.show('danger', 'Ошибка', e instanceof Error ? e.message : ''),
  });

  const leave = useMutation({
    mutationFn: () => api.post<void>('/api/alliances/leave'),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      setView('mine');
      toast.show('info', 'Альянс покинут');
    },
    onError: (e) => toast.show('danger', 'Ошибка', e instanceof Error ? e.message : ''),
  });

  const disband = useMutation({
    mutationFn: (id: string) => api.delete<void>(`/api/alliances/${id}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      setView('mine');
      toast.show('info', 'Альянс распущен');
    },
    onError: (e) => toast.show('danger', 'Ошибка', e instanceof Error ? e.message : ''),
  });

  const myAlliance = mine.data?.alliance ?? null;
  const myMembers = mine.data?.members ?? [];
  const amOwner = !!myAlliance && myMembers.some((m) => m.user_id === myAlliance.owner_id && m.rank === 'owner');

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 20 }}>
      <h2 style={{ margin: 0, fontSize: 18, fontFamily: 'var(--ox-font)', fontWeight: 700 }}>
        🤝 Альянс
      </h2>

      <div className="ox-tabs">
        <button type="button" aria-pressed={view === 'mine'} onClick={() => setView('mine')}>
          Мой альянс
        </button>
        <button type="button" aria-pressed={view === 'list'} onClick={() => setView('list')}>
          Список
        </button>
        {!myAlliance && (
          <button type="button" aria-pressed={view === 'create'} onClick={() => setView('create')}>
            Создать
          </button>
        )}
      </div>

      {view === 'mine' && (
        mine.isLoading
          ? <div className="ox-skeleton" style={{ height: 120 }} />
          : !myAlliance
            ? (
              <div className="ox-panel" style={{ padding: 24, textAlign: 'center', color: 'var(--ox-fg-dim)' }}>
                Вы не состоите в альянсе. Вступите в существующий или создайте свой.
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
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, alignItems: 'start' }}>
          <div className="ox-panel" style={{ overflow: 'hidden' }}>
            {list.isLoading && <div style={{ padding: 16 }}><div className="ox-skeleton" style={{ height: 60 }} /></div>}
            {!list.isLoading && (
              <table className="ox-table" style={{ margin: 0 }}>
                <thead>
                  <tr>
                    <th>[Тег]</th>
                    <th>Название</th>
                    <th>Игроков</th>
                    <th>Тип</th>
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
  const qc = useQueryClient();
  const toast = useToast();
  const [confirmDisband, setConfirmDisband] = useState(false);

  const setOpen = useMutation({
    mutationFn: ({ id, isOpen }: { id: string; isOpen: boolean }) =>
      api.patch<void>(`/api/alliances/${id}/open`, { is_open: isOpen }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
    onError: (e) => toast.show('danger', 'Ошибка', e instanceof Error ? e.message : ''),
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
            <div style={{ fontSize: 12, color: 'var(--ox-fg-muted)' }}>
              Основатель: {alliance.owner_name} · {alliance.member_count} игроков · {alliance.is_open ? '🔓 Открытый' : '🔒 Закрытый'}
            </div>
          </div>
        </div>
        {alliance.description && (
          <div style={{ fontSize: 13, color: 'var(--ox-fg-dim)', borderTop: '1px solid var(--ox-border)', paddingTop: 8, marginTop: 4 }}>
            {alliance.description}
          </div>
        )}
        <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
          {isOwner && (
            <button
              type="button"
              className="btn-ghost btn-sm"
              disabled={setOpen.isPending}
              onClick={() => setOpen.mutate({ id: alliance.id, isOpen: !alliance.is_open })}
            >
              {alliance.is_open ? '🔒 Закрыть (заявки)' : '🔓 Открыть (вход)'}
            </button>
          )}
          {!isOwner && (
            <button type="button" className="btn-ghost btn-sm" disabled={leavePending} onClick={onLeave}>
              Покинуть альянс
            </button>
          )}
          {isOwner && (
            <button type="button" className="btn-ghost btn-sm" style={{ color: 'var(--ox-danger)' }} onClick={() => setConfirmDisband(true)}>
              Распустить
            </button>
          )}
        </div>
      </div>

      {/* Members */}
      <MembersTable alliance={alliance} members={members} isOwner={isOwner} />

      {/* Alliance relations */}
      {isOwner && <RelationsPanel allianceID={alliance.id} />}

      {/* Applications */}
      {isOwner && !alliance.is_open && (
        <div className="ox-panel" style={{ padding: '12px 16px' }}>
          <div style={{ fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 8 }}>
            Заявки на вступление
          </div>
          {apps.isLoading && <div className="ox-skeleton" style={{ height: 40 }} />}
          {!apps.isLoading && (apps.data?.applications ?? []).length === 0 && (
            <div style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>Нет заявок.</div>
          )}
          {(apps.data?.applications ?? []).map((ap) => (
            <div key={ap.id} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '6px 0', borderBottom: '1px solid var(--ox-border)' }}>
              <span style={{ flex: 1, fontSize: 13 }}>
                <b>{ap.username}</b>
                {ap.message && <span style={{ color: 'var(--ox-fg-dim)', marginLeft: 8 }}>{ap.message}</span>}
              </span>
              <button type="button" className="btn btn-sm" disabled={approve.isPending} onClick={() => approve.mutate(ap.id)}>✓ Принять</button>
              <button type="button" className="btn-ghost btn-sm" disabled={reject.isPending} onClick={() => reject.mutate(ap.id)}>✕</button>
            </div>
          ))}
        </div>
      )}

      {confirmDisband && (
        <Confirm
          title="Распустить альянс"
          message="Распустить альянс? Это действие необратимо."
          confirmLabel="Распустить"
          danger
          onConfirm={() => { setConfirmDisband(false); onDisband(); }}
          onCancel={() => setConfirmDisband(false)}
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
  const [message, setMessage] = useState('');
  return (
    <div className="ox-panel" style={{ padding: '16px 20px', display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div>
        <div style={{ fontSize: 15, fontWeight: 700 }}>
          <span style={{ color: 'var(--ox-accent)', fontFamily: 'var(--ox-mono)' }}>[{alliance.tag}]</span>{' '}
          {alliance.name}
        </div>
        <div style={{ fontSize: 12, color: 'var(--ox-fg-muted)', marginTop: 2 }}>
          Основатель: {alliance.owner_name} · {alliance.member_count} игроков · {alliance.is_open ? '🔓 Открытый' : '🔒 Закрытый'}
        </div>
        {alliance.description && (
          <div style={{ fontSize: 13, color: 'var(--ox-fg-dim)', marginTop: 6 }}>{alliance.description}</div>
        )}
      </div>

      <table className="ox-table" style={{ margin: 0, fontSize: 12 }}>
        <thead>
          <tr><th>Игрок</th><th>Ранг</th></tr>
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
              placeholder="Сопроводительное сообщение (необязательно)"
            />
          )}
          <button type="button" className="btn btn-sm" disabled={joining} onClick={() => onJoin(message)}>
            {alliance.is_open ? '🤝 Вступить' : '📨 Подать заявку'}
          </button>
        </div>
      )}
    </div>
  );
}

function CreateForm({ onCreated, onCancel }: { onCreated: () => void; onCancel: () => void }) {
  const toast = useToast();
  const [tag, setTag] = useState('');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');

  const create = useMutation({
    mutationFn: () => api.post<{ alliance: Alliance }>('/api/alliances', { tag, name, description }),
    onSuccess: () => { toast.show('success', 'Альянс создан', `[${tag}] ${name}`); onCreated(); },
    onError: (e: Error) => toast.show('danger', 'Ошибка', e.message),
  });

  return (
    <div className="ox-panel" style={{ padding: 20, maxWidth: 480, display: 'flex', flexDirection: 'column', gap: 14 }}>
      <div style={{ fontSize: 14, fontWeight: 700 }}>Создать альянс</div>

      <div>
        <label style={{ fontSize: 12, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Тег (3–5 символов)</label>
        <input
          value={tag}
          onChange={(e) => setTag(e.target.value.toUpperCase())}
          maxLength={5}
          style={{ width: 100 }}
          placeholder="TAG"
        />
      </div>
      <div>
        <label style={{ fontSize: 12, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Название</label>
        <input value={name} onChange={(e) => setName(e.target.value)} maxLength={64} style={{ width: '100%' }} />
      </div>
      <div>
        <label style={{ fontSize: 12, color: 'var(--ox-fg-dim)', display: 'block', marginBottom: 4 }}>Описание</label>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          style={{ width: '100%', boxSizing: 'border-box', resize: 'vertical' }}
          placeholder="Описание альянса…"
        />
      </div>

      <div style={{ display: 'flex', gap: 8 }}>
        <button
          type="button"
          className="btn btn-sm"
          disabled={create.isPending || tag.length < 3 || name.length < 3}
          onClick={() => create.mutate()}
        >
          {create.isPending ? '…' : '🤝 Создать'}
        </button>
        <button type="button" className="btn-ghost btn-sm" onClick={onCancel}>Отмена</button>
      </div>
    </div>
  );
}

function RelationsPanel({ allianceID }: { allianceID: string }) {
  const qc = useQueryClient();
  const toast = useToast();
  const [targetID, setTargetID] = useState('');
  const [relation, setRelation] = useState<'nap' | 'war' | 'ally'>('nap');

  const rels = useQuery({
    queryKey: ['alliances', allianceID, 'relations'],
    queryFn: () => api.get<{ relations: Relationship[] | null }>(`/api/alliances/${allianceID}/relations`),
    refetchInterval: 30000,
  });

  const propose = useMutation({
    mutationFn: ({ tid, rel }: { tid: string; rel: string }) =>
      api.put<void>(`/api/alliances/${allianceID}/relations/${tid}`, { relation: rel }),
    onSuccess: () => { void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] }); setTargetID(''); },
    onError: (e) => toast.show('danger', 'Ошибка', e instanceof Error ? e.message : ''),
  });

  const remove = useMutation({
    mutationFn: (tid: string) => api.put<void>(`/api/alliances/${allianceID}/relations/${tid}`, { relation: 'none' }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] }),
  });

  const accept = useMutation({
    mutationFn: (initiatorID: string) => api.post<void>(`/api/alliances/${allianceID}/relations/${initiatorID}/accept`),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] }),
  });

  const rejectRel = useMutation({
    mutationFn: (initiatorID: string) => api.delete<void>(`/api/alliances/${allianceID}/relations/${initiatorID}`),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] }),
  });

  const list = rels.data?.relations ?? [];

  return (
    <div className="ox-panel" style={{ padding: '12px 16px' }}>
      <div style={{ fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', marginBottom: 8 }}>
        Отношения с альянсами
      </div>

      {list.length === 0 && !rels.isLoading && (
        <div style={{ fontSize: 13, color: 'var(--ox-fg-dim)', marginBottom: 10 }}>Нет установленных отношений.</div>
      )}

      {list.length > 0 && (
        <table className="ox-table" style={{ margin: '0 0 12px', fontSize: 12 }}>
          <thead>
            <tr><th>Альянс</th><th>Отношение</th><th>Статус</th><th /></tr>
          </thead>
          <tbody>
            {list.map((r) => (
              <tr key={`${r.initiator ? 'out' : 'in'}-${r.target_alliance_id}`}>
                <td style={{ fontFamily: 'var(--ox-mono)' }}>[{r.target_tag}] {r.target_name}</td>
                <td style={{ color: REL_COLOR[r.relation] ?? 'var(--ox-fg-dim)', fontWeight: 700 }}>
                  {REL_LABEL[r.relation] ?? r.relation}
                </td>
                <td style={{ color: r.status === 'pending' ? 'var(--ox-warning)' : 'var(--ox-fg-dim)' }}>
                  {r.initiator ? 'Предложено' : 'Входящее'}{r.status === 'pending' ? ' (ожидает)' : ''}
                </td>
                <td>
                  <div style={{ display: 'flex', gap: 4 }}>
                    {!r.initiator && r.status === 'pending' ? (
                      <>
                        <button type="button" className="btn btn-sm" disabled={accept.isPending} onClick={() => accept.mutate(r.target_alliance_id)}>✓</button>
                        <button type="button" className="btn-ghost btn-sm" disabled={rejectRel.isPending} onClick={() => rejectRel.mutate(r.target_alliance_id)}>✕</button>
                      </>
                    ) : (
                      <button type="button" className="btn-ghost btn-sm" disabled={remove.isPending} onClick={() => remove.mutate(r.target_alliance_id)}>✕</button>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <div style={{ display: 'flex', gap: 6, alignItems: 'center', flexWrap: 'wrap' }}>
        <input
          placeholder="ID альянса"
          value={targetID}
          onChange={(e) => setTargetID(e.target.value)}
          style={{ flex: 1, minWidth: 200, fontFamily: 'var(--ox-mono)', fontSize: '0.85em' }}
        />
        <select value={relation} onChange={(e) => setRelation(e.target.value as typeof relation)}>
          <option value="nap">НЕН (ненападение)</option>
          <option value="ally">СОЮЗ</option>
          <option value="war">ВОЙНА</option>
        </select>
        <button type="button" className="btn btn-sm" disabled={!targetID || propose.isPending} onClick={() => propose.mutate({ tid: targetID, rel: relation })}>
          Предложить
        </button>
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
  const qc = useQueryClient();
  const toast = useToast();
  const [editingUID, setEditingUID] = useState<string | null>(null);
  const [rankDraft, setRankDraft] = useState('');

  const setRank = useMutation({
    mutationFn: ({ uid, name }: { uid: string; name: string }) =>
      api.patch<void>(`/api/alliances/${alliance.id}/members/${uid}/rank`, { rank_name: name }),
    onSuccess: () => { void qc.invalidateQueries({ queryKey: ['alliances'] }); setEditingUID(null); },
    onError: (e) => toast.show('danger', 'Ошибка', e instanceof Error ? e.message : ''),
  });

  return (
    <div className="ox-panel" style={{ overflow: 'hidden' }}>
      <div style={{ padding: '10px 16px 8px', fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', color: 'var(--ox-fg-muted)', borderBottom: '1px solid var(--ox-border)' }}>
        Состав ({members.length})
      </div>
      <table className="ox-table" style={{ margin: 0 }}>
        <thead>
          <tr>
            <th>Игрок</th>
            <th>Ранг</th>
            <th>Вступил</th>
            {isOwner && <th />}
          </tr>
        </thead>
        <tbody>
          {members.map((m) => (
            <tr key={m.user_id}>
              <td style={{ fontWeight: m.rank === 'owner' ? 700 : 400 }}>{m.username}</td>
              <td style={{ color: 'var(--ox-fg-dim)', fontSize: 12 }}>
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
              <td style={{ fontSize: 12, color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>
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
