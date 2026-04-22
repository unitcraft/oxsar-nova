import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';
import { Confirm } from '@/ui/Confirm';

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
  status: string;   // "pending" | "active"
  initiator: boolean;
  set_at: string;
}

export function AllianceScreen() {
  const { tf } = useTranslation();
  const qc = useQueryClient();
  const [view, setView] = useState<'list' | 'mine' | 'create'>('mine');
  const [selectedID, setSelectedID] = useState<string | null>(null);
  const [createError, setCreateError] = useState('');

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
    queryFn: () =>
      api.get<{ alliance: Alliance; members: Member[] }>(`/api/alliances/${selectedID}`),
    enabled: !!selectedID,
  });

  const join = useMutation({
    mutationFn: ({ id, message }: { id: string; message: string }) =>
      api.post<{ status?: string }>(`/api/alliances/${id}/join`, { message }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      setView('mine');
    },
  });

  const leave = useMutation({
    mutationFn: () => api.post<void>('/api/alliances/leave'),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      setView('mine');
    },
  });

  const disband = useMutation({
    mutationFn: (id: string) => api.delete<void>(`/api/alliances/${id}`),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      setView('mine');
    },
  });

  const myAlliance = mine.data?.alliance ?? null;
  const myMembers = mine.data?.members ?? [];
  const myMember = myAlliance
    ? myMembers.find((m: Member) => m.user_id === myAlliance.owner_id && m.rank === 'owner')
    : undefined;
  const amOwner = !!myMember;

  return (
    <section>
      <h2>{tf('global', 'MENU_ALLIANCE', 'Альянс')}</h2>

      <div style={{ marginBottom: 12, display: 'flex', gap: 8 }}>
        <button type="button" onClick={() => setView('mine')}
          style={{ fontWeight: view === 'mine' ? 700 : 400 }}>
          {tf('Main', 'ALLY_MY', 'Мой альянс')}
        </button>
        <button type="button" onClick={() => setView('list')}
          style={{ fontWeight: view === 'list' ? 700 : 400 }}>
          {tf('Main', 'ALLY_LIST', 'Список')}
        </button>
        {!myAlliance && (
          <button type="button" onClick={() => setView('create')}
            style={{ fontWeight: view === 'create' ? 700 : 400 }}>
            {tf('Main', 'ALLY_CREATE', 'Создать')}
          </button>
        )}
      </div>

      {view === 'mine' && (
        <MyAlliancePanel
          alliance={myAlliance}
          members={myMembers}
          isOwner={amOwner}
          loading={mine.isLoading}
          onLeave={() => leave.mutate()}
          onDisband={() => myAlliance && disband.mutate(myAlliance.id)}
          leaveError={leave.error instanceof Error ? leave.error.message : ''}
          disbandError={disband.error instanceof Error ? disband.error.message : ''}
        />
      )}

      {view === 'list' && (
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16 }}>
          <div>
            {list.isLoading && <p>…</p>}
            {list.error instanceof Error && <p className="ox-error">{list.error.message}</p>}
            <table className="ox-table">
              <thead>
                <tr>
                  <th>[{tf('Main', 'ALLY_TAG', 'Тег')}]</th>
                  <th>{tf('Main', 'ALLY_NAME', 'Название')}</th>
                  <th>{tf('Main', 'ALLY_MEMBERS', 'Игроков')}</th>
                  <th>{tf('Main', 'ALLY_OPEN', 'Тип')}</th>
                </tr>
              </thead>
              <tbody>
                {(list.data?.alliances ?? []).map((al) => (
                  <tr key={al.id}
                    style={{ cursor: 'pointer', background: selectedID === al.id ? 'rgba(255,255,255,0.05)' : undefined }}
                    onClick={() => setSelectedID(al.id)}>
                    <td>[{al.tag}]</td>
                    <td>{al.name}</td>
                    <td className="num">{al.member_count}</td>
                    <td>{al.is_open ? '🔓' : '🔒'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div>
            {selectedID && detail.data && (
              <AllianceDetail
                alliance={detail.data.alliance}
                members={detail.data.members ?? []}
                canJoin={!myAlliance}
                joining={join.isPending}
                joinError={join.error instanceof Error ? join.error.message : ''}
                onJoin={(msg) => join.mutate({ id: selectedID, message: msg })}
              />
            )}
          </div>
        </div>
      )}

      {view === 'create' && !myAlliance && (
        <CreateForm
          error={createError}
          onCreated={() => {
            void qc.invalidateQueries({ queryKey: ['alliances'] });
            setView('mine');
          }}
          onError={setCreateError}
          onCancel={() => setView('mine')}
        />
      )}
    </section>
  );
}

function MyAlliancePanel({
  alliance, members, isOwner, loading,
  onLeave, onDisband, leaveError, disbandError,
}: {
  alliance: Alliance | null;
  members: Member[];
  isOwner: boolean;
  loading: boolean;
  onLeave: () => void;
  onDisband: () => void;
  leaveError: string;
  disbandError: string;
}) {
  const { tf } = useTranslation();
  const qc = useQueryClient();
  const [confirmDisband, setConfirmDisband] = useState(false);

  const setOpen = useMutation({
    mutationFn: ({ id, isOpen }: { id: string; isOpen: boolean }) =>
      api.patch<void>(`/api/alliances/${id}/open`, { is_open: isOpen }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
  });

  const approve = useMutation({
    mutationFn: (appID: string) =>
      api.post<void>(`/api/alliances/applications/${appID}/approve`),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
  });

  const reject = useMutation({
    mutationFn: (appID: string) =>
      api.delete<void>(`/api/alliances/applications/${appID}`),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
  });

  const apps = useQuery({
    queryKey: ['alliances', alliance?.id, 'applications'],
    queryFn: () =>
      api.get<{ applications: Application[] | null }>(`/api/alliances/${alliance!.id}/applications`),
    enabled: isOwner && !!alliance && !alliance.is_open,
    refetchInterval: 15000,
  });

  if (loading) return <p>…</p>;
  if (!alliance) return <p>{tf('Main', 'ALLY_NONE', 'Вы не состоите в альянсе.')}</p>;

  return (
    <div>
      <h3>[{alliance.tag}] {alliance.name}</h3>
      {alliance.description && <p>{alliance.description}</p>}
      <p>
        <b>{tf('Main', 'ALLY_OWNER', 'Основатель')}:</b> {alliance.owner_name}{' · '}
        <b>{tf('Main', 'ALLY_MEMBERS', 'Игроков')}:</b> {alliance.member_count}{' · '}
        <b>Тип:</b> {alliance.is_open ? 'Открытый' : 'Закрытый (заявки)'}
      </p>

      {isOwner && (
        <div style={{ marginBottom: 8 }}>
          <button type="button"
            disabled={setOpen.isPending}
            onClick={() => setOpen.mutate({ id: alliance.id, isOpen: !alliance.is_open })}>
            {alliance.is_open ? 'Закрыть (включить заявки)' : 'Открыть (прямой вход)'}
          </button>
        </div>
      )}

      <MembersTable alliance={alliance} members={members} isOwner={isOwner} />

      {isOwner && (
        <RelationsPanel allianceID={alliance.id} />
      )}

      {isOwner && !alliance.is_open && (
        <div style={{ marginTop: 16 }}>
          <h4>Заявки на вступление</h4>
          {apps.isLoading && <p>…</p>}
          {(apps.data?.applications ?? []).length === 0 && !apps.isLoading && (
            <p>Нет заявок.</p>
          )}
          {(apps.data?.applications ?? []).map((ap) => (
            <div key={ap.id} style={{ marginBottom: 8, padding: '6px 10px', border: '1px solid rgba(255,255,255,0.1)', borderRadius: 4 }}>
              <b>{ap.username}</b>
              {ap.message && <span> — {ap.message}</span>}
              <span style={{ float: 'right', display: 'flex', gap: 6 }}>
                <button type="button" onClick={() => approve.mutate(ap.id)}>Принять</button>
                <button type="button" onClick={() => reject.mutate(ap.id)}>Отклонить</button>
              </span>
            </div>
          ))}
        </div>
      )}

      <div style={{ marginTop: 12, display: 'flex', gap: 8 }}>
        {!isOwner && (
          <button type="button" onClick={onLeave}>
            {tf('Main', 'ALLY_LEAVE', 'Покинуть альянс')}
          </button>
        )}
        {isOwner && (
          <button type="button" onClick={() => setConfirmDisband(true)}>
            {tf('Main', 'ALLY_DISBAND', 'Распустить')}
          </button>
        )}
        {confirmDisband && (
          <Confirm
            title={tf('Main', 'ALLY_DISBAND', 'Распустить альянс')}
            message={tf('Main', 'ALLY_DISBAND_CONFIRM', 'Распустить альянс? Это действие необратимо.')}
            confirmLabel={tf('Main', 'ALLY_DISBAND', 'Распустить')}
            danger
            onConfirm={() => { setConfirmDisband(false); onDisband(); }}
            onCancel={() => setConfirmDisband(false)}
          />
        )}
      </div>
      {leaveError && <p className="ox-error">{leaveError}</p>}
      {disbandError && <p className="ox-error">{disbandError}</p>}
    </div>
  );
}

function AllianceDetail({
  alliance, members, canJoin, joining, joinError, onJoin,
}: {
  alliance: Alliance;
  members: Member[];
  canJoin: boolean;
  joining: boolean;
  joinError: string;
  onJoin: (message: string) => void;
}) {
  const { tf } = useTranslation();
  const [message, setMessage] = useState('');
  return (
    <div>
      <h3>[{alliance.tag}] {alliance.name}</h3>
      {alliance.description && <p>{alliance.description}</p>}
      <p>
        <b>{tf('Main', 'ALLY_OWNER', 'Основатель')}:</b> {alliance.owner_name}{' · '}
        <b>{tf('Main', 'ALLY_MEMBERS', 'Игроков')}:</b> {alliance.member_count}{' · '}
        <b>Тип:</b> {alliance.is_open ? 'Открытый' : 'Закрытый'}
      </p>
      <table className="ox-table">
        <thead>
          <tr>
            <th>{tf('Main', 'USERNAME', 'Игрок')}</th>
            <th>{tf('Main', 'ALLY_RANK', 'Ранг')}</th>
          </tr>
        </thead>
        <tbody>
          {members.map((m) => (
            <tr key={m.user_id}>
              <td>{m.username}</td>
              <td>{m.rank}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {canJoin && (
        <div style={{ marginTop: 8 }}>
          {!alliance.is_open && (
            <div style={{ marginBottom: 6 }}>
              <textarea
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                rows={2}
                style={{ width: '100%', boxSizing: 'border-box' }}
                placeholder="Сопроводительное сообщение (необязательно)"
              />
            </div>
          )}
          <button type="button" disabled={joining} onClick={() => onJoin(message)}>
            {alliance.is_open
              ? tf('Main', 'ALLY_JOIN', 'Вступить')
              : 'Подать заявку'}
          </button>
          {joinError && <p className="ox-error">{joinError}</p>}
        </div>
      )}
    </div>
  );
}

function CreateForm({ error, onCreated, onError, onCancel }: {
  error: string;
  onCreated: () => void;
  onError: (e: string) => void;
  onCancel: () => void;
}) {
  const { tf } = useTranslation();
  const [tag, setTag] = useState('');
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');

  const create = useMutation({
    mutationFn: () => api.post<{ alliance: Alliance }>('/api/alliances', { tag, name, description }),
    onSuccess: onCreated,
    onError: (e: Error) => onError(e.message),
  });

  return (
    <div style={{ maxWidth: 480, padding: 12, border: '1px solid rgba(255,255,255,0.15)', borderRadius: 4 }}>
      <h3 style={{ marginTop: 0 }}>{tf('Main', 'ALLY_CREATE', 'Создать альянс')}</h3>
      <div style={{ marginBottom: 8 }}>
        <label>
          {tf('Main', 'ALLY_TAG', 'Тег')} (3–5 символов):{' '}
          <input value={tag} onChange={(e) => setTag(e.target.value.toUpperCase())}
            maxLength={5} style={{ width: 80 }}
            placeholder="TAG" />
        </label>
      </div>
      <div style={{ marginBottom: 8 }}>
        <label>
          {tf('Main', 'ALLY_NAME', 'Название')}:{' '}
          <input value={name} onChange={(e) => setName(e.target.value)}
            maxLength={64} style={{ width: 260 }} />
        </label>
      </div>
      <div style={{ marginBottom: 8 }}>
        <textarea value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3} style={{ width: '100%', boxSizing: 'border-box' }}
          placeholder={tf('Main', 'ALLY_DESC_PLACEHOLDER', 'Описание альянса…')} />
      </div>
      {error && <p className="ox-error">{error}</p>}
      <button type="button"
        disabled={create.isPending || tag.length < 3 || name.length < 3}
        onClick={() => create.mutate()}>
        {tf('Main', 'ALLY_CREATE_BTN', 'Создать')}
      </button>{' '}
      <button type="button" onClick={onCancel}>
        {tf('Main', 'CANCEL', 'Отмена')}
      </button>
    </div>
  );
}

function RelationsPanel({ allianceID }: { allianceID: string }) {
  const qc = useQueryClient();
  const [targetID, setTargetID] = useState('');
  const [relation, setRelation] = useState<'nap' | 'war' | 'ally'>('nap');

  const rels = useQuery({
    queryKey: ['alliances', allianceID, 'relations'],
    queryFn: () =>
      api.get<{ relations: Relationship[] | null }>(`/api/alliances/${allianceID}/relations`),
    refetchInterval: 30000,
  });

  const propose = useMutation({
    mutationFn: ({ tid, rel }: { tid: string; rel: string }) =>
      api.put<void>(`/api/alliances/${allianceID}/relations/${tid}`, { relation: rel }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] });
      setTargetID('');
    },
  });

  const remove = useMutation({
    mutationFn: (tid: string) =>
      api.put<void>(`/api/alliances/${allianceID}/relations/${tid}`, { relation: 'none' }),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] }),
  });

  const accept = useMutation({
    mutationFn: (initiatorID: string) =>
      api.post<void>(`/api/alliances/${allianceID}/relations/${initiatorID}/accept`),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] }),
  });

  const reject = useMutation({
    mutationFn: (initiatorID: string) =>
      api.delete<void>(`/api/alliances/${allianceID}/relations/${initiatorID}`),
    onSuccess: () =>
      void qc.invalidateQueries({ queryKey: ['alliances', allianceID, 'relations'] }),
  });

  const list = rels.data?.relations ?? [];
  const relLabel: Record<string, string> = { nap: 'НЕН', war: 'ВОЙНА', ally: 'СОЮЗ' };
  const statusLabel: Record<string, string> = { active: '', pending: ' (ожидает)' };

  return (
    <div style={{ marginTop: 16 }}>
      <h4>Отношения с альянсами</h4>
      {list.length === 0 ? (
        <p style={{ color: '#888' }}>Нет установленных отношений.</p>
      ) : (
        <table className="ox-table">
          <thead>
            <tr><th>Альянс</th><th>Отношение</th><th>Статус</th><th /></tr>
          </thead>
          <tbody>
            {list.map((r) => (
              <tr key={`${r.initiator ? 'out' : 'in'}-${r.target_alliance_id}`}>
                <td>[{r.target_tag}] {r.target_name}</td>
                <td>{relLabel[r.relation] ?? r.relation}</td>
                <td style={{ color: r.status === 'pending' ? '#f90' : 'inherit' }}>
                  {r.initiator ? 'Предложено' : 'Входящее'}{statusLabel[r.status] ?? ''}
                </td>
                <td style={{ display: 'flex', gap: 4 }}>
                  {!r.initiator && r.status === 'pending' ? (
                    <>
                      <button type="button" disabled={accept.isPending}
                        onClick={() => accept.mutate(r.target_alliance_id)}>
                        ✓
                      </button>
                      <button type="button" disabled={reject.isPending}
                        onClick={() => reject.mutate(r.target_alliance_id)}>
                        ✕
                      </button>
                    </>
                  ) : (
                    <button type="button" disabled={remove.isPending}
                      onClick={() => remove.mutate(r.target_alliance_id)}>
                      ✕
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      <div style={{ marginTop: 8, display: 'flex', gap: 6, alignItems: 'center', flexWrap: 'wrap' }}>
        <input
          placeholder="ID альянса"
          value={targetID}
          onChange={(e) => setTargetID(e.target.value)}
          style={{ width: 280, fontFamily: 'monospace', fontSize: '0.85em' }}
        />
        <select value={relation} onChange={(e) => setRelation(e.target.value as typeof relation)}>
          <option value="nap">НЕН (ненападение)</option>
          <option value="ally">СОЮЗ</option>
          <option value="war">ВОЙНА</option>
        </select>
        <button type="button" disabled={!targetID || propose.isPending}
          onClick={() => propose.mutate({ tid: targetID, rel: relation })}>
          Предложить
        </button>
      </div>
      {propose.isError && (
        <p className="ox-error">{propose.error instanceof Error ? propose.error.message : 'ошибка'}</p>
      )}
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
  const { tf } = useTranslation();
  const qc = useQueryClient();
  const [editingUID, setEditingUID] = useState<string | null>(null);
  const [rankDraft, setRankDraft] = useState('');

  const setRank = useMutation({
    mutationFn: ({ uid, name }: { uid: string; name: string }) =>
      api.patch<void>(`/api/alliances/${alliance.id}/members/${uid}/rank`, { rank_name: name }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      setEditingUID(null);
    },
  });

  return (
    <table className="ox-table">
      <thead>
        <tr>
          <th>{tf('Main', 'USERNAME', 'Игрок')}</th>
          <th>{tf('Main', 'ALLY_RANK', 'Ранг')}</th>
          <th>{tf('Main', 'ALLY_JOINED', 'Вступил')}</th>
          {isOwner && <th />}
        </tr>
      </thead>
      <tbody>
        {members.map((m) => (
          <tr key={m.user_id}>
            <td>{m.username}</td>
            <td>
              {editingUID === m.user_id ? (
                <span style={{ display: 'flex', gap: 4 }}>
                  <input
                    value={rankDraft}
                    onChange={(e) => setRankDraft(e.target.value)}
                    maxLength={32}
                    style={{ width: 140 }}
                    autoFocus
                  />
                  <button type="button" disabled={setRank.isPending}
                    onClick={() => setRank.mutate({ uid: m.user_id, name: rankDraft })}>
                    ✓
                  </button>
                  <button type="button" onClick={() => setEditingUID(null)}>✕</button>
                </span>
              ) : (
                m.rank_name || m.rank
              )}
            </td>
            <td>{new Date(m.joined_at).toLocaleDateString('ru-RU')}</td>
            {isOwner && (
              <td>
                {m.rank !== 'owner' && (
                  <button type="button" style={{ fontSize: '0.8em' }}
                    onClick={() => { setEditingUID(m.user_id); setRankDraft(m.rank_name); }}>
                    ✎
                  </button>
                )}
              </td>
            )}
          </tr>
        ))}
      </tbody>
    </table>
  );
}
