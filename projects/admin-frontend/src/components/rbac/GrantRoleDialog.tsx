// Grant Role Dialog: select role + reason (required) + optional expires_at.
// Audit log (план 52) требует обязательный reason — поэтому не пустой trim.
import { useState, useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { rbacApi, type GrantRoleInput } from '@/lib/api/rbac';
import { ApiError } from '@/lib/api/client';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';

const grantSchema = z.object({
  role: z.string().min(1, 'выберите роль'),
  reason: z
    .string()
    .trim()
    .min(3, 'минимум 3 символа — обязательно для audit'),
  expires_at: z.string().optional(),
});

interface GrantRoleDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  userId: string;
  existingRoles: string[];
}

interface FormState {
  role: string;
  reason: string;
  expiresAt: string;
  fieldErrors: Partial<Record<'role' | 'reason' | 'expires_at', string>>;
  apiError: string | null;
}

const initialState: FormState = {
  role: '',
  reason: '',
  expiresAt: '',
  fieldErrors: {},
  apiError: null,
};

export function GrantRoleDialog({
  open,
  onOpenChange,
  userId,
  existingRoles,
}: GrantRoleDialogProps): React.ReactElement {
  const [state, setState] = useState<FormState>(initialState);
  const queryClient = useQueryClient();

  // Reset при закрытии — чтобы при следующем открытии форма была чистой.
  useEffect(() => {
    if (!open) setState(initialState);
  }, [open]);

  const rolesQuery = useQuery({
    queryKey: ['rbac', 'roles'],
    queryFn: rbacApi.listRoles,
    staleTime: 5 * 60_000,
    enabled: open,
  });

  const grantMutation = useMutation({
    mutationFn: (input: GrantRoleInput) => rbacApi.grantUserRole(userId, input),
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: ['rbac', 'users', userId, 'roles'],
      });
      void queryClient.invalidateQueries({ queryKey: ['rbac', 'audit'] });
      onOpenChange(false);
    },
    onError: (err) => {
      let msg = 'Ошибка';
      if (err instanceof ApiError) {
        msg = err.body?.message || err.body?.error || `HTTP ${err.status}`;
      } else if (err instanceof Error) {
        msg = err.message;
      }
      setState((s) => ({ ...s, apiError: msg }));
    },
  });

  function update<K extends keyof FormState>(key: K, value: FormState[K]): void {
    setState((s) => ({ ...s, [key]: value }));
  }

  function onSubmit(e: React.FormEvent): void {
    e.preventDefault();
    const parsed = grantSchema.safeParse({
      role: state.role,
      reason: state.reason,
      expires_at: state.expiresAt || undefined,
    });
    if (!parsed.success) {
      const fieldErrors: FormState['fieldErrors'] = {};
      for (const issue of parsed.error.issues) {
        const k = issue.path[0];
        if (k === 'role' || k === 'reason' || k === 'expires_at') {
          fieldErrors[k] = issue.message;
        }
      }
      setState((s) => ({ ...s, fieldErrors, apiError: null }));
      return;
    }
    setState((s) => ({ ...s, fieldErrors: {}, apiError: null }));
    const payload: GrantRoleInput = {
      role: parsed.data.role,
      reason: parsed.data.reason,
    };
    if (parsed.data.expires_at) {
      // datetime-local → ISO. Если уже ISO — оставляем.
      const iso = parsed.data.expires_at.includes('T')
        ? new Date(parsed.data.expires_at).toISOString()
        : parsed.data.expires_at;
      payload.expires_at = iso;
    }
    grantMutation.mutate(payload);
  }

  const availableRoles = (rolesQuery.data ?? []).filter(
    (r) => !existingRoles.includes(r.name),
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Grant role</DialogTitle>
          <DialogDescription>
            Назначение роли пользователю записывается в immutable audit log
            (план 52). Reason обязателен.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={onSubmit} className="space-y-3" noValidate>
          <div className="space-y-1">
            <label htmlFor="role" className="text-xs text-muted-foreground">
              Role
            </label>
            <select
              id="role"
              value={state.role}
              onChange={(e) => update('role', e.target.value)}
              disabled={grantMutation.isPending || rolesQuery.isLoading}
              className="flex h-8 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-sm focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:opacity-50"
            >
              <option value="">— выберите роль —</option>
              {availableRoles.map((r) => (
                <option key={r.id} value={r.name}>
                  {r.name}
                  {r.description ? ` — ${r.description}` : ''}
                </option>
              ))}
            </select>
            {state.fieldErrors.role && (
              <p className="text-2xs text-destructive">{state.fieldErrors.role}</p>
            )}
          </div>

          <div className="space-y-1">
            <label htmlFor="reason" className="text-xs text-muted-foreground">
              Reason (audit)
            </label>
            <Input
              id="reason"
              value={state.reason}
              onChange={(e) => update('reason', e.target.value)}
              disabled={grantMutation.isPending}
              placeholder="напр. promotion to support team"
              maxLength={500}
              aria-invalid={Boolean(state.fieldErrors.reason)}
            />
            {state.fieldErrors.reason && (
              <p className="text-2xs text-destructive">{state.fieldErrors.reason}</p>
            )}
          </div>

          <div className="space-y-1">
            <label htmlFor="expiresAt" className="text-xs text-muted-foreground">
              Expires at (опционально)
            </label>
            <Input
              id="expiresAt"
              type="datetime-local"
              value={state.expiresAt}
              onChange={(e) => update('expiresAt', e.target.value)}
              disabled={grantMutation.isPending}
            />
          </div>

          {state.apiError && (
            <div
              role="alert"
              className="rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive"
            >
              {state.apiError}
            </div>
          )}

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() => onOpenChange(false)}
              disabled={grantMutation.isPending}
            >
              Отмена
            </Button>
            <Button type="submit" size="sm" disabled={grantMutation.isPending}>
              {grantMutation.isPending ? 'Назначение…' : 'Grant'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
