// Ф.1.5–1.6: Galaxy + Fleet — космическая часть.
// В сиде у bob есть eve рядом (1:1:9), своя планета (1:1:7), charlie (1:2:7).

import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';
import { goToTab } from '../helpers/nav';

test.describe('galaxy (Ф.1.5)', () => {
  test('bob opens galaxy view', async ({ page }) => {
    await loginAs(page, 'bob');
    await goToTab(page, 'galaxy');
    await expect(page.locator('main.ox-content')).toBeVisible();
    // В галактике показываются координаты вида "1:1:1" и т.п.
    await expect(page.locator('main.ox-content')).toContainText(/\d+\s*:\s*\d+/);
  });

  test('galaxy shows own planet marker for bob', async ({ page }) => {
    await loginAs(page, 'bob');
    await goToTab(page, 'galaxy');
    // На экране должны быть названия планет или никнеймы — bob или его ник
    await expect(page.locator('main.ox-content')).toContainText(/bob|eve|Bob-Home|Eve/i);
  });
});

test.describe('fleet (Ф.1.6)', () => {
  test('bob sees fleet screen', async ({ page }) => {
    await loginAs(page, 'bob');
    await goToTab(page, 'fleet');
    await expect(page.locator('main.ox-content')).toBeVisible();
    // Форма отправки миссий — где-то должно быть "миссия" или выбор кораблей.
    await expect(page.locator('main.ox-content')).toContainText(/флот|миссия|fleet|mission/i);
  });

  test('alice has no fleet — empty state is ok', async ({ page }) => {
    await loginAs(page, 'alice');
    await goToTab(page, 'fleet');
    await expect(page.locator('main.ox-content')).toBeVisible();
  });
});
