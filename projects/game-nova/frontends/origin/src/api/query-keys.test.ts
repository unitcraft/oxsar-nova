// Все 7 экранов — корректность Query-keys для invalidation.
// (План 72 Ф.2 Spring 1)
//
// TanStack Query сравнивает ключи по структурному равенству. Если
// `enqueueBuilding(planet)` invalidate'ит `['buildings', 'queue', planet]`
// и `useQuery({ queryKey: QK.buildingQueue(planet) })` собирает тот же
// ключ — тест проверяет, что обе стороны строят идентичные tuples.

import { describe, it, expect } from 'vitest';
import { QK } from './query-keys';

describe('QK', () => {
  it('planets: статический массив', () => {
    expect(QK.planets()).toEqual(['planets']);
  });

  it('planet(id) включает id', () => {
    expect(QK.planet('abc')).toEqual(['planet', 'abc']);
  });

  it('buildingQueue / shipyardQueue / shipyardInventory параметризованы по planetId', () => {
    expect(QK.buildingQueue('p1')).toEqual(['buildings', 'queue', 'p1']);
    expect(QK.shipyardQueue('p1')).toEqual(['shipyard', 'queue', 'p1']);
    expect(QK.shipyardInventory('p1')).toEqual(['shipyard', 'inventory', 'p1']);
  });

  it('research / fleet / unreadCount — статические', () => {
    expect(QK.research()).toEqual(['research']);
    expect(QK.fleet()).toEqual(['fleet']);
    expect(QK.unreadCount()).toEqual(['messages', 'unread-count']);
  });

  it('galaxy(g, s) различает координаты', () => {
    expect(QK.galaxy(1, 1)).toEqual(['galaxy', 1, 1]);
    expect(QK.galaxy(1, 2)).toEqual(['galaxy', 1, 2]);
    expect(QK.galaxy(1, 1)).not.toEqual(QK.galaxy(2, 1));
  });
});
