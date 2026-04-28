// useResolvedPlanet — резолвит активную планету для экранов
// (план 72 Ф.2 Spring 1).
//
// Алгоритм:
//   1) Если URL-параметр :planetId есть — берём его.
//   2) Иначе — текущая планета из Zustand-store.
//   3) Иначе — первая планета из /api/planets (fallback).
//   4) Если /planets пуст или ещё грузится — null.
//
// Возвращает выбранный id + список планет (для рендера right-rail
// и переключателя). Список грузится один раз через TanStack Query.

import { useQuery } from '@tanstack/react-query';
import { fetchPlanets } from '@/api/planets';
import { QK } from '@/api/query-keys';
import type { Planet } from '@/api/types';
import { useCurrentPlanetStore } from '@/stores/currentPlanet';

export interface ResolvedPlanet {
  planetId: string | null;
  planet: Planet | null;
  planets: Planet[];
  isLoading: boolean;
}

export function useResolvedPlanet(urlPlanetId?: string): ResolvedPlanet {
  const stored = useCurrentPlanetStore((s) => s.planetId);
  const q = useQuery({
    queryKey: QK.planets(),
    queryFn: fetchPlanets,
    staleTime: 30_000,
  });
  const planets = q.data ?? [];
  const candidate =
    urlPlanetId ??
    stored ??
    (planets.length > 0 ? (planets[0]?.id ?? null) : null);
  const planet = planets.find((p) => p.id === candidate) ?? null;
  return {
    planetId: candidate,
    planet,
    planets,
    isLoading: q.isLoading,
  };
}
