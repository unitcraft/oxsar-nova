import { describe, it, expect } from 'vitest';

// UserAgreement — cross-link на portal. Тест проверяет конструкцию URL
// в зависимости от значения VITE_PORTAL_BASE_URL.
function buildAgreementUrl(portalBase: string | undefined): string {
  return `${portalBase ?? ''}/user-agreement`;
}

describe('user-agreement', () => {
  it('с заданной VITE_PORTAL_BASE_URL → абсолютный URL', () => {
    expect(buildAgreementUrl('https://oxsar-nova.ru')).toBe(
      'https://oxsar-nova.ru/user-agreement',
    );
  });

  it('без VITE_PORTAL_BASE_URL → относительный путь', () => {
    expect(buildAgreementUrl(undefined)).toBe('/user-agreement');
  });

  it('пустая строка → относительный путь', () => {
    expect(buildAgreementUrl('')).toBe('/user-agreement');
  });
});
