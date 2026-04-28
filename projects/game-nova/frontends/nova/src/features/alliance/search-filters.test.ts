import { describe, it, expect } from 'vitest';
import {
  EMPTY_FILTERS,
  hasActiveFilters,
  buildSearchQuery,
  type AllianceSearchFilters,
} from './search-filters';

const f = (over: Partial<AllianceSearchFilters>): AllianceSearchFilters => ({
  ...EMPTY_FILTERS,
  ...over,
});

describe('EMPTY_FILTERS', () => {
  it('все поля пустые / all', () => {
    expect(EMPTY_FILTERS).toEqual({ q: '', isOpen: 'all', minMembers: '', maxMembers: '' });
  });
});

describe('hasActiveFilters', () => {
  it('false для EMPTY_FILTERS', () => {
    expect(hasActiveFilters(EMPTY_FILTERS)).toBe(false);
  });
  it('true когда q непустой (после trim)', () => {
    expect(hasActiveFilters(f({ q: 'foo' }))).toBe(true);
    expect(hasActiveFilters(f({ q: '   ' }))).toBe(false); // только пробелы — не активный фильтр
  });
  it('true для isOpen=open|closed', () => {
    expect(hasActiveFilters(f({ isOpen: 'open' }))).toBe(true);
    expect(hasActiveFilters(f({ isOpen: 'closed' }))).toBe(true);
  });
  it('true для непустого min/maxMembers', () => {
    expect(hasActiveFilters(f({ minMembers: '5' }))).toBe(true);
    expect(hasActiveFilters(f({ maxMembers: '100' }))).toBe(true);
  });
});

describe('buildSearchQuery', () => {
  it('EMPTY → пустая строка', () => {
    expect(buildSearchQuery(EMPTY_FILTERS)).toBe('');
  });

  it('q триммится и URL-кодируется', () => {
    expect(buildSearchQuery(f({ q: '  Strikers  ' }))).toBe('q=Strikers');
    expect(buildSearchQuery(f({ q: 'a b' }))).toBe('q=a+b');
  });

  it('пустой q после trim → не отправляется', () => {
    expect(buildSearchQuery(f({ q: '   ' }))).toBe('');
  });

  it('isOpen → is_open=true|false; all → нет ключа', () => {
    expect(buildSearchQuery(f({ isOpen: 'open' }))).toBe('is_open=true');
    expect(buildSearchQuery(f({ isOpen: 'closed' }))).toBe('is_open=false');
    expect(buildSearchQuery(f({ isOpen: 'all' }))).toBe('');
  });

  it('min/maxMembers — валидные числа', () => {
    expect(buildSearchQuery(f({ minMembers: '5', maxMembers: '50' }))).toBe('min_members=5&max_members=50');
  });

  it('невалидные min/maxMembers игнорируются (negative, float, garbage)', () => {
    expect(buildSearchQuery(f({ minMembers: '-1' }))).toBe('');
    expect(buildSearchQuery(f({ minMembers: '3.14' }))).toBe('');
    expect(buildSearchQuery(f({ minMembers: 'abc' }))).toBe('');
    expect(buildSearchQuery(f({ minMembers: '' }))).toBe('');
  });

  it('частично заполненный диапазон — отправляется только заполненная сторона', () => {
    expect(buildSearchQuery(f({ minMembers: '10' }))).toBe('min_members=10');
    expect(buildSearchQuery(f({ maxMembers: '100' }))).toBe('max_members=100');
  });

  it('комбинированный фильтр — все известные ключи', () => {
    const q = buildSearchQuery(
      f({ q: 'pvp', isOpen: 'open', minMembers: '5', maxMembers: '50' }),
    );
    // Порядок query-string фиксирован URLSearchParams — q,is_open,min,max.
    expect(q).toBe('q=pvp&is_open=true&min_members=5&max_members=50');
  });
});
