// Ф.1.1: авторизация. Регистрация, логин, неверный пароль, logout.
// Refresh-токен не тестируем напрямую (он автоматический) — это часть
// smoke-спека: токен подставляется через fixture и работает.

import { test, expect } from '@playwright/test';
import { loginAs, TEST_PASSWORD } from '../fixtures/auth';

test.describe('auth: login screen', () => {
  test('shows login form by default', async ({ page }) => {
    await page.goto('/');
    await expect(page.getByText('OXSAR').first()).toBeVisible();
    await expect(page.getByRole('button', { name: 'Войти', exact: true })).toBeVisible();
  });

  test('login with valid credentials lands on Overview', async ({ page }) => {
    await page.goto('/');
    await page.getByRole('button', { name: 'Войти', exact: true }).first().click();

    // Поле email принимает username тоже
    await page.getByLabel(/e-?mail|логин/i).first().fill('alice');
    await page.getByLabel('Пароль').fill(TEST_PASSWORD);
    await page.getByRole('button', { name: 'Войти', exact: true }).last().click();

    // После логина появляется шапка с логотипом в sidebar-layout.
    await expect(page.locator('.ox-header-logo')).toBeVisible({ timeout: 15_000 });
  });

  test('wrong password shows error and stays on login', async ({ page }) => {
    await page.goto('/');
    await page.getByLabel(/e-?mail|логин/i).first().fill('alice');
    await page.getByLabel('Пароль').fill('wrong-password-999');
    await page.getByRole('button', { name: 'Войти', exact: true }).last().click();

    await expect(page.locator('.ox-error')).toBeVisible({ timeout: 10_000 });
    // Остались на login-экране
    await expect(page.locator('.ox-header-logo')).not.toBeVisible();
  });
});

test.describe('auth: session', () => {
  test('logout returns to login screen', async ({ page }) => {
    await loginAs(page, 'alice');
    await page.getByRole('button', { name: 'Выйти' }).click();
    await expect(page.getByRole('button', { name: 'Войти', exact: true })).toBeVisible({ timeout: 10_000 });
  });

  test('me endpoint returns session user', async ({ page }) => {
    await loginAs(page, 'bob');
    // Имя пользователя отображается в header
    await expect(page.locator('.ox-header-right').getByText('bob')).toBeVisible();
  });
});
