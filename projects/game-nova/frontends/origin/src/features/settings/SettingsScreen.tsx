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
  cancelDeletion,
  changePassword,
  confirmDeletion,
  disableVacation,
  enableVacation,
  fetchSettings,
  requestDeletionCode,
  updateSettings,
} from '@/api/settings';
import { fetchMe } from '@/api/me';
import { QK } from '@/api/query-keys';
import type { ApiError } from '@/api/client';
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

  const settingsQ = useQuery({
    queryKey: QK.settings(),
    queryFn: fetchSettings,
  });

  const [email, setEmail] = useState('');
  const [language, setLanguage] = useState<'ru' | 'en'>('ru');
  const [timezone, setTimezone] = useState('UTC');
  const [profileMsg, setProfileMsg] = useState<string | null>(null);
  const [profileErr, setProfileErr] = useState<string | null>(null);
  // План 72.1.55 Task I (P72.S4.SETTINGS subset 1:1).
  const [showAllConstr, setShowAllConstr] = useState(true);
  const [showAllResearch, setShowAllResearch] = useState(true);
  const [showAllShipyard, setShowAllShipyard] = useState(true);
  const [showAllDefense, setShowAllDefense] = useState(true);
  const [planetOrder, setPlanetOrder] = useState<number>(0);
  // План 72.1.55.E: esps — int 1..99 (default count of spy probes).
  const [esps, setEsps] = useState<number>(5);
  const [ipcheck, setIpcheck] = useState(true);

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

  const [vacationErr, setVacationErr] = useState<string | null>(null);

  // Подплан 72.1.5 A: vacation toggle. /api/me содержит vacation_since
  // (при null — не в отпуске), vacation_unlock_at (NOW+48h после
  // включения), vacation_last_end (для проверки 20-дневного окна
  // на бэке).
  const meQ = useQuery({
    queryKey: QK.me(),
    queryFn: fetchMe,
  });

  const vacationOnMut = useMutation({
    mutationFn: enableVacation,
    onSuccess: () => {
      setVacationErr(null);
      void qc.invalidateQueries({ queryKey: QK.me() });
    },
    onError: (e) => setVacationErr((e as ApiError).message),
  });
  const vacationOffMut = useMutation({
    mutationFn: disableVacation,
    onSuccess: () => {
      setVacationErr(null);
      void qc.invalidateQueries({ queryKey: QK.me() });
    },
    onError: (e) => setVacationErr((e as ApiError).message),
  });

  useEffect(() => {
    if (settingsQ.data) {
      setEmail(settingsQ.data.email);
      setLanguage(settingsQ.data.language);
      setTimezone(settingsQ.data.timezone);
      // План 72.1.55 Task I.
      setShowAllConstr(settingsQ.data.show_all_constructions);
      setShowAllResearch(settingsQ.data.show_all_research);
      setShowAllShipyard(settingsQ.data.show_all_shipyard);
      setShowAllDefense(settingsQ.data.show_all_defense);
      setPlanetOrder(settingsQ.data.planet_order);
      setEsps(settingsQ.data.esps);
      setIpcheck(settingsQ.data.ipcheck);
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
    // План 72.1.30: после ConfirmDeletion ставится grace 7 дней,
    // юзер остаётся залогиненным. Invalidate me чтобы UI показал
    // delete_at + cancel-кнопку. Раньше делали logout().
    onSuccess: () => {
      setDeleteErr(null);
      void qc.invalidateQueries({ queryKey: QK.me() });
    },
    onError: (e) => setDeleteErr((e as ApiError).message),
  });

  // План 72.1.30: cancel pending удаления.
  const cancelDelMut = useMutation({
    mutationFn: cancelDeletion,
    onSuccess: () => {
      setDeleteErr(null);
      setCodeSent(false);
      setDeletionCode('');
      void qc.invalidateQueries({ queryKey: QK.me() });
    },
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
    // План 72.1.55 Task I.
    if (showAllConstr !== settingsQ.data?.show_all_constructions) patch.show_all_constructions = showAllConstr;
    if (showAllResearch !== settingsQ.data?.show_all_research) patch.show_all_research = showAllResearch;
    if (showAllShipyard !== settingsQ.data?.show_all_shipyard) patch.show_all_shipyard = showAllShipyard;
    if (showAllDefense !== settingsQ.data?.show_all_defense) patch.show_all_defense = showAllDefense;
    if (planetOrder !== settingsQ.data?.planet_order) patch.planet_order = planetOrder;
    if (esps !== settingsQ.data?.esps) patch.esps = esps;
    if (ipcheck !== settingsQ.data?.ipcheck) patch.ipcheck = ipcheck;
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
            {/* План 72.1.55 Task I (P72.S4.SETTINGS subset 1:1):
                legacy preferences.tpl поля. Effects (применение
                этих preferences на UI/backend) — отдельные подзадачи
                72.1.55.*. */}
            <tr>
              <td colSpan={2}>
                <hr />
                <b>{t('settings', 'prefsTitle') || 'Игровые настройки'}</b>
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="show-all-constr">
                  {t('settings', 'showAllConstructions') || 'Показывать недоступные постройки'}
                </label>
              </td>
              <td>
                <input
                  type="checkbox"
                  id="show-all-constr"
                  checked={showAllConstr}
                  onChange={(e) => setShowAllConstr(e.target.checked)}
                />
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="show-all-research">
                  {t('settings', 'showAllResearch') || 'Показывать недоступные исследования'}
                </label>
              </td>
              <td>
                <input
                  type="checkbox"
                  id="show-all-research"
                  checked={showAllResearch}
                  onChange={(e) => setShowAllResearch(e.target.checked)}
                />
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="show-all-shipyard">
                  {t('settings', 'showAllShipyard') || 'Показывать недоступные корабли'}
                </label>
              </td>
              <td>
                <input
                  type="checkbox"
                  id="show-all-shipyard"
                  checked={showAllShipyard}
                  onChange={(e) => setShowAllShipyard(e.target.checked)}
                />
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="show-all-defense">
                  {t('settings', 'showAllDefense') || 'Показывать недоступную оборону'}
                </label>
              </td>
              <td>
                <input
                  type="checkbox"
                  id="show-all-defense"
                  checked={showAllDefense}
                  onChange={(e) => setShowAllDefense(e.target.checked)}
                />
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="planet-order">
                  {t('settings', 'planetOrder') || 'Сортировка планет'}
                </label>
              </td>
              <td>
                <select
                  id="planet-order"
                  value={planetOrder}
                  onChange={(e) => setPlanetOrder(Number(e.target.value))}
                >
                  <option value={0}>{t('settings', 'planetOrderDate') || 'По дате колонизации'}</option>
                  <option value={1}>{t('settings', 'planetOrderName') || 'По имени'}</option>
                  <option value={2}>{t('settings', 'planetOrderCoords') || 'По координатам'}</option>
                </select>
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="esps">
                  {t('settings', 'esps') || 'Шпионских зондов по умолчанию'}
                </label>
              </td>
              <td>
                <input
                  type="number"
                  id="esps"
                  min={1}
                  max={99}
                  value={esps}
                  onChange={(e) => {
                    const v = Math.max(1, Math.min(99, Number(e.target.value) || 1));
                    setEsps(v);
                  }}
                  style={{ width: '4em' }}
                />
              </td>
            </tr>
            <tr>
              <td>
                <label htmlFor="ipcheck">
                  {t('settings', 'ipcheck') || 'Уведомлять при логине с другого IP'}
                </label>
              </td>
              <td>
                <input
                  type="checkbox"
                  id="ipcheck"
                  checked={ipcheck}
                  onChange={(e) => setIpcheck(e.target.checked)}
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

      {/* Vacation toggle (план 72.1.5 A). Backend готов через
          POST/DELETE /api/me/vacation. Состояние юзера — из /api/me. */}
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('prefs', 'vacationTitle')}</th>
          </tr>
        </thead>
        <tbody>
          {meQ.isLoading ? (
            <tr><td colSpan={2}>…</td></tr>
          ) : (() => {
            const me = meQ.data;
            if (!me) return null;
            const isOnVacation = me.vacation_since != null;
            if (!isOnVacation) {
              return (
                <tr>
                  <td>{t('prefs', 'vacationOffStatus')}</td>
                  <td>
                    <button
                      type="button"
                      className="button"
                      disabled={vacationOnMut.isPending}
                      onClick={() => vacationOnMut.mutate()}
                    >
                      {t('prefs', 'vacationEnableBtn')}
                    </button>{' '}
                    {vacationErr && <span className="false">{vacationErr}</span>}
                  </td>
                </tr>
              );
            }
            const unlock = me.vacation_unlock_at
              ? new Date(me.vacation_unlock_at)
              : null;
            const tooEarly = unlock != null && unlock.getTime() > Date.now();
            return (
              <tr>
                <td>
                  {t('prefs', 'vacationOnStatus', {
                    since: new Date(me.vacation_since!).toLocaleString(),
                  })}
                  {tooEarly && unlock && (
                    <>
                      <br />
                      <small>
                        {t('prefs', 'vacationCannotEndYet', {
                          unlock: unlock.toLocaleString(),
                        })}
                      </small>
                    </>
                  )}
                </td>
                <td>
                  <button
                    type="button"
                    className="button"
                    disabled={vacationOffMut.isPending || tooEarly}
                    onClick={() => vacationOffMut.mutate()}
                  >
                    {t('prefs', 'vacationDisableBtn')}
                  </button>{' '}
                  {vacationErr && <span className="false">{vacationErr}</span>}
                </td>
              </tr>
            );
          })()}
        </tbody>
      </table>

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
          {/* План 72.1.30: grace 7 дней — pending deletion warning. */}
          {meQ.data?.delete_at ? (
            <tr>
              <td colSpan={2} className="center">
                <p className="false">
                  {t('settings', 'deletionScheduledTitle') ||
                    '⚠ Аккаунт будет удалён'}
                </p>
                <p>
                  {t('settings', 'deletionScheduledBody', {
                    when: new Date(meQ.data.delete_at).toLocaleString('ru-RU'),
                  }) || `Удаление запланировано на ${new Date(meQ.data.delete_at).toLocaleString('ru-RU')}.`}
                </p>
                <button
                  type="button"
                  className="button"
                  disabled={cancelDelMut.isPending}
                  onClick={() => cancelDelMut.mutate()}
                >
                  {cancelDelMut.isPending
                    ? '…'
                    : t('settings', 'cancelDeletionBtn') || 'Отменить удаление'}
                </button>
                {deleteErr && (
                  <div>
                    <span className="false">{deleteErr}</span>
                  </div>
                )}
              </td>
            </tr>
          ) : !dangerOpen ? (
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
