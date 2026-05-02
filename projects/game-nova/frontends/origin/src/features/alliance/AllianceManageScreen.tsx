// S-018 Alliance management / settings (план 72 Ф.3 Spring 2 ч.1).
//
// Pixel-perfect зеркало legacy `templates/standard/manage_ally.tpl`
// (часть с tag/name + memberlist-prefs + open/closed toggle, без блока
// 3 textarea — он вынесен на /alliance/descriptions).
//
// В origin-фронте сохраняются только настройки, которые есть в backend:
//   - is_open (PATCH /api/alliances/{id}/open)
//   - распуск (на странице /alliance/me, кнопка abandonAlliance)
//   - передача лидерства (на странице /alliance/transfer)
//
// Поля legacy `name`/`tag` PATCH backend пока не поддерживает (см.
// openapi.yaml — для /api/alliances/{id} есть только GET/DELETE), поэтому
// они read-only и помечены simplifications.md (P72.S2.B).

import { Link, Navigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  setAllianceOpen,
  broadcastAllianceMail,
  updateAllianceTagName,
} from '@/api/alliance';
import type { ApiError } from '@/api/client';
import { useAuthStore } from '@/stores/auth';
import { useTranslation } from '@/i18n/i18n';
import { useEffect, useState } from 'react';
import { useMyAlliance } from './common';

