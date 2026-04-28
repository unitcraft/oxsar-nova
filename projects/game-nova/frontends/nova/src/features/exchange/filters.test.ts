import { describe, it, expect } from 'vitest';
import {
  EMPTY_FILTERS,
  EXPIRES_OPTIONS,
  MAX_QUANTITY_PER_LOT,
  buildQueryParams,
  errorMessageKey,
  hasActiveFilters,
  validateCreateLot,
  validatePriceRange,
  type ExchangeFilters,
} from './filters';

const f = (over: Partial<ExchangeFilters> = {}): ExchangeFilters => ({
  ...EMPTY_FILTERS,
  ...over,
});

describe('EMPTY_FILTERS', () => {
  it('default status=active, всё остальное пусто', () => {
    expect(EMPTY_FILTERS).toEqual({
      artifactUnitId: null,
      minPrice: '',
      maxPrice: '',
      sellerId: '',
      status: 'active',
    });
  });
});

describe('hasActiveFilters', () => {
  it('false для дефолта', () => {
    expect(hasActiveFilters(EMPTY_FILTERS)).toBe(false);
  });

  it('true при выбранном артефакте', () => {
    expect(hasActiveFilters(f({ artifactUnitId: 300 }))).toBe(true);
  });

  it('true при сменённом статусе', () => {
    expect(hasActiveFilters(f({ status: 'all' }))).toBe(true);
    expect(hasActiveFilters(f({ status: 'sold' }))).toBe(true);
  });

  it('игнорирует пробельные строки в minPrice/sellerId', () => {
    expect(hasActiveFilters(f({ minPrice: '   ' }))).toBe(false);
    expect(hasActiveFilters(f({ sellerId: '\t' }))).toBe(false);
  });
});

describe('buildQueryParams', () => {
  it('дефолт → status=active + limit=50', () => {
    const p = buildQueryParams(EMPTY_FILTERS);
    expect(p.get('status')).toBe('active');
    expect(p.get('limit')).toBe('50');
    // Никаких лишних ключей.
    expect(p.get('artifact_unit_id')).toBeNull();
    expect(p.get('min_price')).toBeNull();
    expect(p.get('max_price')).toBeNull();
    expect(p.get('seller_id')).toBeNull();
    expect(p.get('cursor')).toBeNull();
  });

  it('artifact_unit_id, min/max и sellerId передаются', () => {
    const p = buildQueryParams(f({
      artifactUnitId: 305,
      minPrice: '100',
      maxPrice: '5000',
      sellerId: 'aaaa-bbbb',
    }));
    expect(p.get('artifact_unit_id')).toBe('305');
    expect(p.get('min_price')).toBe('100');
    expect(p.get('max_price')).toBe('5000');
    expect(p.get('seller_id')).toBe('aaaa-bbbb');
  });

  it('status=all → параметр status опускается (бэк вернёт всё)', () => {
    const p = buildQueryParams(f({ status: 'all' }));
    expect(p.get('status')).toBeNull();
  });

  it('cursor подставляется когда передан', () => {
    const p = buildQueryParams(EMPTY_FILTERS, 'opaque-cursor-123', 25);
    expect(p.get('cursor')).toBe('opaque-cursor-123');
    expect(p.get('limit')).toBe('25');
  });

  it('игнорирует нечисловые/нулевые цены', () => {
    const p = buildQueryParams(f({ minPrice: 'abc', maxPrice: '0' }));
    expect(p.get('min_price')).toBeNull();
    expect(p.get('max_price')).toBeNull();
  });
});

describe('validatePriceRange', () => {
  it('null когда обе пусты или валидны', () => {
    expect(validatePriceRange(EMPTY_FILTERS)).toBeNull();
    expect(validatePriceRange(f({ minPrice: '100', maxPrice: '5000' }))).toBeNull();
  });

  it('priceNegative при отрицательной цене', () => {
    expect(validatePriceRange(f({ minPrice: '-10' }))).toBe('priceNegative');
    expect(validatePriceRange(f({ maxPrice: '-1' }))).toBe('priceNegative');
  });

  it('priceRangeInvalid если min > max', () => {
    expect(validatePriceRange(f({ minPrice: '5000', maxPrice: '100' }))).toBe('priceRangeInvalid');
  });

  it('равные min/max — валидно (точное совпадение цены)', () => {
    expect(validatePriceRange(f({ minPrice: '500', maxPrice: '500' }))).toBeNull();
  });
});

describe('validateCreateLot', () => {
  const base = {
    artifactUnitId: 300 as number | null,
    quantity: 5,
    available: 10,
    priceOxsarit: 1000,
    expiresInHours: 24,
  };

  it('null когда всё валидно', () => {
    expect(validateCreateLot(base)).toBeNull();
  });

  it('pickArtefact когда артефакт не выбран', () => {
    expect(validateCreateLot({ ...base, artifactUnitId: null })).toBe('pickArtefact');
  });

  it('qtyMin при quantity < 1', () => {
    expect(validateCreateLot({ ...base, quantity: 0 })).toBe('qtyMin');
    expect(validateCreateLot({ ...base, quantity: -3 })).toBe('qtyMin');
  });

  it('qtyMax при превышении 100 (= MAX_QUANTITY_PER_LOT)', () => {
    expect(MAX_QUANTITY_PER_LOT).toBe(100);
    expect(validateCreateLot({ ...base, quantity: 101, available: 200 })).toBe('qtyMax');
  });

  it('qtyOverAvailable когда qty > available', () => {
    expect(validateCreateLot({ ...base, quantity: 11, available: 10 })).toBe('qtyOverAvailable');
  });

  it('priceMin при нулевой/отрицательной цене', () => {
    expect(validateCreateLot({ ...base, priceOxsarit: 0 })).toBe('priceMin');
    expect(validateCreateLot({ ...base, priceOxsarit: -5 })).toBe('priceMin');
  });

  it('expiresInvalid для произвольного TTL вне списка опций', () => {
    expect(validateCreateLot({ ...base, expiresInHours: 12 })).toBe('expiresInvalid');
    expect(validateCreateLot({ ...base, expiresInHours: 0 })).toBe('expiresInvalid');
  });

  it('каждая EXPIRES_OPTIONS — валидный TTL', () => {
    for (const opt of EXPIRES_OPTIONS) {
      expect(validateCreateLot({ ...base, expiresInHours: opt.hours })).toBeNull();
    }
  });
});

describe('errorMessageKey', () => {
  it('известные коды бэкенда мапятся в ключи i18n exchange.errors.*', () => {
    expect(errorMessageKey('insufficient_oxsarits')).toBe('insufficientOxsarits');
    expect(errorMessageKey('price_cap_exceeded')).toBe('priceCapExceeded');
    expect(errorMessageKey('lot_not_active')).toBe('lotNotActive');
    expect(errorMessageKey('cannot_buy_own_lot')).toBe('cannotBuyOwnLot');
    expect(errorMessageKey('max_active_lots')).toBe('maxActiveLots');
    expect(errorMessageKey('max_quantity')).toBe('maxQuantity');
    expect(errorMessageKey('permit_required')).toBe('permitRequired');
  });

  it('неизвестный код → generic', () => {
    expect(errorMessageKey('unknown_code')).toBe('generic');
    expect(errorMessageKey(undefined)).toBe('generic');
  });
});
