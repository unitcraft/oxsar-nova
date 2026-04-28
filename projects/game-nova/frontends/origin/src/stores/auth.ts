// Auth-store origin-фронта (план 72 Ф.1).
//
// Семантика идентична nova-фронту, но с **отдельным localStorage
// namespace** ('oxsar-origin-auth'), чтобы origin-сессия и
// nova-сессия могли существовать одновременно у одного пользователя
// (разные вселенные — разные токены).
//
// План 36: токен выпускает identity-service, /auth/logout кладёт jti
// в Redis-blacklist. Fire-and-forget: logout мгновенный.

import { create } from 'zustand';

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  userId: string | null;
  setTokens: (p: { access: string; refresh: string; userId: string }) => void;
  logout: () => void;
}

const STORAGE_KEY = 'oxsar-origin-auth';
const STORAGE_VERSION = 1;

interface StoredAuth {
  v?: number;
  access?: string;
  refresh?: string;
  userId?: string;
}

function loadInitial(): Pick<AuthState, 'accessToken' | 'refreshToken' | 'userId'> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return { accessToken: null, refreshToken: null, userId: null };
    const obj = JSON.parse(raw) as StoredAuth;
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
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify({ v: STORAGE_VERSION, access, refresh, userId }),
    );
    set({ accessToken: access, refreshToken: refresh, userId });
  },
  logout: () => {
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
