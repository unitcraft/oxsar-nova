// Минимальные DTO origin-фронта (план 72 Ф.2 Spring 1).
//
// Источник истины — projects/game-nova/api/openapi.yaml. Поскольку
// `openapi-typescript` сейчас выдаёт пустую schema.d.ts (план 72 Ф.1
// зарезервировал её под gen:api), описываем здесь только используемые
// в Spring 1 структуры. Поля — snake_case как в API (R1).
//
// При появлении полноценной schema.d.ts (план 72 Ф.7 финализация) эти
// типы заменятся на алиасы из generated-схемы.

export interface Planet {
  id: string;
  user_id: string;
  is_moon: boolean;
  name: string;
  galaxy: number;
  system: number;
  position: number;
  metal: number;
  silicon: number;
  hydrogen: number;
  last_res_update: string;
}

export interface QueueItem {
  id: string;
  planet_id: string;
  unit_id: number;
  target_level: number;
  start_at: string;
  end_at: string;
  status: string;
}

export interface ShipyardQueueItem {
  id: string;
  planet_id: string;
  unit_id: number;
  count: number;
  per_unit_seconds: number;
  start_at: string;
  end_at: string;
  status: string;
}

export interface ShipyardInventory {
  ships: Record<string, number>;
  defense: Record<string, number>;
}

export interface ResearchOverview {
  queue: QueueItem[];
  levels: Record<string, number>;
}

export interface GalaxyCell {
  position: number;
  has_planet: boolean;
  planet_name: string | null;
  has_moon: boolean;
  moon_name: string | null;
  owner_id: string | null;
  owner_username: string | null;
  debris_metal: number;
  debris_silicon: number;
}

export interface SystemView {
  galaxy: number;
  system: number;
  cells: GalaxyCell[];
}

export interface Coords {
  galaxy: number;
  system: number;
  position: number;
  is_moon?: boolean;
}

export type MissionCode = 6 | 7 | 8 | 9 | 10 | 11 | 12 | 15;

export interface FleetDispatchInput {
  src_planet_id: string;
  dst: Coords;
  ships: Record<string, number>;
  carry_metal?: number;
  carry_silicon?: number;
  carry_hydrogen?: number;
  speed_percent: number;
  mission: MissionCode;
}

export interface Fleet {
  id: string;
  owner_user_id: string;
  src_planet_id: string;
  dst_galaxy: number;
  dst_system: number;
  dst_position: number;
  dst_is_moon: boolean;
  mission: number;
  state: 'outbound' | 'hold' | 'returning' | 'done';
  depart_at: string;
  arrive_at: string;
  return_at: string | null;
  carry: { metal: number; silicon: number; hydrogen: number };
  speed_percent: number;
  ships: Record<string, number>;
}

export interface FleetList {
  fleets: Fleet[];
  slots_used: number;
  slots_max: number;
}

export interface UnreadCount {
  count: number;
}

// ===== Alliance (план 67 backend, план 72 Ф.3 Spring 2 ч.1) =====

export interface Alliance {
  id: string;
  tag: string;
  name: string;
  description: string;
  is_open: boolean;
  owner_id: string;
  owner_name: string;
  member_count: number;
  created_at: string;
}

export interface AllianceMember {
  user_id: string;
  username: string;
  rank: string;
  rank_name: string;
  joined_at: string;
}

export interface AllianceDetail {
  alliance: Alliance;
  members: AllianceMember[];
}

export interface AllianceListResult {
  alliances: Alliance[] | null;
  limit: number;
  offset: number;
}

export interface AllianceListFilters {
  q?: string | undefined;
  is_open?: boolean | undefined;
  min_members?: number | undefined;
  max_members?: number | undefined;
  limit?: number | undefined;
  offset?: number | undefined;
}

export interface AllianceApplication {
  id: string;
  alliance_id: string;
  user_id: string;
  username: string;
  message: string;
  created_at: string;
}

export type AllianceViewer = 'member' | 'applicant' | 'outsider';

export interface AllianceDescriptionView {
  description_external: string;
  description_internal: string;
  description_apply: string;
  description: string;
  viewer: AllianceViewer;
}

export type AlliancePermissionKey =
  | 'can_invite'
  | 'can_kick'
  | 'can_send_global_mail'
  | 'can_manage_diplomacy'
  | 'can_change_description'
  | 'can_propose_relations'
  | 'can_manage_ranks';

export type AlliancePermissionMap = Partial<Record<AlliancePermissionKey, boolean>>;

export interface AllianceRank {
  id: string;
  alliance_id: string;
  name: string;
  position: number;
  permissions: AlliancePermissionMap;
}

export type AllianceRelationStatus =
  | 'protection'
  | 'confederation'
  | 'war'
  | 'trade'
  | 'ceasefire';

export type AllianceRelationState = 'outgoing' | 'incoming' | 'active';

export interface AllianceRelation {
  initiator_id: string;
  target_id: string;
  initiator_tag: string;
  target_tag: string;
  initiator_name: string;
  target_name: string;
  status: AllianceRelationStatus;
  state: AllianceRelationState;
  message: string;
  proposed_at: string;
  established_at: string | null;
}

export interface AllianceAuditEntry {
  id: string;
  alliance_id: string;
  actor_id: string | null;
  actor_name: string;
  action: string;
  target_kind: string | null;
  target_id: string | null;
  target_name: string | null;
  payload: Record<string, unknown>;
  created_at: string;
}

export interface AllianceAuditPage {
  entries: AllianceAuditEntry[] | null;
  limit: number;
  offset: number;
}

export interface AllianceTransferCodeIssued {
  expires_at: string;
  ttl_seconds: number;
}

// ===== Resource market / Artefact market / Repair / Battlestats =====
// (план 72 Ф.3 Spring 2 ч.2)

export type ResourceKind = 'metal' | 'silicon' | 'hydrogen';

export interface MarketRates {
  global_rate: { metal: number; silicon: number; hydrogen: number };
  user_rate: number;
  cooldown_until?: string | null;
}

export interface ExchangeResult {
  delta: { metal: number; silicon: number; hydrogen: number };
  rate: number;
}

export interface ArtMarketOffer {
  id: string;
  artefact_id: string;
  artefact_name: string;
  artefact_type: string;
  seller_id: string;
  seller_name: string;
  price: number;
  created_at: string;
}

export interface BattleStatsTotals {
  total: number;
  wins: number;
  losses: number;
  draws: number;
}
