import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../api/client';
import { useAuthStore } from '../../stores/auth';

interface SettingsData {
  email: string;
  language: string;
  timezone: string;
  vacation_since: string | null;
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

  const passwordMutation = useMutation({
    mutationFn: (body: { current_password: string; new_password: string }) =>
      api.post('/api/settings/password', body),
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
    onError: (e) => setDeleteError(e instanceof Error ? e.message : 'Ошибка запроса кода'),
  });

  const confirmDeleteMutation = useMutation({
    mutationFn: (c: string) => api.delete<void>('/api/me', { code: c }),
    onSuccess: () => { logout(); },
    onError: (e) => setDeleteError(e instanceof Error ? e.message : 'Неверный код'),
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
      setEmailError(e instanceof Error ? e.message : 'Ошибка сохранения');
    }
  }

  async function handlePasswordSave() {
    setPwError('');
    setPwSaved(false);
    if (newPw !== confirmPw) { setPwError('Пароли не совпадают'); return; }
    if (newPw.length < 8) { setPwError('Минимум 8 символов'); return; }
    try {
      await passwordMutation.mutateAsync({ current_password: currentPw, new_password: newPw });
      setPwSaved(true);
      setCurrentPw(''); setNewPw(''); setConfirmPw('');
      setTimeout(() => setPwSaved(false), 3000);
    } catch (e: unknown) {
      setPwError(e instanceof Error ? e.message : 'Неверный текущий пароль');
    }
  }

