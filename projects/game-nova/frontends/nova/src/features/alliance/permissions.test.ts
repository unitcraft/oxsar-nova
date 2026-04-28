import { describe, it, expect } from 'vitest';
import { hasPerm, PERMISSION_KEYS } from './permissions';

describe('hasPerm', () => {
  it('owner всегда true для любого ключа', () => {
    for (const k of PERMISSION_KEYS) {
      expect(hasPerm(true, k, null)).toBe(true);
      expect(hasPerm(true, k, {})).toBe(true);
      expect(hasPerm(true, k, { [k]: false })).toBe(true);
    }
  });

  it('не-owner без rankPerms всегда false', () => {
    for (const k of PERMISSION_KEYS) {
      expect(hasPerm(false, k, null)).toBe(false);
      expect(hasPerm(false, k, undefined)).toBe(false);
      expect(hasPerm(false, k, {})).toBe(false);
    }
  });

  it('не-owner с rank: возвращает значение из карты', () => {
    expect(hasPerm(false, 'can_invite', { can_invite: true })).toBe(true);
    expect(hasPerm(false, 'can_invite', { can_invite: false })).toBe(false);
    expect(hasPerm(false, 'can_kick', { can_invite: true })).toBe(false);
  });
});

describe('PERMISSION_KEYS', () => {
  it('содержит ровно 7 ключей в snake_case (зеркало backend permissions.go)', () => {
    expect(PERMISSION_KEYS).toEqual([
      'can_invite',
      'can_kick',
      'can_send_global_mail',
      'can_manage_diplomacy',
      'can_change_description',
      'can_propose_relations',
      'can_manage_ranks',
    ]);
  });
});
