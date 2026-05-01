// API-модуль alliance origin-фронта (план 72 Ф.3 Spring 2 ч.1).
//
// Backend полностью закрыт планом 67 (коммиты до a149594306). Все
// endpoint'ы из openapi.yaml `/api/alliances/*`. Idempotency-Key (R9)
// прокидывается в мутации, для которых backend поддерживает (descriptions
// PATCH, ranks POST/PATCH, members rank-id PATCH, transfer-leadership).
//
// Семантика идентична nova-варианту (frontends/nova/src/features/alliance/),
// только без обёртки над nova-UI (Confirm/Toast/Modal) — origin-экраны
// рендерят формы pixel-perfect к legacy *.tpl.

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type {
  Alliance,
  AllianceApplication,
  AllianceAuditPage,
  AllianceDescriptionView,
  AllianceDetail,
  AllianceListResult,
  AlliancePermissionMap,
  AllianceRank,
  AllianceRelation,
  AllianceTransferCodeIssued,
} from './types';
// query-builder'ы вынесены в отдельный pure-модуль — это позволяет
// тестировать их без подгрузки auth-store/localStorage.
export {
  buildSearchQuery,
  buildAuditQuery,
  type AuditQuery,
} from '@/features/alliance/queries';

export function fetchAllianceList(qs: string): Promise<AllianceListResult> {
  return api.get<AllianceListResult>(qs ? `/api/alliances?${qs}` : '/api/alliances');
}

export function fetchMyAlliance(): Promise<AllianceDetail | null> {
  return api.get<AllianceDetail | null>('/api/alliances/me').catch((e: { status?: number }) => {
    // В origin-фронте «нет альянса» — это валидное состояние, рендерим экран
    // создания/поиска. nova-API на /api/alliances/me возвращает 404 в этом
    // случае; глушим в null, остальные ошибки пробрасываем.
    if (e?.status === 404) return null;
    throw e;
  });
}

export function fetchAlliance(id: string): Promise<AllianceDetail> {
  return api.get<AllianceDetail>(`/api/alliances/${id}`);
}

