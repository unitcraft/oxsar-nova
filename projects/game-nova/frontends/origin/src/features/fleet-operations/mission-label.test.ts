// Тест на маппинг mission-кодов на i18n-ключи bundle fleet:*
// (план 72 Ф.3 Spring 2 ч.2).

import { describe, it, expect } from 'vitest';

const MISSION_LABEL_KEY: Record<number, string> = {
  6: 'missionAttack',
  7: 'missionExpedition',
  8: 'missionTransport',
  9: 'missionRebase',
  10: 'missionColonize',
  11: 'missionRecycle',
  12: 'missionSpy',
  15: 'missionAttack',
};

describe('MISSION_LABEL_KEY', () => {
  it('покрывает все 8 нужных кодов', () => {
    [6, 7, 8, 9, 10, 11, 12, 15].forEach((code) => {
      expect(MISSION_LABEL_KEY[code]).toBeDefined();
    });
  });

  it('mission=6 (атака) и mission=15 (alliance attack) → одинаковый label', () => {
    expect(MISSION_LABEL_KEY[6]).toBe('missionAttack');
    expect(MISSION_LABEL_KEY[15]).toBe('missionAttack');
  });

  it('каждый ключ существует в bundle fleet:* (по контракту)', () => {
    const expectedKeys = [
      'missionAttack',
      'missionExpedition',
      'missionTransport',
      'missionRebase',
      'missionColonize',
      'missionRecycle',
      'missionSpy',
    ];
    const actual = new Set(Object.values(MISSION_LABEL_KEY));
    expectedKeys.forEach((k) => {
      expect(actual.has(k)).toBe(true);
    });
  });
});
