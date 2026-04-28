import { describe, it, expect } from 'vitest';
import {
  RELATIONS,
  RELATION_LABEL_KEY,
  RELATION_COLOR,
  isKnownRelation,
  relationNeedsAccept,
} from './relations';

describe('RELATIONS', () => {
  it('содержит 5 enum-значений (план 67 D-014 B1)', () => {
    expect(RELATIONS).toEqual(['friend', 'neutral', 'hostile_neutral', 'nap', 'war']);
  });
});

describe('RELATION_LABEL_KEY и RELATION_COLOR покрывают все статусы', () => {
  for (const r of RELATIONS) {
    it(`${r} имеет label-ключ и цвет`, () => {
      expect(RELATION_LABEL_KEY[r]).toBeTruthy();
      expect(RELATION_COLOR[r]).toMatch(/^var\(--/);
    });
  }
});

describe('isKnownRelation', () => {
  it('true для известных значений', () => {
    expect(isKnownRelation('friend')).toBe(true);
    expect(isKnownRelation('war')).toBe(true);
  });
  it('false для неизвестных', () => {
    expect(isKnownRelation('ally')).toBe(false); // legacy, мигрирован → friend
    expect(isKnownRelation('')).toBe(false);
    expect(isKnownRelation('foo')).toBe(false);
  });
});

describe('relationNeedsAccept', () => {
  it('двусторонние отношения требуют согласия (friend, neutral, nap)', () => {
    expect(relationNeedsAccept('friend')).toBe(true);
    expect(relationNeedsAccept('neutral')).toBe(true);
    expect(relationNeedsAccept('nap')).toBe(true);
  });

  it('односторонние отношения принудительны (hostile_neutral, war)', () => {
    expect(relationNeedsAccept('hostile_neutral')).toBe(false);
    expect(relationNeedsAccept('war')).toBe(false);
  });
});
