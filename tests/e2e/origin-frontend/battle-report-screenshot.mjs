// Тест rendering /battle-report/:id с реальными tech-уровнями.
import { chromium } from 'playwright';

const ORIGIN_URL = 'http://localhost:5176';
const OUT = new URL('./screenshots-72.1', import.meta.url).pathname.replace(/^\//, '');
const REPORT_ID = process.argv[2];

if (!REPORT_ID) {
  console.error('Usage: node battle-report-screenshot.mjs <uuid>');
  process.exit(1);
}

const browser = await chromium.launch({ headless: true });
const ctx = await browser.newContext({ viewport: { width: 1280, height: 1400 }, deviceScaleFactor: 2 });
const page = await ctx.newPage();
await page.goto(`${ORIGIN_URL}/battle-report/${REPORT_ID}`, { waitUntil: 'networkidle' });
await page.waitForTimeout(2000);
// Полный скриншот страницы — раунды + итоговый блок.
await page.screenshot({
  path: `${OUT}/origin_battle-report_full.png`,
  fullPage: true,
});
console.log('saved');
await browser.close();
