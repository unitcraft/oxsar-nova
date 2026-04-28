// Дипломатические отношения (план 67 D-014 B1).
//
// 5 enum-значений, источник истины — backend (internal/alliance/service.go,
// `relationNeedsAccept`) и OpenAPI-схема `AllianceRelation`.
// friend (legacy ally), neutral, hostile_neutral, nap, war.

export const RELATIONS = ['friend', 'neutral', 'hostile_neutral', 'nap', 'war'] as const;
export type Relation = (typeof RELATIONS)[number];

// relationLabelKey — i18n-ключ для перевода. Группа: `alliance`.
export const RELATION_LABEL_KEY: Record<Relation, string> = {
  friend: 'relFriend',
  neutral: 'relNeutral',
  hostile_neutral: 'relHostileNeutral',
  nap: 'relNap',
  war: 'relWar',
};

// relationColor — CSS-переменная фона/текста для бейджа.
// Зелёный/серый/жёлтый/синий/красный (промт 5.3).
export const RELATION_COLOR: Record<Relation, string> = {
  friend: 'var(--ox-success)',
  neutral: 'var(--ox-fg-dim)',
  hostile_neutral: 'var(--ox-warning)',
  nap: 'var(--ox-accent)',
  war: 'var(--ox-danger)',
};

// relationNeedsAccept — двусторонние отношения требуют согласия второй
// стороны (status=pending пока initiator один, ok после accept).
// Зеркало backend `relationNeedsAccept` (service.go).
export function relationNeedsAccept(rel: Relation): boolean {
  switch (rel) {
    case 'friend':
    case 'neutral':
    case 'nap':
      return true;
    case 'hostile_neutral':
    case 'war':
      return false;
  }
}

// isKnownRelation — type-guard для произвольной строки от бэка.
export function isKnownRelation(s: string): s is Relation {
  return (RELATIONS as readonly string[]).includes(s);
}
