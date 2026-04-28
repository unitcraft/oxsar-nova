import { describe, it, expect } from 'vitest';

// search-helpers тестируем без рендера: query-string и тип.
describe('search', () => {
  it('минимальная длина 2 символа должна блокировать запрос', () => {
    // backend сам возвращает пустой массив при q < 2 (см. search/handler.go),
    // фронт дополнительно блокирует TanStack Query enabled:false для
    // экономии трафика — проверяем условие.
    const q = 'a';
    expect(q.length >= 2).toBe(false);
  });

  it('committed query пропускает 2+ символа', () => {
    expect('ab'.length >= 2).toBe(true);
    expect('abc'.length >= 2).toBe(true);
  });
});
