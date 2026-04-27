// Protected route guard. Если статус 'unknown' — показываем skeleton
// (идёт первый GET /auth/me). Если 'anonymous' — редирект на /login,
// сохраняя текущий location в state.from.
//
// Permission-guard опционально: если задан requiredPermission и у юзера
// его нет — 403 placeholder (план 53 Ф.3 заменит на полноценный
// PermissionDenied компонент).
import { Navigate, useLocation } from 'react-router-dom';
import { useAuth } from '@/store/auth';
import { Skeleton } from '@/components/ui/skeleton';

interface ProtectedProps {
  children: React.ReactNode;
  requiredPermission?: string;
}

export function Protected({ children, requiredPermission }: ProtectedProps): React.ReactElement {
  const status = useAuth((s) => s.status);
  const hasPermission = useAuth((s) => s.hasPermission);
  const location = useLocation();

  if (status === 'unknown') {
    return (
      <div className="p-8 space-y-3">
        <Skeleton className="h-6 w-48" />
        <Skeleton className="h-4 w-64" />
        <Skeleton className="h-32 w-full max-w-xl" />
      </div>
    );
  }
  if (status === 'anonymous') {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />;
  }
  if (requiredPermission && !hasPermission(requiredPermission)) {
    return (
      <div className="p-8">
        <h1 className="text-lg font-semibold">403 — нет прав</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Требуется permission: <code className="font-mono-sm">{requiredPermission}</code>
        </p>
      </div>
    );
  }
  return <>{children}</>;
}
