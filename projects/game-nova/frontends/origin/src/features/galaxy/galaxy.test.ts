// S-005 Galaxy — навигация по координатам с защитой от выхода за
// допустимые диапазоны. (План 72 Ф.2 Spring 1)
//
// Backend openapi.yaml ограничивает galaxy=1..16, system=1..999.
// На UI должна быть симметричная защита, иначе клик на «‹‹» в
// galaxy=1 отправит запрос с galaxy=0 и backend вернёт 400.
// Локальная функция clamp в GalaxyScreen — экстракт ниже для
// проверки граничных условий.

import { describe, it, expect } from 'vitest';

function clamp(value: number, min: number, max: number): number {
  if (Number.isNaN(value)) return min;
  return Math.max(min, Math.min(max, value));
}

describe('galaxy navigation clamp', () => {
  it('keeps value in [min, max]', () => {
    expect(clamp(5, 1, 16)).toBe(5);
    expect(clamp(0, 1, 16)).toBe(1);
    expect(clamp(99, 1, 16)).toBe(16);
  });

  it('NaN → min', () => {
    expect(clamp(Number.NaN, 1, 999)).toBe(1);
  });

  it('boundaries inclusive', () => {
    expect(clamp(1, 1, 16)).toBe(1);
    expect(clamp(16, 1, 16)).toBe(16);
    expect(clamp(999, 1, 999)).toBe(999);
  });
});
