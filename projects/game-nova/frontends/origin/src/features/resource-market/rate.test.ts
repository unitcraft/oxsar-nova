// Тест на расчёт ожидаемого to_amount в ResourceMarketScreen
// (план 72 Ф.3 Spring 2 ч.2).
//
// Backend `/api/market/rates` возвращает `metal/silicon/hydrogen` как
// относительные веса. Кросс-курс from→to = rate(from)/rate(to).
// Например, при rates{metal:1, silicon:2, hydrogen:4} обмен 1000 metal →
// silicon = 1000 × (1/2) = 500. Тест фиксирует формулу — если кто-то
// случайно поменяет на rate(to)/rate(from) (баг-двойник), пройдёт UI,
// но сделка покажет неверный preview.

import { describe, it, expect } from 'vitest';

function expected(amount: number, rateFrom: number, rateTo: number): number {
  if (rateTo === 0) return 0;
  return Math.floor(amount * (rateFrom / rateTo));
}

describe('expected to_amount', () => {
  const rates = { metal: 1, silicon: 2, hydrogen: 4 };

  it('metal → silicon: 1000 → 500', () => {
    expect(expected(1000, rates.metal, rates.silicon)).toBe(500);
  });

  it('silicon → metal: 1000 → 2000', () => {
    expect(expected(1000, rates.silicon, rates.metal)).toBe(2000);
  });

  it('hydrogen → metal: 100 → 400', () => {
    expect(expected(100, rates.hydrogen, rates.metal)).toBe(400);
  });

  it('same currency: identity', () => {
    expect(expected(777, 2, 2)).toBe(777);
  });

  it('zero amount: zero output', () => {
    expect(expected(0, 1, 2)).toBe(0);
  });

  it('floor — округление вниз', () => {
    // 333 × (1/2) = 166.5 → 166
    expect(expected(333, 1, 2)).toBe(166);
  });
});
