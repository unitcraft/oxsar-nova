import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../api/client';
import { useAuthStore } from '../../stores/auth';
import { useTranslation } from '@/i18n/i18n';

interface SettingsData {
  email: string;
  language: string;
  timezone: string;
  vacation_since: string | null;
}

interface PlanetLite {
  id: string;
  name: string;
  galaxy: number;
  system: number;
  position: number;
  is_moon: boolean;
}

const TIMEZONES = [
  { value: 'UTC', label: 'UTC' },
  { value: 'Europe/Moscow', label: 'Москва (UTC+3)' },
  { value: 'Europe/Kiev', label: 'Киев (UTC+2/+3)' },
  { value: 'Europe/Minsk', label: 'Минск (UTC+3)' },
  { value: 'Asia/Yekaterinburg', label: 'Екатеринбург (UTC+5)' },
  { value: 'Asia/Novosibirsk', label: 'Новосибирск (UTC+7)' },
  { value: 'Asia/Vladivostok', label: 'Владивосток (UTC+10)' },
  { value: 'Asia/Almaty', label: 'Алматы (UTC+6)' },
  { value: 'Europe/Berlin', label: 'Берлин (UTC+1/+2)' },
  { value: 'Europe/London', label: 'Лондон (UTC+0/+1)' },
  { value: 'America/New_York', label: 'Нью-Йорк (UTC-5/-4)' },
  { value: 'America/Los_Angeles', label: 'Лос-Анджелес (UTC-8/-7)' },
  { value: 'Asia/Tokyo', label: 'Токио (UTC+9)' },
];

