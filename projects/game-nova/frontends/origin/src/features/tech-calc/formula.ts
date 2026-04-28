// Формулы калькулятора специализации (план 72 Ф.5 Spring 4 ч.2 — S-047).
//
// Воспроизводит JS-логику legacy `templates/standard/adv_tech_calc.tpl`
// 1:1. Это балансная формула — менять без ADR запрещено (R0 плана 72).

// Дефолтные значения базовой атаки/щита (cfg_tech_def_attack/shield).
export const DEFAULT_ATTACK = 1000;
export const DEFAULT_SHIELD = 1000;

// Матрица атака → щиты: TECH_MATRIX[i][j] — коэффициент урона
// технологии (i+1) по щиту (j+1) в порядке LA/IO/PL.
export const TECH_MATRIX: number[][] = [
  [1.0, 0.5, 0.5],
  [0.5, 1.0, 0.5],
  [0.5, 0.5, 1.0],
];

// Масштаб уровней технологий (cfg_tech_scale_*).
export const TECH_SCALE: number[] = [10, 10, 10];

export interface SideInputs {
  baseAttack: number;
  baseShield: number;
  shots: number;
  tech: [number, number, number];
}

export interface SideResult {
  proc: [number, number, number];
  attack: [number, number, number];
  shieldProc: [number, number, number];
  shield: [number, number, number];
}

export const EMPTY_SIDE: SideInputs = {
  baseAttack: DEFAULT_ATTACK,
  baseShield: DEFAULT_SHIELD,
  shots: 1,
  tech: [0, 0, 0],
};

export function computeSide(s: SideInputs): SideResult {
  const list: number[] = s.tech.map((v, i) =>
    Math.round(Math.abs(v) * (TECH_SCALE[i] ?? 1)),
  );
  const max = Math.max(list[0] ?? 0, list[1] ?? 0, list[2] ?? 0);

  const techEffect: [number, number, number] = [0, 0, 0];
  for (let i = 2; i >= 0; i--) {
    if (list[i] === max) {
      techEffect[i] = list[i] ?? 0;
      break;
    }
  }
  const sum = techEffect[0] + techEffect[1] + techEffect[2];

  const dist: [number, number, number] = [0, 0, 0];
  if (sum > 0) {
    dist[0] = techEffect[0] / sum + (list[0] ?? 0) / 100;
    dist[1] = techEffect[1] / sum + (list[1] ?? 0) / 100;
    dist[2] = techEffect[2] / sum + (list[2] ?? 0) / 100;
  } else {
    dist[0] = 1 / 3 + (list[0] ?? 0) / 100;
    dist[1] = 1 / 3 + (list[1] ?? 0) / 100;
    dist[2] = 1 / 3 + (list[2] ?? 0) / 100;
  }

  const proc: [number, number, number] = [
    Math.round(dist[0] * 100),
    Math.round(dist[1] * 100),
    Math.round(dist[2] * 100),
  ];
  const attack: [number, number, number] = [
    Math.ceil(dist[0] * s.baseAttack),
    Math.ceil(dist[1] * s.baseAttack),
    Math.ceil(dist[2] * s.baseAttack),
  ];

  const shieldDist: [number, number, number] = [0, 0, 0];
  for (let i = 0; i < 3; i++) {
    const row = TECH_MATRIX[i];
    if (!row) continue;
    shieldDist[i] =
      dist[0] * (row[0] ?? 0) + dist[1] * (row[1] ?? 0) + dist[2] * (row[2] ?? 0);
  }
  const shieldProc: [number, number, number] = [
    Math.round(shieldDist[0] * 100),
    Math.round(shieldDist[1] * 100),
    Math.round(shieldDist[2] * 100),
  ];
  const shield: [number, number, number] = [
    Math.ceil(shieldDist[0] * s.baseShield),
    Math.ceil(shieldDist[1] * s.baseShield),
    Math.ceil(shieldDist[2] * s.baseShield),
  ];
  return { proc, attack, shieldProc, shield };
}
