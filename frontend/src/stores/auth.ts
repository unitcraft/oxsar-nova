import { create } from 'zustand';

// Auth-стор хранит только токены и userID. Остальные данные игрока —
// в TanStack Query (server state). Никакого смешения.
interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  userId: string | null;
  setTokens: (p: { access: string; refresh: string; userId: string }) => void;
  logout: () => void;
}

const STORAGE_KEY = 'oxsar-auth';

function loadInitial(): Pick<AuthState, 'accessToken' | 'refreshToken' | 'userId'> {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return { accessToken: null, refreshToken: null, userId: null };
    const obj = JSON.parse(raw) as { access?: string; refresh?: string; userId?: string };
    return {
      accessToken: obj.access ?? null,
      refreshToken: obj.refresh ?? null,
      userId: obj.userId ?? null,
    };
  } catch {
    return { accessToken: null, refreshToken: null, userId: null };
  }
}

export const useAuthStore = create<AuthState>((set) => ({
  ...loadInitial(),
  setTokens: ({ access, refresh, userId }) => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ access, refresh, userId }));
    set({ accessToken: access, refreshToken: refresh, userId });
  },
  logout: () => {
    localStorage.removeItem(STORAGE_KEY);
    set({ accessToken: null, refreshToken: null, userId: null });
  },
}));
