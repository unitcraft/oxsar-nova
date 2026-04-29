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
  // План 72.1 ч.16: legacy-эталон Main (planet block) выводит диаметр,
  // поля, температуру и planet_type (для asset-картинки).
  diameter: number;
  used_fields: number;
  max_fields: number;
  planet_type: string;
  temp_min: number;
  temp_max: number;
  metal: number;
  silicon: number;
  hydrogen: number;
  // Производство и cap — для legacy-TopHeader (per-hour и хранилища).
  metal_per_sec: number;
  silicon_per_sec: number;
  hydrogen_per_sec: number;
  metal_cap: number;
  silicon_cap: number;
  hydrogen_cap: number;
  energy_prod: number;
  energy_cons: number;
  energy_remaining: number;
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
  /** Время следующего уровня каждой технологии (секунды).
   *  Ключ — unit_id как строка. */
  research_seconds?: Record<string, number>;
  /** Стоимость следующего уровня каждой технологии (метал/кремний/водород).
   *  Ключ — unit_id как строка. */
  research_costs?: Record<string, { metal: number; silicon: number; hydrogen: number }>;
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
  metal: number;
  silicon: number;
  hydrogen: number;
  user_rate: number;
}

export interface ExchangeResult {
  from: ResourceKind;
  to: ResourceKind;
  from_amount: number;
  to_amount: number;
  rate: number;
}

export interface ArtMarketOffer {
  id: string;
  seller_id: string;
  unit_id: number;
  price: number;
  created_at: string;
}

export interface ArtMarketCredit {
  credit: number;
}

export interface BattleStatsTotals {
  total: number;
  wins: number;
  losses: number;
  draws: number;
}

// ===== Artefacts (план 72 Ф.4 Spring 3 — S-013) =====

export type ArtefactState =
  | 'held'
  | 'delayed'
  | 'active'
  | 'expired'
  | 'consumed';

export interface Artefact {
  id: string;
  user_id: string;
  planet_id: string | null;
  unit_id: number;
  state: ArtefactState;
  acquired_at: string;
  activated_at: string | null;
  expire_at: string | null;
}

// ===== Highscore / public stats (план 72 Ф.4 — S-023, S-024, S-032) =====

export interface HighscoreEntry {
  user_id: string;
  username: string;
  score: number;
  rank: number;
}

export interface PublicStats {
  online_now: number;
  online_24h: number;
}

// ===== Catalog (план 72 Ф.4 Spring 3) =====

export interface ResCost {
  metal: number;
  silicon: number;
  hydrogen: number;
}

export interface BuildingPreviewRow {
  level: number;
  cost: ResCost;
  build_seconds: number;
  production_per_hour?: number;
  energy_demand?: number;
  energy_output?: number;
}

export interface BuildingCatalogEntry {
  id: number;
  key: string;
  name: string;
  cost_base: ResCost;
  cost_factor: number;
  time_base_seconds: number;
  base_rate_per_hour?: number | null;
  energy_per_level?: number | null;
  energy_output_per_level?: number | null;
  capacity_base?: number | null;
  moon_only?: boolean;
  max_level: number;
  preview: BuildingPreviewRow[];
}

export interface RapidfireEntry {
  target_id: number;
  multiplier: number;
}

export interface ResearchPreviewRow {
  level: number;
  cost: ResCost;
}

export interface UnitCatalogEntry {
  id: number;
  key: string;
  name: string;
  kind: 'ship' | 'defense' | 'research';
  cost: ResCost;
  cost_factor?: number | null;
  attack?: number | null;
  shield?: number | null;
  shell?: number | null;
  cargo?: number | null;
  speed?: number | null;
  fuel?: number | null;
  front?: number | null;
  rapidfire?: RapidfireEntry[];
  preview?: ResearchPreviewRow[];
}

export interface ArtefactEffect {
  type: string;
  field?: string;
  op?: string;
  value?: number;
  active_value?: number;
  inactive_value?: number;
  battle_attack?: number;
  battle_shield?: number;
  battle_shell?: number;
}

export interface ArtefactCatalogEntry {
  id: number;
  key: string;
  name: string;
  effect: ArtefactEffect;
  stackable: boolean;
  max_stacks?: number;
  lifetime_seconds: number;
  delay_seconds?: number;
}

// Techtree (S-021)
export interface TechtreeRequirement {
  kind: 'building' | 'research';
  key: string;
  level: number;
  have: number;
  met: boolean;
}

export interface TechtreeNode {
  key: string;
  kind: 'building' | 'research' | 'ship' | 'defense';
  id: number;
  current_level: number;
  unlocked: boolean;
  requirements: TechtreeRequirement[];
}

export interface Techtree {
  nodes: TechtreeNode[];
}

// Records (S-024)
export interface RecordEntry {
  category: 'building' | 'research' | 'ship' | 'defense' | 'score';
  key: string;
  unit_id?: number;
  holder_id: string;
  holder_name: string;
  value: number;
  my_value: number;
}

// ===== Spring 4 (план 72 Ф.5) =====

// S-034 Friends
export interface Friend {
  user_id: string;
  username: string;
  points: number;
  last_seen?: string;
  alliance_tag?: string;
}

