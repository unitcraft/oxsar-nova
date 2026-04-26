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

export const useAuthStore = create<AuthState>((set) => ({
  ...loadInitial(),
  setTokens: ({ access, refresh, userId }) => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ v: STORAGE_VERSION, access, refresh, userId }));
    set({ accessToken: access, refreshToken: refresh, userId });
  },
  logout: () => {
    localStorage.removeItem(STORAGE_KEY);
    set({ accessToken: null, refreshToken: null, userId: null });
  },
}));
