// Временные ручные типы, пока не сгенерирован OpenAPI-клиент.
// После `npm run gen:api` этот файл удаляется, импорты переезжают на
// schema.d.ts (см. CLAUDE.md).

export interface User {
  id: string;
  username: string;
  email: string;
}

export interface Tokens {
  access: string;
  refresh: string;
  expires: string;
}

export interface AuthResponse {
  user: User;
  tokens: Tokens;
}

export interface Planet {
  id: string;
  user_id: string;
  is_moon: boolean;
  name: string;
  galaxy: number;
  system: number;
  position: number;
  diameter?: number;
  used_fields?: number;
  max_fields?: number; // план 23: лимит полей для постройки

  temp_min?: number;
  temp_max?: number;
  planet_type?: string;
  metal: number;
  silicon: number;
  hydrogen: number;
  last_res_update: string;
  metal_per_sec: number;
  silicon_per_sec: number;
  hydrogen_per_sec: number;
  metal_cap: number;
  silicon_cap: number;
  hydrogen_cap: number;
  energy_prod: number;
  energy_cons: number;
  energy_remaining: number;
  produce_factor?: number;
  build_factor?: number;
  research_factor?: number;
}

export interface IncomingFleet {
  id: string;
  mission: number;
  dst_galaxy: number;
  dst_system: number;
  dst_position: number;
  dst_is_moon: boolean;
  arrive_at: string;
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

export interface UnmetRequirement {
  kind: 'building' | 'research';
  key: string;
  required: number;
  current: number;
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

export interface Inventory {
  ships: Record<string, number>;
  defense: Record<string, number>;
}

export interface ResearchState {
  queue: QueueItem[];
  levels: Record<string, number>;
}

export type ArtefactState = 'held' | 'delayed' | 'active' | 'expired' | 'consumed' | 'listed';

export interface Artefact {
  id: string;
  user_id: string;
  planet_id?: string | null;
  unit_id: number;
  state: ArtefactState;
  acquired_at: string;
  activated_at?: string | null;
  expire_at?: string | null;
}

export interface GalaxyCell {
  position: number;
  has_planet: boolean;
  planet_name?: string | null;
  has_moon: boolean;
  moon_name?: string | null;
  owner_id?: string | null;
  owner_username?: string | null;
  owner_rank?: number | null;
  debris_metal: number;
  debris_silicon: number;
}

export interface SystemView {
  galaxy: number;
  system: number;
  cells: GalaxyCell[];
}

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
  total_energy: number;
}
