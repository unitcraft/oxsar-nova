import { describe, it, expect } from 'vitest';

// settings-валидация локально на фронте: email, password.
function validateEmail(s: string): boolean {
  return s.includes('@') && s.length >= 3 && s.toLowerCase() === s.trim().toLowerCase();
}

function validatePassword(newPw: string, confirm: string): string | null {
  if (newPw !== confirm) return 'mismatch';
  if (newPw.length < 8) return 'too-short';
  return null;
}

describe('settings', () => {
  it('email с @ и >= 3 символа проходит', () => {
    expect(validateEmail('a@b.c')).toBe(true);
  });

  it('email без @ не проходит', () => {
    expect(validateEmail('abc.com')).toBe(false);
  });

  it('пароль < 8 символов отклоняется', () => {
    expect(validatePassword('short', 'short')).toBe('too-short');
  });

  it('пароли не совпадают — mismatch', () => {
    expect(validatePassword('longenough', 'longeNough')).toBe('mismatch');
  });

  it('валидный пароль — null', () => {
    expect(validatePassword('longenough', 'longenough')).toBeNull();
  });
});
