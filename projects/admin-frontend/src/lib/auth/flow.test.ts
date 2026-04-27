import { describe, expect, it, beforeEach } from 'vitest';
import { server, meAnonymous } from '@/test/msw';
import { useAuth } from '@/store/auth';
import { ApiError } from '@/lib/api/client';
import { fetchMe, login, logout } from './flow';

beforeEach(() => {
  useAuth.setState({ status: 'unknown', claims: null, csrfToken: null });
});

describe('login', () => {
  it('successful login populates auth store', async () => {
    await login('admin', 'correct-horse');
    const s = useAuth.getState();
    expect(s.status).toBe('authenticated');
    expect(s.claims?.username).toBe('admin');
    expect(s.claims?.permissions).toContain('users:delete');
    expect(s.csrfToken).toBe('csrf-test-token');
  });

  it('401 throws ApiError and forces anonymous state (no claims set)', async () => {
    await expect(login('admin', 'wrong')).rejects.toBeInstanceOf(ApiError);
    // ApiError на 401 в apiRequest вызывает clearSession() → 'anonymous'.
    expect(useAuth.getState().status).toBe('anonymous');
    expect(useAuth.getState().claims).toBeNull();
  });
});

describe('fetchMe', () => {
  it('authenticated → fills claims', async () => {
    await fetchMe();
    expect(useAuth.getState().status).toBe('authenticated');
  });

  it('401 from /auth/me → setAnonymous, no throw', async () => {
    server.use(meAnonymous);
    await fetchMe();
    expect(useAuth.getState().status).toBe('anonymous');
    expect(useAuth.getState().claims).toBeNull();
  });
});

describe('logout', () => {
  it('clears session even if API succeeds', async () => {
    useAuth.getState().setSession(
      { sub: 'x', username: 'x', roles: [], permissions: [] },
      'csrf',
    );
    await logout();
    expect(useAuth.getState().status).toBe('anonymous');
    expect(useAuth.getState().csrfToken).toBeNull();
  });
});
