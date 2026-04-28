// S-034 Friends — список друзей (план 72 Ф.5 Spring 4).
//
// Pixel-perfect зеркало legacy `templates/standard/buddylist.tpl`:
//   ntable с колонками username / points / alliance / position / status.
//   Удаление через DELETE-кнопку, добавление через input + поиск.
//
// Endpoints: GET /api/friends, POST /api/friends/{userId},
//            DELETE /api/friends/{userId} + GET /api/search для add-flow.

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { addFriend, fetchFriends, removeFriend } from '@/api/friends';
import { search } from '@/api/search';
import { QK } from '@/api/query-keys';
import type { ApiError } from '@/api/client';
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

  const friendsQ = useQuery({
    queryKey: QK.friends(),
    queryFn: fetchFriends,
  });

  const lookupQ = useQuery({
    queryKey: QK.search('player', addQuery),
    queryFn: () => search(addQuery, 'player'),
    enabled: addQuery.length >= 2,
  });

  const addMut = useMutation({
    mutationFn: addFriend,
    onSuccess: () => {
      setAddQuery('');
      setErrMsg(null);
      void qc.invalidateQueries({ queryKey: QK.friends() });
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  const removeMut = useMutation({
    mutationFn: removeFriend,
    onSuccess: () => void qc.invalidateQueries({ queryKey: QK.friends() }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (friendsQ.isLoading) {
    return <div className="idiv">{t('friends', 'loading')}</div>;
  }
  const friends = friendsQ.data?.friends ?? [];
  const candidates = lookupQ.data?.players ?? [];

  return (
    <>
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={5}>{t('friends', 'title')}</th>
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
          {friends.length === 0 ? (
            <tr>
              <td colSpan={5} className="center">
                {t('friends', 'empty')}
              </td>
            </tr>
          ) : (
            friends.map((f) => (
              <tr key={f.user_id}>
                <td>{f.username}</td>
                <td>{Math.round(f.points)}</td>
                <td>{f.alliance_tag ?? '—'}</td>
                <td>{formatLastSeen(f.last_seen, t)}</td>
                <td className="center">
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
                </td>
              </tr>
            ))
          )}
        </tbody>
      </table>

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
