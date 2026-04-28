import { describe, it, expect } from 'vitest';

// Pure-функция формата эффекта officer'а: множитель → "+20%".
function formatEffectPct(multiplier: number): string {
  const pct = Math.round((multiplier - 1) * 100);
  return `${pct > 0 ? '+' : ''}${pct}%`;
}

describe('officer', () => {
  it('эффект 1.20 → "+20%"', () => {
    expect(formatEffectPct(1.2)).toBe('+20%');
  });

  it('эффект 0.80 → "-20%"', () => {
    expect(formatEffectPct(0.8)).toBe('-20%');
  });

  it('эффект 1.00 (нейтральный) → "0%"', () => {
    expect(formatEffectPct(1.0)).toBe('0%');
  });
});
