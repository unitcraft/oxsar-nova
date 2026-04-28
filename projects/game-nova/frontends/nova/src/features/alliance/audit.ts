// Action-константы alliance audit-log (план 67 Ф.5 ч.2, U-013).
//
// Источник истины — backend: internal/alliance/audit.go (18 значений
// + free-form для новых). Здесь — frontend-зеркало для типизации
// фильтров и i18n-резолвинга.

export const AUDIT_ACTIONS = [
  'alliance_created',
  'alliance_disbanded',
  'description_changed',
  'member_joined',
  'member_left',
  'member_kicked',
  'member_rank_assigned',
  'rank_created',
  'rank_updated',
  'rank_deleted',
  'application_approved',
  'application_rejected',
  'relation_proposed',
  'relation_accepted',
  'relation_rejected',
  'relation_cleared',
  'leadership_transferred',
  'open_changed',
] as const;

export type AuditAction = (typeof AUDIT_ACTIONS)[number];

export function isKnownAction(s: string): s is AuditAction {
  return (AUDIT_ACTIONS as readonly string[]).includes(s);
}

// auditActionLabelKey — i18n-ключ для action. Для известного action
// возвращает alliance.audit.action.<name>; для неизвестного — fallback
// alliance.audit.action.unknown (рендерится с самим именем как vars).
export function auditActionLabelKey(action: string): string {
  return isKnownAction(action) ? `audit.action.${action}` : 'audit.action.unknown';
}

// auditTargetKindLabelKey — i18n-ключ для типа цели (member, rank,
// alliance, relation). Для неизвестного — fallback на сам kind.
const KNOWN_TARGET_KINDS = ['member', 'rank', 'alliance', 'relation'] as const;
type AuditTargetKind = (typeof KNOWN_TARGET_KINDS)[number];

export function isKnownTargetKind(s: string): s is AuditTargetKind {
  return (KNOWN_TARGET_KINDS as readonly string[]).includes(s);
}

export function auditTargetKindLabelKey(kind: string): string {
  return isKnownTargetKind(kind) ? `audit.targetKind.${kind}` : 'audit.targetKind.unknown';
}

// formatRelativeTime — короткое представление "5m ago" / "2h ago" /
// "3d ago" для audit-таблицы. Локализуется i18n-ключами audit.time.*.
//
// Возвращает кортеж [key, vars] для t(): фронт сам подставит через i18n.
export function formatRelativeTime(now: Date, then: Date): { key: string; vars: { n: string } } {
  const deltaSec = Math.max(0, Math.floor((now.getTime() - then.getTime()) / 1000));
  if (deltaSec < 60) return { key: 'audit.time.justNow', vars: { n: '' } };
  const m = Math.floor(deltaSec / 60);
  if (m < 60) return { key: 'audit.time.mAgo', vars: { n: String(m) } };
  const h = Math.floor(m / 60);
  if (h < 24) return { key: 'audit.time.hAgo', vars: { n: String(h) } };
  const d = Math.floor(h / 24);
  if (d < 30) return { key: 'audit.time.dAgo', vars: { n: String(d) } };
  const mo = Math.floor(d / 30);
  return { key: 'audit.time.moAgo', vars: { n: String(mo) } };
}
