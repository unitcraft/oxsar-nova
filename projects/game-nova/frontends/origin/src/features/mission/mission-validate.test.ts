// S-006 Mission — собранный selectedShips исключает нули и
// отрицательные. (План 72 Ф.2 Spring 1)
//
// Логика выбрана повторно из MissionScreen.handleSend для проверки
// инвариантов backend FleetDispatch:
//   - ships: непустой dict с count >= 1.
//   - speed_percent: clamped в [10, 100] (см. openapi.yaml).
//   - mission: один из MissionCode (но это TS-уровневая защита).
//
// Если эти инварианты нарушаются на UI — POST /api/fleet возвращает
// 400, и игрок видит «Не удалось отправить» без диагностики.

import { describe, it, expect } from 'vitest';

function buildSelectedShips(
  raw: Record<string, string>,
): Record<string, number> {
  const out: Record<string, number> = {};
  for (const [k, v] of Object.entries(raw)) {
    const n = Math.max(0, Math.floor(Number(v) || 0));
    if (n > 0) out[k] = n;
  }
  return out;
}

function clampSpeed(s: number): number {
  return Math.max(10, Math.min(100, s));
}

describe('mission selected ships', () => {
  it('исключает нули, пустые и отрицательные', () => {
    const r = buildSelectedShips({
      '202': '5',
      '203': '0',
      '204': '',
      '205': '-3',
      '206': '7',
    });
    expect(r).toEqual({ '202': 5, '206': 7 });
  });

  it('floor на дробных', () => {
    expect(buildSelectedShips({ '202': '3.7' })).toEqual({ '202': 3 });
  });

  it('пустой ввод → пустой dict', () => {
    expect(buildSelectedShips({})).toEqual({});
    expect(buildSelectedShips({ '202': '' })).toEqual({});
  });
});

describe('mission speed clamp', () => {
  it('clamps to [10, 100]', () => {
    expect(clampSpeed(0)).toBe(10);
    expect(clampSpeed(50)).toBe(50);
    expect(clampSpeed(150)).toBe(100);
  });
});
