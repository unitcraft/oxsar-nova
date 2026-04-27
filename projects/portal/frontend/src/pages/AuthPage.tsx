import { useState } from 'react';
import { portalApi } from '@/api/portal';
import { useAuthStore } from '@/store/auth';
import { Link } from '@/components/Link';
import styles from './AuthPage.module.css';

export function LoginPage() {
  const { setAuth } = useAuthStore();
  const [login, setLogin] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await portalApi.auth.login(login, password);
      setAuth(res.user, res.tokens);
      window.location.href = '/';
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка входа');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <h1 className={styles.title}>Вход</h1>
        <form className={styles.form} onSubmit={(e) => void submit(e)}>
          <label className={styles.label}>
            Email или имя игрока
            <input
              className={styles.input}
              type="text"
              value={login}
              onChange={(e) => setLogin(e.target.value)}
              autoComplete="username"
              required
            />
          </label>
          <label className={styles.label}>
            Пароль
            <input
              className={styles.input}
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              autoComplete="current-password"
              required
            />
          </label>
          {error && <div className={styles.error}>{error}</div>}
          <button className={styles.submit} type="submit" disabled={loading}>
            {loading ? 'Входим…' : 'Войти'}
          </button>
        </form>
        <p className={styles.switch}>
          Нет аккаунта? <Link href="/register">Зарегистрироваться</Link>
        </p>
      </div>
    </div>
  );
}

export function RegisterPage() {
  const { setAuth } = useAuthStore();
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [consent, setConsent] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await portalApi.auth.register(username, email, password, consent);
      setAuth(res.user, res.tokens);
      window.location.href = '/';
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Ошибка регистрации');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.page}>
      <div className={styles.card}>
        <h1 className={styles.title}>Регистрация</h1>
        <form className={styles.form} onSubmit={(e) => void submit(e)}>
          <label className={styles.label}>
            Имя игрока (3–24 символа)
            <input
              className={styles.input}
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              minLength={3} maxLength={24}
              autoComplete="username"
              required
            />
          </label>
          <label className={styles.label}>
            Email
            <input
              className={styles.input}
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              autoComplete="email"
              required
            />
          </label>
          <label className={styles.label}>
            Пароль (минимум 8 символов)
            <input
              className={styles.input}
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              minLength={8}
              autoComplete="new-password"
              required
            />
          </label>
          <label className={styles.consent}>
            <input
              type="checkbox"
              checked={consent}
              onChange={(e) => setConsent(e.target.checked)}
              required
            />
            <span>
              Я ознакомлен с{' '}
              <Link href="/privacy">Политикой конфиденциальности</Link> и даю
              согласие на обработку моих персональных данных в соответствии с
              Федеральным законом № 152-ФЗ.
            </span>
          </label>
          {error && <div className={styles.error}>{error}</div>}
          <button
            className={styles.submit}
            type="submit"
            disabled={loading || !consent}
          >
            {loading ? 'Создаём аккаунт…' : 'Создать аккаунт'}
          </button>
        </form>
        <p className={styles.switch}>
          Уже есть аккаунт? <Link href="/login">Войти</Link>
        </p>
      </div>
    </div>
  );
}
