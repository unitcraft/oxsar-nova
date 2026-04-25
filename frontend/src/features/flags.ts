// Клиентский feature-flags. План 31 Ф.2.
//
// Источник: /api/features (без auth) — backend читает configs/features.yaml.
// Кэшируем через TanStack Query со staleTime=5мин: флаги меняются
// редко (требуется restart backend), но повторно дёргать /api/features
// при каждом mount не нужно.
//
// Использование:
//
//   import { useFeatureFlag } from '@/features/flags';
//
//   const goalEngine = useFeatureFlag('goal_engine');
//   if (goalEngine) return <NewGoalScreen />;
//   return <LegacyAchievementsScreen />;

import { useQuery } from '@tanstack/react-query';
import { api } from '@/api/client';

export interface Flag {
  enabled: boolean;
  description?: string;
}

export interface FeaturesResponse {
  features: Record<string, Flag>;
  enabled: string[];
}

const QUERY_KEY = ['features'];
const STALE_TIME_MS = 5 * 60 * 1000; // 5 минут

/** Подгрузить весь набор флагов. Используется обычно через useFeatureFlag. */
export function useFeatures() {
  return useQuery({
    queryKey: QUERY_KEY,
    queryFn: () => api.get<FeaturesResponse>('/api/features'),
    staleTime: STALE_TIME_MS,
    // На stale возвращаем кэш мгновенно, фоном обновляем — фронт не моргает.
    refetchOnWindowFocus: false,
  });
}

/**
 * Проверить, включен ли feature flag.
 *
 * Пока запрос загружается — возвращает false (fail-closed: не показываем
 * нестабильную фичу). Это безопаснее, чем рисовать новый UI до
 * подтверждения от сервера.
 *
 * @example
 *   const goalEngine = useFeatureFlag('goal_engine');
 *   return goalEngine ? <NewGoals /> : <OldAchievements />;
 */
export function useFeatureFlag(key: string): boolean {
  const q = useFeatures();
  return q.data?.features?.[key]?.enabled === true;
}

/** Получить все включённые флаги (массив ключей) — для debug-панели. */
export function useEnabledFlagKeys(): string[] {
  const q = useFeatures();
  return q.data?.enabled ?? [];
}
