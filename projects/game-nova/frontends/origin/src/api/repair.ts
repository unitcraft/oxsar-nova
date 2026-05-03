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

// План 72.1.56 B6: legacy 1:1 partial-repair. Если quantity > 0 — чинит
// N штук (backend clamp до damaged_count); 0/опущен — чинит всех damaged.
export function repairUnits(
  planetId: string,
  unitId: number,
  quantity?: number,
): Promise<RepairQueueItem> {
  const body: { unit_id: number; quantity?: number } = { unit_id: unitId };
  if (quantity !== undefined && quantity > 0) {
    body.quantity = quantity;
  }
  return api.post<RepairQueueItem>(
    `/api/planets/${planetId}/repair/repair`,
    body,
    { idempotencyKey: newIdempotencyKey() },
  );
}

// План 72.1.25: cancel задачи в очереди (legacy abortRepair / abortDisassemble).
export function cancelRepairQueue(
  planetId: string,
  queueId: string,
): Promise<void> {
  return api.delete<void>(
    `/api/planets/${planetId}/repair/queue/${queueId}`,
    { idempotencyKey: newIdempotencyKey() },
  );
}

// План 72.1.25: VIP мгновенный старт за credit (legacy startEventVIP).
export interface RepairVIPResult {
  queue_id: string;
  mode: string;
  credit_debit: number;
  new_end_at: string;
}

export function startRepairVIP(
  planetId: string,
  queueId: string,
): Promise<RepairVIPResult> {
  return api.post<RepairVIPResult>(
    `/api/planets/${planetId}/repair/queue/${queueId}/vip`,
    undefined,
    { idempotencyKey: newIdempotencyKey() },
  );
}

// План 72.1.25: legacy `getCreditImmStartShipyard` — публичная формула,
// чтобы UI показал стоимость до клика.
export function vipCreditCost(quantity: number): number {
  if (quantity <= 0) return 10;
  let v = Math.pow(quantity, 0.8);
  v = Math.round(v / 10) * 10;
  if (v < 10) v = 10;
  if (v > 100000) v = 100000;
  return v;
}
