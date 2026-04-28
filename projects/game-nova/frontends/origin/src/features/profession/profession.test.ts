import { describe, it, expect } from 'vitest';

// Pure-функция timeUntil(): когда снова доступна смена профессии.
function timeUntil(iso: string, now: number): string | null {
  const ms = new Date(iso).getTime() - now;
  if (ms <= 0) return null;
  const d = Math.floor(ms / 86_400_000);
  const h = Math.floor((ms % 86_400_000) / 3_600_000);
  if (d > 0) return `${d}д ${h}ч`;
  const m = Math.floor((ms % 3_600_000) / 60_000);
  return `${h}ч ${m}м`;
}

const NOW = new Date('2026-04-28T12:00:00Z').getTime();

describe('profession', () => {
  it('cooldown в прошлом → null (можно менять)', () => {
    expect(timeUntil('2026-04-27T12:00:00Z', NOW)).toBeNull();
  });

  it('cooldown через 5 дней → "5д 0ч"', () => {
    expect(timeUntil('2026-05-03T12:00:00Z', NOW)).toBe('5д 0ч');
  });

  it('cooldown через 3 часа 15 минут → "3ч 15м"', () => {
    expect(timeUntil('2026-04-28T15:15:00Z', NOW)).toBe('3ч 15м');
  });

  it('фильтр bonus/malus: 0 не показываем', () => {
    const bonus = { metalmine: 5, gun: 0, silicon_lab: -2 };
    const filtered = Object.entries(bonus).filter(([, v]) => v !== 0);
    expect(filtered).toEqual([
      ['metalmine', 5],
      ['silicon_lab', -2],
    ]);
  });
});
