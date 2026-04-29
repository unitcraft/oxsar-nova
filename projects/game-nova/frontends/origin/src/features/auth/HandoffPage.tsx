// HandoffPage обрабатывает редирект из портала в эту вселенную.
// План 72.2 — universe session handoff.
//
// URL: /auth/handoff?code=<one-time-handoff-token>
//
// Flow:
//   1. Юзер на портале клик «Играть».
//   2. Portal-backend POST identity /auth/universe-token → handoff_token.
//   3. Portal возвращает redirect_url: <game-uri>/auth/handoff?code=<token>.
//   4. Браузер открыл этот URL — мы здесь.
//   5. POST /auth/token/exchange (vite-proxy → identity) → access+refresh.
//   6. Сохранили в auth-store, редиректим на /.
//
// /auth/* проксируется vite в identity-service (см. vite.config.ts).

import { useEffect, useState } from 'react';
import { useAuthStore } from '@/stores/auth';

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
          // Чистим URL и редиректим на главную.
          window.history.replaceState(null, '', '/');
          window.location.href = '/';
        }
      },
      (e) => setError(e instanceof Error ? e.message : 'unknown'),
    );

    async function exchange(handoffCode: string): Promise<boolean> {
      const res = await fetch('/auth/token/exchange', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ code: handoffCode }),
      });
      if (!res.ok) {
        const body = (await res.json().catch(() => ({}))) as {
          error?: { message?: string };
        };
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

  // Ссылка обратно на портал — для случая когда code просрочен.
  const portalURL = (import.meta.env['VITE_PORTAL_URL'] as string) || '/';

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: '#000',
        color: '#fff',
        fontFamily: 'sans-serif',
      }}
    >
      <div
        style={{
          padding: '32px 48px',
          textAlign: 'center',
          minWidth: 320,
          border: '1px solid #333',
          background: '#0a0a0a',
        }}
      >
        {error ? (
          <>
            <div style={{ fontSize: 32, marginBottom: 8 }}>⚠️</div>
            <div style={{ fontSize: 18, marginBottom: 16 }}>
              Ошибка перехода во вселенную
            </div>
            <div style={{ marginBottom: 16, color: '#f55' }}>{error}</div>
            <a href={portalURL} style={{ color: '#3af' }}>
              Вернуться на портал
            </a>
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
