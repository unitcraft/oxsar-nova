// S-001 Main / S-007 Empire — формат ресурсов и координат.
// (План 72 Ф.2 Spring 1)
//
// Покрытие: formatNumber (для метал/кремний/водород на Main + Empire),
// formatCoords (для шапки Main, для строк Empire), formatDuration
// (для секций «События» Main + очереди Constructions/Research/Shipyard).
//
// Pixel-perfect соответствие легаси (план 72.1 §20.12):
//   • formatNumber — разделитель тысяч точка `.` (legacy ru: `409.600`,
//     `4.500.000`); legacy `Functions.inc.php::fNumber` =
//     number_format($n, 0, DECIMAL_POINT, THOUSANDS_SEPERATOR='.').
//   • formatDuration — `HH:MM:SS` (с днями `Nд HH:MM:SS`), как в
//     legacy resource.tpl/required_res_table (на скрине Время «00:06:23»).

import { describe, it, expect } from 'vitest';
import {
  formatNumber,
  formatCoords,
  formatDuration,
  secondsUntil,
} from './format';

describe('formatNumber', () => {
  it('тысячи отделяет точкой (legacy ru: 409.600 / 4.500.000)', () => {
    expect(formatNumber(1234567)).toBe('1.234.567');
    expect(formatNumber(123)).toBe('123');
    expect(formatNumber(0)).toBe('0');
  });

  it('отрицательные числа сохраняют знак', () => {
    expect(formatNumber(-1500)).toBe('-1.500');
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
  it('секунды только → 00:00:SS', () => {
    expect(formatDuration(45)).toBe('00:00:45');
  });
  it('минуты + секунды → 00:MM:SS', () => {
    expect(formatDuration(125)).toBe('00:02:05');
  });
  it('часы + минуты + секунды → HH:MM:SS', () => {
    expect(formatDuration(3725)).toBe('01:02:05');
  });
  it('дни выводятся когда > 86400с → Nд HH:MM:SS', () => {
    expect(formatDuration(90000)).toBe('1д 01:00:00');
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
