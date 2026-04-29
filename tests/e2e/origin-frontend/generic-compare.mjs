// Универсальный скрипт сравнения скриншотов экранов origin vs legacy.
// Запуск: node generic-compare.mjs <originPath> <legacyClass>
// Пример: node generic-compare.mjs /profession Profession
//
// Сохраняет в screenshots-72.1/origin_<key>.png и legacy_<key>.png
// где <key> — последний сегмент originPath.

import { chromium } from 'playwright';
import { execSync } from 'child_process';

const ORIGIN_URL = 'http://localhost:5176';
const LEGACY_URL = 'http://localhost:8092';
const IDENTITY_URL = 'http://localhost:9000';
const OUT = new URL('./screenshots-72.1', import.meta.url).pathname.replace(/^\//, '');

const [originPath, legacyClass] = process.argv.slice(2);
if (!originPath || !legacyClass) {
  console.error('Usage: node generic-compare.mjs <originPath> <legacyClass>');
  process.exit(1);
}
const key = originPath.replace(/^\//, '').replace(/[/?=&]/g, '_') || 'index';

const browser = await chromium.launch({ headless: true });

// --- ORIGIN ---
{
  const raw = execSync(
    `curl -s -X POST ${IDENTITY_URL}/auth/login -H "Content-Type: application/json" -d "{\\"login\\":\\"test\\",\\"password\\":\\"DevPass123\\"}"`,
    { encoding: 'utf8' }
  );
  const data = JSON.parse(raw);
  const access = data.access_token ?? '';
  const refresh = data.refresh_token ?? '';
  const userId = data.user?.id ?? '';

  const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await ctx.newPage();
  await ctx.addInitScript(({ access, refresh, userId }) => {
    localStorage.setItem('oxsar-origin-auth', JSON.stringify({ v: 1, access, refresh, userId }));
  }, { access, refresh, userId });

  await page.goto(`${ORIGIN_URL}${originPath}`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(3000);
  await page.screenshot({ path: `${OUT}/origin_${key}.png`, fullPage: true });
  console.log(`origin ${originPath} → origin_${key}.png`);

  await ctx.close();
}

// --- LEGACY ---
{
  const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await ctx.newPage();
  await page.goto(`${LEGACY_URL}/dev-login.php`, { waitUntil: 'domcontentloaded', timeout: 60000 });
  await page.waitForTimeout(2000);

  await page.goto(`${LEGACY_URL}/game.php?go=${legacyClass}`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(3000);
  await page.screenshot({ path: `${OUT}/legacy_${key}.png`, fullPage: true });
  console.log(`legacy ?go=${legacyClass} → legacy_${key}.png`);

  await ctx.close();
}

await browser.close();
