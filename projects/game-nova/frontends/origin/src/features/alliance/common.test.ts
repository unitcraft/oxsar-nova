// Тесты на общие утилиты alliance (план 72 Ф.3 Spring 2 ч.1).

import { describe, it, expect } from 'vitest';
import {
  hasPerm,
  PERMISSION_KEYS,
  relationStatusKey,
} from './permissions';

describe('hasPerm', () => {
  it('owner всегда имеет любое право, независимо от rankPerms', () => {
    expect(hasPerm(true, 'can_manage_ranks', null)).toBe(true);
    expect(hasPerm(true, 'can_kick', { can_kick: false })).toBe(true);
    expect(hasPerm(true, 'can_change_description')).toBe(true);
  });

  it('не-owner без rankPerms всегда не имеет прав', () => {
    expect(hasPerm(false, 'can_manage_ranks', null)).toBe(false);
    expect(hasPerm(false, 'can_kick', undefined)).toBe(false);
    expect(hasPerm(false, 'can_invite', {})).toBe(false);
  });

  it('не-owner с явным разрешением получает true', () => {
    expect(hasPerm(false, 'can_kick', { can_kick: true })).toBe(true);
    expect(
      hasPerm(false, 'can_send_global_mail', { can_send_global_mail: true }),
    ).toBe(true);
  });

  it('не-owner с явным запретом получает false', () => {
    expect(hasPerm(false, 'can_kick', { can_kick: false })).toBe(false);
  });
});

describe('PERMISSION_KEYS', () => {
  it('содержит все 7 ключей backend (план 67 D-014)', () => {
    expect(PERMISSION_KEYS).toEqual([
      'can_invite',
      'can_kick',
      'can_send_global_mail',
      'can_manage_diplomacy',
      'can_change_description',
      'can_propose_relations',
      'can_manage_ranks',
    ]);
    expect(PERMISSION_KEYS.length).toBe(7);
  });
});

describe('relationStatusKey', () => {
  it('маппит 5 enum-статусов backend на ключи bundle alliance:', () => {
    expect(relationStatusKey('protection')).toBe('protection');
    expect(relationStatusKey('confederation')).toBe('confederation');
    expect(relationStatusKey('war')).toBe('war');
    expect(relationStatusKey('trade')).toBe('tradeAgreement');
    expect(relationStatusKey('ceasefire')).toBe('ceasefire');
  });

  it('возвращает status для неизвестного значения', () => {
    expect(relationStatusKey('unknown')).toBe('status');
    expect(relationStatusKey('')).toBe('status');
  });
});
