import { useState } from 'react';
import { useAuthStore } from '@/stores/auth';
import { api } from '@/api/client';
import type { AuthResponse } from '@/api/types';
import { useTranslation } from '@/i18n/i18n';

type Mode = 'login' | 'register';

export function LoginScreen() {
  const setTokens = useAuthStore((s) => s.setTokens);
  const [mode, setMode] = useState<Mode>('login');
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const { t } = useTranslation('auth');

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      // План 36 Ф.11: фронтенд ходит в auth-service (через vite proxy /auth/* в dev,
      // через nginx auth.oxsar-nova.ru в prod). На login auth-service ждёт поле
      // login (email или username), на register — username + email + password.
      const path = mode === 'login' ? '/auth/login' : '/auth/register';
      const body =
        mode === 'login'
          ? { login: email, password }
          : { username, email, password };
      const res = await api.post<AuthResponse>(path, body);
      setTokens({
        access: res.tokens.access,
        refresh: res.tokens.refresh,
        userId: res.user.id,
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'unknown error');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      background: 'var(--ox-bg)',
      padding: '24px 16px',
    }}>
      <div style={{ width: '100%', maxWidth: 400 }}>

        {/* Логотип */}
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <div style={{ fontSize: 48, marginBottom: 8 }}>🚀</div>
          <div style={{
            fontSize: 26, fontWeight: 700,
            fontFamily: 'var(--ox-font)',
            color: 'var(--ox-accent)',
            letterSpacing: '0.06em',
          }}>
            OXSAR
          </div>
          <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)', marginTop: 4 }}>
            {t('tagline')}
          </div>
        </div>

        <div className="ox-panel" style={{ padding: '28px 32px' }}>
          <div className="ox-tabs" style={{ marginBottom: 24 }}>
            <button type="button" aria-pressed={mode === 'login'} onClick={() => setMode('login')}>
              {t('loginTab')}
            </button>
            <button type="button" aria-pressed={mode === 'register'} onClick={() => setMode('register')}>
              {t('registerTab')}
            </button>
          </div>

          <form className="ox-form" onSubmit={submit}>
            {mode === 'register' && (
              <label>
                <span>{t('usernameLabel')}</span>
                <input
                  required
                  minLength={3}
                  maxLength={24}
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  autoComplete="username"
                  placeholder={t('usernamePlaceholder')}
                />
              </label>
            )}
            <label>
              <span>{mode === 'login' ? t('emailOrLogin') : t('emailLabel')}</span>
              <input
                required
                type={mode === 'login' ? 'text' : 'email'}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                autoComplete="email"
                placeholder={mode === 'login' ? t('loginPlaceholder') : t('emailPlaceholder')}
              />
            </label>
            <label>
              <span>{t('passwordLabel')}</span>
              <input
                required
                minLength={8}
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
                placeholder={t('passwordPlaceholder')}
              />
            </label>
            {error && <div className="ox-error">{error}</div>}
            <button type="submit" className="btn" disabled={loading} style={{ width: '100%', marginTop: 8 }}>
              {loading ? '…' : mode === 'login' ? t('loginButton') : t('registerButton')}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
