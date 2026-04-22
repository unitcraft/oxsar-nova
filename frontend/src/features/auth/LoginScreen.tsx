import { useState } from 'react';
import { useAuthStore } from '@/stores/auth';
import { api } from '@/api/client';
import type { AuthResponse } from '@/api/types';

type Mode = 'login' | 'register';

export function LoginScreen() {
  const setTokens = useAuthStore((s) => s.setTokens);
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
          <div style={{ fontSize: 12, color: 'var(--ox-fg-muted)', marginTop: 4 }}>
            космическая стратегия
          </div>
        </div>

        <div className="ox-panel" style={{ padding: '28px 32px' }}>
          <div className="ox-tabs" style={{ marginBottom: 24 }}>
            <button type="button" aria-pressed={mode === 'login'} onClick={() => setMode('login')}>
              Войти
            </button>
            <button type="button" aria-pressed={mode === 'register'} onClick={() => setMode('register')}>
              Регистрация
            </button>
          </div>

          <form className="ox-form" onSubmit={submit}>
            {mode === 'register' && (
              <label>
                <span>Имя пользователя</span>
                <input
                  required
                  minLength={3}
                  maxLength={24}
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  autoComplete="username"
                  placeholder="от 3 до 24 символов"
                />
              </label>
            )}
            <label>
              <span>{mode === 'login' ? 'E-Mail или логин' : 'E-Mail'}</span>
              <input
                required
                type={mode === 'login' ? 'text' : 'email'}
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                autoComplete="email"
                placeholder={mode === 'login' ? 'example@mail.com или alice' : 'example@mail.com'}
              />
            </label>
            <label>
              <span>Пароль</span>
              <input
                required
                minLength={8}
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                autoComplete={mode === 'login' ? 'current-password' : 'new-password'}
                placeholder="минимум 8 символов"
              />
            </label>
            {error && <div className="ox-error">{error}</div>}
            <button type="submit" className="btn" disabled={loading} style={{ width: '100%', marginTop: 8 }}>
              {loading ? '…' : mode === 'login' ? 'Войти' : 'Зарегистрироваться'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
