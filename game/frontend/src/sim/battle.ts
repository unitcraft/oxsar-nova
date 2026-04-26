// Клиентский офлайн-симулятор боя. Должен быть ПОЛНОСТЬЮ эквивалентен
// серверному internal/battle (§5.7 ТЗ): те же формулы, тот же seed-
// контракт, тот же RNG.
//
// Статус: каркас. Порт формул синхронно с Go-реализацией.

export interface BattleInput {
  seed: number;
  rounds?: number;
  attackers: BattleSide[];
  defenders: BattleSide[];
  isMoon?: boolean;
}

export interface BattleSide {
  userId: string;
  username?: string;
  units: BattleUnit[];
}

export interface BattleUnit {
  unitId: number;
  quantity: number;
  front: number;
  attack: number;
  shield: number;
  shell: number;
}

export interface BattleReport {
  seed: number;
  winner: 'attackers' | 'defenders' | 'draw';
  rounds: number;
}

// Плейсхолдер — «никто не умер», для верификации API сигнатуры.
// Реальная реализация приходит вместе с portом Java-движка (M4).
export function simulate(input: BattleInput): BattleReport {
  return {
    seed: input.seed,
    winner: 'draw',
    rounds: 0,
  };
}
