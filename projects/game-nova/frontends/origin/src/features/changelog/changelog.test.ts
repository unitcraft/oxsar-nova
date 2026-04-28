import { describe, it, expect } from 'vitest';
import { parseChangelog } from './parse';

const SAMPLE = `# История обновлений

## 1.0.0 — 2026-05-01

- Релиз.

## 0.9.0 — 2026-04-28

- Beta.
- Фикс багов.
`;

describe('changelog', () => {
  it('парсит ## заголовки в releases', () => {
    const r = parseChangelog(SAMPLE);
    expect(r).toHaveLength(2);
    expect(r[0]?.version).toBe('1.0.0 — 2026-05-01');
    expect(r[1]?.version).toBe('0.9.0 — 2026-04-28');
  });

  it('тело между заголовками собирается в changes', () => {
    const r = parseChangelog(SAMPLE);
    expect(r[1]?.changes).toContain('- Beta.');
    expect(r[1]?.changes).toContain('- Фикс багов.');
  });

  it('пустой markdown → пустой список', () => {
    expect(parseChangelog('# нет релизов')).toEqual([]);
  });
});
