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
//
// План 72.1.55 Task D (P72.S2.D 1:1): backend теперь возвращает
// effective_perms на self-Member DTO (если viewer = self). Передавайте
// `selfMember.effective_perms` в rankPerms — функция корректно
// разрезолвит для не-owner'ов с rank_id.
//
// Если effective_perms нет (старый клиент / viewer не self) — fallback
// на isOwner: только owner получает true. Backend защищает 403'ом
// если игрок попытается сделать запрещённое действие.
export function hasPerm(
  isOwner: boolean,
  perm: AlliancePermissionKey,
  rankPerms?: AlliancePermissionMap | null,
): boolean {
  if (isOwner) return true;
  if (!rankPerms) return false;
  return rankPerms[perm] === true;
}

// findSelfPerms — извлекает effective_perms текущего viewer'а из
// списка членов. Backend заполняет это поле только для self-row.
// Возвращает null если self-row не найден (anonymous viewer).
//
// План 72.1.55 Task D (P72.S2.D 1:1).
export function findSelfPerms(
  members: { user_id: string; effective_perms?: AlliancePermissionMap }[],
  selfUserID: string | null,
): AlliancePermissionMap | null {
  if (!selfUserID) return null;
  for (const m of members) {
    if (m.user_id === selfUserID && m.effective_perms) {
      return m.effective_perms;
    }
  }
  return null;
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
