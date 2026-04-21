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

export interface Inventory {
  ships: Record<string, number>;
  defense: Record<string, number>;
}

export interface ResearchState {
  queue: QueueItem[];
  levels: Record<string, number>;
}

export type ArtefactState = 'held' | 'delayed' | 'active' | 'expired' | 'consumed';

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
  debris_metal: number;
  debris_silicon: number;
}

export interface SystemView {
  galaxy: number;
  system: number;
  cells: GalaxyCell[];
}
