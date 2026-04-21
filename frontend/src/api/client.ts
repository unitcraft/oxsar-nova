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
    throw err;
  }
  if (res.status === 204) return undefined as unknown as T;
  return (await res.json()) as T;
}

export const api = {
  get: <T,>(path: string) => request<T>(path),
  post: <T,>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : null }),
  patch: <T,>(path: string, body?: unknown) =>
    request<T>(path, { method: 'PATCH', body: body ? JSON.stringify(body) : null }),
  delete: <T,>(path: string) => request<T>(path, { method: 'DELETE' }),
};
