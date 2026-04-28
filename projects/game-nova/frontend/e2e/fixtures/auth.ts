// Login-фикстура: дёргает backend /api/auth/login напрямую и кладёт токены
// в localStorage под ключом `oxsar-auth` (v=2) — в том же формате, что пишет
// Zustand-стор в src/stores/auth.ts. Так форма логина не прогоняется в каждом
// спеке, экономится время и тест не зависит от верстки login-экрана.

import { type APIRequestContext, type Page, expect } from '@playwright/test';

export const TEST_PASSWORD = 'test-password-123';

export const TEST_USERS = {
  admin: { username: 'admin', userId: '00000000-0000-0000-0000-000000000001' },
  alice: { username: 'alice', userId: '00000000-0000-0000-0000-000000000002' },
  bob: { username: 'bob', userId: '00000000-0000-0000-0000-000000000003' },
  eve: { username: 'eve', userId: '00000000-0000-0000-0000-000000000004' },
  charlie: { username: 'charlie', userId: '00000000-0000-0000-0000-000000000005' },
} as const;

export type TestUserName = keyof typeof TEST_USERS;

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080';

// План 63: identity отвечает по RFC 6749 §5.1 (плоский access_token / refresh_token).
interface AuthResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user: { id: string; username: string };
}

// Кешируем токены на весь прогон — backend имеет ratelimit на /api/auth/login
// (см. authRL в main.go). Без кеша 110 тестов × retries = 429.
const tokenCache = new Map<string, AuthResponse>();

async function fetchTokens(request: APIRequestContext, username: string): Promise<AuthResponse> {
  const cached = tokenCache.get(username);
  if (cached) return cached;

  // Backend принимает username в поле `email` (см. auth.Service.Login:
  // WHERE email=$1 OR lower(username)=$1).
  const res = await request.post(`${BACKEND_URL}/api/auth/login`, {
    data: { email: username, password: TEST_PASSWORD },
  });
  expect(res.ok(), `login ${username} failed (${res.status()})`).toBe(true);
  const data = (await res.json()) as AuthResponse;
  tokenCache.set(username, data);
  return data;
}

export async function loginAs(page: Page, user: TestUserName): Promise<void> {
  const resp = await fetchTokens(page.request, TEST_USERS[user].username);

  // Подменяем localStorage ДО первой загрузки SPA — тогда Zustand-стор
  // прочитает токены сразу в loadInitial().
  await page.addInitScript(
    ([access, refresh, userId]) => {
      localStorage.setItem(
        'oxsar-auth',
        JSON.stringify({ v: 2, access, refresh, userId }),
      );
    },
    [resp.access_token, resp.refresh_token, resp.user.id] as const,
  );

  await page.goto('/');
  // Ждём, пока приложение пройдёт загрузку.
  await expect(page.locator('.ox-header-logo')).toBeVisible({ timeout: 10_000 });
}
