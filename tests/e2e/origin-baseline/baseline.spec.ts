import { test, expect, type BrowserContext, type Page } from '@playwright/test';
import * as path from 'node:path';
import { SCREENS, SMOKE_SCREEN_IDS, type Screen } from './screens';

/**
 * baseline.spec.ts — снятие эталонных скриншотов с running legacy-php
 * (порт :8092 по умолчанию). Часть плана 73 Ф.1+Ф.2.
 *
 * Перед запуском:
 *   - legacy-php стек должен быть поднят (см. take-screenshots.sh).
 *   - Dev-логин активен (`/dev-login.php` ставит JWT-cookie для userid=1).
 *
 * Поведение:
 *   - SMOKE=1 (default) — снимает только SMOKE_SCREEN_IDS (7 экранов).
 *   - SMOKE=0 — снимает все SCREENS (22 экрана).
 *   - PNG сохраняются в `screenshots/<id>-<name>.png`.
 *
 * Маски (план 73 §«Конвенции»): countdown-блоки в шапке/очередях
 * меняются между запусками. На уровне snapshot маскируем ничем —
 * пиксел-дифф (Ф.3) применит маски через `mask:` опцию.
 */

const SMOKE = process.env.SMOKE !== '0';
const SCREENSHOT_DIR = path.join(__dirname, 'screenshots');

const targetScreens: ReadonlyArray<Screen> = SMOKE
  ? SCREENS.filter((s) => SMOKE_SCREEN_IDS.includes(s.id))
  : SCREENS;

/**
 * Внешние домены, которые legacy-php тянет из CDN/трекеров и которые в
 * dev-инстансе не нужны (висят / медленные / шумят в скриншоте). Блокируем,
 * чтобы page.goto не таймаутился на networkidle.
 *
 * jQuery 1.5.1 / jQuery-UI 1.8.14 — headless Chromium CDN-загрузка
 * 5-30с, иногда полностью таймаутит. Без них страница не интерактивна,
 * но статический layout рендерится → скриншот валиден для baseline.
 */
const BLOCKED_HOSTS = [
  'ajax.googleapis.com',
  'fonts.googleapis.com',
  'fonts.gstatic.com',
  'counter.yadro.ru',
  'www.liveinternet.ru',
  'cakeuniverse.ru',
];

async function blockExternal(context: BrowserContext): Promise<void> {
  await context.route('**/*', (route) => {
    const host = new URL(route.request().url()).hostname;
    if (BLOCKED_HOSTS.some((h) => host === h || host.endsWith(`.${h}`))) {
      void route.abort();
      return;
    }
    void route.continue();
  });
}

async function devLogin(context: BrowserContext, page: Page): Promise<void> {
  // /dev-login.php ставит cookie `oxsar-jwt` (alg=none) и редиректит на ?go=Main.
  // waitUntil 'load' (не networkidle) — networkidle ждёт CDN-запросы.
  const resp = await page.goto('/dev-login.php', { waitUntil: 'load' });
  expect(resp, 'dev-login response is null').not.toBeNull();
  // Acceptable: редирект на Main (200) или прямой 302→200. После waitUntil
  // domcontentloaded — page.url() уже на ?go=Main.
  const url = page.url();
  expect(url, `dev-login did not redirect into game (current url=${url})`).toMatch(
    /[?&]go=Main\b|game\.php/,
  );
  // Sanity: cookie реально появилась.
  const cookies = await context.cookies();
  const jwt = cookies.find((c) => c.name === 'oxsar-jwt');
  expect(jwt, 'oxsar-jwt cookie not set after dev-login').toBeDefined();
}

test.describe('origin-baseline screenshots', () => {
  test.beforeEach(async ({ context, page }) => {
    await blockExternal(context);
    await devLogin(context, page);
  });

  for (const screen of targetScreens) {
    test(`${screen.id} ${screen.description}`, async ({ page }) => {
      const resp = await page.goto(screen.path, { waitUntil: 'load' });
      expect(resp?.status(), `${screen.path} status`).toBeLessThan(400);
      // Дать DOM/CSS отрендериться (без JS-countdown — он заблокирован вместе с jQuery).
      await page.waitForTimeout(500);
      const file = path.join(SCREENSHOT_DIR, `${screen.id.toLowerCase()}-${screen.name}.png`);
      await page.screenshot({ path: file, fullPage: true });
    });
  }
});