  return (
    <div style={{ maxWidth: 600, margin: '0 auto', padding: 24, display: 'flex', flexDirection: 'column', gap: 24 }}>
      <h2 style={{ margin: 0, fontSize: 20, fontWeight: 700 }}>Настройки аккаунта</h2>

      {/* Профиль */}
      <section className="ox-panel" style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}>
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ox-fg-muted)' }}>Профиль</h3>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>Email</label>
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
              Сохранить
            </button>
          </div>
          {emailSaved && <span style={{ fontSize: 12, color: 'var(--ox-success)' }}>✓ Email обновлён</span>}
          {emailError && <span style={{ fontSize: 12, color: 'var(--ox-danger)' }}>{emailError}</span>}
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>Язык</label>
          <select
            className="ox-input"
            value={data.language}
            onChange={(e) => void updateMutation.mutateAsync({ language: e.target.value })}
          >
            <option value="ru">Русский</option>
            <option value="en">English</option>
          </select>
        </div>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
          <label style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>Часовой пояс</label>
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

      {/* Безопасность */}
      <section className="ox-panel" style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 16 }}>
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ox-fg-muted)' }}>Безопасность</h3>

        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>Текущий пароль</label>
            <input
              type="password"
              className="ox-input"
              value={currentPw}
              onChange={(e) => setCurrentPw(e.target.value)}
              autoComplete="current-password"
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>Новый пароль</label>
            <input
              type="password"
              className="ox-input"
              value={newPw}
              onChange={(e) => setNewPw(e.target.value)}
              autoComplete="new-password"
            />
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            <label style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>Повторите новый пароль</label>
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
            Сменить пароль
          </button>
          {pwSaved && <span style={{ fontSize: 12, color: 'var(--ox-success)' }}>✓ Пароль изменён</span>}
          {pwError && <span style={{ fontSize: 12, color: 'var(--ox-danger)' }}>{pwError}</span>}
        </div>
      </section>

      {/* Режим отпуска */}
      <section className="ox-panel" style={{ padding: 20, display: 'flex', flexDirection: 'column', gap: 12 }}>
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ox-fg-muted)' }}>Режим отпуска</h3>

        <p style={{ margin: 0, fontSize: 13, color: 'var(--ox-fg-dim)', lineHeight: 1.6 }}>
          В режиме отпуска вы защищены от атак, но не можете строить, исследовать, отправлять флот или торговать.
          Минимальный интервал между отпусками — 20 дней.
        </p>

        {isOnVacation ? (
          <>
            <div style={{ fontSize: 13, color: 'var(--ox-accent)' }}>
              ✈ Отпуск активен с {new Date(data.vacation_since!).toLocaleDateString('ru-RU')}
            </div>
            <button
              type="button"
              className="btn"
              style={{ alignSelf: 'flex-start' }}
              disabled={vacationUnsetMutation.isPending}
              onClick={() => void vacationUnsetMutation.mutateAsync()}
            >
              Выйти из отпуска
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
                Включить режим отпуска
              </button>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                <p style={{ margin: 0, fontSize: 13, color: 'var(--ox-warn, #f59e0b)', fontWeight: 600 }}>
                  ⚠ Вы уверены? Вся активность будет заморожена.
                </p>
                <div style={{ display: 'flex', gap: 8 }}>
                  <button
                    type="button"
                    className="btn btn-danger"
                    disabled={vacationSetMutation.isPending}
                    onClick={() => void vacationSetMutation.mutateAsync()}
                  >
                    Да, уйти в отпуск
                  </button>
                  <button
                    type="button"
                    className="btn-ghost"
                    onClick={() => setVacationConfirm(false)}
                  >
                    Отмена
                  </button>
                </div>
                {vacationSetMutation.isError && (
                  <span style={{ fontSize: 12, color: 'var(--ox-danger)' }}>
                    {vacationSetMutation.error instanceof Error ? vacationSetMutation.error.message : 'Ошибка'}
                  </span>
                )}
              </div>
            )}
          </>
        )}
      </section>

      {/* Опасная зона */}
      <section style={{
        padding: 20,
        border: '1px solid var(--ox-danger)',
        borderRadius: 6,
        display: 'flex', flexDirection: 'column', gap: 12,
        background: 'rgba(239,68,68,0.03)',
      }}>
        <h3 style={{ margin: 0, fontSize: 14, fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--ox-danger)' }}>
          ⚠ Опасная зона
        </h3>

        {!dangerOpen ? (
          <button
            type="button"
            className="btn-ghost"
            style={{ alignSelf: 'flex-start', color: 'var(--ox-danger)' }}
            onClick={() => setDangerOpen(true)}
          >
            🗑 Удалить аккаунт…
          </button>
        ) : (
          <>
            <p style={{ margin: 0, fontSize: 13, color: 'var(--ox-fg-dim)', lineHeight: 1.6 }}>
              Аккаунт будет удалён навсегда: планеты, флоты, сообщения и участие в альянсе
              исчезнут. Восстановление невозможно. Для подтверждения требуется одноразовый
              код, который будет отправлен вам в системные сообщения.
            </p>

            {!codeSent ? (
              <div style={{ display: 'flex', gap: 8 }}>
                <button
                  type="button"
                  className="btn btn-danger"
                  disabled={requestCodeMutation.isPending}
                  onClick={() => requestCodeMutation.mutate()}
                >
                  {requestCodeMutation.isPending ? '…' : 'Получить код подтверждения'}
                </button>
                <button type="button" className="btn-ghost" onClick={() => setDangerOpen(false)}>
                  Отмена
                </button>
              </div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                <div style={{ fontSize: 12, color: 'var(--ox-fg-muted)' }}>
                  Код отправлен в сообщения (папка «Система»). Действителен до{' '}
                  <span style={{ fontFamily: 'var(--ox-mono)', color: 'var(--ox-fg)' }}>
                    {codeExpires ? new Date(codeExpires).toLocaleTimeString('ru-RU') : '—'}
                  </span>.
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
                    {confirmDeleteMutation.isPending ? '…' : '🗑 Удалить аккаунт навсегда'}
                  </button>
                  <button
                    type="button"
                    className="btn-ghost"
                    onClick={() => { setDangerOpen(false); setCodeSent(false); setCode(''); setDeleteError(''); }}
                  >
                    Отмена
                  </button>
                </div>
              </div>
            )}

            {deleteError && (
              <span style={{ fontSize: 12, color: 'var(--ox-danger)' }}>{deleteError}</span>
            )}
          </>
        )}
      </section>
    </div>
  );
}
