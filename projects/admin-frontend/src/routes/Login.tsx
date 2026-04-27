import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { z } from 'zod';
import { ApiError } from '@/lib/api/client';
import { login } from '@/lib/auth/flow';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

const loginSchema = z.object({
  username: z.string().min(1, 'обязательное поле').max(64),
  password: z.string().min(1, 'обязательное поле').max(128),
});

interface FormState {
  username: string;
  password: string;
  error: string | null;
  fieldErrors: Partial<Record<'username' | 'password', string>>;
  submitting: boolean;
}

const initialState: FormState = {
  username: '',
  password: '',
  error: null,
  fieldErrors: {},
  submitting: false,
};

export function Login(): React.ReactElement {
  const navigate = useNavigate();
  const location = useLocation();
  const [state, setState] = useState<FormState>(initialState);

  function update<K extends keyof FormState>(key: K, value: FormState[K]): void {
    setState((s) => ({ ...s, [key]: value }));
  }

  async function onSubmit(e: React.FormEvent): Promise<void> {
    e.preventDefault();
    const parsed = loginSchema.safeParse({
      username: state.username,
      password: state.password,
    });
    if (!parsed.success) {
      const fieldErrors: FormState['fieldErrors'] = {};
      for (const issue of parsed.error.issues) {
        const key = issue.path[0];
        if (key === 'username' || key === 'password') {
          fieldErrors[key] = issue.message;
        }
      }
      setState((s) => ({ ...s, fieldErrors, error: null }));
      return;
    }

    setState((s) => ({ ...s, submitting: true, error: null, fieldErrors: {} }));
    try {
      await login(parsed.data.username, parsed.data.password);
      const fallback = '/';
      const from = (location.state as { from?: string } | null)?.from ?? fallback;
      navigate(from, { replace: true });
    } catch (err) {
      let msg = 'Ошибка входа';
      if (err instanceof ApiError) {
        if (err.status === 401) {
          msg = 'Неверный логин или пароль';
        } else if (err.status === 502) {
          msg = 'Identity-сервис недоступен';
        } else {
          msg = err.body?.message || err.body?.error || `HTTP ${err.status}`;
        }
      }
      setState((s) => ({ ...s, error: msg, submitting: false }));
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="w-full max-w-sm">
        <CardHeader>
          <div className="flex flex-col gap-0.5">
            <h1 className="text-base font-semibold text-foreground">oxsar-nova admin</h1>
            <CardTitle>вход</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} className="space-y-3" noValidate>
            <div className="space-y-1">
              <label htmlFor="username" className="text-xs text-muted-foreground">
                Логин
              </label>
              <Input
                id="username"
                type="text"
                autoComplete="username"
                autoFocus
                value={state.username}
                onChange={(e) => update('username', e.target.value)}
                disabled={state.submitting}
                aria-invalid={Boolean(state.fieldErrors.username)}
              />
              {state.fieldErrors.username && (
                <p className="text-2xs text-destructive">{state.fieldErrors.username}</p>
              )}
            </div>
            <div className="space-y-1">
              <label htmlFor="password" className="text-xs text-muted-foreground">
                Пароль
              </label>
              <Input
                id="password"
                type="password"
                autoComplete="current-password"
                value={state.password}
                onChange={(e) => update('password', e.target.value)}
                disabled={state.submitting}
                aria-invalid={Boolean(state.fieldErrors.password)}
              />
              {state.fieldErrors.password && (
                <p className="text-2xs text-destructive">{state.fieldErrors.password}</p>
              )}
            </div>
            {state.error && (
              <div
                role="alert"
                className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive"
              >
                {state.error}
              </div>
            )}
            <Button type="submit" disabled={state.submitting} className="w-full">
              {state.submitting ? 'Вход…' : 'Войти'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
