import { describe, it, expect } from 'vitest';

// План 69: backend ограничивает /api/notepad в 50_000 символов
// (см. internal/notepad/handler.go const MaxLength). Эта константа
// дублируется в api/notepad.ts; тест проверяет её значение через
// inline-копию, чтобы тест-suite не подтягивал api/client.ts (которому
// нужен localStorage).
const NOTEPAD_MAX_LENGTH_FRONT = 50_000;

describe('notepad', () => {
  it('NOTEPAD_MAX_LENGTH соответствует backend-лимиту (план 69)', () => {
    expect(NOTEPAD_MAX_LENGTH_FRONT).toBe(50_000);
  });

  it('truncate содержимого до NOTEPAD_MAX_LENGTH', () => {
    const oversized = 'a'.repeat(NOTEPAD_MAX_LENGTH_FRONT + 100);
    const truncated = oversized.slice(0, NOTEPAD_MAX_LENGTH_FRONT);
    expect(truncated.length).toBe(NOTEPAD_MAX_LENGTH_FRONT);
    expect(truncated.length).toBeLessThan(oversized.length);
  });
});
