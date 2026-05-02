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
  updateAlliancePrefs,
  type AlliancePrefsInput,
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

  // План 72.1.54 (P72.S2.ALLIANCE_PREFS 1:1): legacy updateAllyPrefs.
  const [logoInput, setLogoInput] = useState('');
  const [homepageInput, setHomepageInput] = useState('');
  const [foundernameInput, setFoundernameInput] = useState('');
  const [showMemberInput, setShowMemberInput] = useState(true);
  const [showHomepageInput, setShowHomepageInput] = useState(true);
  const [memberlistSortInput, setMemberlistSortInput] = useState(0);
  const updatePrefs = useMutation({
    mutationFn: (prefs: AlliancePrefsInput) => updateAlliancePrefs(allianceID, prefs),
    onSuccess: () => {
      setErrMsg(null);
      void qc.invalidateQueries({ queryKey: ['alliances'] });
    },
    onError: (e) => setErrMsg((e as ApiError).message),
  });
  // Init prefs из alliance data — sync на смену al.
  useEffect(() => {
    if (my.data) {
      const a = my.data.alliance;
      setLogoInput(a.logo ?? '');
      setHomepageInput(a.homepage ?? '');
      setFoundernameInput(a.foundername ?? '');
      setShowMemberInput(a.show_member ?? true);
      setShowHomepageInput(a.show_homepage ?? true);
      setMemberlistSortInput(a.memberlist_sort ?? 0);
    }
  }, [my.data]);

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

      {/* План 72.1.54 (P72.S2.ALLIANCE_PREFS 1:1): legacy
          updateAllyPrefs — logo/homepage/foundername/show_member/
          show_homepage/memberlist_sort. */}
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('alliance', 'preferences') || 'Настройки альянса'}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <label htmlFor="logo-input">
                {t('alliance', 'logoUrl') || 'Логотип (URL картинки)'}
              </label>
            </td>
            <td>
              <input
                id="logo-input"
                type="url"
                value={logoInput}
                maxLength={128}
                placeholder="https://example.com/logo.png"
                onChange={(e) => setLogoInput(e.target.value)}
                disabled={updatePrefs.isPending}
                style={{ width: '24em' }}
              />
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="homepage-input">
                {t('alliance', 'homepageUrl') || 'Сайт альянса (URL)'}
              </label>
            </td>
            <td>
              <input
                id="homepage-input"
                type="url"
                value={homepageInput}
                maxLength={128}
                placeholder="https://..."
                onChange={(e) => setHomepageInput(e.target.value)}
                disabled={updatePrefs.isPending}
                style={{ width: '24em' }}
              />{' '}
              <input
                type="checkbox"
                id="show-homepage"
                checked={showHomepageInput}
                onChange={(e) => setShowHomepageInput(e.target.checked)}
                disabled={updatePrefs.isPending}
              />{' '}
              <label htmlFor="show-homepage">
                {t('alliance', 'showHomepage') || 'Показывать всем'}
              </label>
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="foundername-input">
                {t('alliance', 'foundername') || 'Title основателя'}
              </label>
            </td>
            <td>
              <input
                id="foundername-input"
                type="text"
                value={foundernameInput}
                maxLength={64}
                onChange={(e) => setFoundernameInput(e.target.value)}
                disabled={updatePrefs.isPending}
                style={{ width: '20em' }}
              />
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="show-member">
                {t('alliance', 'showMember') || 'Memberlist виден всем'}
              </label>
            </td>
            <td>
              <input
                type="checkbox"
                id="show-member"
                checked={showMemberInput}
                onChange={(e) => setShowMemberInput(e.target.checked)}
                disabled={updatePrefs.isPending}
              />
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="memberlist-sort">
                {t('alliance', 'memberlistSort') || 'Сортировка memberlist'}
              </label>
            </td>
            <td>
              <select
                id="memberlist-sort"
                value={memberlistSortInput}
                onChange={(e) => setMemberlistSortInput(Number(e.target.value))}
                disabled={updatePrefs.isPending}
              >
                <option value={0}>{t('alliance', 'sortRank') || 'По рангу'}</option>
                <option value={1}>{t('alliance', 'sortName') || 'По имени'}</option>
                <option value={2}>{t('alliance', 'sortPoints') || 'По очкам'}</option>
                <option value={3}>{t('alliance', 'sortJoined') || 'По дате вступления'}</option>
                <option value={4}>{t('alliance', 'sortLastSeen') || 'По активности'}</option>
              </select>
            </td>
          </tr>
          <tr>
            <td colSpan={2} className="center">
              <button
                type="button"
                className="button"
                disabled={updatePrefs.isPending}
                onClick={() => {
                  updatePrefs.mutate({
                    logo: logoInput.trim() || null,
                    homepage: homepageInput.trim() || null,
                    foundername: foundernameInput.trim() || null,
                    show_member: showMemberInput,
                    show_homepage: showHomepageInput,
                    memberlist_sort: memberlistSortInput,
                  });
                }}
              >
                {t('alliance', 'saveBtn') || 'Сохранить'}
              </button>
            </td>
          </tr>
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
