// API-модуль catalog endpoints (план 72 Ф.4 Spring 3, S-014/S-018/
// S-019 + S-021 techtree + S-024 records).
//
// Endpoints (openapi.yaml):
//   GET /api/buildings/catalog/{type}    → BuildingCatalogEntry
//   GET /api/units/catalog/{type}        → UnitCatalogEntry (ship|defense|research)
//   GET /api/artefacts/catalog/{type}    → ArtefactCatalogEntry
//   GET /api/techtree[?planet_id=...]    → Techtree
//   GET /api/records                     → { records: RecordEntry[] }
//
// Catalog данные не меняются в runtime — используем `staleTime: Infinity`
// (или большое значение) на стороне TanStack Query для агрессивного
// кеширования.

import { api } from './client';
import type {
  ArtefactCatalogEntry,
  BuildingCatalogEntry,
  RecordEntry,
  Techtree,
  UnitCatalogEntry,
} from './types';

export function fetchBuildingCatalog(
  type: string | number,
): Promise<BuildingCatalogEntry> {
  return api.get<BuildingCatalogEntry>(`/api/buildings/catalog/${type}`);
}

export function fetchUnitCatalog(
  type: string | number,
): Promise<UnitCatalogEntry> {
  return api.get<UnitCatalogEntry>(`/api/units/catalog/${type}`);
}

export function fetchArtefactCatalog(
  type: string | number,
): Promise<ArtefactCatalogEntry> {
  return api.get<ArtefactCatalogEntry>(`/api/artefacts/catalog/${type}`);
}

export function fetchTechtree(planetId?: string): Promise<Techtree> {
  const qs = planetId
    ? `?planet_id=${encodeURIComponent(planetId)}`
    : '';
  return api.get<Techtree>(`/api/techtree${qs}`);
}

export function fetchRecords(): Promise<{ records: RecordEntry[] | null }> {
  return api.get<{ records: RecordEntry[] | null }>('/api/records');
}
