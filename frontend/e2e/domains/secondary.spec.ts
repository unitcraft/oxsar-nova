// Ф.3: второстепенные экраны — smoke «открывается, не падает».

import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';
import { goToTab } from '../helpers/nav';

const SECONDARY_TABS = [
  'empire',
  'techtree',
  'battlestats',
  'records',
  'notepad',
  'referral',
  'friends',
  'settings',
  'planet-options',
  'resource',
] as const;

test.describe('Ф.3 secondary screens', () => {
  for (const tab of SECONDARY_TABS) {
    test(`${tab} opens for bob`, async ({ page }) => {
      await loginAs(page, 'bob');
      await goToTab(page, tab);
      await expect(page.locator('main.ox-content')).toBeVisible();
      // Сам факт рендера без Error boundary уже проверяется в smoke;
      // тут — что навигация на конкретный хэш работает.
      expect(new URL(page.url()).hash).toBe(`#${tab}`);
    });
  }
});

test.describe('Ф.3.12 global search (Ctrl+K)', () => {
  test('Ctrl+K opens the search dialog', async ({ page }) => {
    await loginAs(page, 'bob');
    await page.keyboard.press('Control+K');
    // Поле поиска становится видимым (любой input внутри overlay).
    await expect(page.getByPlaceholder(/поиск|search/i).first()).toBeVisible({ timeout: 5_000 });
  });
});

test.describe('Ф.3.13 admin', () => {
  test('admin tab visible for superadmin', async ({ page }) => {
    await loginAs(page, 'admin');
    await goToTab(page, 'admin');
    await expect(page.locator('main.ox-content')).toBeVisible();
    await expect(page.locator('main.ox-content')).toContainText(/админ|admin|stats|пользоват/i);
  });

  test('admin tab missing for regular player', async ({ page }) => {
    await loginAs(page, 'bob');
    // Попытка перейти на #admin — компонент скрыт условием {tab==='admin' && isAdmin}.
    await goToTab(page, 'admin');
    // main виден, но содержимого admin-экрана нет.
    await expect(page.locator('main.ox-content')).not.toContainText(/бан|role change/i);
  });
});
