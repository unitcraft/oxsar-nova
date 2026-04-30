// AuthGate — компонент-обёртка для всех защищённых маршрутов.
// План 72.2: на game-фронтах нет собственного LoginScreen — единственный
// способ войти это handoff-flow с портала. Если юзер пришёл без токена —
// редирект на портал, где он залогинится и кликнет «Играть».
//
// Exceptions:
//   - /auth/handoff?code=... обрабатывается до AuthGate (вне Routes).
//   - /battle-report/:id (план 72.1 ч.20.11): анонимный публичный
//     просмотр. AuthGate пропускает без редиректа — отчёт грузится
//     через анонимный backend endpoint, остальная часть UI (sidebar/
//     planets) скрыта если нет токена.

import type { ReactNode } from 'react';
import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth';

const PORTAL_URL =
  (import.meta.env['VITE_PORTAL_URL'] as string) || 'http://localhost:5174';

// Анонимные routes — не редиректят на portal даже без auth.
const PUBLIC_ROUTES = [
  /^\/battle-report\/[0-9a-f-]+$/,
];

function isPublicRoute(pathname: string): boolean {
  return PUBLIC_ROUTES.some((re) => re.test(pathname));
}

export function AuthGate({ children }: { children: ReactNode }) {
  const accessToken = useAuthStore((s) => s.accessToken);
  const { pathname } = useLocation();
  const isPublic = isPublicRoute(pathname);

  useEffect(() => {
    if (accessToken === null && !isPublic) {
      window.location.replace(PORTAL_URL);
    }
  }, [accessToken, isPublic]);

  if (accessToken === null && !isPublic) {
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
