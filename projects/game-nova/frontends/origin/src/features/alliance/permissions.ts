// Pure-функции прав/маппинга статусов alliance (план 72 Ф.3 Spring 2 ч.1).
//
// Вынесено в отдельный модуль без react-зависимостей, чтобы
// тестировать без jsdom.

import type {
  AlliancePermissionKey,
  AlliancePermissionMap,
} from '@/api/types';

export const PERMISSION_KEYS: readonly AlliancePermissionKey[] = [
  'can_invite',
  'can_kick',
  'can_send_global_mail',
  'can_manage_diplomacy',
  'can_change_description',
  'can_propose_relations',
  'can_manage_ranks',
] as const;

// hasPerm — резолвит permission для пользователя в контексте альянса.
// owner всегда имеет все права (builtin-fallback в бэке).
// План 67 Ф.5 trade-off (P67.S5.B): Member DTO без rank_id, поэтому
// гранулярная проверка для не-owner'а не работает. Backend защитит 403.
export function hasPerm(
  isOwner: boolean,
  perm: AlliancePermissionKey,
  rankPerms?: AlliancePermissionMap | null,
): boolean {
  if (isOwner) return true;
  if (!rankPerms) return false;
  return rankPerms[perm] === true;
}

// relationStatusKey — i18n-key из bundle alliance: для статуса
// дипотношения (см. configs/i18n/ru.yml: protection/confederation/war/
// tradeAgreement/ceasefire).
export function relationStatusKey(status: string): string {
  switch (status) {
    case 'protection':
      return 'protection';
    case 'confederation':
      return 'confederation';
    case 'war':
      return 'war';
    case 'trade':
      return 'tradeAgreement';
    case 'ceasefire':
      return 'ceasefire';
    default:
      return 'status';
  }
}
