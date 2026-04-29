import { chromium } from 'playwright';
import { execSync } from 'child_process';

const ORIGIN_URL = 'http://localhost:5176';
const LEGACY_URL = 'http://localhost:8092';
const IDENTITY_URL = 'http://localhost:9000';
const OUT = 'C:/Users/Евгений/OneDrive/Рабочий стол';

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

  await page.goto(ORIGIN_URL, { waitUntil: 'networkidle' });
  await page.waitForTimeout(3000);

  // Три скриншота: полная страница, левое меню, topHeader
  await page.screenshot({ path: `${OUT}/origin_full.png`, fullPage: true });
  await page.screenshot({ path: `${OUT}/origin_viewport.png` });

  // Клип левого меню
  await page.screenshot({ path: `${OUT}/origin_menu.png`, clip: { x: 0, y: 0, width: 200, height: 900 } });
  // Клип topHeader
  await page.screenshot({ path: `${OUT}/origin_header.png`, clip: { x: 0, y: 0, width: 1920, height: 120 } });

  console.log('origin screenshots saved');
  await ctx.close();
}

// --- LEGACY ---
{
  const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await ctx.newPage();
  await page.goto(`${LEGACY_URL}/dev-login.php`);
  await page.waitForTimeout(2000);
  await page.screenshot({ path: `${OUT}/legacy_full.png`, fullPage: true });
  await page.screenshot({ path: `${OUT}/legacy_viewport.png` });
  await page.screenshot({ path: `${OUT}/legacy_menu.png`, clip: { x: 0, y: 0, width: 200, height: 900 } });
  await page.screenshot({ path: `${OUT}/legacy_header.png`, clip: { x: 0, y: 0, width: 1920, height: 120 } });
  console.log('legacy screenshots saved');
  await ctx.close();
}

await browser.close();