export function SettingsScreen() {
  const { t } = useTranslation('settings');
  const qc = useQueryClient();
  const { data, isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: () => api.get<SettingsData>('/api/settings'),
  });

  const updateMutation = useMutation({
    mutationFn: (body: Partial<{ email: string; language: string; timezone: string }>) =>
      api.put('/api/settings', body),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['settings'] }),
  });

  const vacationSetMutation = useMutation({
    mutationFn: () => api.post('/api/me/vacation', {}),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['settings'] }),
  });

  const vacationUnsetMutation = useMutation({
    mutationFn: () => api.delete('/api/me/vacation'),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['settings'] }),
  });

  // План 36 Critical-6: смена пароля переехала в auth-service.
  // Хеш живёт в auth-db, в game-db password_hash IS NULL.
  // /auth/* через vite proxy дёргает auth-service.
  const passwordMutation = useMutation({
    mutationFn: (body: { current: string; new: string }) =>
      api.post('/auth/password', body),
  });

  const [email, setEmail] = useState('');
  const [emailSaved, setEmailSaved] = useState(false);
  const [emailError, setEmailError] = useState('');

  const [currentPw, setCurrentPw] = useState('');
  const [newPw, setNewPw] = useState('');
  const [confirmPw, setConfirmPw] = useState('');
  const [pwSaved, setPwSaved] = useState(false);
  const [pwError, setPwError] = useState('');

  const [vacationConfirm, setVacationConfirm] = useState(false);

  const [dangerOpen, setDangerOpen] = useState(false);
  const [codeSent, setCodeSent] = useState(false);
  const [codeExpires, setCodeExpires] = useState<string>('');
  const [code, setCode] = useState('');
  const [deleteError, setDeleteError] = useState('');
  const logout = useAuthStore((s) => s.logout);

  const requestCodeMutation = useMutation({
    mutationFn: () => api.post<{ expires_at: string }>('/api/me/deletion/code'),
    onSuccess: (r) => { setCodeSent(true); setCodeExpires(r.expires_at); setDeleteError(''); },
    onError: (e) => setDeleteError(e instanceof Error ? e.message : t('requestCodeErr')),
  });

  const confirmDeleteMutation = useMutation({
    mutationFn: (c: string) => api.delete<void>('/api/me', { code: c }),
    onSuccess: () => { logout(); },
    onError: (e) => setDeleteError(e instanceof Error ? e.message : t('confirmDeleteErr')),
  });

  if (isLoading || !data) {
    return (
      <div style={{ padding: 24 }}>
        <div className="ox-skeleton" style={{ height: 400, borderRadius: 8 }} />
      </div>
    );
  }

  const isOnVacation = data.vacation_since !== null;

  async function handleEmailSave() {
    setEmailError('');
    setEmailSaved(false);
    try {
      await updateMutation.mutateAsync({ email: email || data.email });
      setEmailSaved(true);
      setTimeout(() => setEmailSaved(false), 3000);
    } catch (e: unknown) {
      setEmailError(e instanceof Error ? e.message : t('pwSaveError'));
    }
  }

  async function handlePasswordSave() {
    setPwError('');
    setPwSaved(false);
    if (newPw !== confirmPw) { setPwError(t('pwMismatch')); return; }
    if (newPw.length < 8) { setPwError(t('pwTooShort')); return; }
    try {
      await passwordMutation.mutateAsync({ current: currentPw, new: newPw });
      setPwSaved(true);
      setCurrentPw(''); setNewPw(''); setConfirmPw('');
      setTimeout(() => setPwSaved(false), 3000);
    } catch (e: unknown) {
      setPwError(e instanceof Error ? e.message : t('pwWrongCurrent'));
    }
  }

  return (
    <div style={{ maxWidth: 600, margin: '0 auto', padding: 24, display: 'flex', flexDirection: 'column', gap: 24 }}>
      <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700 }}>{t('title')}</h2>

      <section className="ox-panel" style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}>
        <h3 style={{ margin: 0, fontSize: 16, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ox-fg-muted)' }}>{t('sectionProfile')}</h3>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={{ fontSize: 15, color: 'var(--ox-fg-dim)' }}>{t('labelEmail')}</label>
          <div style={{ display: 'flex', gap: 8 }}>
            <input
              type="email"
              className="ox-input"
              style={{ flex: 1 }}
              placeholder={data.email}
              value={email}
              onChange={(e) => { setEmail(e.target.value); setEmailSaved(false); }}
            />
            <button
              type="button"
              className="btn"
              disabled={updateMutation.isPending || !email}
              onClick={() => void handleEmailSave()}
            >
              {t('saveBtn')}
            </button>
          </div>
          {emailSaved && <span style={{ fontSize: 14, color: 'var(--ox-success)' }}>{t('emailSaved')}</span>}
          {emailError && <span style={{ fontSize: 14, color: 'var(--ox-danger)' }}>{emailError}</span>}
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={{ fontSize: 15, color: 'var(--ox-fg-dim)' }}>{t('labelLanguage')}</label>
          <select
            className="ox-input"
            value={data.language}
            onChange={(e) => void updateMutation.mutateAsync({ language: e.target.value })}
          >
            <option value="ru">{t('langRu')}</option>
            <option value="en">{t('langEn')}</option>
          </select>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={{ fontSize: 15, color: 'var(--ox-fg-dim)' }}>{t('labelTimezone')}</label>
          <select
            className="ox-input"
            value={data.timezone}
            onChange={(e) => void updateMutation.mutateAsync({ timezone: e.target.value })}
          >
            {TIMEZONES.map((tz) => (
              <option key={tz.value} value={tz.value}>{tz.label}</option>
            ))}
          </select>
        </div>
      </section>

      <section className="ox-panel" style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}>
        <h3 style={{ margin: 0, fontSize: 16, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ox-fg-muted)' }}>{t('sectionSecurity')}</h3>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={{ fontSize: 15, color: 'var(--ox-fg-dim)' }}>{t('labelCurrentPw')}</label>
            <input
              type="password"
              className="ox-input"
              value={currentPw}
              onChange={(e) => setCurrentPw(e.target.value)}
              autoComplete="current-password"
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={{ fontSize: 15, color: 'var(--ox-fg-dim)' }}>{t('labelNewPw')}</label>
            <input
              type="password"
              className="ox-input"
              value={newPw}
              onChange={(e) => setNewPw(e.target.value)}
              autoComplete="new-password"
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={{ fontSize: 15, color: 'var(--ox-fg-dim)' }}>{t('labelConfirmPw')}</label>
            <input
              type="password"
              className="ox-input"
              value={confirmPw}
              onChange={(e) => setConfirmPw(e.target.value)}
              autoComplete="new-password"
            />
          </div>
          <button
            type="button"
            className="btn"
            style={{ alignSelf: 'flex-start' }}
            disabled={passwordMutation.isPending || !currentPw || !newPw || !confirmPw}
            onClick={() => void handlePasswordSave()}
          >
            {t('changePwBtn')}
          </button>
          {pwSaved && <span style={{ fontSize: 14, color: 'var(--ox-success)' }}>{t('pwSaved')}</span>}
          {pwError && <span style={{ fontSize: 14, color: 'var(--ox-danger)' }}>{pwError}</span>}
        </div>
      </section>

      <section className="ox-panel" style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}>
        <h3 style={{ margin: 0, fontSize: 16, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ox-fg-muted)' }}>{t('sectionVacation')}</h3>

        <p style={{ margin: 0, fontSize: 15, color: 'var(--ox-fg-dim)', lineHeight: 1.6 }}>
          {t('vacationDesc')}
        </p>

        {isOnVacation ? (
          <>
            <div style={{ fontSize: 15, color: 'var(--ox-accent)' }}>
              {t('vacationActive', { date: new Date(data.vacation_since!).toLocaleDateString('ru-RU') })}
            </div>
            <button
              type="button"
              className="btn"
              style={{ alignSelf: 'flex-start' }}
              disabled={vacationUnsetMutation.isPending}
              onClick={() => void vacationUnsetMutation.mutateAsync()}
            >
              {t('leaveVacationBtn')}
            </button>
          </>
        ) : (
          <>
            {!vacationConfirm ? (
              <button
                type="button"
                className="btn"
                style={{ alignSelf: 'flex-start' }}
                onClick={() => setVacationConfirm(true)}
              >
                {t('startVacationBtn')}
              </button>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                <p style={{ margin: 0, fontSize: 15, color: 'var(--ox-warn, #f59e0b)', fontWeight: 600 }}>
                  {t('vacationWarning')}
                </p>
                <div style={{ display: 'flex', gap: 8 }}>
                  <button
                    type="button"
                    className="btn btn-danger"
                    disabled={vacationSetMutation.isPending}
                    onClick={() => void vacationSetMutation.mutateAsync()}
                  >
                    {t('vacationConfirmBtn')}
                  </button>
                  <button
                    type="button"
                    className="btn-ghost"
                    onClick={() => setVacationConfirm(false)}
                  >
                    {t('vacationCancelBtn')}
                  </button>
                </div>
                {vacationSetMutation.isError && (
                  <span style={{ fontSize: 14, color: 'var(--ox-danger)' }}>
                    {vacationSetMutation.error instanceof Error ? vacationSetMutation.error.message : t('pwSaveError')}
                  </span>
                )}
              </div>
            )}
          </>
        )}
      </section>

      <PlanetOrderSection />

      <section style={{
        padding: 20,
        border: '1px solid var(--ox-danger)',
        borderRadius: 6,
        display: 'flex', flexDirection: 'column', gap: 12,
        background: 'rgba(239,68,68,0.03)',
      }}>
        <h3 style={{ margin: 0, fontSize: 16, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ox-danger)' }}>
          {t('sectionDanger')}
        </h3>

        {!dangerOpen ? (
          <button
            type="button"
            className="btn-ghost"
            style={{ alignSelf: 'flex-start', color: 'var(--ox-danger)' }}
            onClick={() => setDangerOpen(true)}
          >
            {t('deleteAccountBtn')}
          </button>
        ) : (
          <>
            <p style={{ margin: 0, fontSize: 15, color: 'var(--ox-fg-dim)', lineHeight: 1.6 }}>
              {t('deleteAccountDesc')}
            </p>

            {!codeSent ? (
              <div style={{ display: 'flex', gap: 8 }}>
                <button
                  type="button"
                  className="btn btn-danger"
                  disabled={requestCodeMutation.isPending}
                  onClick={() => requestCodeMutation.mutate()}
                >
                  {requestCodeMutation.isPending ? '…' : t('requestCodeBtn')}
                </button>
                <button type="button" className="btn-ghost" onClick={() => setDangerOpen(false)}>
                  {t('cancelBtn')}
                </button>
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)' }}>
                  {t('codeSentMsg', { time: codeExpires ? new Date(codeExpires).toLocaleTimeString('ru-RU') : '—' })}
                </div>
                <input
                  type="text"
                  placeholder="XXXXXXXX"
                  value={code}
                  onChange={(e) => setCode(e.target.value.toUpperCase().slice(0, 8))}
                  style={{
                    fontFamily: 'var(--ox-mono)', fontSize: 18, letterSpacing: '0.15em',
                    padding: '10px 12px', maxWidth: 200, textTransform: 'uppercase',
                  }}
                  maxLength={8}
                />
                <div style={{ display: 'flex', gap: 8 }}>
                  <button
                    type="button"
                    className="btn btn-danger"
                    disabled={code.length !== 8 || confirmDeleteMutation.isPending}
                    onClick={() => confirmDeleteMutation.mutate(code)}
                  >
                    {confirmDeleteMutation.isPending ? '…' : t('confirmDeleteBtn')}
                  </button>
                  <button
                    type="button"
                    className="btn-ghost"
                    onClick={() => { setDangerOpen(false); setCodeSent(false); setCode(''); setDeleteError(''); }}
                  >
                    {t('cancelBtn')}
                  </button>
                </div>
              </div>
            )}

            {deleteError && (
              <span style={{ fontSize: 14, color: 'var(--ox-danger)' }}>{deleteError}</span>
            )}
          </>
        )}
      </section>
    </div>
  );
}

