// Ф.1.2–1.4 + Ф.1.8: «ядро планеты» — Overview, Buildings, Research, Shipyard.
// Каждая вкладка открывается, показывает ожидаемый UI, не падает.

import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';
import { goToTab } from '../helpers/nav';

test.describe('overview (Ф.1.8)', () => {
  test('bob sees resources and planet info', async ({ page }) => {
    await loginAs(page, 'bob');
    await goToTab(page, 'overview');

    // Ресурсы в шапке
    await expect(page.locator('.ox-header-resources')).toBeVisible();
    // Baza planet switcher
    await expect(page.locator('.ox-planet-switcher')).toBeVisible();
    // Имя планеты bob
    await expect(page.locator('.ox-planet-switcher')).toContainText('Bob-Home');
  });

  test('alice sees energy warning when under-powered or just empty state', async ({ page }) => {
    await loginAs(page, 'alice');
    await goToTab(page, 'overview');
    await expect(page.locator('main.ox-content')).toBeVisible();
  });
});

test.describe('buildings (Ф.1.2)', () => {
  test('alice sees building list', async ({ page }) => {
    await loginAs(page, 'alice');
    await goToTab(page, 'buildings');

    // Заголовок / карточки зданий
    await expect(page.locator('main.ox-content')).toBeVisible();
    // metal_mine — один из стартовых
    await expect(page.locator('main.ox-content')).toContainText(/металл|metal/i);
  });

  test('bob sees upgraded buildings', async ({ page }) => {
    await loginAs(page, 'bob');
    await goToTab(page, 'buildings');
    // У bob metal_mine уровня 20 — где-то должно быть число 20 или "Ур. 20"
    await expect(page.locator('main.ox-content')).toContainText(/20/);
  });
});

test.describe('research (Ф.1.3)', () => {
  test('bob sees research list', async ({ page }) => {
    await loginAs(page, 'bob');
    await goToTab(page, 'research');
    await expect(page.locator('main.ox-content')).toBeVisible();
    await expect(page.locator('main.ox-content')).toContainText(/исслед|research|компьютер/i);
  });
});

test.describe('shipyard (Ф.1.4)', () => {
  test('bob sees shipyard list with ships', async ({ page }) => {
    await loginAs(page, 'bob');
    await goToTab(page, 'shipyard');
    await expect(page.locator('main.ox-content')).toBeVisible();
    // У bob есть флот — должны увидеть числа кораблей (100 легких истребителей).
    await expect(page.locator('main.ox-content')).toContainText(/корабл|истреб|fighter|ship/i);
  });
});
