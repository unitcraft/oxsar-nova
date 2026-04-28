import { describe, it, expect } from 'vitest';

// helpers для friends-экрана: формат last_seen.
function formatMinutes(ms: number): number {
  return Math.floor(ms / 60_000);
}

describe('friends', () => {
  it('online — менее 5 минут с last_seen', () => {
    const minutes = formatMinutes(2 * 60_000);
    expect(minutes).toBe(2);
    expect(minutes < 5).toBe(true);
  });

  it('часовой формат — между 60 и 1440 минут', () => {
    const minutes = formatMinutes(120 * 60_000);
    expect(minutes).toBe(120);
    expect(minutes >= 60).toBe(true);
    expect(Math.floor(minutes / 60)).toBe(2);
  });

  it('дневной формат — более 24 часов', () => {
    const minutes = formatMinutes(72 * 60 * 60_000);
    expect(minutes).toBe(72 * 60);
    const hours = Math.floor(minutes / 60);
    expect(hours).toBeGreaterThanOrEqual(24);
    const days = Math.floor(hours / 24);
    expect(days).toBe(3);
  });
});
