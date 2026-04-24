// Ф.1.1: авторизация. Регистрация, логин, неверный пароль, logout.
// Refresh-токен не тестируем напрямую (он автоматический) — это часть
// smoke-спека: токен подставляется через fixture и работает.

import { test, expect } from '@playwright/test';
import { loginAs, TEST_PASSWORD } from '../fixtures/auth';

test.describe('auth: login screen', () => {
  test('shows login form by default', async ({ page }) => {
    await page.goto('/');
    // Логотип "OXSAR" в центре экрана. `.first()` — страховка от
    // повторов, если кто-то добавит ещё один.
    await expect(page.getByText('OXSAR').first()).toBeVisible();
    // На login-экране submit-кнопка имеет type=submit. Это отличает её
    // от таба «Войти» (type=button).
    await expect(page.locator('form button[type="submit"]')).toBeVisible();
  });

  test('login with valid credentials lands on Overview', async ({ page }) => {
    await page.goto('/');
    await page.getByLabel(/e-?mail|логин/i).first().fill('alice');
    await page.getByLabel('Пароль').fill(TEST_PASSWORD);
    await page.locator('form button[type="submit"]').click();

    await expect(page.locator('.ox-header-logo')).toBeVisible({ timeout: 15_000 });
  });

  test('wrong password shows error and stays on login', async ({ page }) => {
    await page.goto('/');
    await page.getByLabel(/e-?mail|логин/i).first().fill('alice');
    await page.getByLabel('Пароль').fill('wrong-password-999');
    await page.locator('form button[type="submit"]').click();

    await expect(page.locator('.ox-error')).toBeVisible({ timeout: 10_000 });
    await expect(page.locator('.ox-header-logo')).not.toBeVisible();
  });
});

test.describe('auth: session', () => {
  test('logout returns to login screen', async ({ page }) => {
    await loginAs(page, 'alice');
    await page.getByRole('button', { name: 'Выйти' }).click();
    // После logout login-форма — submit-кнопка уникальна, таб тоже,
    // но safer селектор — форма.
    await expect(page.locator('form button[type="submit"]')).toBeVisible({ timeout: 10_000 });
  });

  test('me endpoint returns session user', async ({ page }) => {
    await loginAs(page, 'bob');
    // В правой части шапки есть маленький серый span с именем юзера.
    // Ограничиваем селектор на прямого потомка с text=bob (exact),
    // чтобы не зацепить Bob-Home.
    await expect(page.locator('.ox-header-right span').filter({ hasText: /^bob$/ }).first()).toBeVisible();
  });
});
