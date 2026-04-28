// Тесты на utility-функции API-модуля alliance (план 72 Ф.3 Spring 2 ч.1).
//
// Сетевые вызовы здесь не покрываем (они тонкие обёртки над api.* и
// проверяются интеграционно). Тестируем сериализацию query-параметров —
// это место, где легко словить регрессию (порядок ключей, кодирование,
// отсутствие undefined).

import { describe, it, expect } from 'vitest';
import {
  buildAuditQuery,
  buildSearchQuery,
} from '@/features/alliance/queries';

describe('buildSearchQuery', () => {
  it('пустой фильтр → пустая строка', () => {
    expect(buildSearchQuery({})).toBe('');
  });

  it('q обрезается, but не нормализуется', () => {
    expect(buildSearchQuery({ q: '  Strikers  ' })).toBe('q=Strikers');
  });

  it('пустой/whitespace q не попадает в query', () => {
    expect(buildSearchQuery({ q: '   ' })).toBe('');
    expect(buildSearchQuery({ q: '' })).toBe('');
  });

  it('is_open пробрасывается как строка', () => {
    expect(buildSearchQuery({ is_open: true })).toBe('is_open=true');
    expect(buildSearchQuery({ is_open: false })).toBe('is_open=false');
  });

  it('диапазон + пагинация склеиваются', () => {
    const qs = buildSearchQuery({
      q: 'Wolf',
      is_open: true,
      min_members: 5,
      max_members: 50,
      limit: 25,
      offset: 50,
    });
    // URLSearchParams сохраняет порядок set() — фиксируем как контракт.
    expect(qs).toBe(
      'q=Wolf&is_open=true&min_members=5&max_members=50&limit=25&offset=50',
    );
  });
});

describe('buildAuditQuery', () => {
  it('пустой фильтр → пустая строка', () => {
    expect(buildAuditQuery({})).toBe('');
  });

  it('action + offset склеиваются в правильном порядке', () => {
    expect(buildAuditQuery({ action: 'member_kicked', offset: 100 })).toBe(
      'action=member_kicked&offset=100',
    );
  });

  it('все поля', () => {
    expect(
      buildAuditQuery({
        action: 'leadership_transferred',
        actor_id: 'aaa',
        limit: 50,
        offset: 0,
      }),
    ).toBe('action=leadership_transferred&actor_id=aaa&limit=50&offset=0');
  });
});
