import { lazy, Suspense, useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { fetchMe } from '@/lib/auth/flow';
import { Login } from '@/routes/Login';
import { Dashboard } from '@/routes/Dashboard';
import { Protected } from '@/routes/Protected';
import { RootLayout } from '@/components/layout/RootLayout';
import { Placeholder } from '@/routes/Placeholder';
import { Skeleton } from '@/components/ui/skeleton';

// Lazy-load страниц с большими таблицами/графиками: разнесёт bundle
// по chunks (план 53 §Производительность). Сейчас — placeholder, в
// Ф.4-7 каждая будет вытащена в отдельный module.
const RouteFallback = (): React.ReactElement => (
  <div className="space-y-3">
    <Skeleton className="h-6 w-48" />
    <Skeleton className="h-4 w-64" />
    <Skeleton className="h-32 w-full max-w-xl" />
  </div>
);

const UsersLookupPage = lazy(async () => ({
  default: (await import('@/routes/UsersLookup')).UsersLookup,
}));
const UserDetailPage = lazy(async () => ({
  default: (await import('@/routes/UserDetail')).UserDetail,
}));
const RolesPage = lazy(async () => ({
  default: (await import('@/routes/Roles')).Roles,
}));
const AuditPage = lazy(async () => ({
  default: (await import('@/routes/Audit')).Audit,
}));
const BillingStub = lazy(async () => ({
  default: () => <Placeholder title="Billing" phase="план 54" description="платежи, возвраты, лимиты" />,
}));
const GameOpsStub = lazy(async () => ({
  default: () => <Placeholder title="Game-ops" phase="Ф.6" description="dead events, planet/fleet operations" />,
}));
const SettingsStub = lazy(async () => ({
  default: () => <Placeholder title="Settings" phase="Ф.8" description="2FA, security, sessions" />,
}));

export function App(): React.ReactElement {
  useEffect(() => {
    void fetchMe();
  }, []);

  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        element={
          <Protected>
            <RootLayout />
          </Protected>
        }
      >
        <Route path="/" element={<Dashboard />} />
        <Route
          path="/users"
          element={<Navigate to="/users/lookup" replace />}
        />
        <Route
          path="/users/lookup"
          element={
            <Suspense fallback={<RouteFallback />}>
              <UsersLookupPage />
            </Suspense>
          }
        />
        <Route
          path="/users/:id"
          element={
            <Suspense fallback={<RouteFallback />}>
              <UserDetailPage />
            </Suspense>
          }
        />
        <Route
          path="/roles"
          element={
            <Suspense fallback={<RouteFallback />}>
              <RolesPage />
            </Suspense>
          }
        />
        <Route
          path="/audit"
          element={
            <Suspense fallback={<RouteFallback />}>
              <AuditPage />
            </Suspense>
          }
        />
        <Route
          path="/billing/*"
          element={
            <Suspense fallback={<RouteFallback />}>
              <BillingStub />
            </Suspense>
          }
        />
        <Route
          path="/game-ops/*"
          element={
            <Suspense fallback={<RouteFallback />}>
              <GameOpsStub />
            </Suspense>
          }
        />
        <Route
          path="/settings/*"
          element={
            <Suspense fallback={<RouteFallback />}>
              <SettingsStub />
            </Suspense>
          }
        />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
