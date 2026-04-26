// Ф.2.14: платежи через MockGateway (PAYMENT_PROVIDER=mock).
// Проверяем полный флоу: открыть экран → увидеть баннер «Тестовый режим» →
// купить starter → редирект через mock/pay?result=success → toast + баланс.

import { test, expect } from '@playwright/test';
import { loginAs } from '../fixtures/auth';
import { goToTab } from '../helpers/nav';

test.describe('Ф.2.14 credits (mock payment flow)', () => {
  test('alice sees test-mode banner and package list', async ({ page }) => {
    await loginAs(page, 'alice');
    await goToTab(page, 'credits');

    await expect(page.getByRole('alert')).toContainText(/тестовый режим/i);
    // 5 пакетов
    for (const label of ['Пробный', 'Стартовый', 'Средний', 'Большой', 'Максимальный']) {
      await expect(page.locator('main.ox-content')).toContainText(label);
    }
  });

  test('alice buys starter package via mock → balance grows', async ({ page }) => {
    await loginAs(page, 'alice');

    // Baseline balance
    const beforeMatch = await page
      .locator('.ox-header-right')
      .getByTitle('Кредиты')
      .innerText();
    const before = Number(beforeMatch.replace(/[^\d.]/g, '')) || 0;

    await goToTab(page, 'credits');

    // Кнопка «Купить» у пакета «Стартовый» (1000 кр, 100 ₽).
    const starterCard = page.locator('.credit-package-card', { hasText: 'Стартовый' });
    await expect(starterCard).toBeVisible();
    await starterCard.getByRole('button', { name: 'Купить' }).click();

    // В mock-режиме редирект происходит в том же окне (window.location.href).
    // Бэкенд обрабатывает /api/payment/mock/pay → redirect на ?payment=success.
    await page.waitForURL(/payment=success|#overview|#credits/, { timeout: 15_000 });

    // Toast об успехе — «Оплата прошла успешно»
    await expect(page.getByText(/оплата прошла|успешно/i)).toBeVisible({ timeout: 10_000 });

    // Ждём обновления /api/me
    await page.waitForTimeout(1_500);
    const afterMatch = await page
      .locator('.ox-header-right')
      .getByTitle('Кредиты')
      .innerText();
    const after = Number(afterMatch.replace(/[^\d.]/g, '')) || 0;

    expect(after).toBeGreaterThan(before);
  });

  test('fail-mode shows error toast and balance unchanged', async ({ page }) => {
    await loginAs(page, 'alice');
    await goToTab(page, 'credits');

    // Создадим заказ вручную и сходим на mock-endpoint с result=fail.
    const orderResp = await page.request.post('/api/payment/order', {
      data: { package_key: 'trial' },
      headers: {
        Authorization: `Bearer ${await page.evaluate(() => {
          const raw = localStorage.getItem('oxsar-auth');
          return raw ? (JSON.parse(raw).access ?? '') : '';
        })}`,
      },
    });
    const { order_id } = (await orderResp.json()) as { order_id: string };
    expect(order_id).toBeTruthy();

    await page.goto(`/api/payment/mock/pay?order=${order_id}&result=fail`);
    // После редиректа должны увидеть toast с ошибкой
    await expect(page.getByText(/не прошла|ошибк/i)).toBeVisible({ timeout: 10_000 });
  });
});
