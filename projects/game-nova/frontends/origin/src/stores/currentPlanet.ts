// «Активная планета» origin-фронта (план 72 Ф.2 Spring 1).
//
// В legacy-PHP активная планета хранится в сессии на сервере и
// меняется при клике в правом списке #planets. На SPA это локальный
// state — Zustand с persist в localStorage (отдельный namespace,
// независимый от auth).
//
// Если planetId не выбран явно — экраны используют первую планету
// из /api/planets.

import { create } from 'zustand';

interface CurrentPlanetState {
  planetId: string | null;
  set: (id: string | null) => void;
}

const STORAGE_KEY = 'oxsar-origin-current-planet';

function loadInitial(): string | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw && raw.length > 0 ? raw : null;
  } catch {
    return null;
  }
}

export const useCurrentPlanetStore = create<CurrentPlanetState>((set) => ({
  planetId: loadInitial(),
  set: (id) => {
    if (id === null) {
      localStorage.removeItem(STORAGE_KEY);
    } else {
      localStorage.setItem(STORAGE_KEY, id);
    }
    set({ planetId: id });
  },
}));
