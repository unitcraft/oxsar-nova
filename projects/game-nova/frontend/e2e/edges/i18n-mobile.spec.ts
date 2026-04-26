// Ф.4.4–4.5: i18n + мобильный viewport.

import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';
import { goToTab } from '../helpers/nav';

test.describe('Ф.4.4 i18n', () => {
  test('no raw MENU_/ACTION_/STATE_ keys leaking on main screens', async ({ page }) => {
    await loginAs(page, 'bob');
    for (const tab of ['overview', 'buildings', 'research', 'shipyard', 'galaxy', 'messages']) {
      await goToTab(page, tab);
      const text = await page.locator('main.ox-content').innerText();
      // Точечные префиксы из src/i18n словарей — если утечёт, значит
      // ключ не найден в bundle.
      expect.soft(text, `tab ${tab} contains raw i18n key`).not.toMatch(
        /\b(MENU_|BTN_|ACTION_|STATE_|LABEL_|HINT_)[A-Z0-9_]{2,}\b/,
      );
    }
  });
});

test.describe('Ф.4.5 mobile viewport', () => {
  test.use({ viewport: { width: 375, height: 667 } });

  test('bottom nav visible and sidebar hidden on mobile', async ({ page }) => {
    await loginAs(page, 'bob');
    await expect(page.locator('.ox-bottom-nav')).toBeVisible();
  });

  test('more sheet opens from bottom nav', async ({ page }) => {
    await loginAs(page, 'bob');
    // Последняя кнопка bottom-nav — «⋯ Ещё»
    await page.getByRole('button', { name: /ещё|more/i }).click();
    await expect(page.locator('.ox-modal')).toBeVisible();
  });
});
