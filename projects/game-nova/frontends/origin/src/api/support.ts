// API-модуль support origin-фронта (план 72 Ф.5 Spring 4 ч.2 — S-045).
//
// КРОСС-СЕРВИСНЫЙ ENDPOINT: /api/reports НЕ принадлежит game-nova-backend.
// Согласно плану 56 reports переехал в portal-backend (единый реестр
// жалоб для всех вселенных). Поэтому fetch идёт на
// `${VITE_PORTAL_BASE_URL}/api/reports`, не на game-nova /api/*.
//
// В dev requires VITE_PORTAL_BASE_URL=http://localhost:8090. В проде
// origin портала должен быть в ALLOWED_ORIGINS portal-backend'а
// (cross-origin fetch с Authorization-header Bearer JWT).
//
// portal-backend сам управляет дедупликацией обращений по комбинации
// (target_type, target_id, user_id, reason) — R9 Idempotency-Key
// здесь не передаём, см. план 56 §reports.
//
// Эта реализация мирорит nova-фронт ReportButton.tsx (план 56 коммит
// 37ae65b430), но на отдельном экране /support с богатой формой.

import { useAuthStore } from '@/stores/auth';

const PORTAL_BASE =
  (import.meta.env['VITE_PORTAL_BASE_URL'] as string | undefined) ?? '';

const REPORT_ENDPOINT = `${PORTAL_BASE}/api/reports`;

export type SupportReason =
  | 'profanity'
  | 'extremism'
  | 'drugs'
  | 'spam'
  | 'impersonation'
  | 'cheat'
  | 'other';

export interface SupportSubmitRequest {
  // backend portal-backend ожидает target_type/target_id для жалоб; для
  // support-формы (S-045) target_type='support', target_id='self'.
  reason: SupportReason;
  comment: string;
  // дополнительные поля для тех. поддержки — пакуются в comment как
  // структурированный текст (browser, page, time, fleet, ...). Это
  // временно и совместимо с текущим portal-backend; план 56 не
  // вводил отдельный shape для support-формы.
}

export async function submitSupport(req: SupportSubmitRequest): Promise<void> {
  const token = useAuthStore.getState().accessToken;
  const res = await fetch(REPORT_ENDPOINT, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
    body: JSON.stringify({
      target_type: 'support',
      target_id: 'self',
      reason: req.reason,
      comment: req.comment,
    }),
  });
  if (!res.ok) {
    let message = `HTTP ${res.status}`;
    try {
      const body = (await res.json()) as { error?: { message?: string } };
      if (body.error?.message) message = body.error.message;
    } catch {
      // ignore
    }
    throw new Error(message);
  }
}
