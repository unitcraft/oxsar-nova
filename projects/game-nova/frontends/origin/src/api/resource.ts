import { api } from './client';
import type { ResourceReport } from './types';

export function fetchResourceReport(planetId: string): Promise<ResourceReport> {
  return api.get<ResourceReport>(`/api/planets/${planetId}/resource-report`);
}

export function updateResourceFactors(
  planetId: string,
  factors: Record<string, number>,
): Promise<{ status: string }> {
  return api.post<{ status: string }>(
    `/api/planets/${planetId}/resource-update`,
    { factors },
  );
}