export function createAlliance(input: {
  tag: string;
  name: string;
  description?: string;
}): Promise<Alliance> {
  return api.post<Alliance>('/api/alliances', input, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function joinAlliance(id: string, message: string): Promise<{ status?: string }> {
  return api.post<{ status?: string }>(
    `/api/alliances/${id}/join`,
    { message },
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function leaveAlliance(): Promise<void> {
  return api.post<void>('/api/alliances/leave', undefined, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function disbandAlliance(id: string): Promise<void> {
  return api.delete<void>(`/api/alliances/${id}`, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function setAllianceOpen(id: string, isOpen: boolean): Promise<void> {
  return api.patch<void>(
    `/api/alliances/${id}/open`,
    { is_open: isOpen },
    { idempotencyKey: newIdempotencyKey() },
  );
}

// Applications

export function fetchAllianceApplications(
  id: string,
): Promise<{ applications: AllianceApplication[] | null }> {
  return api.get<{ applications: AllianceApplication[] | null }>(
    `/api/alliances/${id}/applications`,
  );
}

export function approveApplication(appID: string): Promise<void> {
  return api.post<void>(
    `/api/alliances/applications/${appID}/approve`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function rejectApplication(appID: string): Promise<void> {
  return api.delete<void>(`/api/alliances/applications/${appID}`, {
    idempotencyKey: newIdempotencyKey(),
  });
}

// Members

export function kickMember(allianceID: string, userID: string): Promise<void> {
  return api.delete<void>(
    `/api/alliances/${allianceID}/members/${userID}`,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function setMemberRank(
  allianceID: string,
  userID: string,
  rankName: string,
): Promise<void> {
  return api.patch<void>(
    `/api/alliances/${allianceID}/members/${userID}/rank`,
    { rank_name: rankName },
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function setMemberRankID(
  allianceID: string,
  userID: string,
  rankID: string | null,
): Promise<void> {
  return api.patch<void>(
    `/api/alliances/${allianceID}/members/${userID}/rank-id`,
    { rank_id: rankID },
    { idempotencyKey: newIdempotencyKey() },
  );
}

// Descriptions (план 67 Ф.2, U-015 — 3 описания)

export function fetchDescriptions(
  allianceID: string,
): Promise<AllianceDescriptionView> {
  return api.get<AllianceDescriptionView>(
    `/api/alliances/${allianceID}/descriptions`,
  );
}

export function updateDescriptions(
  allianceID: string,
  patch: {
    description_external?: string;
    description_internal?: string;
    description_apply?: string;
  },
): Promise<void> {
  return api.patch<void>(
    `/api/alliances/${allianceID}/descriptions`,
    patch,
    { idempotencyKey: newIdempotencyKey() },
  );
}

// Ranks (план 67 Ф.2, U-005 — гранулярные permissions)

export function fetchRanks(allianceID: string): Promise<{ ranks: AllianceRank[] | null }> {
  return api.get<{ ranks: AllianceRank[] | null }>(
    `/api/alliances/${allianceID}/ranks`,
  );
}

export function createRank(
  allianceID: string,
  input: { name: string; position?: number; permissions?: AlliancePermissionMap },
): Promise<{ rank: AllianceRank }> {
  return api.post<{ rank: AllianceRank }>(
    `/api/alliances/${allianceID}/ranks`,
    input,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function updateRank(
  allianceID: string,
  rankID: string,
  patch: { name?: string; position?: number; permissions?: AlliancePermissionMap },
): Promise<void> {
  return api.patch<void>(
    `/api/alliances/${allianceID}/ranks/${rankID}`,
    patch,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function deleteRank(allianceID: string, rankID: string): Promise<void> {
  return api.delete<void>(
    `/api/alliances/${allianceID}/ranks/${rankID}`,
    { idempotencyKey: newIdempotencyKey() },
  );
}

// Diplomacy (план 67 D-014 B1 — 5 enum-статусов)

export function fetchRelations(
  allianceID: string,
): Promise<{ relations: AllianceRelation[] | null }> {
  return api.get<{ relations: AllianceRelation[] | null }>(
    `/api/alliances/${allianceID}/relations`,
  );
}

export function proposeRelation(
  allianceID: string,
  input: { target_id: string; status: string; message?: string | undefined },
): Promise<void> {
  return api.put<void>(
    `/api/alliances/${allianceID}/relations`,
    input,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function acceptRelation(
  allianceID: string,
  initiatorID: string,
): Promise<void> {
  return api.put<void>(
    `/api/alliances/${allianceID}/relations/${initiatorID}/accept`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function rejectRelation(
  allianceID: string,
  initiatorID: string,
): Promise<void> {
  return api.post<void>(
    `/api/alliances/${allianceID}/relations/${initiatorID}/reject`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function breakRelation(
  allianceID: string,
  initiatorID: string,
): Promise<void> {
  return api.delete<void>(
    `/api/alliances/${allianceID}/relations/${initiatorID}`,
    { idempotencyKey: newIdempotencyKey() },
  );
}

// Audit-log (план 67 Ф.2, U-013)

export function fetchAuditLog(
  allianceID: string,
  qs: string,
): Promise<AllianceAuditPage> {
  const url = qs
    ? `/api/alliances/${allianceID}/audit?${qs}`
    : `/api/alliances/${allianceID}/audit`;
  return api.get<AllianceAuditPage>(url);
}

// Transfer leadership (план 67 Ф.3 — 2-step с email-кодом)

export function requestTransferCode(
  allianceID: string,
  newOwnerID: string,
  idempotencyKey: string,
): Promise<AllianceTransferCodeIssued> {
  return api.post<AllianceTransferCodeIssued>(
    `/api/alliances/${allianceID}/transfer-leadership/code`,
    { new_owner_id: newOwnerID },
    { idempotencyKey },
  );
}

export function confirmTransfer(
  allianceID: string,
  newOwnerID: string,
  code: string,
  idempotencyKey: string,
): Promise<void> {
  return api.post<void>(
    `/api/alliances/${allianceID}/transfer-leadership`,
    { new_owner_id: newOwnerID, code },
    { idempotencyKey },
  );
}

// План 72.1.43: legacy `Alliance::globalMail` — рассылка всем участникам.
export function broadcastAllianceMail(
  allianceID: string,
  title: string,
  body: string,
): Promise<void> {
  return api.post<void>(
    `/api/alliances/${allianceID}/broadcast`,
    { title, body },
  );
}

// План 72.1.43: legacy updateAllyTag + updateAllyName (PATCH).
export function updateAllianceTagName(
  allianceID: string,
  tag?: string,
  name?: string,
): Promise<void> {
  const payload: { tag?: string; name?: string } = {};
  if (tag) payload.tag = tag;
  if (name) payload.name = name;
  return api.patch<void>(`/api/alliances/${allianceID}`, payload);
}
