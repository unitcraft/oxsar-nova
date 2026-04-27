// Поиск юзера по UUID. Полноценный список ждёт sub-плана 53-users-list
// (нужен endpoint в identity или aggregator в admin-bff).
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Search } from 'lucide-react';
import { z } from 'zod';
import { PageHeader } from '@/components/layout/PageHeader';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';

const uuidSchema = z.string().regex(
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
  'некорректный UUID',
);

export function UsersLookup(): React.ReactElement {
  const [value, setValue] = useState('');
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  function onSubmit(e: React.FormEvent): void {
    e.preventDefault();
    const trimmed = value.trim();
    const parsed = uuidSchema.safeParse(trimmed);
    if (!parsed.success) {
      setError(parsed.error.issues[0]?.message ?? 'некорректный UUID');
      return;
    }
    setError(null);
    navigate(`/users/${trimmed}`);
  }

  return (
    <>
      <PageHeader
        title="Users"
        description="поиск юзера по UUID — полноценный список ждёт расширения identity API"
      />
      <Card className="max-w-xl">
        <CardContent className="pt-4">
          <form onSubmit={onSubmit} className="space-y-3" noValidate>
            <div className="space-y-1">
              <label htmlFor="user-uuid" className="text-xs text-muted-foreground">
                User UUID
              </label>
              <div className="flex gap-2">
                <Input
                  id="user-uuid"
                  value={value}
                  onChange={(e) => setValue(e.target.value)}
                  placeholder="00000000-0000-0000-0000-000000000000"
                  className="font-mono-sm"
                  autoFocus
                  aria-invalid={Boolean(error)}
                />
                <Button type="submit" size="sm">
                  <Search className="h-3.5 w-3.5" aria-hidden="true" />
                  Найти
                </Button>
              </div>
              {error && <p className="text-2xs text-destructive">{error}</p>}
            </div>
          </form>
          <p className="mt-3 text-2xs text-muted-foreground">
            Полноценный список юзеров с фильтрами и поиском по
            username/email — отдельный sub-план (53-users-list), требует
            нового endpoint в identity.
          </p>
        </CardContent>
      </Card>
    </>
  );
}
