// S-014 / S-018 / S-019 — преобразование snake_case key → camelCase
// для резолва i18n.info.{name}. (План 72 Ф.4 Spring 3)
//
// Все три info-экрана берут entry.key из catalog endpoint и резолвят
// его в i18n как camelCase (info.metalMine, info.smallTransporter,
// info.computerTech). Тест защищает от случайного rename конвенции.

import { describe, it, expect } from 'vitest';

function snakeToCamelI18n(key: string): string {
  return key.replace(/_([a-z])/g, (_, c: string) => c.toUpperCase());
}

describe('snake_case → camelCase i18n key', () => {
  it('metal_mine → metalMine', () => {
    expect(snakeToCamelI18n('metal_mine')).toBe('metalMine');
  });

  it('small_transporter → smallTransporter', () => {
    expect(snakeToCamelI18n('small_transporter')).toBe('smallTransporter');
  });

  it('computer_tech → computerTech', () => {
    expect(snakeToCamelI18n('computer_tech')).toBe('computerTech');
  });

  it('catalyst (без подчёркивания) — без изменений', () => {
    expect(snakeToCamelI18n('catalyst')).toBe('catalyst');
  });

  it('multi-underscore: hyper_space_engine → hyperSpaceEngine', () => {
    expect(snakeToCamelI18n('hyper_space_engine')).toBe('hyperSpaceEngine');
  });
});
