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
const ctx = await browser.newContext({ viewport: { width: 1920, height: 1080 } });
const page = await ctx.newPage();
await ctx.addInitScript(({ a, r, u }) => {
  localStorage.setItem('oxsar-origin-auth', JSON.stringify({ v: 1, access: a, refresh: r, userId: u }));
  // Pre-populate last-sim чтобы сразу увидеть результат + кнопку
}, { a: data.access_token, r: data.refresh_token, u: data.user?.id ?? '' });

await page.goto(`${ORIGIN_URL}/simulator`, { waitUntil: 'networkidle' });
await page.waitForTimeout(2000);

// Заполняю input для light_fighter атакующего и rocket_launcher защитника.
// Используем имена input'ов name="31" / name="31_d" / name="31_p" (атакующий).
// Defender: name="43" / "43_d" / "43_p". Но компонент использует id'ы.
// Проще: программно нажать кнопку «Установить флот» для атакующего (поставит весь
// inventory) и «Установить оборону» для защитника. У test-юзера inv может быть пуст.
// Поэтому делаю напрямую через React-events: ставим qty 100 в первый input строки атакующего light_fighter.

// Самый надёжный путь — установить значения через прямые DOM манипуляции и dispatchEvent.
// React onChange должен сработать.
const setFirstField = async (rowText, columnIdx, value) => {
  const row = page.locator(`tr:has-text("${rowText}")`).first();
  const inputs = await row.locator('input[type="text"]').all();
  if (inputs[columnIdx]) {
    await inputs[columnIdx].fill(value);
  }
};

// columnIdx: 0,1,2 — атакующий qty/dmg/percent; 3,4,5 — защитник.
await setFirstField('Легкий истребитель', 0, '100');
await setFirstField('Ракетная установка', 3, '50');

await page.click('input[type="submit"][name="simulate"]');
await page.waitForTimeout(2000);

await page.screenshot({ path: `${OUT}/origin_simulator_with_result.png`, fullPage: true });
console.log('saved');
await browser.close();
