import { create } from 'zustand';
import type { AuthUser, TokenResponse } from '@/api/types';

interface AuthState {
  user: AuthUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  // План 63: setAuth принимает RFC 6749-формат напрямую — user из tokens.user
  // (если присутствует) либо переданный отдельно (для refresh без user).
  setAuth: (tokens: TokenResponse, fallbackUser?: AuthUser) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: localStorage.getItem('access_token'),
  refreshToken: localStorage.getItem('refresh_token'),

  setAuth: (tokens, fallbackUser) => {
    const user = tokens.user ?? fallbackUser ?? null;
    localStorage.setItem('access_token', tokens.access_token);
    localStorage.setItem('refresh_token', tokens.refresh_token);
    set({
      user,
      accessToken: tokens.access_token,
      refreshToken: tokens.refresh_token,
    });
  },

  clearAuth: () => {
    // План 36 Critical-4: revoke refresh через /auth/logout (fire-and-forget).
    const refresh = get().refreshToken;
    if (refresh) {
      void fetch('/auth/logout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh }),
      }).catch(() => undefined);
    }
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    set({ user: null, accessToken: null, refreshToken: null });
  },
}));
