import { useState } from 'react';
import { useAuthStore } from '@/stores/auth';
import { api } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';
import type { AuthResponse } from '@/api/types';

type Mode = 'login' | 'register';

export function LoginScreen() {
  const setTokens = useAuthStore((s) => s.setTokens);
  const { t } = useTranslation();
  const [mode, setMode] = useState<Mode>('login');
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const path = mode === 'login' ? '/api/auth/login' : '/api/auth/register';
      const body =
        mode === 'login' ? { email, password } : { username, email, password };
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
    <div>
      <div className="ox-tabs">
        <button type="button" aria-pressed={mode === 'login'} onClick={() => setMode('login')}>
          {t('Registration', 'LOGIN')}
        </button>
        <button type="button" aria-pressed={mode === 'register'} onClick={() => setMode('register')}>
          {t('Registration', 'REGISTRATION')}
        </button>
      </div>

      <form className="ox-form" onSubmit={submit}>
        {mode === 'register' && (
          <label>
            <span>{t('Registration', 'USERNAME')}</span>
            <input
              required
              minLength={3}
              maxLength={24}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoComplete="username"
            />
          </label>
        )}
        <label>
          <span>{t('global', 'EMAIL')}</span>
          <input
            required
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoComplete="email"
          />
        </label>
        <label>
          <span>{t('Registration', 'PASSWORD')}</span>
          <input
            required
            minLength={8}
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
          />
        </label>
        {error && <div className="ox-error">{error}</div>}
        <button type="submit" disabled={loading}>
          {loading ? '…' : t('Registration', mode === 'login' ? 'LOGIN' : 'REGISTRATION')}
        </button>
      </form>
    </div>
  );
}
