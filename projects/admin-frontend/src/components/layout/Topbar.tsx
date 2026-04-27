// Topbar: 48px height, logo + title слева, user menu справа.
// Cmd+K placeholder в центре (план 53 Ф.10 даст полноценный command-palette).
import { useNavigate } from 'react-router-dom';
import { LogOut, User } from 'lucide-react';
import { useAuth } from '@/store/auth';
import { logout } from '@/lib/auth/flow';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

export function Topbar(): React.ReactElement {
  const claims = useAuth((s) => s.claims);
  const navigate = useNavigate();

  async function onLogout(): Promise<void> {
    await logout();
    navigate('/login', { replace: true });
  }

  return (
    <header className="flex h-topbar items-center justify-between border-b bg-card px-4">
      <div className="flex items-center gap-3">
        <span className="text-sm font-semibold">oxsar-nova</span>
        <Badge variant="secondary" className="text-2xs">admin</Badge>
      </div>

      <div className="flex flex-1 max-w-xl items-center px-6">
        <button
          type="button"
          disabled
          className="w-full text-left rounded-md border bg-background px-2.5 py-1 text-2xs text-muted-foreground"
          aria-label="Search (coming soon)"
        >
          <span>Поиск… </span>
          <kbd className="ml-1 rounded bg-muted px-1 py-0.5 font-mono-sm text-[10px]">⌘K</kbd>
        </button>
      </div>

      <div className="flex items-center gap-2">
        {claims && (
          <span className="flex items-center gap-1.5 text-xs text-muted-foreground">
            <User className="h-3.5 w-3.5" aria-hidden="true" />
            <span className="font-mono-sm">{claims.username}</span>
          </span>
        )}
        <Button variant="ghost" size="sm" onClick={onLogout} aria-label="Logout">
          <LogOut className="h-3.5 w-3.5" aria-hidden="true" />
          <span className="text-xs">Выйти</span>
        </Button>
      </div>
    </header>
  );
}
