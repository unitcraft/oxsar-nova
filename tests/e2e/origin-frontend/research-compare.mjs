import { chromium } from 'playwright';
import { execSync } from 'child_process';

const ORIGIN_URL = 'http://localhost:5176';
const LEGACY_URL = 'http://localhost:8092';
const IDENTITY_URL = 'http://localhost:9000';
const OUT = 'C:/Users/Евгений/OneDrive/Рабочий стол';

const browser = await chromium.launch({ headless: true });

// --- ORIGIN /research ---
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

  await page.goto(`${ORIGIN_URL}/research`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(3000);
  await page.screenshot({ path: `${OUT}/origin_research.png`, fullPage: true });
  console.log('origin /research saved');
  await ctx.close();
}

// --- LEGACY Research ---
{
  const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await ctx.newPage();
  await page.goto(`${LEGACY_URL}/dev-login.php`);
  await page.waitForTimeout(1500);
  await page.goto(`${LEGACY_URL}/game.php?go=Research`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(3000);
  await page.screenshot({ path: `${OUT}/legacy_research.png`, fullPage: true });
  console.log('legacy Research saved');
  await ctx.close();
}

await browser.close();
