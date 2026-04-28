// API-модуль messages origin-фронта (план 72 Ф.2 Spring 1).
//
// Из mailbox в Spring 1 нужен только счётчик непрочитанных для S-001
// Main («Есть непрочитанные сообщения»).
//
// Endpoint (openapi.yaml):
//   GET /api/messages/unread-count

import { api } from './client';
import type { UnreadCount } from './types';

export function fetchUnreadCount(): Promise<UnreadCount> {
  return api.get<UnreadCount>('/api/messages/unread-count');
}
