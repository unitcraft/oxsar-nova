// Smoke: логинимся как bob (прокачанный игрок) и alice (новичок),
// открываем поочерёдно все вкладки. На каждой проверяем:
//   - нет console.error
//   - нет текста "Error boundary" / "Что-то пошло не так"
//   - expectNoLayoutIssues (Ф.4.6)
//
// Если smoke падает — остальные спеки не запускать.

import { test, expect } from '@playwright/test';
import { loginAs, type TestUserName } from './fixtures/auth';
import { expectNoLayoutIssues } from './helpers/layout';

// Вкладки из src/App.tsx. `admin` виден только под ролью admin/superadmin;
// `unit-info` требует параметр и тестируется отдельно.
const TABS = [
  'overview', 'buildings', 'research', 'shipyard', 'repair',
  'artefacts', 'galaxy', 'fleet', 'market', 'rockets',
  'art-market', 'officers', 'achievements', 'score',
  'messages', 'alliance', 'chat', 'sim',
  'planet-options', 'resource', 'credits',
  'profession', 'empire', 'settings', 'referral',
  'notepad', 'techtree', 'battlestats', 'friends', 'records',
] as const;

const ERROR_PATTERNS = [
  /error boundary/i,
  /что-то пошло не так/i,
  /something went wrong/i,
];

function setupErrorMonitors(page: import('@playwright/test').Page): { errors: string[] } {
  const errors: string[] = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      const text = msg.text();
      // Игнорируем ожидаемые warnings от dev-режима и третьесторонних либ.
      if (text.includes('Download the React DevTools')) return;
      if (text.includes('Warning:')) return; // deprecation warnings React
      errors.push(text);
    }
  });
  page.on('pageerror', (err) => {
    errors.push(`[pageerror] ${err.message}`);
  });
  return { errors };
}

async function visitTab(
  page: import('@playwright/test').Page,
  tab: string,
  errors: string[],
): Promise<void> {
  await page.goto(`/#${tab}`);
  // Даём React + Suspense прогрузить lazy-chunk и отрендерить.
  await page.waitForLoadState('networkidle', { timeout: 10_000 }).catch(() => {
    // networkidle может не наступать из-за WS/poll — это ок.
  });
  await page.waitForTimeout(500);

  for (const re of ERROR_PATTERNS) {
    await expect
      .soft(page.locator('body'), `tab "${tab}": error text on screen`)
      .not.toHaveText(re);
  }
  await expectNoLayoutIssues(page, tab);

  // Копим — репортим в конце теста, чтобы видеть все упавшие вкладки разом.
  if (errors.length > 0) {
    console.warn(`[${tab}] console errors:\n  ${errors.join('\n  ')}`);
  }
}

function smokeForUser(userName: TestUserName): void {
  test(`smoke: all tabs render for ${userName}`, async ({ page }) => {
    const { errors } = setupErrorMonitors(page);
    await loginAs(page, userName);

    for (const tab of TABS) {
      const before = errors.length;
      await visitTab(page, tab, errors);
      const added = errors.slice(before);
      expect.soft(added, `tab "${tab}" produced console errors`).toEqual([]);
    }

    // Фейлим тест, если были soft-assertion'ы.
    expect(test.info().errors).toEqual([]);
  });
}

test.describe('smoke', () => {
  smokeForUser('bob');
  smokeForUser('alice');
});
