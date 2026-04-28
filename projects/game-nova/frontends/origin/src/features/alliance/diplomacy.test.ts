// Тесты на форматирование строк дипотношений (план 72 Ф.3 Spring 2 ч.1).
//
// Backend возвращает initiator_id/target_id; UI должен показывать
// «противоположную сторону» относительно текущего альянса. Логика
// маленькая, но критичная — игрок не должен видеть собственный тег
// в столбце «контрагент».

import { describe, it, expect } from 'vitest';
import type {
  AllianceRelation,
  AllianceRelationState,
  AllianceRelationStatus,
} from '@/api/types';

function counterpart(rel: AllianceRelation, myAllianceID: string) {
  if (rel.initiator_id === myAllianceID) {
    return { tag: rel.target_tag, name: rel.target_name };
  }
  return { tag: rel.initiator_tag, name: rel.initiator_name };
}

function makeRel(over: Partial<AllianceRelation>): AllianceRelation {
  return {
    initiator_id: 'A',
    target_id: 'B',
    initiator_tag: 'AAA',
    target_tag: 'BBB',
    initiator_name: 'Alpha',
    target_name: 'Bravo',
    status: 'protection' as AllianceRelationStatus,
    state: 'active' as AllianceRelationState,
    message: '',
    proposed_at: '2026-04-28T00:00:00Z',
    established_at: null,
    ...over,
  };
}

describe('counterpart()', () => {
  it('я — initiator, контрагент = target', () => {
    expect(counterpart(makeRel({}), 'A')).toEqual({ tag: 'BBB', name: 'Bravo' });
  });

  it('я — target, контрагент = initiator', () => {
    expect(counterpart(makeRel({}), 'B')).toEqual({
      tag: 'AAA',
      name: 'Alpha',
    });
  });

  it('я ни с какой из сторон → отдаём initiator (нейтральное наблюдение)', () => {
    expect(counterpart(makeRel({}), 'C')).toEqual({
      tag: 'AAA',
      name: 'Alpha',
    });
  });
});
