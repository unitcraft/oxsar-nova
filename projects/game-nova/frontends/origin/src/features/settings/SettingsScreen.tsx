// S-042 Settings — настройки аккаунта (план 72 Ф.5 Spring 4).
//
// Pixel-perfect зеркало legacy `templates/standard/preferences.tpl`:
//   ntable с разделами USER_DATA / ADVANCED_PREFERENCES / DELETE_ACCOUNT.
//   В origin-фронте оставляем только реально поддерживаемые backend-ом
//   настройки: email, language, timezone, vacation, password, deletion.
//   Legacy-only поля (skin/template/IP-check/show_all_*) — не реализуем,
//   они описаны в simplifications.md как «legacy-only, P72.S4.SETTINGS».
//
// Endpoints:
//   GET/PUT /api/settings
//   POST /api/me/deletion/code
//   DELETE /api/me  body: {code}
//   POST /auth/password — identity-service

import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  changePassword,
  confirmDeletion,
  fetchSettings,
  requestDeletionCode,
  updateSettings,
} from '@/api/settings';
import { QK } from '@/api/query-keys';
import type { ApiError } from '@/api/client';
import { useAuthStore } from '@/stores/auth';
import { useTranslation } from '@/i18n/i18n';

const TIMEZONES = [
  'UTC',
  'Europe/Moscow',
  'Europe/Kiev',
  'Europe/Minsk',
  'Asia/Yekaterinburg',
  'Asia/Novosibirsk',
  'Asia/Vladivostok',
  'Asia/Almaty',
  'Europe/Berlin',
  'Europe/London',
  'America/New_York',
  'America/Los_Angeles',
  'Asia/Tokyo',
];

