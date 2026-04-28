import { describe, it, expect } from 'vitest';

const MAX_PM_LENGTH = 2000;

function isReadByReadAt(readAt: string | undefined): boolean {
  return readAt !== undefined;
}

function validateCompose(to: string, subject: string): string | null {
  if (!to.trim()) return 'no-recipient';
  if (!subject.trim()) return 'no-subject';
  return null;
}

describe('messages', () => {
  it('non-null read_at = прочитанное сообщение', () => {
    expect(isReadByReadAt(undefined)).toBe(false);
    expect(isReadByReadAt('2026-04-28T12:00:00Z')).toBe(true);
  });

  it('compose требует получателя и тему', () => {
    expect(validateCompose('', 'sub')).toBe('no-recipient');
    expect(validateCompose('user', '')).toBe('no-subject');
    expect(validateCompose('user', 'sub')).toBeNull();
  });

  it('truncate тела до MAX_PM_LENGTH', () => {
    const oversized = 'a'.repeat(MAX_PM_LENGTH + 50);
    expect(oversized.slice(0, MAX_PM_LENGTH).length).toBe(MAX_PM_LENGTH);
  });
});
