// S-034 Friends — список друзей и запросов (план 72 Ф.5 Spring 4,
// расширен 72.1.14: двусторонний accept-flow с подтверждением).
//
// Pixel-perfect зеркало legacy `templates/standard/buddylist.tpl` с
// учётом legacy `Friends.class.php`/`buddylist.accepted=0/1`:
//   - Mutual-друзья: основная таблица (как раньше).
//   - Incoming pending: входящие запросы — кнопки Accept/Decline.
//   - Outgoing pending: отправленные мной — кнопка Cancel.
//
// Endpoints: GET /api/friends?pending=...,
// POST /api/friends/{userId}, POST /api/friends/{userId}/accept,
// DELETE /api/friends/{userId}.

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import {
  acceptFriend,
  addFriend,
  fetchFriends,
  removeFriend,
} from '@/api/friends';
import { search } from '@/api/search';
import { QK } from '@/api/query-keys';
import type { ApiError } from '@/api/client';
import type { Friend } from '@/api/types';
import { useTranslation } from '@/i18n/i18n';

function formatLastSeen(
  lastSeen: string | undefined,
  t: (g: string, k: string, v?: Record<string, string | number>) => string,
): string {
  if (!lastSeen) return '—';
  const minutes = Math.floor((Date.now() - new Date(lastSeen).getTime()) / 60_000);
  if (minutes < 5) return t('friends', 'statusOnline');
  if (minutes < 60) return t('friends', 'statusMinAgo', { n: minutes });
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return t('friends', 'statusHourAgo', { n: hours });
  const days = Math.floor(hours / 24);
  return t('friends', 'statusDayAgo', { n: days });
}

