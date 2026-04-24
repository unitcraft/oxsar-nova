// Навигация между вкладками в SPA. Используем hash-роутинг (#tab), чтобы
// не зависеть от состояния sidebar'а (на mobile он скрыт за bottom-nav).

import { type Page, expect } from '@playwright/test';

export async function goToTab(page: Page, tab: string): Promise<void> {
  await page.goto(`/#${tab}`);
  // networkidle часто не наступает из-за WebSocket/polling — короткий таймаут
  // + затем ждём DOM-элемент. Не тратим время на полноценный idle.
  await page.waitForLoadState('networkidle', { timeout: 2_000 }).catch(() => {});
  // 20s чтобы пережить медленный vite dev-transform на mobile-project'е
  // (первая загрузка lazy-chunk'а может быть ~5-15 секунд).
  await expect(page.locator('main.ox-content')).toBeVisible({ timeout: 20_000 });
}
