import { defineConfig, devices } from '@playwright/test';

const BACKEND_URL = process.env.BACKEND_URL ?? 'http://localhost:8080';
const FRONTEND_URL = process.env.FRONTEND_URL ?? 'http://localhost:5173';

// Playwright-конфиг для E2E-тестов UI.
//
// - `webServer` не поднимаем автоматически: dev-стек (backend+worker+frontend)
//   запускается вручную через `make dev-up backend-run worker-run frontend-run`,
//   или в CI отдельным шагом. Это даёт контроль над логами и предотвращает
//   таймауты из-за холодного старта goose-миграций.
// - Сид БД выполняется из фикстуры (e2e/fixtures/seed.ts) через HTTP к
//   backend-хосту или через отдельный шаг `make test-seed` в CI.
export default defineConfig({
  testDir: './e2e',
  // 60s — mobile viewport + vite dev server + первый lazy-chunk = медленно.
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: false, // общая БД — чтобы тесты не портили друг другу состояние
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: process.env.CI ? [['github'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: FRONTEND_URL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    // extraHTTPHeaders убран: Chromium шлёт их и на cross-origin запросы
    // (например, fonts.gstatic.com), ломая CORS-preflight и плодя
    // console.error'ы, которые ловит smoke-спек.
  },
  projects: [
    {
      name: 'desktop',
      use: { ...devices['Desktop Chrome'], viewport: { width: 1440, height: 900 } },
    },
    {
      name: 'mobile',
      use: { ...devices['Pixel 5'] },
    },
  ],
  metadata: { backendURL: BACKEND_URL },
});
