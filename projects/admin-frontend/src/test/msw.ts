// MSW server для unit/integration тестов: моки admin-bff endpoints.
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';

interface LoginBody {
  username: string;
  password: string;
}

const validUser = {
  username: 'admin',
  password: 'correct-horse',
};

export const server = setupServer(
  http.post('/auth/login', async ({ request }) => {
    const body = (await request.json()) as LoginBody;
    if (body.username === validUser.username && body.password === validUser.password) {
      return HttpResponse.json({
        sub: 'user-uuid-1',
        username: 'admin',
        roles: ['admin', 'superadmin'],
        permissions: ['users:read', 'users:delete', 'roles:grant'],
        csrf_token: 'csrf-test-token',
      });
    }
    return HttpResponse.json(
      { error: 'invalid_credentials' },
      { status: 401 },
    );
  }),

  http.post('/auth/logout', () => {
    return new HttpResponse(null, { status: 204 });
  }),

  http.get('/auth/me', () => {
    return HttpResponse.json({
      sub: 'user-uuid-1',
      username: 'admin',
      roles: ['admin'],
      permissions: ['users:read'],
      csrf_token: 'csrf-test-token',
    });
  }),
);

export const meAnonymous = http.get('/auth/me', () =>
  HttpResponse.json({ error: 'no_session' }, { status: 401 }),
);
