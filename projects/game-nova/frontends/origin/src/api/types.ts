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
