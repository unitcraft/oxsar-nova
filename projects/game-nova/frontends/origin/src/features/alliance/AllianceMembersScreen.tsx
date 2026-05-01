// S-012 Alliance members (план 72 Ф.3 Spring 2 ч.1).
//
// Pixel-perfect зеркало legacy `templates/standard/memberlist.tpl`.
// Список членов альянса с рангами / онлайн-статусом / кнопкой kick для
// owner'а или пользователя с can_kick.
//
// Endpoint: GET /api/alliances/{id} — там же members[] (rank, joined_at,
// rank_name). Owner может выгнать через DELETE /members/{userID}.

import { useState } from 'react';
import { Navigate, Link } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { kickMember } from '@/api/alliance';
import type { ApiError } from '@/api/client';
import { useAuthStore } from '@/stores/auth';
import { useTranslation } from '@/i18n/i18n';
import { useMyAlliance } from './common';

export function AllianceMembersScreen() {
  const { t } = useTranslation();
  const my = useMyAlliance();
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.userId);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const kick = useMutation({
    mutationFn: ({ allianceID, uid }: { allianceID: string; uid: string }) =>
      kickMember(allianceID, uid),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (my.isLoading) return <div className="idiv">…</div>;
  if (!my.data) return <Navigate to="/alliance" replace />;

  const al = my.data.alliance;
  const members = my.data.members;
  const canKick = !!userId && userId === al.owner_id;
  // План 72.1.45 §3: online + points → 6 колонок (5/6 без kick/с).
  const colspan = canKick ? 6 : 5;

  // Сортировка: owner — первый, остальные по очкам (нет данных) →
  // по дате вступления.
  const sorted = [...members].sort((a, b) => {
    if (a.user_id === al.owner_id) return -1;
    if (b.user_id === al.owner_id) return 1;
    return a.joined_at.localeCompare(b.joined_at);
  });

  return (
    <>
      <div className="idiv">
        <Link to="/alliance/me">← {al.tag}</Link>
      </div>
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={colspan}>
              {t('alliance', 'memberList')} ({t('alliance', 'totalMembers')}:{' '}
              {al.member_count})
            </th>
          </tr>
          <tr>
            <td>{t('alliance', 'newbie')}</td>
            <td>{t('alliance', 'rank')}</td>
            <td>{t('alliance', 'joinDate')}</td>
            <td>{t('alliance', 'memberPoints') || 'Очки'}</td>
            <td>{t('alliance', 'memberOnline') || 'Активность'}</td>
            {canKick && <td />}
          </tr>
        </thead>
        <tbody>
          {sorted.map((m) => {
            // План 72.1.45 §3: online (5 минут) / недавно (1 ч) / давно.
            const lastSeenMs = m.last_seen ? new Date(m.last_seen).getTime() : 0;
            const ageMs = lastSeenMs ? Date.now() - lastSeenMs : Infinity;
            const onlineDot =
              ageMs < 5 * 60_000
                ? '🟢'
                : ageMs < 60 * 60_000
                  ? '🟡'
                  : '⚫';
            const lastSeenLabel = lastSeenMs
              ? new Date(m.last_seen!).toLocaleString('ru-RU')
              : '—';
            return (
            <tr key={m.user_id}>
              <td>{m.username}</td>
              <td>{m.rank_name || m.rank}</td>
              <td className="center">
                {new Date(m.joined_at).toLocaleDateString('ru-RU')}
              </td>
              <td className="center">{(m.points ?? 0).toLocaleString('ru-RU')}</td>
              <td className="center" title={lastSeenLabel}>
                {onlineDot}
              </td>
              {canKick && (
                <td className="center">
                  {m.user_id !== al.owner_id && (
                    <button
                      type="button"
                      className="button"
                      disabled={kick.isPending}
                      title={t('alliance', 'remove')}
                      onClick={() => {
                        if (
                          window.confirm(
                            `${t('alliance', 'remove')} ${m.username}?`,
                          )
                        ) {
                          kick.mutate({ allianceID: al.id, uid: m.user_id });
                        }
                      }}
                    >
                      ✕
                    </button>
                  )}
                </td>
              )}
            </tr>
            );
          })}
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
