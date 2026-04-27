import { create } from 'zustand';
import type { AuthUser } from '@/api/types';

interface AuthState {
  user: AuthUser | null;
  accessToken: string | null;
  refreshToken: string | null;
  setAuth: (user: AuthUser, tokens: { access: string; refresh: string }) => void;
  clearAuth: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: localStorage.getItem('access_token'),
  refreshToken: localStorage.getItem('refresh_token'),

  setAuth: (user, tokens) => {
    localStorage.setItem('access_token', tokens.access);
    localStorage.setItem('refresh_token', tokens.refresh);
    set({ user, accessToken: tokens.access, refreshToken: tokens.refresh });
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
