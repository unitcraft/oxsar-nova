// Тест workflow план 72.1 ч.20.11:
// 1. Симулятор → Симулировать → редирект на /battle-report/{id}
// 2. Тот же URL открыт анонимно (без localStorage auth) — отчёт виден
import { chromium } from 'playwright';
import { execSync } from 'child_process';

const ORIGIN_URL = 'http://localhost:5176';
const IDENTITY_URL = 'http://localhost:9000';
const OUT = new URL('./screenshots-72.1', import.meta.url).pathname.replace(/^\//, '');

const raw = execSync(
  `curl -s -X POST ${IDENTITY_URL}/auth/login -H "Content-Type: application/json" -d "{\\"login\\":\\"test\\",\\"password\\":\\"DevPass123\\"}"`,
  { encoding: 'utf8' }
);
const data = JSON.parse(raw);

const browser = await chromium.launch({ headless: true });

// 1. Авторизованный симулятор
{
  const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
  const page = await ctx.newPage();
  await ctx.addInitScript(({ a, r, u }) => {
    localStorage.setItem('oxsar-origin-auth', JSON.stringify({ v: 1, access: a, refresh: r, userId: u }));
  }, { a: data.access_token, r: data.refresh_token, u: data.user?.id ?? '' });

  await page.goto(`${ORIGIN_URL}/simulator`, { waitUntil: 'networkidle' });
  await page.waitForTimeout(2000);

  // Заполняем atacker light_fighter qty + defender rocket_launcher qty
  // Атакующий light_fighter — первый input в строке (qty атакующего).
  const aLF = await page.locator('tr:has-text("Легкий истребитель") input[type="text"]').first();
  await aLF.fill('100');
  // Защитник rocket_launcher — это оборона, у неё только 3 input (атакующий "—").
  // Берём первый input в строке (это qty защитника).
  const dRL = await page.locator('tr:has-text("Ракетная установка") input[type="text"]').first();
  await dRL.fill('50');

  await page.click('input[type="submit"][name="simulate"]');
  await page.waitForURL(/\/battle-report\/[0-9a-f-]+/, { timeout: 10000 });
  const url = page.url();
  console.log('redirected to:', url);
  await page.waitForTimeout(2000);
  await page.screenshot({ path: `${OUT}/origin_battle-report_after-sim.png`, fullPage: true });
  await ctx.close();

  // Извлекаем id из URL.
  const m = url.match(/\/battle-report\/([0-9a-f-]+)/);
  if (m) {
    // 2. Анонимный context — без localStorage auth
    const anonCtx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
    const anonPage = await anonCtx.newPage();
    await anonPage.goto(`${ORIGIN_URL}/battle-report/${m[1]}`, { waitUntil: 'networkidle' });
    await anonPage.waitForTimeout(2000);
    await anonPage.screenshot({ path: `${OUT}/origin_battle-report_anonymous.png`, fullPage: true });
    console.log('anonymous view saved');
    await anonCtx.close();
  }
}

await browser.close();
