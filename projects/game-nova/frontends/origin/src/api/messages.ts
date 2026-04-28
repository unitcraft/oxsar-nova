// API-модуль messages origin-фронта.
//
// Spring 1: только счётчик непрочитанных для S-001 Main.
// Spring 4 (план 72 Ф.5 — S-035): полный CRUD сообщений.
//
// Endpoints (openapi.yaml):
//   GET    /api/messages                — inbox (получено мной)
//   GET    /api/messages/sent           — отправленные мной
//   POST   /api/messages                — отправить (Idempotency-Key)
//   DELETE /api/messages                — удалить все
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
  UnreadCount,
} from './types';

export function fetchUnreadCount(): Promise<UnreadCount> {
  return api.get<UnreadCount>('/api/messages/unread-count');
}

export function fetchMessages(folder: MessageFolder): Promise<MessagesList> {
  const path = folder === 'sent' ? '/api/messages/sent' : '/api/messages';
  return api.get<MessagesList>(path);
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
