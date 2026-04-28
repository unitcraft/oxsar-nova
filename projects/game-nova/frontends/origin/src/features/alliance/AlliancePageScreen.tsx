// S-008/S-011 Alliance page (внешний просмотр + Apply, план 72 Ф.3
// Spring 2 ч.1).
//
// Pixel-perfect зеркало legacy `templates/standard/allypage_own.tpl`
// (внешняя страница альянса с join-формой) + apply.tpl (форма заявки
// для closed-альянса).
//
// Поведение по openapi `/api/alliances/{id}` + `/api/alliances/{id}/descriptions`:
//   - Любой залогиненный игрок может открыть страницу.
//   - Если открыт — кнопка [Вступить] (POST /join, без message).
//   - Если закрыт — поле message + кнопка [Подать заявку].
//   - Если игрок уже состоит в этом же альянсе — редирект на
//     /alliance/me (там полный owner/member функционал).

import { useState } from 'react';
import { useNavigate, useParams, Navigate } from 'react-router-dom';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  fetchAlliance,
  fetchDescriptions,
  joinAlliance,
} from '@/api/alliance';
import type { ApiError } from '@/api/client';
import { QK } from '@/api/query-keys';
import { useTranslation } from '@/i18n/i18n';
import { useMyAlliance } from './common';

export function AlliancePageScreen() {
  const { id } = useParams<{ id: string }>();
  const { t } = useTranslation();
  const qc = useQueryClient();
  const navigate = useNavigate();
  const my = useMyAlliance();

  const detail = useQuery({
    queryKey: QK.alliance(id ?? ''),
    queryFn: () => fetchAlliance(id ?? ''),
    enabled: !!id,
  });

  const descr = useQuery({
    queryKey: QK.allianceDescriptions(id ?? ''),
    queryFn: () => fetchDescriptions(id ?? ''),
    enabled: !!id,
  });

  const [message, setMessage] = useState('');
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const join = useMutation({
    mutationFn: () => joinAlliance(id ?? '', message),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      navigate('/alliance/me');
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (!id) return <Navigate to="/alliance/list" replace />;
  if (detail.isLoading) return <div className="idiv">…</div>;
  if (!detail.data) return <div className="idiv">—</div>;

  // Если игрок состоит в этом же альянсе — кидаем на свою страницу с
  // полным набором действий (manage / members / ranks / ...).
  if (my.data && my.data.alliance.id === detail.data.alliance.id) {
    return <Navigate to="/alliance/me" replace />;
  }

  const al = detail.data.alliance;
  const inAnother = !!my.data && my.data.alliance.id !== al.id;
  const externalText = descr.data?.description_external ?? al.description ?? '';
  const applyText = descr.data?.description_apply ?? '';

  return (
    <form
      method="post"
      onSubmit={(ev) => {
        ev.preventDefault();
        if (!join.isPending && !inAnother) join.mutate();
      }}
    >
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('alliance', 'alliances')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td style={{ width: '30%' }}>{t('alliance', 'tag')}</td>
            <td>{al.tag}</td>
          </tr>
          <tr>
            <td>{t('alliance', 'name')}</td>
            <td>{al.name}</td>
          </tr>
          <tr>
            <td>{t('alliance', 'member')}</td>
            <td>{al.member_count}</td>
          </tr>
          <tr>
            <td>{t('alliance', 'founder')}</td>
            <td>{al.owner_name}</td>
          </tr>
          {externalText && (
            <tr>
              <td colSpan={2} className="center">
                <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>
                  {externalText}
                </pre>
              </td>
            </tr>
          )}
          {!my.data && al.is_open && (
            <tr>
              <td colSpan={2} className="center">
                <input
                  type="submit"
                  name="enter"
                  value={t('alliance', 'join')}
                  className="button"
                  disabled={join.isPending}
                />
              </td>
            </tr>
          )}
          {!my.data && !al.is_open && (
            <>
              {applyText && (
                <tr>
                  <td colSpan={2}>
                    <pre style={{ whiteSpace: 'pre-wrap', margin: 0 }}>
                      {applyText}
                    </pre>
                  </td>
                </tr>
              )}
              <tr>
                <td colSpan={2}>
                  <textarea
                    cols={60}
                    rows={10}
                    className="center"
                    name="application"
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    maxLength={4000}
                    placeholder={t('alliance', 'applicationText')}
                  />
                </td>
              </tr>
              <tr>
                <td colSpan={2} className="center">
                  <input
                    type="submit"
                    name="apply"
                    value={t('alliance', 'applyBtn')}
                    className="button"
                    disabled={join.isPending}
                  />
                </td>
              </tr>
            </>
          )}
          {inAnother && (
            <tr>
              <td colSpan={2} className="center">
                <span className="false">
                  {t('alliance', 'applicationInProgress')}
                </span>
              </td>
            </tr>
          )}
          {errMsg && (
            <tr>
              <td colSpan={2} className="center">
                <span className="false">{errMsg}</span>
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </form>
  );
}
