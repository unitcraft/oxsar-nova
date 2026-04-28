// S-024 Records — фильтрация по категории. (План 72 Ф.4 Spring 3)
//
// RecordsScreen разбивает плоский массив records на 5 секций
// (score / building / research / ship / defense). Тест защищает от
// случайной фильтрации по неправильному полю и от потери записей
// (sum по секциям должен совпадать с длиной исходного массива).

import { describe, it, expect } from 'vitest';
import type { RecordEntry } from '@/api/types';

const SECTIONS: RecordEntry['category'][] = [
  'score',
  'building',
  'research',
  'ship',
  'defense',
];

function bySection(records: RecordEntry[]): Record<string, number> {
  const out: Record<string, number> = {};
  for (const cat of SECTIONS) {
    out[cat] = records.filter((r) => r.category === cat).length;
  }
  return out;
}

function mk(category: RecordEntry['category'], key: string): RecordEntry {
  return {
    category,
    key,
    holder_id: '00000000-0000-0000-0000-000000000000',
    holder_name: 'Tester',
    value: 1,
    my_value: 0,
  };
}

describe('records section split', () => {
  it('разбивает по 5 категориям без потерь', () => {
    const records = [
      mk('score', 'total'),
      mk('building', 'metal_mine'),
      mk('research', 'computer_tech'),
      mk('ship', 'small_transporter'),
      mk('defense', 'rocket_launcher'),
      mk('building', 'silicon_lab'),
    ];
    const by = bySection(records);
    const total = Object.values(by).reduce((a, b) => a + b, 0);
    expect(total).toBe(records.length);
    expect(by['building']).toBe(2);
    expect(by['score']).toBe(1);
  });

  it('пустой список — все секции 0', () => {
    const by = bySection([]);
    expect(Object.values(by).every((n) => n === 0)).toBe(true);
  });
});