export interface FriendsList {
  friends: Friend[];
}

// S-035 Messages
export type MessageFolder = 'inbox' | 'sent';

export interface Message {
  id: string;
  from_user_id?: string;
  from_username: string;
  subject: string;
  body: string;
  folder: number;
  created_at: string;
  read_at?: string;
  battle_report_id?: string;
  espionage_report_id?: string;
  expedition_report_id?: string;
}

export interface MessagesList {
  messages: Message[] | null;
}

export interface MessageCompose {
  to: string; // username получателя (backend ожидает поле "to")
  subject: string;
  body: string;
}

// S-036 / S-037 Chat
export type ChatChannelKind = 'global' | 'alliance';

export interface ChatMessage {
  id: string;
  channel: string;
  author_id: string;
  author_name: string;
  body: string;
  created_at: string;
  edited_at?: string;
  kind?: 'msg' | 'edit' | 'delete';
}

export interface ChatUnreadCount {
  channel: string;
  unread: number;
  last_read_at?: string;
}

// S-038 Notepad
export interface NotepadContent {
  content: string;
  updated_at: string;
}

// S-039 Search
export interface SearchPlayer {
  user_id: string;
  username: string;
  alliance_tag?: string;
  points: number;
}

export interface SearchAlliance {
  tag: string;
  name: string;
  members: number;
  points: number;
}

export interface SearchPlanet {
  planet_id: string;
  name: string;
  galaxy: number;
  system: number;
  position: number;
  owner: string;
}

export interface SearchResults {
  players: SearchPlayer[];
  alliances: SearchAlliance[];
  planets: SearchPlanet[];
}

export type SearchType = 'player' | 'alliance' | 'planet';

// S-042 Settings
export interface SettingsResponse {
  email: string;
  language: 'ru' | 'en';
  timezone: string;
  vacation_since: string | null;
}

export interface SettingsUpdate {
  email?: string;
  language?: 'ru' | 'en';
  timezone?: string;
}

export interface DeletionCodeResponse {
  expires_at: string;
}

export interface DeletionConfirmRequest {
  code: string;
}

// ===== Resource report (план 72.1 — /resource экран) =====

export interface ResourceBuilding {
  unit_id: number;
  name: string;
  level: number;
  prod_metal: number;
  prod_silicon: number;
  prod_hydrogen: number;
  cons_energy: number;
  factor: number;
  allow_factor: boolean;
}

export interface ResourceReport {
  planet_id: string;
  planet_name: string;
  buildings: ResourceBuilding[];
  basic_metal: number;
  basic_silicon: number;
  basic_hydrogen: number;
  storage_metal: number;
  storage_silicon: number;
  storage_hydrogen: number;
  metal_total: number;
  silicon_total: number;
  hydrogen_total: number;
  metal_per_hour: number;
  silicon_per_hour: number;
  hydrogen_per_hour: number;
  metal_per_day: number;
  silicon_per_day: number;
  hydrogen_per_day: number;
  metal_per_week: number;
  silicon_per_week: number;
  hydrogen_per_week: number;
}

// ===== Repair / Disassemble queue =====

export interface RepairQueueItem {
  id: string;
  planet_id: string;
  unit_id: number;
  count: number;
  mode: 'repair' | 'disassemble';
  start_at: string;
  end_at: string;
  status: string;
}

// ===== Exchange lots (Биржа) =====

export interface ExchangeLot {
  id: string;
  seller_user_id: string;
  seller_username?: string | null;
  artifact_unit_id: number;
  quantity: number;
  price_oxsarit: number;
  unit_price_oxsarit?: number;
  status: 'active' | 'sold' | 'cancelled' | 'expired';
  expires_at: string;
  created_at: string;
}

export interface ExchangeLotsResult {
  lots: ExchangeLot[];
  next_cursor?: string | null;
}

// ===== Spring 4 ч.2 (план 72 Ф.5) =====

// S-040 Officer — каталог + state. Backend (internal/officer/service.go)
// возвращает Entry с полным набором полей (title/description/duration_days/
// cost_credit/effect/activated_at/expires_at). OpenAPI обновлён под это
// в Spring 4 ч.2.
export interface Officer {
  key: string;
  title: string;
  description: string;
  duration_days: number;
  cost_credit: number;
  effect?: Record<string, number> | null;
  activated_at?: string | null;
  expires_at?: string | null;
}

export interface OfficersList {
  officers: Officer[] | null;
}

export interface OfficerActivateRequest {
  auto_renew?: boolean;
}

// S-041 Profession — backend (internal/profession/service.go) DTO:
// {key, label, bonus, malus} для list + {profession, label,
// next_change_allowed} для me. Bonus/malus — мапа техн.ключ → дельта
// уровня (например, metalmine: +5, gun: -3).
export interface Profession {
  key: string;
  label: string;
  bonus?: Record<string, number>;
  malus?: Record<string, number>;
}

export interface ProfessionsList {
  professions: Profession[];
}

export interface ProfessionInfo {
  profession: string;
  label: string;
  next_change_allowed?: string | null;
}
