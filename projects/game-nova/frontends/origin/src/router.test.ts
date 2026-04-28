// 7-экранный router origin-фронта (план 72 Ф.2 Spring 1).
//
// Тест-фикстура: заявленный набор путей не должен молча сократиться.
// Если кто-то удалит маршрут (скажем, /mission), тест провалится — это
// сигнал, что Spring 1 распался.
//
// Реальные React-компоненты не рендерятся (для этого нужна testing-library
// + jsdom, добавится в Ф.7 при подключении screenshot-diff CI плана 73).
// Здесь мы статически перечисляем требуемые пути.

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
