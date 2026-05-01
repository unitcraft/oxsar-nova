// API-модуль messages origin-фронта.
//
// Spring 1: только счётчик непрочитанных для S-001 Main.
// Spring 4 (план 72 Ф.5 — S-035): полный CRUD сообщений.
// План 72.1.17: legacy folder routing (12 системных папок).
//
// Endpoints (openapi.yaml):
//   GET    /api/messages?folder=N       — inbox или конкретная папка
//   GET    /api/messages/folders        — список папок с счётчиками (план 72.1.17)
//   GET    /api/messages/sent           — отправленные (backwards-compat)
//   POST   /api/messages                — отправить (Idempotency-Key)
//   DELETE /api/messages?folder=N       — удалить все (или в папке)
//   DELETE /api/messages/{id}           — удалить одно (Idempotency-Key)
//   GET    /api/messages/unread-count   — счётчик
//   POST   /api/messages/{id}/read      — пометить прочитанным
//   GET    /api/{battle,espionage,expedition}-reports/{id} — отчёты

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type {
  MessagesList,
  MessageCompose,
  MessageFolder,
  MessageFoldersList,
  UnreadCount,
} from './types';

export function fetchUnreadCount(): Promise<UnreadCount> {
  return api.get<UnreadCount>('/api/messages/unread-count');
}

// План 72.1.17: legacy folder constants (config/consts.php:508-518) +
// SYSTEM=12 (расширение oxsar-nova для welcome/inactivity).
export const FOLDER_INBOX = 1;
export const FOLDER_SENT = 2;
export const FOLDER_FLEET = 3;
export const FOLDER_SPY = 4;
export const FOLDER_BATTLE = 5;
export const FOLDER_ALLIANCE = 6;
export const FOLDER_ARTEFACTS = 7;
export const FOLDER_CREDIT = 8;
export const FOLDER_EXPEDITION = 9;
export const FOLDER_RECYCLER = 10;
export const FOLDER_SURVEILLANCE = 11;
export const FOLDER_SYSTEM = 12;

export function fetchMessages(
  folderOrId: MessageFolder | number,
): Promise<MessagesList> {
  // Backwards-compat: 'sent' → /api/messages/sent (отдельный endpoint),
  // 'inbox' → /api/messages без фильтра, число → ?folder=<id>.
  if (folderOrId === 'sent') {
    return api.get<MessagesList>('/api/messages/sent');
  }
  if (typeof folderOrId === 'number') {
    return api.get<MessagesList>(
      `/api/messages?folder=${folderOrId}`,
    );
  }
  return api.get<MessagesList>('/api/messages');
}

// План 72.1.17: список системных папок с счётчиками.
export function fetchMessageFolders(): Promise<MessageFoldersList> {
  return api.get<MessageFoldersList>('/api/messages/folders');
}

export function sendMessage(payload: MessageCompose): Promise<void> {
  return api.post<void>('/api/messages', payload, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function markMessageRead(id: string): Promise<void> {
  return api.post<void>(
    `/api/messages/${encodeURIComponent(id)}/read`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function deleteMessage(id: string): Promise<void> {
  return api.delete<void>(`/api/messages/${encodeURIComponent(id)}`, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function deleteAllMessages(): Promise<void> {
  return api.delete<void>('/api/messages', {
    idempotencyKey: newIdempotencyKey(),
  });
}
