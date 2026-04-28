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
  // Spring 2 ч.1 — alliance
  alliancesMe: () => ['alliances', 'me'] as const,
  alliancesSearch: (qs: string) => ['alliances', 'search', qs] as const,
  alliance: (id: string) => ['alliances', id] as const,
  allianceApplications: (id: string) =>
    ['alliances', id, 'applications'] as const,
  allianceDescriptions: (id: string) =>
    ['alliances', id, 'descriptions'] as const,
  allianceRanks: (id: string) => ['alliances', id, 'ranks'] as const,
  allianceRelations: (id: string) =>
    ['alliances', id, 'relations'] as const,
  allianceAudit: (id: string, qs: string) =>
    ['alliances', id, 'audit', qs] as const,
  // Spring 2 ч.2 — resource/market/repair/battlestats
  marketRates: () => ['market', 'rates'] as const,
  artMarketOffers: () => ['art-market', 'offers'] as const,
  artMarketCredit: () => ['art-market', 'credit'] as const,
  battlestats: () => ['battlestats'] as const,
};
