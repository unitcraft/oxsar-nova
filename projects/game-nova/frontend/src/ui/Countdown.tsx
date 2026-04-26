import { useEffect, useState } from 'react';

function fmt(sec: number): string {
  if (sec <= 0) return '00:00:00';
  const h = Math.floor(sec / 3600);
  const m = Math.floor((sec % 3600) / 60);
  const s = sec % 60;
  if (h > 0) return `${String(h).padStart(2,'0')}:${String(m).padStart(2,'0')}:${String(s).padStart(2,'0')}`;
  return `${String(m).padStart(2,'0')}:${String(s).padStart(2,'0')}`;
}

interface CountdownProps {
  finishAt: string; // ISO timestamp
  onDone?: () => void;
}

export function Countdown({ finishAt, onDone }: CountdownProps) {
  const [sec, setSec] = useState(() => Math.max(0, Math.round((new Date(finishAt).getTime() - Date.now()) / 1000)));

  useEffect(() => {
    if (sec <= 0) { onDone?.(); return; }
    const t = setInterval(() => {
      setSec((s) => {
        if (s <= 1) { clearInterval(t); onDone?.(); return 0; }
        return s - 1;
      });
    }, 1000);
    return () => clearInterval(t);
  }, [finishAt, onDone, sec]);

  const urgent = sec < 60;
  return <span className={`ox-timer${urgent ? ' urgent' : ''}`}>{fmt(sec)}</span>;
}
