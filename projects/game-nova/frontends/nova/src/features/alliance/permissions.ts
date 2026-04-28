// Гранулярные permissions альянса (план 67 D-014, U-005).
//
// Источник истины — backend: internal/alliance/permissions.go
// (7 ключей, snake_case, owner-fallback). Здесь — frontend-зеркало.

export const PERMISSION_KEYS = [
  'can_invite',
  'can_kick',
  'can_send_global_mail',
  'can_manage_diplomacy',
  'can_change_description',
  'can_propose_relations',
  'can_manage_ranks',
] as const;

export type PermissionKey = (typeof PERMISSION_KEYS)[number];

export type PermissionMap = Partial<Record<PermissionKey, boolean>>;

// hasPerm — резолвит permission для пользователя в контексте альянса.
// owner всегда имеет все права (builtin-fallback в бэке).
// Прочие builtin-роли (member и т.п.) без custom-ranga получают false.
//
// План 67 Ф.5 trade-off (см. docs/simplifications.md):
// frontend пока не получает rank_id с бэка (Member DTO без rank_id),
// поэтому гранулярная проверка для не-owner'а не работает; кнопка
// management видна только owner'у. Для не-owner'а возвращаем false по
// всем permissions кроме builtin'ов, не требующих прав. Бэкенд-проверки
// идут по полной — если кнопка вдруг будет видна, 403 защитит от
// неавторизованных мутаций.
export function hasPerm(
  isOwner: boolean,
  perm: PermissionKey,
  rankPerms?: PermissionMap | null,
): boolean {
  if (isOwner) return true;
  if (!rankPerms) return false;
  return rankPerms[perm] === true;
}
