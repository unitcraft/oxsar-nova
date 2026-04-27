// Dashboard — главный экран admin-консоли.
// Ф.3 даёт каркас (карточки-метрики + claims summary). Реальные метрики
// заполняются в дальнейших фазах (план 53 §Dashboard).
import { useAuth } from '@/store/auth';
import { PageHeader } from '@/components/layout/PageHeader';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

export function Dashboard(): React.ReactElement {
  const claims = useAuth((s) => s.claims);

  return (
    <>
      <PageHeader
        title="Dashboard"
        description="обзор системы и быстрые метрики"
      />

      <div className="grid grid-cols-1 gap-3 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader>
            <CardTitle>Active users (24h)</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold tracking-tight">—</div>
            <p className="mt-1 text-2xs text-muted-foreground">метрика добавится позже</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Revenue today</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold tracking-tight">—</div>
            <p className="mt-1 text-2xs text-muted-foreground">план 54</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Pending reports</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold tracking-tight">—</div>
            <p className="mt-1 text-2xs text-muted-foreground">план 48</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Dead events</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold tracking-tight">—</div>
            <p className="mt-1 text-2xs text-muted-foreground">план 53 Ф.6</p>
          </CardContent>
        </Card>
      </div>

      {claims && (
        <Card className="mt-4 max-w-2xl">
          <CardHeader>
            <CardTitle>Текущая сессия</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            <div>
              <div className="text-2xs uppercase tracking-wide text-muted-foreground">
                user
              </div>
              <div className="font-mono-sm">{claims.username}</div>
            </div>
            <div>
              <div className="text-2xs uppercase tracking-wide text-muted-foreground">
                roles ({claims.roles.length})
              </div>
              <div className="mt-1 flex flex-wrap gap-1.5">
                {claims.roles.length === 0 ? (
                  <span className="text-xs text-muted-foreground">— нет ролей —</span>
                ) : (
                  claims.roles.map((r) => <Badge key={r}>{r}</Badge>)
                )}
              </div>
            </div>
            <div>
              <div className="text-2xs uppercase tracking-wide text-muted-foreground">
                permissions ({claims.permissions.length})
              </div>
              {claims.permissions.length === 0 ? (
                <p className="mt-1 text-xs text-muted-foreground">— нет permissions —</p>
              ) : (
                <ul className="mt-1 grid grid-cols-2 gap-x-4 font-mono-sm text-xs">
                  {claims.permissions.map((p) => (
                    <li key={p}>{p}</li>
                  ))}
                </ul>
              )}
            </div>
          </CardContent>
        </Card>
      )}
    </>
  );
}
