import { useEffect, useState } from 'react';
import { useAuthStore } from '@/stores/auth';

/**
 * HandoffPage обрабатывает редирект из Universe Switcher другой вселенной.
 *
 * URL: /auth/handoff?code=<one-time-handoff-token>
 *
 * Flow (план 36 Ф.5/Ф.8):
 *   1. Universe Switcher в uni01 кликнул «uni02».
 *   2. uni01-backend дёрнул identity-service /auth/universe-token → handoff_token.
 *   3. uni01-backend вернул redirect_url: <uni02>/auth/handoff?code=<token>.
 *   4. Браузер открыл этот URL — мы здесь.
 *   5. Дёргаем /auth/token/exchange с handoff_token → новые JWT (access+refresh).
 *   6. Сохраняем в localStorage, редиректим на /.
 *
 * /auth/* проксируется vite в identity-service (см. vite.config.ts).
 */
export function HandoffPage() {
  const setTokens = useAuthStore((s) => s.setTokens);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const params = new URLSearchParams(window.location.search);
    const code = params.get('code');
    if (!code) {
      setError('Отсутствует параметр code в URL');
      return;
    }
    void exchange(code).then(
      (ok) => {
        if (ok) {
          // Чистим URL и редиректим на главную (history.replaceState чтобы
          // /auth/handoff?code=... не остался в истории).
          window.history.replaceState(null, '', '/');
          window.location.href = '/';
        }
      },
      (e) => setError(e instanceof Error ? e.message : 'unknown'),
    );

    async function exchange(handoffToken: string): Promise<boolean> {
      const res = await fetch('/auth/token/exchange', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code: handoffToken }),
      });
      if (!res.ok) {
        const body = (await res.json().catch(() => ({}))) as { error?: { message?: string } };
        throw new Error(body.error?.message ?? `HTTP ${res.status}`);
      }
      const data = (await res.json()) as {
        access_token: string;
        refresh_token: string;
        token_type: string;
        expires_in: number;
        user: { id: string };
      };
      setTokens({
        access: data.access_token,
        refresh: data.refresh_token,
        userId: data.user.id,
      });
      return true;
    }
  }, [setTokens]);

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'var(--ox-bg)',
        fontFamily: 'var(--ox-font)',
      }}
    >
      <div className="ox-panel" style={{ padding: '32px 48px', textAlign: 'center', minWidth: 320 }}>
        {error ? (
          <>
            <div style={{ fontSize: 32, marginBottom: 8 }}>⚠️</div>
            <div style={{ fontSize: 18, marginBottom: 16 }}>Ошибка перехода во вселенную</div>
            <div className="ox-error" style={{ marginBottom: 16 }}>{error}</div>
            <a href="/" className="btn">Вернуться на главную</a>
          </>
        ) : (
          <>
            <div style={{ fontSize: 32, marginBottom: 8 }}>🚀</div>
            <div style={{ fontSize: 18 }}>Переход во вселенную…</div>
          </>
        )}
      </div>
    </div>
  );
}
