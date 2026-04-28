import { defineConfig, devices } from '@playwright/test';

const LEGACY_URL = process.env.LEGACY_URL ?? 'http://localhost:8092';

export default defineConfig({
  testDir: '.',
  testMatch: /baseline\.spec\.ts$/,
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  workers: 1,
  // Первый goto после холодного старта legacy-php (MySQL row cache, opcache,
  // memcached prewarm) занимает ~30-60s; повторы — ~3-4s. retries=1 покрывает.
  retries: 1,
  reporter: [['list']],
  use: {
    baseURL: LEGACY_URL,
    trace: 'off',
    screenshot: 'off',
    video: 'off',
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 1,
    locale: 'ru-RU',
    timezoneId: 'Europe/Moscow',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'], viewport: { width: 1440, height: 900 } },
    },
  ],
  metadata: { legacyURL: LEGACY_URL },
});
