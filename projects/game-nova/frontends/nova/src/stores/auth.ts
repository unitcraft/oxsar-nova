import { create } from 'zustand';

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  userId: string | null;
  setTokens: (p: { access: string; refresh: string; userId: string }) => void;
  logout: () => void;
}

const STORAGE_KEY = 'oxsar-auth';
const STORAGE_VERSION = 2; // bump при изменении формата

function loadInitial(): Pick<AuthState, 'accessToken' | 'refreshToken' | 'userId'> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return { accessToken: null, refreshToken: null, userId: null };
    const obj = JSON.parse(raw) as { v?: number; access?: string; refresh?: string; userId?: string };
    if (obj.v !== STORAGE_VERSION) {
      localStorage.removeItem(STORAGE_KEY);
      return { accessToken: null, refreshToken: null, userId: null };
    }
    return {
      accessToken: obj.access ?? null,
      refreshToken: obj.refresh ?? null,
      userId: obj.userId ?? null,
    };
  } catch {
    localStorage.removeItem(STORAGE_KEY);
    return { accessToken: null, refreshToken: null, userId: null };
  }
}

export const useAuthStore = create<AuthState>((set, get) => ({
  ...loadInitial(),
  setTokens: ({ access, refresh, userId }) => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ v: STORAGE_VERSION, access, refresh, userId }));
    set({ accessToken: access, refreshToken: refresh, userId });
  },
  logout: () => {
    // План 36 Critical-4: revoke refresh-token через /auth/logout —
    // identity-service кладёт его jti в Redis-blacklist. После этого
    // украденный refresh-token не работает до истечения TTL.
    //
    // Fire-and-forget: не ждём ответа, чтобы logout был мгновенным.
    // Если запрос упал (offline / identity-service недоступен) — токен всё
    // равно стирается локально, на сервере он истечёт по TTL (7d/30d).
    const refresh = get().refreshToken;
    if (refresh) {
      void fetch('/auth/logout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh }),
      }).catch(() => undefined);
    }
    localStorage.removeItem(STORAGE_KEY);
    set({ accessToken: null, refreshToken: null, userId: null });
  },
}));
