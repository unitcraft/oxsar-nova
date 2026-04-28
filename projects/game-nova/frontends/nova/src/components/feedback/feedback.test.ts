// Unit-тесты для pure-функций UX-микрологики (план 71).
// Покрывают: дефицит ресурсов (X-001), знак производства (X-002),
// энергодефицит (X-010), added_level (X-013), статус артефактов
// (X-008), слоты флота (X-007).
//
// Все функции pure → тестируются стандартным vitest без DOM.

import { describe, it, expect } from 'vitest';
import {
  computeDeficit,
  hasAnyDeficit,
  canAfford,
  numKind,
  energyKind,
  addedLevelKind,
  formatAddedLevel,
  slotsState,
  artefactStatusKind,
  expiryUrgency,
} from './feedback';

describe('computeDeficit (X-001)', () => {
  it('returns zero deficit when fully covered', () => {
    expect(computeDeficit(
      { metal: 100, silicon: 50, hydrogen: 0 },
      { metal: 1000, silicon: 200, hydrogen: 50 },
    )).toEqual({ metal: 0, silicon: 0, hydrogen: 0 });
  });

  it('reports difference when short on one resource', () => {
    expect(computeDeficit(
      { metal: 1000, silicon: 0, hydrogen: 0 },
      { metal: 250, silicon: 0, hydrogen: 0 },
    )).toEqual({ metal: 750, silicon: 0, hydrogen: 0 });
  });

  it('clamps to zero, not negative', () => {
    const d = computeDeficit(
      { metal: 100, silicon: 100, hydrogen: 100 },
      { metal: 50,  silicon: 200, hydrogen: 50 },
    );
    expect(d.silicon).toBe(0);
    expect(d.metal).toBe(50);
    expect(d.hydrogen).toBe(50);
  });

  it('handles all-three deficit', () => {
    expect(computeDeficit(
      { metal: 100, silicon: 200, hydrogen: 300 },
      { metal: 0, silicon: 0, hydrogen: 0 },
    )).toEqual({ metal: 100, silicon: 200, hydrogen: 300 });
  });
});

describe('hasAnyDeficit', () => {
  it('false when all zero', () => {
    expect(hasAnyDeficit({ metal: 0, silicon: 0, hydrogen: 0 })).toBe(false);
  });
  it('true when any positive', () => {
    expect(hasAnyDeficit({ metal: 0, silicon: 1, hydrogen: 0 })).toBe(true);
    expect(hasAnyDeficit({ metal: 1, silicon: 0, hydrogen: 0 })).toBe(true);
    expect(hasAnyDeficit({ metal: 0, silicon: 0, hydrogen: 1 })).toBe(true);
  });
});

describe('canAfford', () => {
  it('true when have ≥ cost across the board', () => {
    expect(canAfford(
      { metal: 100, silicon: 50, hydrogen: 0 },
      { metal: 100, silicon: 50, hydrogen: 0 },
    )).toBe(true);
  });
  it('false on any deficit', () => {
    expect(canAfford(
      { metal: 100, silicon: 50, hydrogen: 0 },
      { metal: 100, silicon: 49, hydrogen: 0 },
    )).toBe(false);
  });
});

describe('numKind (X-002)', () => {
  it('classifies positive/negative/zero', () => {
    expect(numKind(10)).toBe('positive');
    expect(numKind(-3)).toBe('negative');
    expect(numKind(0)).toBe('zero');
  });
  it('handles non-integer values', () => {
    expect(numKind(0.0001)).toBe('positive');
    expect(numKind(-0.0001)).toBe('negative');
  });
});

describe('energyKind (X-010)', () => {
  it('zero is deficit (origin: <= 0)', () => {
    expect(energyKind(0)).toBe('deficit');
  });
  it('negative is deficit', () => {
    expect(energyKind(-1)).toBe('deficit');
  });
  it('positive is surplus', () => {
    expect(energyKind(1)).toBe('surplus');
  });
});

describe('addedLevelKind / formatAddedLevel (X-013)', () => {
  it('+2 -> positive', () => {
    expect(addedLevelKind(2)).toBe('positive');
    expect(formatAddedLevel(2)).toBe('+2');
  });
  it('-1 -> negative', () => {
    expect(addedLevelKind(-1)).toBe('negative');
    expect(formatAddedLevel(-1)).toBe('-1');
  });
  it('0 -> none + empty string', () => {
    expect(addedLevelKind(0)).toBe('none');
    expect(formatAddedLevel(0)).toBe('');
  });
});

describe('slotsState (X-007)', () => {
  it('full when used >= max', () => {
    expect(slotsState(5, 5)).toBe('full');
    expect(slotsState(6, 5)).toBe('full');
  });
  it('almost when 1 slot left', () => {
    expect(slotsState(4, 5)).toBe('almost');
  });
  it('ok when more than one free', () => {
    expect(slotsState(3, 5)).toBe('ok');
    expect(slotsState(0, 5)).toBe('ok');
  });
  it('full when max <= 0 (degenerate)', () => {
    expect(slotsState(0, 0)).toBe('full');
  });
});

describe('artefactStatusKind (X-008)', () => {
  it('maps known states', () => {
    expect(artefactStatusKind('active')).toBe('active');
    expect(artefactStatusKind('delayed')).toBe('charging');
    expect(artefactStatusKind('listed')).toBe('listed');
    expect(artefactStatusKind('expired')).toBe('gone');
    expect(artefactStatusKind('consumed')).toBe('gone');
    expect(artefactStatusKind('held')).toBe('idle');
  });
  it('unknown state -> idle (safe fallback)', () => {
    expect(artefactStatusKind('whatever')).toBe('idle');
  });
});

describe('expiryUrgency (X-008)', () => {
  const now = 1_700_000_000_000; // фиксированный момент

  it('returns none when no expiry', () => {
    expect(expiryUrgency(null, now)).toBe('none');
    expect(expiryUrgency(undefined, now)).toBe('none');
  });
  it('imminent within 1 hour', () => {
    const t = new Date(now + 30 * 60 * 1000).toISOString(); // 30 мин
    expect(expiryUrgency(t, now)).toBe('imminent');
  });
  it('soon within 1 day', () => {
    const t = new Date(now + 5 * 60 * 60 * 1000).toISOString(); // 5 часов
    expect(expiryUrgency(t, now)).toBe('soon');
  });
  it('ok beyond 1 day', () => {
    const t = new Date(now + 3 * 24 * 60 * 60 * 1000).toISOString(); // 3 дня
    expect(expiryUrgency(t, now)).toBe('ok');
  });
  it('past expiry counts as imminent (left ≤ 0)', () => {
    const t = new Date(now - 1000).toISOString();
    expect(expiryUrgency(t, now)).toBe('imminent');
  });
});