export function FriendsScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const [addQuery, setAddQuery] = useState('');
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const mutualQ = useQuery({
    queryKey: QK.friends('mutual'),
    queryFn: () => fetchFriends('mutual'),
  });
  const incomingQ = useQuery({
    queryKey: QK.friends('incoming'),
    queryFn: () => fetchFriends('incoming'),
  });
  const outgoingQ = useQuery({
    queryKey: QK.friends('outgoing'),
    queryFn: () => fetchFriends('outgoing'),
  });

  const lookupQ = useQuery({
    queryKey: QK.search('player', addQuery),
    queryFn: () => search(addQuery, 'player'),
    enabled: addQuery.length >= 2,
  });

  function invalidateAll() {
    void qc.invalidateQueries({ queryKey: ['friends'] });
  }

  const addMut = useMutation({
    mutationFn: addFriend,
    onSuccess: () => {
      setAddQuery('');
      setErrMsg(null);
      invalidateAll();
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const acceptMut = useMutation({
    mutationFn: acceptFriend,
    onSuccess: () => {
      setErrMsg(null);
      invalidateAll();
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const removeMut = useMutation({
    mutationFn: removeFriend,
    onSuccess: invalidateAll,
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (mutualQ.isLoading) {
    return <div className="idiv">{t('friends', 'loading')}</div>;
  }
  const mutual = mutualQ.data?.friends ?? [];
  const incoming = incomingQ.data?.friends ?? [];
  const outgoing = outgoingQ.data?.friends ?? [];
  const candidates = lookupQ.data?.players ?? [];

  return (
    <>
      {/* Mutual-друзья */}
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={5}>
              {t('friends', 'tabAccepted')} ({mutual.length})
            </th>
          </tr>
          <tr>
            <th>{t('friends', 'colPlayer')}</th>
            <th>{t('overview', 'points') || 'Очки'}</th>
            <th>{t('alliance', 'alliance') || 'Альянс'}</th>
            <th>{t('friends', 'colStatus')}</th>
            <th>{t('friends', 'colActions')}</th>
          </tr>
        </thead>
        <tbody>
          {mutual.length === 0 ? (
            <tr>
              <td colSpan={5} className="center">
                {t('friends', 'empty')}
              </td>
            </tr>
          ) : (
            mutual.map((f) => (
              <FriendRow
                key={f.user_id}
                f={f}
                tools={
                  <>
                    <Link to={`/msg/compose?to=${encodeURIComponent(f.username)}`}>
                      {t('message', 'newMessage') || 'Сообщение'}
                    </Link>
                    {' · '}
                    <button
                      type="button"
                      className="button"
                      disabled={removeMut.isPending}
                      onClick={() => {
                        if (
                          window.confirm(
                            `${t('friends', 'removeBtn')} ${f.username}?`,
                          )
                        ) {
                          removeMut.mutate(f.user_id);
                        }
                      }}
                    >
                      ✕
                    </button>
                  </>
                }
                t={t}
              />
            ))
          )}
        </tbody>
      </table>

      {/* Входящие запросы — рендерим только если есть */}
      {incoming.length > 0 && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={5}>
                {t('friends', 'tabIncoming')} ({incoming.length})
              </th>
            </tr>
            <tr>
              <th>{t('friends', 'colPlayer')}</th>
              <th>{t('overview', 'points') || 'Очки'}</th>
              <th>{t('alliance', 'alliance') || 'Альянс'}</th>
              <th>{t('friends', 'colStatus')}</th>
              <th>{t('friends', 'colActions')}</th>
            </tr>
          </thead>
          <tbody>
            {incoming.map((f) => (
              <FriendRow
                key={f.user_id}
                f={f}
                tools={
                  <>
                    <button
                      type="button"
                      className="button"
                      disabled={acceptMut.isPending}
                      onClick={() => acceptMut.mutate(f.user_id)}
                    >
                      ✓ {t('friends', 'acceptBtn')}
                    </button>
                    {' · '}
                    <button
                      type="button"
                      className="button"
                      disabled={removeMut.isPending}
                      onClick={() => removeMut.mutate(f.user_id)}
                    >
                      ✕ {t('friends', 'declineBtn')}
                    </button>
                  </>
                }
                t={t}
              />
            ))}
          </tbody>
        </table>
      )}

      {/* Отправленные запросы — рендерим только если есть */}
      {outgoing.length > 0 && (
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={5}>
                {t('friends', 'tabOutgoing')} ({outgoing.length})
              </th>
            </tr>
            <tr>
              <th>{t('friends', 'colPlayer')}</th>
              <th>{t('overview', 'points') || 'Очки'}</th>
              <th>{t('alliance', 'alliance') || 'Альянс'}</th>
              <th>{t('friends', 'colStatus')}</th>
              <th>{t('friends', 'colActions')}</th>
            </tr>
          </thead>
          <tbody>
            {outgoing.map((f) => (
              <FriendRow
                key={f.user_id}
                f={f}
                tools={
                  <button
                    type="button"
                    className="button"
                    disabled={removeMut.isPending}
                    onClick={() => removeMut.mutate(f.user_id)}
                  >
                    ✕ {t('friends', 'cancelRequestBtn')}
                  </button>
                }
                t={t}
              />
            ))}
          </tbody>
        </table>
      )}

      {/* Add-form */}
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('friends', 'addBtn')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <input
                type="text"
                value={addQuery}
                placeholder={t('friends', 'addPlaceholder')}
                onChange={(e) => setAddQuery(e.target.value)}
                maxLength={64}
              />
            </td>
            <td className="center">
              <span className="small">{t('search', 'hint')}</span>
            </td>
          </tr>
          {addQuery.length >= 2 && (
            <tr>
              <td colSpan={2}>
                {lookupQ.isFetching && !lookupQ.data ? (
                  <span>{t('search', 'searching')}</span>
                ) : candidates.length === 0 ? (
                  <span>{t('search', 'notFound')}</span>
                ) : (
                  <ul style={{ margin: 0, paddingLeft: 16 }}>
                    {candidates.slice(0, 8).map((c) => (
                      <li key={c.user_id}>
                        {c.username}{' '}
                        <button
                          type="button"
                          className="button"
                          disabled={addMut.isPending}
                          onClick={() => addMut.mutate(c.user_id)}
                        >
                          + {t('friends', 'addBtn')}
                        </button>
                      </li>
                    ))}
                  </ul>
                )}
              </td>
            </tr>
          )}
        </tbody>
      </table>

      {errMsg && (
        <div className="idiv">
          <span className="false">{errMsg}</span>
        </div>
      )}
    </>
  );
}

function FriendRow({
  f,
  tools,
  t,
}: {
  f: Friend;
  tools: React.ReactNode;
  t: (g: string, k: string, v?: Record<string, string | number>) => string;
}) {
  return (
    <tr>
      <td>{f.username}</td>
      <td>{Math.round(f.points)}</td>
      <td>{f.alliance_tag ?? '—'}</td>
      <td>{formatLastSeen(f.last_seen, t)}</td>
      <td className="center">{tools}</td>
    </tr>
  );
}
