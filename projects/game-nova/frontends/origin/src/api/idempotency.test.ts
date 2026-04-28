// S-002 Constructions / S-003 Research / S-004 Shipyard / S-006 Mission —
// Idempotency-Key обязательно у каждой mutation (R9 ТЗ §16.10).
// (План 72 Ф.2 Spring 1)

import { describe, it, expect } from 'vitest';
import { newIdempotencyKey } from './idempotency';

describe('newIdempotencyKey', () => {
  it('возвращает непустую строку', () => {
    const k = newIdempotencyKey();
    expect(typeof k).toBe('string');
    expect(k.length).toBeGreaterThan(8);
  });

  it('два вызова дают разные значения (anti-double-submit)', () => {
    const a = newIdempotencyKey();
    const b = newIdempotencyKey();
    expect(a).not.toBe(b);
  });

  it('1000 ключей уникальны (нет коллизий)', () => {
    const set = new Set<string>();
    for (let i = 0; i < 1000; i++) set.add(newIdempotencyKey());
    expect(set.size).toBe(1000);
  });
});
