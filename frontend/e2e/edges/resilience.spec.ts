// Ф.4.1–4.2: сетевые ошибки и пустые состояния.

import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';
import { goToTab } from '../helpers/nav';

test.describe('Ф.4.1 network errors', () => {
  test('500 from server — user sees something, not a white screen', async ({ page }) => {
    await loginAs(page, 'bob');
    // Интерсепт конкретный endpoint
    await page.route('**/api/planets/*/buildings', (route) =>
      route.fulfill({ status: 500, body: JSON.stringify({ error: { code: 'internal', message: 'oops' } }) }),
    );
    await goToTab(page, 'buildings');
    // main должен быть виден, даже если данные не пришли
    await expect(page.locator('main.ox-content')).toBeVisible();
  });

  test('401 mid-session triggers logout', async ({ page }) => {
    await loginAs(page, 'bob');
    await page.route('**/api/planets', (route) =>
      route.fulfill({
        status: 401,
        body: JSON.stringify({ error: { code: 'unauthorized', message: 'expired' } }),
      }),
    );
    // Любой обновляющий запрос вызовет 401 → логаут
    await page.goto('/');
    await expect(page.locator('form button[type="submit"]')).toBeVisible({ timeout: 10_000 });
  });
});

test.describe('Ф.4.2 empty states', () => {
  test('alice: inbox is not empty (has welcome), but achievements likely empty', async ({ page }) => {
    await loginAs(page, 'alice');
    await goToTab(page, 'achievements');
    await expect(page.locator('main.ox-content')).toBeVisible();
  });

  test('alice: no artefacts — empty state is friendly', async ({ page }) => {
    await loginAs(page, 'alice');
    await goToTab(page, 'artefacts');
    await expect(page.locator('main.ox-content')).toBeVisible();
    // Не падает и не показывает "undefined"
    await expect(page.locator('main.ox-content')).not.toContainText(/undefined|null|NaN/);
  });
});
