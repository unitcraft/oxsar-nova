// Тонкая обёртка fetch с прозрачным прокидыванием токена. Типы
// генерируются из OpenAPI (npm run gen:api) в api/schema.d.ts; после
// первой генерации заменим any на конкретные типы.
//
// Пока schema.d.ts не сгенерирован, оставляем явные минимальные типы
// для критичных полей, чтобы не лить any по коду (§17.3 ТЗ).

import { useAuthStore } from '@/stores/auth';

export interface ApiError extends Error {
  status: number;
  code: string;
}

// MutationOpts — расширенные опции для PATCH/POST/PUT/DELETE.
// idempotencyKey: значение HTTP-заголовка `Idempotency-Key` (R9 ТЗ —
// см. §16.10). Бэкенд хранит ответ под этим ключом 24ч; повторный
// запрос с тем же ключом не выполнится дважды.
// headers: произвольные дополнительные заголовки (если потребуется
// в будущем).
export interface MutationOpts {
  idempotencyKey?: string;
  headers?: Record<string, string>;
}

function withMutationHeaders(opts?: MutationOpts): Record<string, string> | undefined {
  if (!opts) return undefined;
  const h: Record<string, string> = { ...(opts.headers ?? {}) };
  if (opts.idempotencyKey !== undefined) h['Idempotency-Key'] = opts.idempotencyKey;
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
      const body = (await res.json()) as { error?: { code?: string; message?: string } };
      if (body.error) {
        code = body.error.code ?? code;
        message = body.error.message ?? message;
      }
    } catch {
      // ignore parse errors
    }
    const err = new Error(message) as ApiError;
    err.status = res.status;
    err.code = code;
    if (res.status === 401) {
      useAuthStore.getState().logout();
    }
    throw err;
  }
  if (res.status === 204) return undefined as unknown as T;
  return (await res.json()) as T;
}

function buildInit(method: string, body?: unknown, opts?: MutationOpts): RequestInit {
  const init: RequestInit = { method, body: body ? JSON.stringify(body) : null };
  const headers = withMutationHeaders(opts);
  if (headers) init.headers = headers;
  return init;
}

export const api = {
  get: <T,>(path: string) => request<T>(path),
  post: <T,>(path: string, body?: unknown, opts?: MutationOpts) =>
    request<T>(path, buildInit('POST', body, opts)),
  put: <T,>(path: string, body?: unknown, opts?: MutationOpts) =>
    request<T>(path, buildInit('PUT', body, opts)),
  patch: <T,>(path: string, body?: unknown, opts?: MutationOpts) =>
    request<T>(path, buildInit('PATCH', body, opts)),
  delete: <T,>(path: string, body?: unknown, opts?: MutationOpts) =>
    request<T>(path, buildInit('DELETE', body, opts)),
};

// genIdempotencyKey — короткий уникальный ключ для одной мутации.
// Использует crypto.randomUUID если доступен (HTTPS / localhost / dev).
// Fallback — Math.random+timestamp (достаточно для дедупа в окне 24ч,
// бэкенд хеширует ключ перед хранением).
export function genIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  return `${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 10)}`;
}
