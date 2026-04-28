import { describe, it, expect } from 'vitest';
import { computeSide, EMPTY_SIDE, TECH_MATRIX } from './formula';

describe('AdvTechCalculator', () => {
  it('Без технологий — равномерное распределение 1/3', () => {
    const r = computeSide(EMPTY_SIDE);
    // proc[i] ≈ 33% для каждого канала.
    expect(r.proc[0]).toBe(33);
    expect(r.proc[1]).toBe(33);
    expect(r.proc[2]).toBe(33);
  });

  it('С перевесом ПЛ (технология 5) — победитель ПЛ', () => {
    const r = computeSide({
      ...EMPTY_SIDE,
      tech: [0, 0, 5],
    });
    // Канал ПЛ должен иметь максимальное распределение.
    expect(r.proc[2]).toBeGreaterThan(r.proc[0]);
    expect(r.proc[2]).toBeGreaterThan(r.proc[1]);
  });

  it('Атака = baseAttack × dist (победитель получает > 100% при бонусе уровня)', () => {
    // tech=[0,0,5] → list=[0,0,50], max=50, dist[2]=1+50/100=1.5,
    // dist[0]=dist[1]=0; attack[2] = ceil(1.5 × 100) = 150.
    const r = computeSide({
      ...EMPTY_SIDE,
      baseAttack: 100,
      tech: [0, 0, 5],
    });
    expect(r.attack[0]).toBe(0);
    expect(r.attack[1]).toBe(0);
    expect(r.attack[2]).toBe(150);
  });

  it('Матрица 3×3, диагональ 1.0', () => {
    expect(TECH_MATRIX[0]?.[0]).toBe(1);
    expect(TECH_MATRIX[1]?.[1]).toBe(1);
    expect(TECH_MATRIX[2]?.[2]).toBe(1);
  });
});
