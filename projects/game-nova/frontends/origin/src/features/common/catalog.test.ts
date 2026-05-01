// S-002 Constructions / S-003 Research / S-004 Shipyard — каталог
// юнитов разделён на правильные группы. (План 72 Ф.2 Spring 1)
//
// Если в каталог добавлен новый ship, но он попал в group 'building'
// — экран Shipyard не покажет его, а Constructions «отрисует» как
// здание. Тест ловит это до появления в UI.

import { describe, it, expect } from 'vitest';
import { CATALOG, catalogByGroup, findCatalog } from './catalog';

describe('CATALOG', () => {
  it('каждая запись имеет валидную группу', () => {
    const groups = new Set(['building', 'research', 'ship', 'defense']);
    for (const e of CATALOG) {
      expect(groups.has(e.group)).toBe(true);
    }
  });

  it('id-ы уникальны во всём каталоге', () => {
    const ids = new Set<number>();
    for (const e of CATALOG) {
      expect(ids.has(e.id)).toBe(false);
      ids.add(e.id);
    }
  });

  it('каждая запись имеет i18n ключ "group.key"', () => {
    for (const e of CATALOG) {
      expect(e.i18n).toMatch(/^[a-z]+\.[a-zA-Z0-9]+$/);
    }
  });

  it('catalogByGroup возвращает только заявленную группу', () => {
    const ships = catalogByGroup('ship');
    expect(ships.length).toBeGreaterThan(0);
    for (const e of ships) {
      expect(e.group).toBe('ship');
    }
  });

  it('findCatalog находит по unit_id', () => {
    // Cruiser (legacy unit_id=33) — корабль из каталога origin.
    const e = findCatalog(33);
    expect(e?.group).toBe('ship');
    expect(findCatalog(99999)).toBeUndefined();
  });

  it('Spring 1 минимум: 4 группы, ≥3 юнитов в каждой', () => {
    expect(catalogByGroup('building').length).toBeGreaterThanOrEqual(3);
    expect(catalogByGroup('research').length).toBeGreaterThanOrEqual(3);
    expect(catalogByGroup('ship').length).toBeGreaterThanOrEqual(3);
    expect(catalogByGroup('defense').length).toBeGreaterThanOrEqual(3);
  });
});
