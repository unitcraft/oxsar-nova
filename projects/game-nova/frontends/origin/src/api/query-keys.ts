// TanStack Query keys для origin-фронта (план 72 Ф.2 Spring 1).
//
// Конвенция: первый сегмент — домен ('planets', 'research', ...),
// остальные — параметризация (planetId, galaxy/system). Invalidation
// делается по корневому домену либо по точечному ключу.

export const QK = {
  planets: () => ['planets'] as const,
  planet: (id: string) => ['planet', id] as const,
  buildingQueue: (planetId: string) => ['buildings', 'queue', planetId] as const,
  research: () => ['research'] as const,
  shipyardQueue: (planetId: string) => ['shipyard', 'queue', planetId] as const,
  shipyardInventory: (planetId: string) =>
    ['shipyard', 'inventory', planetId] as const,
  galaxy: (g: number, s: number) => ['galaxy', g, s] as const,
  fleet: () => ['fleet'] as const,
  unreadCount: () => ['messages', 'unread-count'] as const,
};
