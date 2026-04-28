// Alliance search filters (план 67 Ф.5 ч.2, U-012).
//
// Backend: GET /api/alliances?q=&is_open=&min_members=&max_members=&limit=&offset=
// Все параметры опциональны; пустые не отправляются.

export interface AllianceSearchFilters {
  q: string;
  isOpen: 'all' | 'open' | 'closed';
  minMembers: string; // строка — input хранит сырой ввод; пустая = не задано
  maxMembers: string;
}

export const EMPTY_FILTERS: AllianceSearchFilters = {
  q: '',
  isOpen: 'all',
  minMembers: '',
  maxMembers: '',
};

// hasActiveFilters — true если пользователь хоть что-то ввёл; нужно
// для UI-индикатора и принятия решения о cache-key (debounce-defer).
export function hasActiveFilters(f: AllianceSearchFilters): boolean {
  return (
    f.q.trim().length > 0 ||
    f.isOpen !== 'all' ||
    f.minMembers.trim().length > 0 ||
    f.maxMembers.trim().length > 0
  );
}

// buildSearchQuery — собирает query-string. Невалидные числа
// игнорируются (не отправляются), пустые поля — тоже. Это упрощает
// UX: можно частично заполнить min без max.
export function buildSearchQuery(f: AllianceSearchFilters): string {
  const params = new URLSearchParams();
  const q = f.q.trim();
  if (q) params.set('q', q);
  if (f.isOpen === 'open') params.set('is_open', 'true');
  if (f.isOpen === 'closed') params.set('is_open', 'false');
  const min = parseNonNegativeInt(f.minMembers);
  if (min !== null) params.set('min_members', String(min));
  const max = parseNonNegativeInt(f.maxMembers);
  if (max !== null) params.set('max_members', String(max));
  return params.toString();
}

function parseNonNegativeInt(s: string): number | null {
  const t = s.trim();
  if (!t) return null;
  if (!/^\d+$/.test(t)) return null;
  const n = Number.parseInt(t, 10);
  if (!Number.isFinite(n) || n < 0) return null;
  return n;
}
