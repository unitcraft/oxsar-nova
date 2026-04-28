// Auth store: claims summary текущей сессии.
//
// План 53 (BFF, ревизия 2026-04-27): JWT в браузере НЕ хранится. Сессия
// живёт на admin-bff (Redis), браузер видит только opaque admin_session
// HttpOnly cookie. UI получает summary (sub/username/roles/permissions)
// через GET /auth/me и использует его для permission-guards.
import { create } from 'zustand';

export interface AuthClaims {
  sub: string;
  username: string;
  roles: string[];
  permissions: string[];
}

export type AuthStatus = 'unknown' | 'authenticated' | 'anonymous';

export interface AuthState {
  status: AuthStatus;
  claims: AuthClaims | null;
  csrfToken: string | null;
  setSession: (claims: AuthClaims, csrfToken: string) => void;
  clearSession: () => void;
  setAnonymous: () => void;
  hasPermission: (perm: string) => boolean;
  hasRole: (role: string) => boolean;
}

export const useAuth = create<AuthState>((set, get) => ({
  status: 'unknown',
  claims: null,
  csrfToken: null,
  setSession: (claims, csrfToken) =>
    set({ status: 'authenticated', claims, csrfToken }),
  clearSession: () =>
    set({ status: 'anonymous', claims: null, csrfToken: null }),
  setAnonymous: () => set({ status: 'anonymous', claims: null, csrfToken: null }),
  hasPermission: (perm) => get().claims?.permissions?.includes(perm) ?? false,
  hasRole: (role) => get().claims?.roles?.includes(role) ?? false,
}));
