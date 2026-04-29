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

// План 72.2: refresh-flow. На 401 пытаемся обновить access через
// /auth/refresh (refresh-токен из store), затем retry оригинальный
// запрос. Если refresh упал → logout (auth-store очистит токены, App.tsx
// сам редиректит на портал).
//
// Глобальный refreshInFlight предотвращает race-condition: если 5
// параллельных запросов получили 401, они ждут одного refresh-вызова,
// а не делают пять.
let refreshInFlight: Promise<boolean> | null = null;

async function tryRefresh(): Promise<boolean> {
  if (refreshInFlight) return refreshInFlight;
  refreshInFlight = (async () => {
    const refresh = useAuthStore.getState().refreshToken;
    if (!refresh) return false;
    try {
      const res = await fetch('/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh }),
      });
      if (!res.ok) return false;
      const data = (await res.json()) as {
        access_token: string;
        refresh_token: string;
      };
      const userId = useAuthStore.getState().userId ?? '';
      useAuthStore.getState().setTokens({
        access: data.access_token,
        refresh: data.refresh_token,
        userId,
      });
      return true;
    } catch {
      return false;
    } finally {
      refreshInFlight = null;
    }
  })();
  return refreshInFlight;
}

async function doFetch(path: string, init: RequestInit | undefined, token: string | null): Promise<Response> {
  return fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  let token = useAuthStore.getState().accessToken;
  let res = await doFetch(path, init, token);

  // На 401 — попытка refresh и один retry.
  if (res.status === 401 && token) {
    const refreshed = await tryRefresh();
    if (refreshed) {
      token = useAuthStore.getState().accessToken;
      res = await doFetch(path, init, token);
    }
  }

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
      // Refresh failed (или не было refresh-token). Чистим store —
      // App.tsx увидит accessToken === null и редиректит на портал.
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
