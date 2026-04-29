import { chromium } from 'playwright';
import { execSync } from 'child_process';

const ORIGIN_URL = 'http://localhost:5176';
const LEGACY_URL = 'http://localhost:8092';
const IDENTITY_URL = 'http://localhost:9000';
// Скриншоты сохраняются в папку с именем плана, по которому идёт работа.
// Текущий план: docs/plans/72.1-post-remaster-stabilization.md
const OUT = new URL('./screenshots-72.1', import.meta.url).pathname.replace(/^\//, '');

const browser = await chromium.launch({ headless: true });

// --- ORIGIN /constructions ---
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

  await page.goto(`${ORIGIN_URL}/constructions`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(3000);
  await page.screenshot({ path: `${OUT}/origin_constructions.png`, fullPage: true });
  console.log('origin /constructions saved');
  await ctx.close();
}

// --- LEGACY Construction ---
{
  const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await ctx.newPage();
  await page.goto(`${LEGACY_URL}/dev-login.php`);
  await page.waitForTimeout(1500);
  await page.goto(`${LEGACY_URL}/game.php?go=Constructions`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(3000);
  await page.screenshot({ path: `${OUT}/legacy_constructions.png`, fullPage: true });
  console.log('legacy Construction saved');
  await ctx.close();
}

await browser.close();
