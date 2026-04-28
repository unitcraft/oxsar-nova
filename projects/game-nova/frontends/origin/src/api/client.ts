// HTTP-клиент origin-фронта (план 72 Ф.1).
//
// Тонкая обёртка над fetch с прозрачным прокидыванием Bearer-токена
// и разворачиванием стандартного error-envelope nova-API (R1: snake_case
// JSON). Шаблон взят из nova-фронта (api/client.ts) и адаптирован под
// origin-store (отдельный localStorage namespace).
//
// Идентификаторы: nova-API ожидает Idempotency-Key (R9 ТЗ §16.10) для
// PATCH/POST/PUT/DELETE; origin-фронт сразу пишется на nova-имена API
// без backend-адаптеров (R6 плана 72).
//
// Конкретные типы Request/Response будут подтянуты из
// `src/api/schema.d.ts` после `npm run gen:api` (генерируется по
// `projects/game-nova/api/openapi.yaml`).

import { useAuthStore } from '@/stores/auth';

export interface ApiError extends Error {
  status: number;
  code: string;
}

export interface MutationOpts {
  idempotencyKey?: string;
  headers?: Record<string, string>;
}

interface ErrorEnvelope {
  error?: { code?: string; message?: string };
}

function withMutationHeaders(
  opts?: MutationOpts,
): Record<string, string> | undefined {
  if (!opts) return undefined;
  const h: Record<string, string> = { ...(opts.headers ?? {}) };
  if (opts.idempotencyKey !== undefined) {
    h['Idempotency-Key'] = opts.idempotencyKey;
  }
  return Object.keys(h).length > 0 ? h : undefined;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = useAuthStore.getState().accessToken;
  const res = await fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });

  if (!res.ok) {
    let code = 'http_error';
    let message = `HTTP ${res.status}`;
    try {
      const body = (await res.json()) as ErrorEnvelope;
      if (body.error) {
        code = body.error.code ?? code;
        message = body.error.message ?? message;
      }
    } catch {
      // body не JSON — оставляем дефолт
    }
    const err = new Error(message) as ApiError;
    err.status = res.status;
    err.code = code;
    if (res.status === 401) {
      useAuthStore.getState().logout();
    }
    throw err;
  }

  // 204 No Content
  if (res.status === 204) {
    return undefined as T;
  }
  return (await res.json()) as T;
}

export const api = {
  get: <T>(path: string): Promise<T> => request<T>(path, { method: 'GET' }),
  post: <T>(path: string, body?: unknown, opts?: MutationOpts): Promise<T> => {
    const extra = withMutationHeaders(opts);
    return request<T>(path, {
      method: 'POST',
      body: body !== undefined ? JSON.stringify(body) : null,
      ...(extra ? { headers: extra } : {}),
    });
  },
  put: <T>(path: string, body?: unknown, opts?: MutationOpts): Promise<T> => {
    const extra = withMutationHeaders(opts);
    return request<T>(path, {
      method: 'PUT',
      body: body !== undefined ? JSON.stringify(body) : null,
      ...(extra ? { headers: extra } : {}),
    });
  },
  patch: <T>(path: string, body?: unknown, opts?: MutationOpts): Promise<T> => {
    const extra = withMutationHeaders(opts);
    return request<T>(path, {
      method: 'PATCH',
      body: body !== undefined ? JSON.stringify(body) : null,
      ...(extra ? { headers: extra } : {}),
    });
  },
  delete: <T>(path: string, opts?: MutationOpts): Promise<T> => {
    const extra = withMutationHeaders(opts);
    return request<T>(path, {
      method: 'DELETE',
      ...(extra ? { headers: extra } : {}),
    });
  },
};