export function AllianceManageScreen() {
  const { t } = useTranslation();
  const my = useMyAlliance();
  const qc = useQueryClient();
  const userId = useAuthStore((s) => s.userId);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const allianceID = my.data?.alliance.id ?? '';

  const setOpen = useMutation({
    mutationFn: (isOpen: boolean) => setAllianceOpen(allianceID, isOpen),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['alliances'] }),
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  // План 72.1.43: tag/name edit + broadcast.
  const [tagInput, setTagInput] = useState('');
  const [nameInput, setNameInput] = useState('');
  const [bcTitle, setBcTitle] = useState('');
  const [bcBody, setBcBody] = useState('');
  // Init inputs from alliance data — useEffect зависит от al.tag/name.
  useEffect(() => {
    if (my.data) {
      setTagInput(my.data.alliance.tag);
      setNameInput(my.data.alliance.name);
    }
  }, [my.data]);

  const updateTagName = useMutation({
    mutationFn: ({ tag, name }: { tag?: string; name?: string }) =>
      updateAllianceTagName(allianceID, tag, name),
    onSuccess: () => {
      setErrMsg(null);
      void qc.invalidateQueries({ queryKey: ['alliances'] });
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });
  const broadcast = useMutation({
    mutationFn: () => broadcastAllianceMail(allianceID, bcTitle, bcBody),
    onSuccess: () => {
      setErrMsg(null);
      setBcTitle('');
      setBcBody('');
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });

  if (my.isLoading) return <div className="idiv">…</div>;
  if (!my.data) return <Navigate to="/alliance" replace />;

  const al = my.data.alliance;
  const isOwner = !!userId && userId === al.owner_id;
  if (!isOwner) {
    return (
      <div className="idiv">
        <span className="false">{t('alliance', 'allianceManagement')}</span>
      </div>
    );
  }

  return (
    <>
      <div className="idiv">
        <Link to="/alliance/me">← {al.tag}</Link>
      </div>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('alliance', 'allianceManagement')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <label htmlFor="tag-input">{t('alliance', 'allianceTag')}</label>
            </td>
            <td>
              {/* План 72.1.43: tag editable. Submit через кнопку ниже. */}
              <input
                id="tag-input"
                type="text"
                value={tagInput}
                maxLength={5}
                onChange={(e) => setTagInput(e.target.value)}
                disabled={updateTagName.isPending}
              />
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="name-input">{t('alliance', 'allianceName')}</label>
            </td>
            <td>
              <input
                id="name-input"
                type="text"
                value={nameInput}
                maxLength={64}
                onChange={(e) => setNameInput(e.target.value)}
                disabled={updateTagName.isPending}
                style={{ width: '20em' }}
              />
            </td>
          </tr>
          <tr>
            <td colSpan={2} className="center">
              {/* План 72.1.52 (72.1.6 P3 closure): client-side
                  validation tag (3-5 символов, ASCII alphanumeric) и
                  name (3-64 символа). Backend `validateTag` /
                  `Create` сентинели дублируют, но FE-кнопка не должна
                  даже пытаться отправить пустое/негодное значение. */}
              <button
                type="button"
                className="button"
                disabled={
                  updateTagName.isPending ||
                  (tagInput === al.tag && nameInput === al.name) ||
                  tagInput.trim().length < 3 ||
                  tagInput.trim().length > 5 ||
                  !/^[A-Za-z0-9]+$/.test(tagInput.trim()) ||
                  nameInput.trim().length < 3 ||
                  nameInput.trim().length > 64
                }
                onClick={() => {
                  const payload: { tag?: string; name?: string } = {};
                  if (tagInput !== al.tag) payload.tag = tagInput;
                  if (nameInput !== al.name) payload.name = nameInput;
                  updateTagName.mutate(payload);
                }}
              >
                {t('alliance', 'saveBtn') || 'Сохранить'}
              </button>
            </td>
          </tr>
          <tr>
            <td colSpan={2} className="center">
              [{' '}
              <Link to="/alliance/descriptions">
                {t('alliance', 'externAllianceText')}
              </Link>{' '}
              ] [{' '}
              <Link to="/alliance/diplomacy">{t('alliance', 'diplomacy')}</Link>{' '}
              ] [{' '}
              <Link to="/alliance/ranks">
                {t('alliance', 'rightManagement')}
              </Link>{' '}
              ]
            </td>
          </tr>
        </tbody>
      </table>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('alliance', 'memberList')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <label htmlFor="apps">
                {t('alliance', 'enableApplications')}
              </label>
            </td>
            <td>
              <input
                type="checkbox"
                id="apps"
                checked={al.is_open}
                disabled={setOpen.isPending}
                onChange={(e) => setOpen.mutate(e.target.checked)}
              />{' '}
              {al.is_open
                ? t('alliance', 'labelOpen')
                : t('alliance', 'labelClosed')}
            </td>
          </tr>
          {errMsg && (
            <tr>
              <td colSpan={2} className="center">
                <span className="false">{errMsg}</span>
              </td>
            </tr>
          )}
        </tbody>
      </table>

      <table className="ntable">
        <thead>
          <tr>
            <th>{t('alliance', 'referFounderStatus')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td className="center">
              <Link to="/alliance/transfer">
                {t('alliance', 'referFounderStatus')}
              </Link>
            </td>
          </tr>
        </tbody>
      </table>

      {/* План 72.1.43: legacy `Alliance::globalMail`. */}
      <table className="ntable">
        <thead>
          <tr>
            <th>📨 {t('alliance', 'globalMail') || 'Рассылка'}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <input
                type="text"
                placeholder={t('alliance', 'broadcastTitle') || 'Заголовок'}
                value={bcTitle}
                maxLength={120}
                onChange={(e) => setBcTitle(e.target.value)}
                disabled={broadcast.isPending}
                style={{ width: '100%' }}
              />
              <textarea
                placeholder={t('alliance', 'broadcastBody') || 'Сообщение'}
                value={bcBody}
                rows={4}
                maxLength={2000}
                onChange={(e) => setBcBody(e.target.value)}
                disabled={broadcast.isPending}
                style={{ width: '100%', marginTop: 4 }}
              />
              <div style={{ marginTop: 4, textAlign: 'right' }}>
                <button
                  type="button"
                  className="button"
                  disabled={
                    broadcast.isPending || !bcTitle || !bcBody
                  }
                  onClick={() => broadcast.mutate()}
                >
                  {broadcast.isPending
                    ? '…'
                    : t('alliance', 'broadcastBtn') || 'Отправить всем'}
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </>
  );
}
