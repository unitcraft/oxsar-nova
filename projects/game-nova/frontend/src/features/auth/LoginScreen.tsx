import { useState } from 'react';
import { useAuthStore } from '@/stores/auth';
import { api } from '@/api/client';
import type { AuthResponse } from '@/api/types';
import { useTranslation } from '@/i18n/i18n';
import { AgeRating } from '@/components/AgeRating';

type Mode = 'login' | 'register';

// План 44: Privacy Policy живёт на портале (oxsar-nova.ru/privacy). В проде
// game-nova на subdomain — нужен абсолютный URL. В dev (vite) — относительный.
// План 47: к ней добавились Договор-оферта, Правила игры и Политика возврата
// на том же портале.
const PORTAL_BASE =
  (import.meta.env['VITE_PORTAL_BASE_URL'] as string | undefined) ?? '';
const PORTAL_PRIVACY_URL =
  (import.meta.env['VITE_PORTAL_PRIVACY_URL'] as string | undefined) ??
  (PORTAL_BASE ? `${PORTAL_BASE}/privacy` : '/privacy');
const PORTAL_OFFER_URL = PORTAL_BASE ? `${PORTAL_BASE}/offer` : '/offer';
const PORTAL_GAME_RULES_URL = PORTAL_BASE ? `${PORTAL_BASE}/game-rules` : '/game-rules';
const PORTAL_REFUND_URL = PORTAL_BASE ? `${PORTAL_BASE}/refund` : '/refund';

export function LoginScreen() {
  const setTokens = useAuthStore((s) => s.setTokens);
  const [mode, setMode] = useState<Mode>('login');
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [consent, setConsent] = useState(false);
  const [termsConsent, setTermsConsent] = useState(false);
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
      // План 44 (152-ФЗ): на register дополнительно отдаём consent_accepted —
      // подтверждение согласия с обработкой ПДн.
      // План 47: terms_accepted — акцепт Договора-оферты, Правил игры и
      // Политики возврата.
      const path = mode === 'login' ? '/auth/login' : '/auth/register';
      const body =
        mode === 'login'
          ? { login: email, password }
          : {
              username,
              email,
              password,
              consent_accepted: consent,
              terms_accepted: termsConsent,
            };
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
          <div style={{ fontSize: 14, color: 'var(--ox-fg-muted)', marginTop: 4, display: 'flex', justifyContent: 'center', alignItems: 'center', gap: 8 }}>
            {t('tagline')}
            <AgeRating size="sm" />
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
            {mode === 'register' && (
              <label
                style={{
                  display: 'flex', alignItems: 'flex-start', gap: 8,
                  fontSize: 13, color: 'var(--ox-fg-muted)', lineHeight: 1.5,
                  marginTop: 4,
                }}
              >
                <input
                  type="checkbox"
                  checked={termsConsent}
                  onChange={(e) => setTermsConsent(e.target.checked)}
                  required
                  style={{ marginTop: 3, flexShrink: 0 }}
                />
                <span>
                  {t('termsConsentPrefix')}{' '}
                  <a href={PORTAL_OFFER_URL} target="_blank" rel="noreferrer" style={{ color: 'var(--ox-accent)' }}>
                    {t('termsConsentOffer')}
                  </a>
                  {t('termsConsentSeparator1')}
                  <a href={PORTAL_GAME_RULES_URL} target="_blank" rel="noreferrer" style={{ color: 'var(--ox-accent)' }}>
                    {t('termsConsentRules')}
                  </a>
                  {t('termsConsentSeparator2')}
                  <a href={PORTAL_REFUND_URL} target="_blank" rel="noreferrer" style={{ color: 'var(--ox-accent)' }}>
                    {t('termsConsentRefund')}
                  </a>
                  {t('termsConsentSuffix')}
                </span>
              </label>
            )}
            {mode === 'register' && (
              <label
                style={{
                  display: 'flex', alignItems: 'flex-start', gap: 8,
                  fontSize: 13, color: 'var(--ox-fg-muted)', lineHeight: 1.5,
                  marginTop: 4,
                }}
              >
                <input
                  type="checkbox"
                  checked={consent}
                  onChange={(e) => setConsent(e.target.checked)}
                  required
                  style={{ marginTop: 3, flexShrink: 0 }}
                />
                <span>
                  {t('consentPrefix')}{' '}
                  <a href={PORTAL_PRIVACY_URL} target="_blank" rel="noreferrer" style={{ color: 'var(--ox-accent)' }}>
                    {t('consentLinkText')}
                  </a>
                  {t('consentSuffix')}
                </span>
              </label>
            )}
            {error && <div className="ox-error">{error}</div>}
            <button
              type="submit"
              className="btn"
              disabled={loading || (mode === 'register' && (!consent || !termsConsent))}
              style={{ width: '100%', marginTop: 8 }}
            >
              {loading ? '…' : mode === 'login' ? t('loginButton') : t('registerButton')}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
