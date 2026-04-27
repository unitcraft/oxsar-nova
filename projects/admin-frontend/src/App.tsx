import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

export function App(): React.ReactElement {
  return (
    <div className="min-h-screen p-8">
      <header className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold">oxsar-nova admin</h1>
          <p className="text-xs text-muted-foreground">
            план 53 — admin-консоль (Ф.1 скаффолдинг)
          </p>
        </div>
        <Badge variant="warning">dev preview</Badge>
      </header>

      <Card className="max-w-xl">
        <CardHeader>
          <CardTitle>Stack</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          <p>React 18 + TypeScript strict + Vite + Tailwind + shadcn/ui.</p>
          <p>
            <span className="text-muted-foreground">Auth:</span> OAuth2 PKCE (план 53 Ф.2).
          </p>
          <p>
            <span className="text-muted-foreground">RBAC:</span> permissions из JWT (план 52).
          </p>
          <div className="pt-2">
            <Button variant="outline" size="sm">
              dummy action
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
