// API-модуль friends origin-фронта (план 72 Ф.5 Spring 4 — S-034).
//
// Endpoints (openapi.yaml):
//   GET    /api/friends                  — список друзей
//   POST   /api/friends/{userId}         — добавить (Idempotency-Key)
//   DELETE /api/friends/{userId}         — удалить (Idempotency-Key)
//
// Backend реализует «односторонний» friend-list (без подтверждения второй
// стороной): запись в таблицу friends(user_id, friend_id) добавляется
// сразу. Это соответствует legacy behavior buddylist.

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type { FriendsList } from './types';

export function fetchFriends(): Promise<FriendsList> {
  return api.get<FriendsList>('/api/friends');
}

export function addFriend(userId: string): Promise<void> {
  return api.post<void>(`/api/friends/${encodeURIComponent(userId)}`, undefined, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function removeFriend(userId: string): Promise<void> {
  return api.delete<void>(`/api/friends/${encodeURIComponent(userId)}`, {
    idempotencyKey: newIdempotencyKey(),
  });
}
