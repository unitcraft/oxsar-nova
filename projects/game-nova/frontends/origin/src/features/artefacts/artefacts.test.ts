// S-013 Artefacts — группировка инвентаря по статусу.
// (План 72 Ф.4 Spring 3)
//
// Перепроверка чистой функции группировки артефактов: active+delayed
// идут в группу 'active', остальное — в 'held' / 'other'. UI рендерит
// только non-empty группы.

import { describe, it, expect } from 'vitest';
import type { Artefact } from '@/api/types';

type GroupKey = 'active' | 'held' | 'other';

function groupArtefacts(items: Artefact[]): Array<{ key: GroupKey; count: number }> {
  const buckets: Record<GroupKey, number> = { active: 0, held: 0, other: 0 };
  for (const a of items) {
    if (a.state === 'active' || a.state === 'delayed') buckets.active += 1;
    else if (a.state === 'held') buckets.held += 1;
    else buckets.other += 1;
  }
  return (['active', 'held', 'other'] as const).map((key) => ({
    key,
    count: buckets[key],
  }));
}

function mk(state: Artefact['state']): Artefact {
  return {
    id: '00000000-0000-0000-0000-000000000000',
    user_id: '00000000-0000-0000-0000-000000000001',
    planet_id: null,
    unit_id: 3001,
    state,
    acquired_at: '2026-04-28T00:00:00Z',
    activated_at: null,
    expire_at: null,
  };
}

describe('groupArtefacts', () => {
  it('пустой инвентарь — все группы пустые', () => {
    const groups = groupArtefacts([]);
    expect(groups.every((g) => g.count === 0)).toBe(true);
  });

  it('active+delayed → active; held → held; expired/consumed → other', () => {
    const groups = groupArtefacts([
      mk('active'),
      mk('delayed'),
      mk('held'),
      mk('expired'),
      mk('consumed'),
    ]);
    expect(groups.find((g) => g.key === 'active')?.count).toBe(2);
    expect(groups.find((g) => g.key === 'held')?.count).toBe(1);
    expect(groups.find((g) => g.key === 'other')?.count).toBe(2);
  });
});