function PlanetOrderSection() {
  const { t } = useTranslation('settings');
  const qc = useQueryClient();
  const q = useQuery({
    queryKey: ['planets'],
    queryFn: () => api.get<{ planets: PlanetLite[] }>('/api/planets'),
  });

  const reorder = useMutation({
    mutationFn: (ids: string[]) => api.patch<void>('/api/planets/order', { planet_ids: ids }),
    onSuccess: () => void qc.invalidateQueries({ queryKey: ['planets'] }),
  });

  const [order, setOrder] = useState<PlanetLite[]>([]);
  const [dragIdx, setDragIdx] = useState<number | null>(null);

  if (q.data && order.length === 0 && q.data.planets.length > 0) {
    setOrder(q.data.planets.filter((p) => !p.is_moon));
  }

  function onDragStart(i: number) { setDragIdx(i); }
  function onDragOver(e: React.DragEvent) { e.preventDefault(); }
  function onDrop(i: number) {
    if (dragIdx === null || dragIdx === i) return;
    const copy = [...order];
    const [moved] = copy.splice(dragIdx, 1);
    if (moved) copy.splice(i, 0, moved);
    setOrder(copy);
    setDragIdx(null);
    reorder.mutate(copy.map((p) => p.id));
  }

  if (q.isLoading || !q.data || order.length <= 1) return null;

  return (
    <section className="ox-panel" style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}>
      <h3 style={{ margin: 0, fontSize: 16, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ox-fg-muted)' }}>
        {t('sectionPlanetOrder')}
      </h3>
      <p style={{ margin: 0, fontSize: 14, color: 'var(--ox-fg-dim)' }}>
        {t('planetOrderDesc')}
      </p>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
        {order.map((p, i) => (
          <div
            key={p.id}
            draggable
            onDragStart={() => onDragStart(i)}
            onDragOver={onDragOver}
            onDrop={() => onDrop(i)}
            style={{
              display: 'flex', alignItems: 'center', gap: 10,
              padding: '8px 12px',
              background: dragIdx === i ? 'rgba(99,217,255,0.08)' : 'var(--ox-bg-panel)',
              border: '1px solid var(--ox-border)', borderRadius: 4,
              cursor: 'grab', userSelect: 'none',
            }}
          >
            <span style={{ color: 'var(--ox-fg-muted)', fontFamily: 'var(--ox-mono)' }}>⋮⋮</span>
            <span style={{ flex: 1 }}>🪐 {p.name}</span>
            <span style={{ fontFamily: 'var(--ox-mono)', fontSize: 13, color: 'var(--ox-fg-muted)' }}>
              [{p.galaxy}:{p.system}:{p.position}]
            </span>
          </div>
        ))}
      </div>
      {reorder.isPending && <span style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>{t('planetOrderSaving')}</span>}
    </section>
  );
}
