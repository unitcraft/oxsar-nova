// S-021 Techtree — секции и сортировка узлов. (План 72 Ф.4 Spring 3)
//
// Логика: внутри секции — разблокированные сверху, дальше по id ASC.
// Это критично для UX: игрок видит то что доступно сразу, locked
// уходят вниз.

import { describe, it, expect } from 'vitest';
import type { TechtreeNode } from '@/api/types';

function sortSection(items: TechtreeNode[]): TechtreeNode[] {
  return [...items].sort((a, b) => {
    if (a.unlocked !== b.unlocked) return a.unlocked ? -1 : 1;
    return a.id - b.id;
  });
}

function mk(id: number, unlocked: boolean): TechtreeNode {
  return {
    key: `unit_${id}`,
    kind: 'building',
    id,
    current_level: 0,
    unlocked,
    requirements: [],
  };
}

describe('techtree section sort', () => {
  it('unlocked сверху, locked снизу', () => {
    const sorted = sortSection([mk(1, false), mk(2, true), mk(3, false)]);
    expect(sorted[0]?.id).toBe(2);
    expect(sorted[0]?.unlocked).toBe(true);
    expect(sorted.slice(1).every((n) => !n.unlocked)).toBe(true);
  });

  it('внутри одного статуса — по id ASC', () => {
    const sorted = sortSection([mk(5, true), mk(1, true), mk(3, true)]);
    expect(sorted.map((n) => n.id)).toEqual([1, 3, 5]);
  });

  it('пустой массив — пустой результат', () => {
    expect(sortSection([])).toEqual([]);
  });
});
