// API-модуль chat origin-фронта (план 72 Ф.5 Spring 4 — S-036/S-037).
//
// Endpoints (openapi.yaml):
//   GET    /api/chat/{kind}/history    — последние N сообщений канала
//   GET    /api/chat/{kind}/ws         — WebSocket upgrade (kind=global|alliance)
//   POST   /api/chat/{kind}/send       — REST fallback отправки (Idempotency-Key)
//   GET    /api/chat/{kind}/unread     — непрочитано / last_read_at
//   POST   /api/chat/{kind}/read       — отметить как прочитанное
//   PATCH  /api/chat/messages/{id}     — отредактировать (5 мин окно)
//   DELETE /api/chat/messages/{id}     — soft-delete (5 мин окно)
//
// План 72 Ф.5: BBCode из legacy выкидывается — render как plain text
// (см. simplifications.md, P72.S4.BBCODE). TipTap-интеграция отложена
// в Ф.8 плана 72.

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type {
  ChatChannelKind,
  ChatMessage,
  ChatUnreadCount,
} from './types';

export function fetchChatHistory(
  kind: ChatChannelKind,
): Promise<ChatMessage[]> {
  return api
    .get<ChatMessage[] | null>(`/api/chat/${kind}/history`)
    .then((list) => list ?? []);
}

export function sendChatMessage(
  kind: ChatChannelKind,
  body: string,
): Promise<ChatMessage> {
  return api.post<ChatMessage>(
    `/api/chat/${kind}/send`,
    { body },
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function fetchChatUnread(
  kind: ChatChannelKind,
): Promise<ChatUnreadCount> {
  return api.get<ChatUnreadCount>(`/api/chat/${kind}/unread`);
}

export function markChatRead(kind: ChatChannelKind): Promise<{
  channel: string;
  last_read_at: string;
}> {
  return api.post<{ channel: string; last_read_at: string }>(
    `/api/chat/${kind}/read`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function editChatMessage(
  id: string,
  body: string,
): Promise<ChatMessage> {
  return api.patch<ChatMessage>(
    `/api/chat/messages/${encodeURIComponent(id)}`,
    { body },
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function deleteChatMessage(id: string): Promise<void> {
  return api.delete<void>(`/api/chat/messages/${encodeURIComponent(id)}`, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function chatWsUrl(kind: ChatChannelKind, token: string): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host;
  return `${proto}//${host}/api/chat/${kind}/ws?token=${encodeURIComponent(token)}`;
}
