// Auth flow: login / logout / me — через admin-bff endpoints.
//
// План 53 BFF: cookies (admin_session HttpOnly + admin_csrf для double-
// submit) ставит сам admin-bff. Frontend получает claims summary в
// JSON и кладёт в Zustand.
import { ApiError, apiRequest } from '@/lib/api/client';
import { useAuth, type AuthClaims } from '@/store/auth';

interface MeResponse extends AuthClaims {
  csrf_token: string;
}

export async function login(username: string, password: string): Promise<void> {
  const body = await apiRequest<MeResponse>('/auth/login', {
    method: 'POST',
    body: { username, password },
  });
  useAuth.getState().setSession(
    {
      sub: body.sub,
      username: body.username,
      roles: body.roles ?? [],
      permissions: body.permissions ?? [],
    },
    body.csrf_token,
  );
}

export async function logout(): Promise<void> {
  try {
    await apiRequest<void>('/auth/logout', { method: 'POST' });
  } catch (err) {
    // Logout best-effort — даже при ошибке локально чистим сессию.
    if (!(err instanceof ApiError)) throw err;
  } finally {
    useAuth.getState().clearSession();
  }
}

export async function fetchMe(): Promise<void> {
  try {
    const body = await apiRequest<MeResponse>('/auth/me');
    useAuth.getState().setSession(
      {
        sub: body.sub,
        username: body.username,
        roles: body.roles ?? [],
        permissions: body.permissions ?? [],
      },
      body.csrf_token,
    );
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      useAuth.getState().setAnonymous();
      return;
    }
    throw err;
  }
}
