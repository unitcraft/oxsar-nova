// Временный dashboard placeholder. План 53 Ф.3 даст полноценный
// layout (sidebar + topbar) и метрики; пока отображаем claims.
import { useAuth } from '@/store/auth';
import { logout } from '@/lib/auth/flow';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

export function Dashboard(): React.ReactElement {
  const claims = useAuth((s) => s.claims);
  const navigate = useNavigate();

  async function onLogout(): Promise<void> {
    await logout();
    navigate('/login', { replace: true });
  }

  if (!claims) return <></>;

  return (
    <div className="min-h-screen p-8">
      <header className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">oxsar-nova admin</h1>
          <p className="text-xs text-muted-foreground">
            план 53 Ф.2 — BFF auth flow рабочий
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Badge variant="secondary">
            <span className="font-mono-sm">{claims.username}</span>
          </Badge>
          <Button variant="outline" size="sm" onClick={onLogout}>
            Выйти
          </Button>
        </div>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-3xl">
        <Card>
          <CardHeader>
            <CardTitle>Roles</CardTitle>
          </CardHeader>
          <CardContent className="space-y-1">
            {claims.roles.length === 0 ? (
              <p className="text-xs text-muted-foreground">— нет ролей —</p>
            ) : (
              <div className="flex flex-wrap gap-1.5">
                {claims.roles.map((r) => (
                  <Badge key={r}>{r}</Badge>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Permissions ({claims.permissions.length})</CardTitle>
          </CardHeader>
          <CardContent>
            {claims.permissions.length === 0 ? (
              <p className="text-xs text-muted-foreground">— нет permissions —</p>
            ) : (
              <ul className="space-y-0.5 font-mono-sm text-xs">
                {claims.permissions.map((p) => (
                  <li key={p}>{p}</li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
