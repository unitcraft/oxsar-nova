import { api } from './client';
import { newIdempotencyKey } from './idempotency';
import type { RepairQueueItem } from './types';

export async function fetchRepairQueue(
  planetId: string,
): Promise<RepairQueueItem[]> {
  const res = await api.get<{ queue: RepairQueueItem[] }>(
    `/api/planets/${planetId}/repair/queue`,
  );
  return res.queue ?? [];
}

export async function fetchDamagedShips(
  planetId: string,
): Promise<{ unit_id: number; count: number; damaged: number; shell_percent: number }[]> {
  const res = await api.get<{
    damaged: { unit_id: number; count: number; damaged: number; shell_percent: number }[];
  }>(`/api/planets/${planetId}/repair/damaged`);
  return res.damaged ?? [];
}

export function disassembleUnits(
  planetId: string,
  unitId: number,
  count: number,
): Promise<RepairQueueItem> {
  return api.post<RepairQueueItem>(
    `/api/planets/${planetId}/repair/disassemble`,
    { unit_id: unitId, count },
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function repairUnits(
  planetId: string,
  unitId: number,
): Promise<RepairQueueItem> {
  return api.post<RepairQueueItem>(
    `/api/planets/${planetId}/repair/repair`,
    { unit_id: unitId },
    { idempotencyKey: newIdempotencyKey() },
  );
}
