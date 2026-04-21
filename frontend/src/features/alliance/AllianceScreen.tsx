import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

interface Alliance {
  id: string;
  tag: string;
  name: string;
  description: string;
  owner_id: string;
  owner_name: string;
  member_count: number;
  created_at: string;
}

interface Member {
  user_id: string;
  username: string;
  rank: string;
  joined_at: string;
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
    mutationFn: (id: string) => api.post<void>(`/api/alliances/${id}/join`),
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
  const amOwner = !!myAlliance && !!mine.data?.members?.find(
    (m) => m.rank === 'owner' && m.user_id === myAlliance.owner_id,
  );

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
                onJoin={() => join.mutate(selectedID)}
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

  if (loading) return <p>…</p>;
  if (!alliance) return <p>{tf('Main', 'ALLY_NONE', 'Вы не состоите в альянсе.')}</p>;

  return (
    <div>
      <h3>[{alliance.tag}] {alliance.name}</h3>
      {alliance.description && <p>{alliance.description}</p>}
      <p>
        <b>{tf('Main', 'ALLY_OWNER', 'Основатель')}:</b> {alliance.owner_name}{' · '}
        <b>{tf('Main', 'ALLY_MEMBERS', 'Игроков')}:</b> {alliance.member_count}
      </p>
      <table className="ox-table">
        <thead>
          <tr>
            <th>{tf('Main', 'USERNAME', 'Игрок')}</th>
            <th>{tf('Main', 'ALLY_RANK', 'Ранг')}</th>
            <th>{tf('Main', 'ALLY_JOINED', 'Вступил')}</th>
          </tr>
        </thead>
        <tbody>
          {members.map((m) => (
            <tr key={m.user_id}>
              <td>{m.username}</td>
              <td>{m.rank}</td>
              <td>{new Date(m.joined_at).toLocaleDateString('ru-RU')}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <div style={{ marginTop: 12, display: 'flex', gap: 8 }}>
        {!isOwner && (
          <button type="button" onClick={onLeave}>
            {tf('Main', 'ALLY_LEAVE', 'Покинуть альянс')}
          </button>
        )}
        {isOwner && (
          <button type="button"
            onClick={() => { if (window.confirm(tf('Main', 'ALLY_DISBAND_CONFIRM', 'Распустить альянс? Это действие необратимо.'))) onDisband(); }}>
            {tf('Main', 'ALLY_DISBAND', 'Распустить')}
          </button>
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
  onJoin: () => void;
}) {
  const { tf } = useTranslation();
  return (
    <div>
      <h3>[{alliance.tag}] {alliance.name}</h3>
      {alliance.description && <p>{alliance.description}</p>}
      <p>
        <b>{tf('Main', 'ALLY_OWNER', 'Основатель')}:</b> {alliance.owner_name}{' · '}
        <b>{tf('Main', 'ALLY_MEMBERS', 'Игроков')}:</b> {alliance.member_count}
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
          <button type="button" disabled={joining} onClick={onJoin}>
            {tf('Main', 'ALLY_JOIN', 'Вступить')}
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
