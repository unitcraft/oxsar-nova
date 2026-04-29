// AuthGate — компонент-обёртка для всех защищённых маршрутов.
// План 72.2: на game-фронтах нет собственного LoginScreen — единственный
// способ войти это handoff-flow с портала. Если юзер пришёл без токена —
// редирект на портал, где он залогинится и кликнет «Играть».
//
// Один exception: /auth/handoff?code=... обрабатывается до AuthGate
// (см. router.tsx — HandoffPage вне Routes-обёртки с гейтом).

import type { ReactNode } from 'react';
import { useEffect } from 'react';
import { useAuthStore } from '@/stores/auth';

const PORTAL_URL =
  (import.meta.env['VITE_PORTAL_URL'] as string) || 'http://localhost:5174';

export function AuthGate({ children }: { children: ReactNode }) {
  const accessToken = useAuthStore((s) => s.accessToken);

  useEffect(() => {
    if (accessToken === null) {
      // Редирект на портал. Используем replace, чтобы game-домен не
      // попал в browser history (юзер вернётся «назад» на портал, а
      // не на пустой game-фронт).
      window.location.replace(PORTAL_URL);
    }
  }, [accessToken]);

  if (accessToken === null) {
    return (
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: '#888',
          fontFamily: 'sans-serif',
        }}
      >
        Перенаправление на портал…
      </div>
    );
  }

  return <>{children}</>;
}
