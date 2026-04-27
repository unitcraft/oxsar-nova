// Auth store: JWT держим в memory (план 53 §Auth-flow). Refresh-токен —
// в httpOnly cookie, недоступен из JS. Permissions/roles берутся из
// claims access-токена и используются permission-guards в UI.
import { create } from 'zustand';

export interface AuthClaims {
  sub: string;
  username: string;
  roles: string[];
  permissions: string[];
  exp: number;
  iat: number;
  jti: string;
}

export interface AuthState {
  accessToken: string | null;
  claims: AuthClaims | null;
  setSession: (token: string, claims: AuthClaims) => void;
  clearSession: () => void;
  hasPermission: (perm: string) => boolean;
  hasRole: (role: string) => boolean;
}

export const useAuth = create<AuthState>((set, get) => ({
  accessToken: null,
  claims: null,
  setSession: (token, claims) => set({ accessToken: token, claims }),
  clearSession: () => set({ accessToken: null, claims: null }),
  hasPermission: (perm) => get().claims?.permissions.includes(perm) ?? false,
  hasRole: (role) => get().claims?.roles.includes(role) ?? false,
}));
