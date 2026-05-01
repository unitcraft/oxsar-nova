// ConstructionProgress — план 72.1.45 §4.
//
// Pixel-perfect зеркало legacy `constructions.tpl` L.24-50 +
// `ExtRepair.class.php` L.244-249: progressbar для очереди задач,
// показывает % прогресса и обратный отсчёт. Каждый event_percent_timeout ms
// инкремент value на 1.
//
// Используется в ConstructionsScreen, ShipyardScreen, ResearchScreen,
// RepairScreen — везде где есть очередь задач с start_at/end_at.

import { useEffect, useState } from 'react';
import { formatDuration } from '@/lib/format';

export interface ConstructionProgressProps {
  startAt: string;
  endAt: string;
  /** Опциональная подпись внутри/рядом с баром. */
  label?: string;
}

export function ConstructionProgress({
  startAt,
  endAt,
  label,
}: ConstructionProgressProps) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  const startMs = new Date(startAt).getTime();
  const endMs = new Date(endAt).getTime();
  const totalMs = endMs - startMs;
  const elapsedMs = now - startMs;
  const remainingMs = endMs - now;

  const pct =
    totalMs > 0 ? Math.max(0, Math.min(100, Math.floor((elapsedMs * 100) / totalMs))) : 0;
  const remainingSec = Math.max(0, Math.floor(remainingMs / 1000));

  return (
    <div
      style={{
        position: 'relative',
        height: 14,
        background: '#222',
        border: '1px solid #444',
        borderRadius: 2,
      }}
      title={label}
    >
      <div
        style={{
          width: `${pct}%`,
          height: '100%',
          background: '#3a7',
          transition: 'width 1s linear',
        }}
      />
      <div
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          fontSize: '10px',
          color: '#fff',
          textShadow: '0 0 2px #000',
          pointerEvents: 'none',
        }}
      >
        {pct}% · {formatDuration(remainingSec)}
      </div>
    </div>
  );
}
