// Pure-функции для UX-микрологики (план 71, X-NNN записи).
// Воспроизводит accumulated UX origin/legacy: цвета производства,
// дефицит ресурсов с показом "(нужно X)", энергодефицит,
// added_level бонусы/штрафы, статус артефактов, "X из Y" слоты.
//
// Функции pure: не имеют side-effects и не зависят от React,
// тестируются как обычные ts-функции в vitest.
//
// Источник логики — projects/game-origin-php/src/templates/standard/
// (resource.tpl, required_res_table.tpl, artefacts.tpl). См.
// docs/research/origin-vs-nova/nova-ui-backlog.md X-001..X-022.

// Cost — денежные значения юнита/постройки для проверки дефицита.
// Совместим с api/types Cost: { metal, silicon, hydrogen }.
export interface Cost {
  metal: number;
  silicon: number;
  hydrogen: number;
}

// ResourcesAvailable — что игроку доступно на планете сейчас.
export interface ResourcesAvailable {
  metal: number;
  silicon: number;
  hydrogen: number;
}

// ResourceDeficit — недобор по каждому ресурсу (>= 0).
export interface ResourceDeficit {
  metal: number;
  silicon: number;
  hydrogen: number;
}

// computeDeficit считает дефицит «cost − have», обрезая по нулю.
// X-001: метка `(нужно X)` — это именно эти значения.
export function computeDeficit(cost: Cost, have: ResourcesAvailable): ResourceDeficit {
  return {
    metal:    Math.max(0, cost.metal    - have.metal),
    silicon:  Math.max(0, cost.silicon  - have.silicon),
    hydrogen: Math.max(0, cost.hydrogen - have.hydrogen),
  };
}

export function hasAnyDeficit(d: ResourceDeficit): boolean {
  return d.metal > 0 || d.silicon > 0 || d.hydrogen > 0;
}

// canAfford — упрощённая «целая стоимость покрыта» проверка.
export function canAfford(cost: Cost, have: ResourcesAvailable): boolean {
  return have.metal >= cost.metal && have.silicon >= cost.silicon && have.hydrogen >= cost.hydrogen;
}

// numKind — для X-002: знак производства/потребления.
// Возвращает 'positive' (производство), 'negative' (потребление,
// расход), 'zero'. Цвет назначается в React-компоненте через
// CSS-переменные.
export type NumKind = 'positive' | 'negative' | 'zero';
export function numKind(value: number): NumKind {
  if (value > 0) return 'positive';
  if (value < 0) return 'negative';
  return 'zero';
}

// energyKind — для X-010 (энергодефицит). totalEnergy <= 0 →
// 'deficit', > 0 → 'surplus'. Origin: `<= 0` (равенство тоже красит
// в красный, потому что любое потребление при нулевом производстве
// уже срыв).
export type EnergyKind = 'deficit' | 'surplus';
export function energyKind(totalEnergy: number): EnergyKind {
  return totalEnergy <= 0 ? 'deficit' : 'surplus';
}

// addedLevelKind — для X-013. added_level ∈ ℤ; > 0 — бонус (+),
// < 0 — штраф (−), 0 — не показываем.
export type AddedLevelKind = 'positive' | 'negative' | 'none';
export function addedLevelKind(added: number): AddedLevelKind {
  if (added > 0) return 'positive';
  if (added < 0) return 'negative';
  return 'none';
}

// formatAddedLevel — текстовое представление: «+2» или «-1»; для
// нуля — пустая строка (вызывающий код решает не рендерить).
export function formatAddedLevel(added: number): string {
  if (added === 0) return '';
  if (added > 0) return `+${added}`;
  return String(added); // уже с минусом
}

// slotsState — для X-007 (слоты флота/экспедиции).
// Возвращает 'full' если used >= max (нельзя послать), 'almost' если
// осталось ≤ 1 слота (предупреждение), 'ok' иначе.
export type SlotsState = 'full' | 'almost' | 'ok';
export function slotsState(used: number, max: number): SlotsState {
  if (max <= 0) return 'full';
  if (used >= max) return 'full';
  if (max - used <= 1) return 'almost';
  return 'ok';
}

// artefactStatusKind — для X-008. Маппит state в визуальный статус.
// Origin: 'active' (зелёный), 'delayed' (заряжается, красный/false),
// 'listed' (на бирже), 'expired' (серый), 'consumed' (серый).
// 'held' (на руках, нейтрально).
export type ArtefactStatusKind = 'active' | 'charging' | 'listed' | 'idle' | 'gone';
export function artefactStatusKind(state: string): ArtefactStatusKind {
  if (state === 'active') return 'active';
  if (state === 'delayed') return 'charging';
  if (state === 'listed') return 'listed';
  if (state === 'expired' || state === 'consumed') return 'gone';
  return 'idle';
}

// expiryUrgency — для X-008 (истекающий артефакт).
// expireAt — ISO-строка, now — Date.now() (для тестируемости).
// 'imminent' если ≤ 1 час, 'soon' если ≤ 1 день, 'ok' иначе.
// null/undefined — 'none' (нет TTL).
export type ExpiryUrgency = 'none' | 'ok' | 'soon' | 'imminent';
export function expiryUrgency(expireAt: string | null | undefined, nowMs: number): ExpiryUrgency {
  if (!expireAt) return 'none';
  const left = new Date(expireAt).getTime() - nowMs;
  if (left <= 60 * 60 * 1000) return 'imminent';
  if (left <= 24 * 60 * 60 * 1000) return 'soon';
  return 'ok';
}
