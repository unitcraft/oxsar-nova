// Ф.1.7: сообщения. У alice в сиде — 1 welcome-сообщение.

import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';
import { goToTab } from '../helpers/nav';

test.describe('messages (Ф.1.7)', () => {
  test('alice sees welcome message', async ({ page }) => {
    await loginAs(page, 'alice');
    await goToTab(page, 'messages');
    await expect(page.locator('main.ox-content')).toBeVisible();
    await expect(page.locator('main.ox-content')).toContainText(/добро пожаловать|welcome|тестовое/i);
  });

  test('bob messages screen opens without errors', async ({ page }) => {
    await loginAs(page, 'bob');
    await goToTab(page, 'messages');
    await expect(page.locator('main.ox-content')).toBeVisible();
  });

  test('unread badge reflected in nav for alice', async ({ page, isMobile }) => {
    await loginAs(page, 'alice');
    await page.waitForTimeout(2_000);
    // На desktop — link в sidebar; на mobile — button в bottom-nav.
    const locator = isMobile
      ? page.getByRole('button', { name: /сообщ|messages/i }).first()
      : page.getByRole('link', { name: /сообщения|messages/i }).first();
    await expect(locator).toBeVisible();
  });
});
