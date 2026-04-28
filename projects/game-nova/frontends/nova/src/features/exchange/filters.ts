// Логика фильтров биржи артефактов (план 76).
// Чистый модуль без React — тестируется через Vitest (см. filters.test.ts).
// Бэкенд (план 68) принимает фильтры в query-string GET /api/exchange/lots:
//   artifact_unit_id, min_price, max_price, seller_id, status, cursor, limit.

export type LotStatus = 'active' | 'sold' | 'cancelled' | 'expired';
export type StatusFilter = LotStatus | 'all';

export interface ExchangeFilters {
  artifactUnitId: number | null;
  minPrice: string;
  maxPrice: string;
  sellerId: string;
  status: StatusFilter;
}

export const EMPTY_FILTERS: ExchangeFilters = {
  artifactUnitId: null,
  minPrice: '',
  maxPrice: '',
  sellerId: '',
  status: 'active',
};

// hasActiveFilters — true, если хоть один фильтр отличается от дефолта.
// Дефолт: status=active, остальное пусто.
export function hasActiveFilters(f: ExchangeFilters): boolean {
  if (f.artifactUnitId !== null) return true;
  if (f.minPrice.trim() !== '') return true;
  if (f.maxPrice.trim() !== '') return true;
  if (f.sellerId.trim() !== '') return true;
  if (f.status !== 'active') return true;
  return false;
}

// buildQueryParams — превращает фильтры в URLSearchParams для GET-запроса.
// Пропускает пустые/дефолтные значения, числа парсит безопасно.
export function buildQueryParams(f: ExchangeFilters, cursor?: string, limit = 50): URLSearchParams {
  const p = new URLSearchParams();
  if (f.artifactUnitId !== null) p.set('artifact_unit_id', String(f.artifactUnitId));
  const min = parseIntOrNull(f.minPrice);
  const max = parseIntOrNull(f.maxPrice);
  if (min !== null && min > 0) p.set('min_price', String(min));
  if (max !== null && max > 0) p.set('max_price', String(max));
  const seller = f.sellerId.trim();
  if (seller !== '') p.set('seller_id', seller);
  // status=all — backend параметр не передаём (вернёт все статусы).
  if (f.status !== 'all') p.set('status', f.status);
  if (cursor) p.set('cursor', cursor);
  p.set('limit', String(limit));
  return p;
}

function parseIntOrNull(s: string): number | null {
  const trimmed = s.trim();
  if (trimmed === '') return null;
  const n = Number(trimmed);
  if (!Number.isFinite(n)) return null;
  return Math.floor(n);
}

// validatePriceRange — клиентская проверка диапазона цен.
// Возвращает код ошибки i18n (exchange.validation.*) или null.
export function validatePriceRange(f: ExchangeFilters): string | null {
  const min = parseIntOrNull(f.minPrice);
  const max = parseIntOrNull(f.maxPrice);
  if (min !== null && min < 0) return 'priceNegative';
  if (max !== null && max < 0) return 'priceNegative';
  if (min !== null && max !== null && min > max) return 'priceRangeInvalid';
  return null;
}

// EXPIRES_OPTIONS — допустимые TTL лота, выровнено по плану 68
// (handler принимает 1..168 часов).
export const EXPIRES_OPTIONS: ReadonlyArray<{ hours: number; tKey: string }> = [
  { hours: 1,   tKey: 'expires1h' },
  { hours: 6,   tKey: 'expires6h' },
  { hours: 24,  tKey: 'expires1d' },
  { hours: 72,  tKey: 'expires3d' },
  { hours: 168, tKey: 'expires7d' },
];

// MAX_QUANTITY_PER_LOT и MAX_ACTIVE_LOTS — клиентские зеркала backend-лимитов
// (план 68 §Ф.6: balance config). Используются для inline-валидации формы;
// в любом случае backend верифицирует и вернёт 422.
export const MAX_QUANTITY_PER_LOT = 100;
export const MAX_ACTIVE_LOTS = 10;

// validateCreateLot — синхронная проверка формы перед POST.
// Возвращает первый код ошибки или null.
export interface CreateLotInput {
  artifactUnitId: number | null;
  quantity: number;
  available: number;
  priceOxsarit: number;
  expiresInHours: number;
}

export function validateCreateLot(in_: CreateLotInput): string | null {
  if (in_.artifactUnitId === null) return 'pickArtefact';
  if (!Number.isFinite(in_.quantity) || in_.quantity < 1) return 'qtyMin';
  if (in_.quantity > MAX_QUANTITY_PER_LOT) return 'qtyMax';
  if (in_.quantity > in_.available) return 'qtyOverAvailable';
  if (!Number.isFinite(in_.priceOxsarit) || in_.priceOxsarit < 1) return 'priceMin';
  const opt = EXPIRES_OPTIONS.find((o) => o.hours === in_.expiresInHours);
  if (!opt) return 'expiresInvalid';
  return null;
}

// errorMessageKey — маппинг backend error.code → i18n exchange.errors.*
// Ключи уже добавлены планом 68 в configs/i18n/{ru,en}.yml.
export function errorMessageKey(code: string | undefined): string {
  switch (code) {
    case 'insufficient_artefacts': return 'insufficientArtefacts';
    case 'insufficient_oxsarits':  return 'insufficientOxsarits';
    case 'price_cap_exceeded':     return 'priceCapExceeded';
    case 'permit_required':        return 'permitRequired';
    case 'lot_not_active':         return 'lotNotActive';
    case 'not_a_seller':           return 'notASeller';
    case 'cannot_buy_own_lot':     return 'cannotBuyOwnLot';
    case 'max_active_lots':        return 'maxActiveLots';
    case 'max_quantity':           return 'maxQuantity';
    case 'invalid_expiry':         return 'invalidExpiry';
    default:                       return 'generic';
  }
}
