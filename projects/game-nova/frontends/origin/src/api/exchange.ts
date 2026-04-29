import { api } from './client';
import type { ExchangeLotsResult } from './types';

export function fetchExchangeLots(params?: {
  artifact_unit_id?: number;
  status?: string;
  limit?: number;
  cursor?: string;
}): Promise<ExchangeLotsResult> {
  const qs = new URLSearchParams();
  if (params?.artifact_unit_id != null)
    qs.set('artifact_unit_id', String(params.artifact_unit_id));
  if (params?.status) qs.set('status', params.status);
  if (params?.limit != null) qs.set('limit', String(params.limit));
  if (params?.cursor) qs.set('cursor', params.cursor);
  const query = qs.toString();
  return api.get<ExchangeLotsResult>(`/api/exchange/lots${query ? `?${query}` : ''}`);
}
