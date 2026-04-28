// X-021 (план 71): счётчик новых достижений в navbar.
//
// Backend ещё не возвращает поле new_count (план 70 — реактивация
// achievements отложена). До того момента считаем «новыми» те
// unlocked_at, что произошли после последнего захода игрока на
// экран achievements. Метку храним в localStorage. Когда план 70
// даст серверный счётчик — заменим источник, контракт hook'а
// сохранится.

import { useEffect, useMemo } from 'react';

const STORAGE_KEY = 'oxsar.achievements.lastSeenAt';

export function useNewAchievementCount(
  list: ReadonlyArray<{ unlocked_at?: string | null }>,
): number {
  return useMemo(() => {
    const lastSeen = readLastSeen();
    let count = 0;
    for (const e of list) {
      if (!e.unlocked_at) continue;
      if (new Date(e.unlocked_at).getTime() > lastSeen) count++;
    }
    return count;
  }, [list]);
}

// useMarkAchievementsSeen вызывается компонентом экрана достижений
// при mount — ставит метку «всё, что было до этого момента,
// прочитано». При следующем mount счётчик обнуляется.
export function useMarkAchievementsSeen(): void {
  useEffect(() => {
    try {
      window.localStorage.setItem(STORAGE_KEY, String(Date.now()));
    } catch {
      // QuotaExceeded / private mode — некритично, в худшем случае
      // счётчик не очистится и игрок увидит «X новых» лишний раз.
    }
  }, []);
}

function readLastSeen(): number {
  if (typeof window === 'undefined') return 0;
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) return 0;
    const n = Number(raw);
    return Number.isFinite(n) ? n : 0;
  } catch {
    return 0;
  }
}
