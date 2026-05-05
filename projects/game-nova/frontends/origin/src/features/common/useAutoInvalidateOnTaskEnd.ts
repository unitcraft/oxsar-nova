// План 72.1.59: общий хук для авто-инвалидации очередей по окончании
// ближайшей задачи.
//
// Проблема: legacy `Page.class.php` пересчитывает события на каждом
// запросе, поэтому в legacy F5 = страница свежая. У нас экраны
// строятся на TanStack-Query кешах без `refetchInterval` — когда
// задача доезжает до 100%, фронт остаётся со старым snapshot'ом.
//
// Решение: следим за `end_at` задач и ставим setTimeout на ближайший.
// При срабатывании дёргаем переданный invalidate-callback, экран
// перерисовывается со свежими данными.
//
// Используется в: ConstructionsScreen, ResearchScreen, BuildPanel
// (Shipyard/Defense), RepairScreen, DisassembleScreen, MonitorPlanet,
// FleetOperations.

import { useEffect, useRef } from 'react';

interface TaskWithEnd {
  end_at: string;
}

export function useAutoInvalidateOnTaskEnd(
  tasks: ReadonlyArray<TaskWithEnd>,
  invalidate: () => void,
): void {
  // Храним последнюю версию invalidate в ref, чтобы не перезапускать
  // эффект при каждом рендере (inline-функции пересоздаются каждый раз).
  const invalidateRef = useRef(invalidate);
  useEffect(() => { invalidateRef.current = invalidate; });

  useEffect(() => {
    if (tasks.length === 0) return;
    const earliestEnd = Math.min(
      ...tasks.map((t) => new Date(t.end_at).getTime()),
    );
    const delay = earliestEnd - Date.now();
    if (delay <= 0) {
      invalidateRef.current();
      return;
    }
    // +500ms запас, чтобы backend event-loop успел обработать tick.
    const id = setTimeout(() => invalidateRef.current(), delay + 500);
    return () => clearTimeout(id);
  }, [tasks]);
}
