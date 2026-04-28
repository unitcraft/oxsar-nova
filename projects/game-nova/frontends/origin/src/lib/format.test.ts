// S-001 Main / S-007 Empire — формат ресурсов и координат.
// (План 72 Ф.2 Spring 1)
//
// Покрытие: formatNumber (для метал/кремний/водород на Main + Empire),
// formatCoords (для шапки Main, для строк Empire), formatDuration
// (для секций «События» Main + очереди Constructions/Research/Shipyard).

import { describe, it, expect } from 'vitest';
import {
  formatNumber,
  formatCoords,
  formatDuration,
  secondsUntil,
} from './format';

const NBSP = ' ';

describe('formatNumber', () => {
  it('тысячи отделяет неразрывным пробелом (NBSP)', () => {
    expect(formatNumber(1234567)).toBe(`1${NBSP}234${NBSP}567`);
    expect(formatNumber(123)).toBe('123');
    expect(formatNumber(0)).toBe('0');
  });

  it('отрицательные числа сохраняют знак', () => {
    expect(formatNumber(-1500)).toBe(`-1${NBSP}500`);
  });

  it('NaN и Infinity → "—"', () => {
    expect(formatNumber(Number.NaN)).toBe('—');
    expect(formatNumber(Number.POSITIVE_INFINITY)).toBe('—');
  });

  it('отбрасывает дробную часть (legacy fNumber поведение)', () => {
    expect(formatNumber(123.9)).toBe('123');
  });
});

describe('formatCoords', () => {
  it('форматирует [g:s:p] как в legacy', () => {
    expect(formatCoords(1, 42, 7)).toBe('[1:42:7]');
  });
});

describe('formatDuration', () => {
  it('секунды только', () => {
    expect(formatDuration(45)).toBe('45с');
  });
  it('минуты + секунды', () => {
    expect(formatDuration(125)).toBe('2м 5с');
  });
  it('часы + минуты + секунды', () => {
    expect(formatDuration(3725)).toBe('1ч 2м 5с');
  });
  it('дни выводятся когда > 86400с', () => {
    expect(formatDuration(90000)).toBe('1д 1ч 0м 0с');
  });
  it('некорректные данные → "—"', () => {
    expect(formatDuration(-5)).toBe('—');
    expect(formatDuration(Number.NaN)).toBe('—');
  });
});

describe('secondsUntil', () => {
  it('будущее ISO → положительное', () => {
    const future = new Date(Date.now() + 5_000).toISOString();
    const sec = secondsUntil(future);
    expect(sec).toBeGreaterThan(0);
    expect(sec).toBeLessThanOrEqual(5);
  });

  it('прошлое ISO → 0', () => {
    const past = new Date(Date.now() - 10_000).toISOString();
    expect(secondsUntil(past)).toBe(0);
  });

  it('некорректная строка → 0', () => {
    expect(secondsUntil('not-a-date')).toBe(0);
  });
});
