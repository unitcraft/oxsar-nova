// Fetch wrapper для admin-bff.
//
// План 53 BFF: все запросы идут на /api/* (через admin-bff на том же
// домене), credentials: include — браузер автоматически шлёт
// admin_session cookie. Для state-changing методов ставим header
// X-CSRF-Token (читаем из admin_csrf cookie через JS).
//
// Ошибки:
// - 401 → useAuth.clearSession() и пробрасываем ApiError со status=401
//   (UI может перенаправить на /login).
// - не-2xx → ApiError со статусом и парсером тела.
import { useAuth } from '@/store/auth';

const CSRF_COOKIE = 'admin_csrf';

export interface ApiErrorBody {
  error?: string;
  message?: string;
}

export class ApiError extends Error {
  public override readonly name = 'ApiError';
  public readonly status: number;
  public readonly body: ApiErrorBody | null;

  constructor(status: number, body: ApiErrorBody | null, message?: string) {
    super(message ?? body?.message ?? body?.error ?? `HTTP ${status}`);
    this.status = status;
    this.body = body;
  }
}

function readCookie(name: string): string | null {
  const prefix = `${name}=`;
  const parts = document.cookie.split(';');
  for (const raw of parts) {
    const c = raw.trim();
    if (c.startsWith(prefix)) {
      return decodeURIComponent(c.slice(prefix.length));
    }
  }
  return null;
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  body?: unknown;
  signal?: AbortSignal;
}

export async function apiRequest<T>(
  path: string,
  opts: RequestOptions = {},
): Promise<T> {
  const method = opts.method ?? 'GET';
  const headers: Record<string, string> = {
    Accept: 'application/json',
  };
  if (opts.body !== undefined) {
    headers['Content-Type'] = 'application/json';
  }
  if (method !== 'GET') {
    const csrf = readCookie(CSRF_COOKIE);
    if (csrf) {
      headers['X-CSRF-Token'] = csrf;
    }
  }

  const init: RequestInit = {
    method,
    credentials: 'include',
    headers,
  };
  if (opts.body !== undefined) {
    init.body = JSON.stringify(opts.body);
  }
  if (opts.signal) {
    init.signal = opts.signal;
  }

  const resp = await fetch(path, init);

  if (resp.status === 401) {
    useAuth.getState().clearSession();
    throw new ApiError(401, await safeReadJson(resp));
  }
  if (!resp.ok) {
    throw new ApiError(resp.status, await safeReadJson(resp));
  }
  if (resp.status === 204) {
    return undefined as T;
  }
  const ct = resp.headers.get('Content-Type') ?? '';
  if (ct.includes('application/json')) {
    return (await resp.json()) as T;
  }
  return (await resp.text()) as unknown as T;
}

async function safeReadJson(resp: Response): Promise<ApiErrorBody | null> {
  try {
    return (await resp.json()) as ApiErrorBody;
  } catch {
    return null;
  }
}
