// TanStack Query keys для origin-фронта (план 72 Ф.2 Spring 1).
//
// Конвенция: первый сегмент — домен ('planets', 'research', ...),
// остальные — параметризация (planetId, galaxy/system). Invalidation
// делается по корневому домену либо по точечному ключу.

export const QK = {
  planets: () => ['planets'] as const,
  planet: (id: string) => ['planet', id] as const,
  buildingQueue: (planetId: string) => ['buildings', 'queue', planetId] as const,
  buildingsOverview: (planetId: string) =>
    ['buildings', 'overview', planetId] as const,
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
  // Spring 3 (Ф.4) — artefacts / records / stats / catalog
  artefacts: () => ['artefacts'] as const,
  highscore: () => ['highscore'] as const,
  highscoreMe: () => ['highscore', 'me'] as const,
  publicStats: () => ['stats'] as const,
  buildingCatalog: (type: string | number) =>
    ['catalog', 'building', type] as const,
  unitCatalog: (type: string | number) => ['catalog', 'unit', type] as const,
  artefactCatalog: (type: string | number) =>
    ['catalog', 'artefact', type] as const,
  techtree: (planetId?: string) => ['techtree', planetId ?? ''] as const,
  records: () => ['records'] as const,
  // Spring 4 (Ф.5) — communication / notes / search / settings
  friends: () => ['friends'] as const,
  messages: (folder: 'inbox' | 'sent') => ['messages', folder] as const,
  message: (id: string) => ['messages', 'detail', id] as const,
  chatHistory: (kind: 'global' | 'alliance') =>
    ['chat', kind, 'history'] as const,
  chatUnread: (kind: 'global' | 'alliance') =>
    ['chat', kind, 'unread'] as const,
  notepad: () => ['notepad'] as const,
  search: (type: string, q: string) => ['search', type, q] as const,
  settings: () => ['settings'] as const,
  // Spring 4 ч.2 — premium / static / utilities
  officers: () => ['officers'] as const,
  professions: () => ['professions'] as const,
  professionMe: () => ['professions', 'me'] as const,
  // Plan 72.1: /api/me для TopHeader (credit) и MainScreen (profession).
  me: () => ['me'] as const,
  // Новые экраны: resource / disassemble / exchange
  resourceReport: (planetId: string) => ['resource-report', planetId] as const,
  repairQueue: (planetId: string) => ['repair', 'queue', planetId] as const,
  repairDamaged: (planetId: string) => ['repair', 'damaged', planetId] as const,
  exchangeLots: (params: string) => ['exchange', 'lots', params] as const,
};
