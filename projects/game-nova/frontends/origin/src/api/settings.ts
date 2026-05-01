// API-модуль settings origin-фронта (план 72 Ф.5 Spring 4 — S-042).
//
// Endpoints (openapi.yaml):
//   GET  /api/settings                 — текущие настройки
//   PUT  /api/settings                 — обновить email/language/timezone
//   POST /api/me/deletion/code         — запросить код удаления (на email)
//   DELETE /api/me                     — подтвердить удаление по коду
//   POST   /api/me/vacation            — включить отпуск (план 72.1.5 A)
//   DELETE /api/me/vacation            — выключить отпуск
//
// Смена пароля живёт в identity-service (POST /auth/password); в origin-
// фронте реализуется отдельным вызовом через тот же fetch на /auth-prefix.

import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type {
  DeletionCodeResponse,
  SettingsResponse,
  SettingsUpdate,
} from './types';

export function fetchSettings(): Promise<SettingsResponse> {
  return api.get<SettingsResponse>('/api/settings');
}

export function updateSettings(patch: SettingsUpdate): Promise<void> {
  return api.put<void>('/api/settings', patch, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function requestDeletionCode(): Promise<DeletionCodeResponse> {
  return api.post<DeletionCodeResponse>('/api/me/deletion/code', undefined, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function confirmDeletion(code: string): Promise<void> {
  // backend ждёт DELETE /api/me с body {code}. fetch-DELETE с body
  // допустим в современных браузерах; backend парсит JSON-тело.
  // План 72.1.30: после ConfirmDeletion ставится grace 7 дней
  // (delete_at), а не немедленный soft-delete. Юзер остаётся
  // залогиненным; UI показывает grace-warning + cancel кнопку.
  return api.deleteWithBody<void>(
    '/api/me',
    { code },
    { idempotencyKey: newIdempotencyKey() },
  );
}

// План 72.1.30: отмена pending удаления в grace-period.
// Backend (settings/delete.go::CancelDeletion):
//   400 если no pending или grace истёк / 404 если уже deleted_at.
export function cancelDeletion(): Promise<void> {
  return api.post<void>('/api/me/deletion/cancel', undefined, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export interface ChangePasswordPayload {
  current_password: string;
  new_password: string;
}

// identity-service endpoint, не game-nova. Origin-фронт ходит на общий
// JWT-сервис как nova; в Vite-конфиге origin /auth → identity 9000.
export function changePassword(payload: ChangePasswordPayload): Promise<void> {
  return api.post<void>('/auth/password', payload, {
    idempotencyKey: newIdempotencyKey(),
  });
}

// План 72.1.5 A: vacation toggle. Backend готов (auth/vacation.go).
// Возможные ошибки см. handler.go::SetVacation / UnsetVacation:
//   POST 409 vacation already active
//   POST 409 vacation blocked by pending events
//   POST 400 vacation interval not met (20 days)
//   DELETE 400 vacation not active
//   DELETE 409 vacation must last at least 48h before you can exit
export function enableVacation(): Promise<void> {
  return api.post<void>('/api/me/vacation', undefined, {
    idempotencyKey: newIdempotencyKey(),
  });
}

export function disableVacation(): Promise<void> {
  return api.delete<void>('/api/me/vacation');
}