export function SettingsScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const logout = useAuthStore((s) => s.logout);

  const settingsQ = useQuery({
    queryKey: QK.settings(),
    queryFn: fetchSettings,
  });

  const [email, setEmail] = useState('');
  const [language, setLanguage] = useState<'ru' | 'en'>('ru');
  const [timezone, setTimezone] = useState('UTC');
  const [profileMsg, setProfileMsg] = useState<string | null>(null);
  const [profileErr, setProfileErr] = useState<string | null>(null);

  const [currentPw, setCurrentPw] = useState('');
  const [newPw, setNewPw] = useState('');
  const [confirmPw, setConfirmPw] = useState('');
  const [pwMsg, setPwMsg] = useState<string | null>(null);
  const [pwErr, setPwErr] = useState<string | null>(null);

  const [dangerOpen, setDangerOpen] = useState(false);
  const [codeSent, setCodeSent] = useState(false);
  const [codeExpires, setCodeExpires] = useState<string>('');
  const [deletionCode, setDeletionCode] = useState('');
  const [deleteErr, setDeleteErr] = useState<string | null>(null);

  useEffect(() => {
    if (settingsQ.data) {
      setEmail(settingsQ.data.email);
      setLanguage(settingsQ.data.language);
      setTimezone(settingsQ.data.timezone);
    }
  }, [settingsQ.data]);

  const updateMut = useMutation({
    mutationFn: updateSettings,
    onSuccess: () => {
      setProfileMsg(t('settings', 'emailSaved'));
      setProfileErr(null);
      void qc.invalidateQueries({ queryKey: QK.settings() });
      setTimeout(() => setProfileMsg(null), 3000);
    },
    onError: (e) => {
      setProfileErr((e as ApiError).message);
      setProfileMsg(null);
    },
  });

  const passwordMut = useMutation({
    mutationFn: changePassword,
    onSuccess: () => {
      setPwMsg(t('settings', 'pwSaved'));
      setPwErr(null);
      setCurrentPw('');
      setNewPw('');
      setConfirmPw('');
      setTimeout(() => setPwMsg(null), 3000);
    },
    onError: (e) => {
      setPwErr((e as ApiError).message ?? t('settings', 'pwSaveError'));
      setPwMsg(null);
    },
  });

  const requestCodeMut = useMutation({
    mutationFn: requestDeletionCode,
    onSuccess: (r) => {
      setCodeSent(true);
      setCodeExpires(r.expires_at);
      setDeleteErr(null);
    },
    onError: (e) => setDeleteErr((e as ApiError).message),
  });

  const confirmMut = useMutation({
    mutationFn: confirmDeletion,
    onSuccess: () => logout(),
    onError: (e) => setDeleteErr((e as ApiError).message),
  });

  if (settingsQ.isLoading) {
    return <div className="idiv">…</div>;
  }
  if (!settingsQ.data) {
    return <div className="idiv">{t('settings', 'pwSaveError')}</div>;
  }

  function saveProfile(e: React.FormEvent) {
    e.preventDefault();
    const patch: Parameters<typeof updateSettings>[0] = {};
    if (email !== settingsQ.data?.email) patch.email = email;
    if (language !== settingsQ.data?.language) patch.language = language;
    if (timezone !== settingsQ.data?.timezone) patch.timezone = timezone;
    if (Object.keys(patch).length === 0) return;
    updateMut.mutate(patch);
  }

  function savePassword(e: React.FormEvent) {
    e.preventDefault();
    setPwErr(null);
    if (newPw !== confirmPw) {
      setPwErr(t('settings', 'pwMismatch'));
      return;
    }
    if (newPw.length < 8) {
      setPwErr(t('settings', 'pwTooShort'));
      return;
    }
    passwordMut.mutate({ current_password: currentPw, new_password: newPw });
  }

  return (
    <>
      <form onSubmit={saveProfile}>
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={2}>{t('settings', 'sectionProfile')}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>
                <label htmlFor="email">{t('settings', 'labelEmail')}</label>
              </td>
              <td>
                <input
                  type="email"
                  id="email"
                  name="email"
                  maxLength={50}
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="language">
                  {t('settings', 'labelLanguage')}
                </label>
              </td>
              <td>
                <select
                  id="language"
                  name="language"
                  value={language}
                  onChange={(e) =>
                    setLanguage(e.target.value as 'ru' | 'en')
                  }
                >
                  <option value="ru">{t('settings', 'langRu')}</option>
                  <option value="en">{t('settings', 'langEn')}</option>
                </select>
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="timezone">
                  {t('settings', 'labelTimezone')}
                </label>
              </td>
              <td>
                <select
                  id="timezone"
                  name="timezone"
                  value={timezone}
                  onChange={(e) => setTimezone(e.target.value)}
                >
                  {TIMEZONES.map((tz) => (
                    <option key={tz} value={tz}>
                      {tz}
                    </option>
                  ))}
                </select>
              </td>
            </tr>
          </tbody>
          <tfoot>
            <tr>
              <td colSpan={2} className="center">
                <input
                  type="submit"
                  className="button"
                  value={t('settings', 'saveBtn')}
                  disabled={updateMut.isPending}
                />{' '}
                {profileMsg && <span className="true">{profileMsg}</span>}
                {profileErr && <span className="false">{profileErr}</span>}
              </td>
            </tr>
          </tfoot>
        </table>
      </form>

      <form onSubmit={savePassword}>
        <table className="ntable">
          <thead>
            <tr>
              <th colSpan={2}>{t('settings', 'sectionSecurity')}</th>
            </tr>
          </thead>
          <tbody>
            <tr>
              <td>
                <label htmlFor="cur-pw">
                  {t('settings', 'labelCurrentPw')}
                </label>
              </td>
              <td>
                <input
                  type="password"
                  id="cur-pw"
                  autoComplete="current-password"
                  value={currentPw}
                  onChange={(e) => setCurrentPw(e.target.value)}
                />
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="new-pw">{t('settings', 'labelNewPw')}</label>
              </td>
              <td>
                <input
                  type="password"
                  id="new-pw"
                  autoComplete="new-password"
                  value={newPw}
                  onChange={(e) => setNewPw(e.target.value)}
                />
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="conf-pw">
                  {t('settings', 'labelConfirmPw')}
                </label>
              </td>
              <td>
                <input
                  type="password"
                  id="conf-pw"
                  autoComplete="new-password"
                  value={confirmPw}
                  onChange={(e) => setConfirmPw(e.target.value)}
                />
              </td>
            </tr>
          </tbody>
          <tfoot>
            <tr>
              <td colSpan={2} className="center">
                <input
                  type="submit"
                  className="button"
                  value={t('settings', 'changePwBtn')}
                  disabled={
                    passwordMut.isPending ||
                    !currentPw ||
                    !newPw ||
                    !confirmPw
                  }
                />{' '}
                {pwMsg && <span className="true">{pwMsg}</span>}
                {pwErr && <span className="false">{pwErr}</span>}
              </td>
            </tr>
          </tfoot>
        </table>
      </form>

      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>
              <span className="false">
                {t('prefs', 'deleteAccount') || 'Удалить аккаунт'}
              </span>
            </th>
          </tr>
        </thead>
        <tbody>
          {!dangerOpen ? (
            <tr>
              <td colSpan={2} className="center">
                <button
                  type="button"
                  className="button"
                  onClick={() => setDangerOpen(true)}
                >
                  {t('prefs', 'deleteAccount') || 'Удалить аккаунт'}
                </button>
              </td>
            </tr>
          ) : !codeSent ? (
            <tr>
              <td colSpan={2} className="center">
                <p>
                  {t('settings', 'vacationDesc') ||
                    'Запросите код по email, затем введите его для подтверждения удаления.'}
                </p>
                <button
                  type="button"
                  className="button"
                  disabled={requestCodeMut.isPending}
                  onClick={() => requestCodeMut.mutate()}
                >
                  {requestCodeMut.isPending
                    ? '…'
                    : t('settings', 'deletionCode.title')}
                </button>{' '}
                <button
                  type="button"
                  className="button"
                  onClick={() => setDangerOpen(false)}
                >
                  {t('alliance', 'cancelBtn') || 'Отмена'}
                </button>
                {deleteErr && (
                  <div>
                    <span className="false">{deleteErr}</span>
                  </div>
                )}
              </td>
            </tr>
          ) : (
            <tr>
              <td colSpan={2} className="center">
                <p>
                  {t('settings', 'deletionCode.body', {
                    code: '—',
                    expiresAt: codeExpires
                      ? new Date(codeExpires).toLocaleString('ru-RU')
                      : '—',
                  })}
                </p>
                <input
                  type="text"
                  value={deletionCode}
                  maxLength={8}
                  placeholder="XXXXXXXX"
                  onChange={(e) =>
                    setDeletionCode(
                      e.target.value.toUpperCase().slice(0, 8),
                    )
                  }
                  style={{ fontFamily: 'monospace', letterSpacing: '0.15em' }}
                />{' '}
                <button
                  type="button"
                  className="button"
                  disabled={
                    deletionCode.length !== 8 || confirmMut.isPending
                  }
                  onClick={() => confirmMut.mutate(deletionCode)}
                >
                  {confirmMut.isPending
                    ? '…'
                    : t('prefs', 'deleteAccount') || 'Удалить'}
                </button>{' '}
                <button
                  type="button"
                  className="button"
                  onClick={() => {
                    setDangerOpen(false);
                    setCodeSent(false);
                    setDeletionCode('');
                    setDeleteErr(null);
                  }}
                >
                  {t('alliance', 'cancelBtn') || 'Отмена'}
                </button>
                {deleteErr && (
                  <div>
                    <span className="false">{deleteErr}</span>
                  </div>
                )}
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </>
  );
}
