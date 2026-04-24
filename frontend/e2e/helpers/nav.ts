// Навигация между вкладками в SPA. Используем hash-роутинг (#tab), чтобы
// не зависеть от состояния sidebar'а (на mobile он скрыт за bottom-nav).

import { type Page, expect } from '@playwright/test';

export async function goToTab(page: Page, tab: string): Promise<void> {
  await page.goto(`/#${tab}`);
  // Ждём, пока Suspense-fallback исчезнет (любой скелетон в main-контент).
  await page.waitForLoadState('networkidle', { timeout: 10_000 }).catch(() => {});
  await page.waitForTimeout(300);
  await expect(page.locator('main.ox-content')).toBeVisible();
}
