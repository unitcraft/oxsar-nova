// Расчёт прогресса задачи с start_at / end_at.
// Используется в ConstructionProgress, BuildPanel, ResearchScreen,
// RepairScreen, DisassembleScreen.

export function calcRemainingSec(endMs: number, now: number): number {
  return Math.max(0, Math.floor((endMs - now) / 1000));
}

// pct для отображения цифрами (+1s компенсирует задержку тика).
export function calcPct(startMs: number, endMs: number, now: number): number {
  const totalMs = endMs - startMs;
  if (totalMs <= 0) return 0;
  return Math.max(0, Math.min(100, Math.floor((now - startMs + 1000) * 100 / totalMs)));
}

// barPct для CSS-полоски с анимацией (+2s * progress компенсирует отставание transition).
export function calcBarPct(startMs: number, endMs: number, now: number): number {
  const totalMs = endMs - startMs;
  if (totalMs <= 0) return 0;
  const progress = Math.max(0, Math.min(1, (now - startMs + 1000) / totalMs));
  return Math.max(0, Math.min(100, Math.floor(
    (now - startMs + 2000 * progress) * 100 / totalMs,
  )));
}
