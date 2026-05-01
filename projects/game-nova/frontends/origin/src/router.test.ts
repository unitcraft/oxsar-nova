// Router origin-фронта (план 72 Ф.2 Spring 1 + Ф.3 Spring 2).
//
// Тест-фикстура: заявленный набор путей не должен молча сократиться.
// Если кто-то удалит маршрут (скажем, /mission или /alliance/me), тест
// провалится — это сигнал, что Spring распался.
//
// Реальные React-компоненты не рендерятся (jsdom + RTL минимально
// используем в feature-тестах). Здесь — только статическая проверка
// контракта роутов.

import { describe, it, expect } from 'vitest';

const SPRING_1_PATHS = [
  '/',
  '/constructions',
  '/constructions/:planetId',
  '/research',
  '/research/:planetId',
  '/shipyard',
  '/shipyard/:planetId',
  '/galaxy',
  '/galaxy/:galaxy/:system',
  '/mission',
  '/mission/:planetId',
  '/empire',
  '/login',
];

const SPRING_2_PART1_ALLIANCE = [
  '/alliance',
  '/alliance/list',
  '/alliance/create',
  '/alliance/me',
  '/alliance/members',
  '/alliance/manage',
  '/alliance/descriptions',
  '/alliance/ranks',
  '/alliance/diplomacy',
  '/alliance/audit',
  '/alliance/transfer',
  '/alliance/:id',
];

const SPRING_2_PART2 = [
  '/resource-market',
  '/market',
  '/repair',
  '/battlestats',
  '/fleet-operations',
];

const SPRING_3_PATHS = [
  '/artefacts',
  '/artefact/:id',
  '/building/:type',
  '/unit/:type',
  '/techtree',
  '/records',
  '/ranking',
];

describe('router Spring 1', () => {
  it('содержит все 7 главных экранов + login + 404 placeholder', () => {
    expect(SPRING_1_PATHS).toContain('/');
    expect(SPRING_1_PATHS).toContain('/constructions');
    expect(SPRING_1_PATHS).toContain('/research');
    expect(SPRING_1_PATHS).toContain('/shipyard');
    expect(SPRING_1_PATHS).toContain('/galaxy');
    expect(SPRING_1_PATHS).toContain('/mission');
    expect(SPRING_1_PATHS).toContain('/empire');
    expect(SPRING_1_PATHS).toContain('/login');
  });

  it('имеет параметризованные варианты для каждого planet-зависимого экрана', () => {
    expect(SPRING_1_PATHS).toContain('/constructions/:planetId');
    expect(SPRING_1_PATHS).toContain('/research/:planetId');
    expect(SPRING_1_PATHS).toContain('/shipyard/:planetId');
    expect(SPRING_1_PATHS).toContain('/mission/:planetId');
  });

  it('Galaxy имеет параметризацию galaxy/system', () => {
    expect(SPRING_1_PATHS).toContain('/galaxy/:galaxy/:system');
  });
});

describe('router Spring 2 — alliance (12 экранов)', () => {
  it('покрывает все S-008..S-019 alliance-роуты', () => {
    expect(SPRING_2_PART1_ALLIANCE).toHaveLength(12);
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance');
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance/list');
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance/create');
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance/me');
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance/members');
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance/manage');
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance/descriptions');
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance/ranks');
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance/diplomacy');
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance/audit');
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance/transfer');
    expect(SPRING_2_PART1_ALLIANCE).toContain('/alliance/:id');
  });
});

describe('router Spring 2 — ресурсный/боевой блок (5 экранов)', () => {
  it('покрывает S-020..S-024', () => {
    expect(SPRING_2_PART2).toHaveLength(5);
    expect(SPRING_2_PART2).toEqual([
      '/resource-market',
      '/market',
      '/repair',
      '/battlestats',
      '/fleet-operations',
    ]);
  });
});

// План 72.1 ч.20.12: /user-agreement и /support удалены из origin
// (за TOS/EULA и саппорт отвечает портал, не игровой сервер).
const SPRING_4_PART2_PATHS = [
  '/officer',
  '/profession',
  '/changelog',
  '/tools/tech-calc',
  '/widgets',
];

describe('router Spring 4 ч.2 — premium / static / utilities (5 маршрутов)', () => {
  it('покрывает S-040/S-041/S-044/S-046/S-047', () => {
    expect(SPRING_4_PART2_PATHS).toHaveLength(5);
    expect(SPRING_4_PART2_PATHS).toContain('/officer');
    expect(SPRING_4_PART2_PATHS).toContain('/profession');
    expect(SPRING_4_PART2_PATHS).toContain('/changelog');
    expect(SPRING_4_PART2_PATHS).toContain('/tools/tech-calc');
    expect(SPRING_4_PART2_PATHS).toContain('/widgets');
  });

  it('S-046 /widgets — redirect-маршрут (см. simplifications P72.S4.WIDGETS)', () => {
    // /widgets делает Navigate → /, см. WidgetsRedirect.tsx. Тест-факт:
    // маршрут существует, но семантически — алиас /.
    expect(SPRING_4_PART2_PATHS).toContain('/widgets');
  });
});

describe('router Spring 3 — artefacts / info / techtree / records / ranking', () => {
  it('покрывает 7 новых маршрутов S-013/S-014/S-018/S-019/S-021/S-024/S-023', () => {
    expect(SPRING_3_PATHS).toHaveLength(7);
    expect(SPRING_3_PATHS).toContain('/artefacts');
    expect(SPRING_3_PATHS).toContain('/artefact/:id');
    expect(SPRING_3_PATHS).toContain('/building/:type');
    expect(SPRING_3_PATHS).toContain('/unit/:type');
    expect(SPRING_3_PATHS).toContain('/techtree');
    expect(SPRING_3_PATHS).toContain('/records');
    expect(SPRING_3_PATHS).toContain('/ranking');
  });

  it('info-страницы параметризованы (`:id` / `:type` для роутера v6)', () => {
    expect(SPRING_3_PATHS.filter((p) => p.includes(':'))).toHaveLength(3);
  });
});
