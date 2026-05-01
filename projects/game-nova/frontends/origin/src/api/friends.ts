// API-модуль friends origin-фронта (план 72 Ф.5 Spring 4 — S-034,
// расширен 72.1.14: двусторонний accept-flow с подтверждением).
//
// Endpoints (openapi.yaml):
//   GET    /api/friends?pending=...            — список (mutual / pending)
//   POST   /api/friends/{userId}               — отправить запрос
//   POST   /api/friends/{userId}/accept        — принять входящий запрос
//   DELETE /api/friends/{userId}               — удалить (или отменить запрос)
//
// Двусторонняя модель: legacy buddylist.accepted=0/1. По умолчанию
// /api/friends возвращает только mutual (accepted=true), отдельные
// разделы UI запрашивают ?pending=incoming или ?pending=outgoing.

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type { FriendsList } from './types';

export type FriendsFilter = 'mutual' | 'incoming' | 'outgoing' | 'all';

function withPending(filter: FriendsFilter): string {
  if (filter === 'mutual') return '/api/friends';
  return `/api/friends?pending=${filter}`;
}

export function fetchFriends(filter: FriendsFilter = 'mutual'): Promise<FriendsList> {
  return api.get<FriendsList>(withPending(filter));
}

export function addFriend(userId: string): Promise<void> {
  return api.post<void>(`/api/friends/${encodeURIComponent(userId)}`, undefined, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function acceptFriend(userId: string): Promise<void> {
  return api.post<void>(
    `/api/friends/${encodeURIComponent(userId)}/accept`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function removeFriend(userId: string): Promise<void> {
  return api.delete<void>(`/api/friends/${encodeURIComponent(userId)}`, {
    idempotencyKey: newIdempotencyKey(),
  });
}
