import { chromium } from 'playwright';
import { execSync } from 'child_process';

const ORIGIN_URL = 'http://localhost:5176';
const IDENTITY_URL = 'http://localhost:9000';
const OUT = new URL('./screenshots-72.1', import.meta.url).pathname.replace(/^\//, '');

const browser = await chromium.launch({ headless: true });
const raw = execSync(
  `curl -s -X POST ${IDENTITY_URL}/auth/login -H "Content-Type: application/json" -d "{\\"login\\":\\"test\\",\\"password\\":\\"DevPass123\\"}"`,
  { encoding: 'utf8' }
);
const data = JSON.parse(raw);
const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
const page = await ctx.newPage();
await ctx.addInitScript(({ a, r, u }) => {
  localStorage.setItem('oxsar-origin-auth', JSON.stringify({ v: 1, access: a, refresh: r, userId: u }));
}, { a: data.access_token, r: data.refresh_token, u: data.user?.id ?? '' });
await page.goto(`${ORIGIN_URL}/tools/tech-calc`, { waitUntil: 'networkidle' });
await page.waitForTimeout(3000);
await page.screenshot({ path: `${OUT}/origin_tech-calc.png`, fullPage: true });
console.log('done');
await browser.close();
