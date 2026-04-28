// Общие хуки alliance-экранов origin (план 72 Ф.3 Spring 2 ч.1).
//
// Pure-функции (hasPerm, PERMISSION_KEYS, relationStatusKey) живут в
// permissions.ts — сюда они re-export'ятся для удобства импорта.

import { useQuery } from '@tanstack/react-query';
import { fetchMyAlliance } from '@/api/alliance';
import { QK } from '@/api/query-keys';
import type { AllianceDetail } from '@/api/types';

export {
  PERMISSION_KEYS,
  hasPerm,
  relationStatusKey,
} from './permissions';

export function useMyAlliance(): {
  data: AllianceDetail | null;
  isLoading: boolean;
} {
  const q = useQuery({
    queryKey: QK.alliancesMe(),
    queryFn: fetchMyAlliance,
    staleTime: 30_000,
    refetchInterval: 30_000,
  });
  return { data: q.data ?? null, isLoading: q.isLoading };
}
