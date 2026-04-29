import { chromium } from 'playwright';
import { execSync } from 'child_process';

const ORIGIN_URL = 'http://localhost:5176';
const LEGACY_URL = 'http://localhost:8092';
const IDENTITY_URL = 'http://localhost:9000';
// Текущий план: docs/plans/72.1-post-remaster-stabilization.md
const OUT = new URL('./screenshots-72.1', import.meta.url).pathname.replace(/^\//, '');

const browser = await chromium.launch({ headless: true });

// --- ORIGIN /shipyard ---
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

  await page.goto(`${ORIGIN_URL}/shipyard`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(3000);
  await page.screenshot({ path: `${OUT}/origin_shipyard.png`, fullPage: true });
  console.log('origin /shipyard saved');

  await page.goto(`${ORIGIN_URL}/defense`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(2000);
  await page.screenshot({ path: `${OUT}/origin_defense.png`, fullPage: true });
  console.log('origin /defense saved');

  await ctx.close();
}

// --- LEGACY Shipyard / Defense ---
{
  const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await ctx.newPage();
  await page.goto(`${LEGACY_URL}/dev-login.php`, { waitUntil: 'domcontentloaded', timeout: 60000 });
  await page.waitForTimeout(2000);

  await page.goto(`${LEGACY_URL}/game.php?go=Shipyard`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(3000);
  await page.screenshot({ path: `${OUT}/legacy_shipyard.png`, fullPage: true });
  console.log('legacy Shipyard saved');

  await page.goto(`${LEGACY_URL}/game.php?go=Defense`, { waitUntil: 'domcontentloaded' });
  await page.waitForTimeout(3000);
  await page.screenshot({ path: `${OUT}/legacy_defense.png`, fullPage: true });
  console.log('legacy Defense saved');

  await ctx.close();
}

await browser.close();
