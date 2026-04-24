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

interface AuthResponse {
  user: { id: string; username: string };
  tokens: { access: string; refresh: string };
}

async function fetchTokens(request: APIRequestContext, username: string): Promise<AuthResponse> {
  const res = await request.post(`${BACKEND_URL}/api/auth/login`, {
    data: { username, password: TEST_PASSWORD },
  });
  expect(res.ok(), `login ${username} failed (${res.status()})`).toBe(true);
  return (await res.json()) as AuthResponse;
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
    [resp.tokens.access, resp.tokens.refresh, resp.user.id] as const,
  );

  await page.goto('/');
  // Ждём, пока приложение пройдёт загрузку.
  await expect(page.locator('.ox-header-logo')).toBeVisible({ timeout: 10_000 });
}
