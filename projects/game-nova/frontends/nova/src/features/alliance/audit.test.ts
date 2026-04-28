import { describe, it, expect } from 'vitest';
import {
  AUDIT_ACTIONS,
  isKnownAction,
  isKnownTargetKind,
  auditActionLabelKey,
  auditTargetKindLabelKey,
  formatRelativeTime,
} from './audit';

describe('AUDIT_ACTIONS', () => {
  it('содержит ровно 18 known actions (зеркало backend audit.go)', () => {
    expect(AUDIT_ACTIONS).toHaveLength(18);
    // спот-чек ключевых event-ов из плана
    expect(AUDIT_ACTIONS).toContain('alliance_created');
    expect(AUDIT_ACTIONS).toContain('leadership_transferred');
    expect(AUDIT_ACTIONS).toContain('relation_proposed');
    expect(AUDIT_ACTIONS).toContain('member_kicked');
  });

  it('все имена в snake_case', () => {
    for (const a of AUDIT_ACTIONS) {
      expect(a).toMatch(/^[a-z]+(_[a-z]+)*$/);
    }
  });
});

describe('isKnownAction', () => {
  it('true для известных', () => {
    expect(isKnownAction('alliance_created')).toBe(true);
    expect(isKnownAction('leadership_transferred')).toBe(true);
  });
  it('false для неизвестных', () => {
    expect(isKnownAction('foo_bar')).toBe(false);
    expect(isKnownAction('')).toBe(false);
    expect(isKnownAction('Alliance_Created')).toBe(false); // case-sensitive
  });
});

describe('auditActionLabelKey', () => {
  it('для известного action — конкретный ключ', () => {
    expect(auditActionLabelKey('alliance_created')).toBe('audit.action.alliance_created');
  });
  it('для неизвестного — fallback unknown', () => {
    expect(auditActionLabelKey('something_new')).toBe('audit.action.unknown');
    expect(auditActionLabelKey('')).toBe('audit.action.unknown');
  });
});

describe('isKnownTargetKind / auditTargetKindLabelKey', () => {
  it('известные target_kind', () => {
    for (const k of ['member', 'rank', 'alliance', 'relation']) {
      expect(isKnownTargetKind(k)).toBe(true);
      expect(auditTargetKindLabelKey(k)).toBe(`audit.targetKind.${k}`);
    }
  });
  it('неизвестные → fallback unknown', () => {
    expect(isKnownTargetKind('mailbox')).toBe(false);
    expect(auditTargetKindLabelKey('mailbox')).toBe('audit.targetKind.unknown');
  });
});

describe('formatRelativeTime', () => {
  const now = new Date('2026-04-28T12:00:00Z');

  it('< 60 сек → justNow', () => {
    const then = new Date(now.getTime() - 30 * 1000);
    expect(formatRelativeTime(now, then)).toEqual({ key: 'audit.time.justNow', vars: { n: '' } });
  });

  it('минуты', () => {
    const then = new Date(now.getTime() - 5 * 60 * 1000);
    expect(formatRelativeTime(now, then)).toEqual({ key: 'audit.time.mAgo', vars: { n: '5' } });
  });

  it('часы', () => {
    const then = new Date(now.getTime() - 3 * 60 * 60 * 1000);
    expect(formatRelativeTime(now, then)).toEqual({ key: 'audit.time.hAgo', vars: { n: '3' } });
  });

  it('дни', () => {
    const then = new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000);
    expect(formatRelativeTime(now, then)).toEqual({ key: 'audit.time.dAgo', vars: { n: '5' } });
  });

  it('месяцы (~30 дней округление)', () => {
    const then = new Date(now.getTime() - 60 * 24 * 60 * 60 * 1000);
    expect(formatRelativeTime(now, then)).toEqual({ key: 'audit.time.moAgo', vars: { n: '2' } });
  });

  it('будущее (часы из будущего отрицательны) → justNow (clamp на 0)', () => {
    const future = new Date(now.getTime() + 60 * 60 * 1000);
    expect(formatRelativeTime(now, future)).toEqual({ key: 'audit.time.justNow', vars: { n: '' } });
  });

  it('граничные значения 60 сек / 60 мин / 24 ч — округление вниз корректно', () => {
    expect(formatRelativeTime(now, new Date(now.getTime() - 60 * 1000)))
      .toEqual({ key: 'audit.time.mAgo', vars: { n: '1' } });
    expect(formatRelativeTime(now, new Date(now.getTime() - 60 * 60 * 1000)))
      .toEqual({ key: 'audit.time.hAgo', vars: { n: '1' } });
    expect(formatRelativeTime(now, new Date(now.getTime() - 24 * 60 * 60 * 1000)))
      .toEqual({ key: 'audit.time.dAgo', vars: { n: '1' } });
  });
});
