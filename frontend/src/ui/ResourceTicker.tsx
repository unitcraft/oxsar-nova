import { useEffect, useRef, useState } from 'react';

function fmtNum(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(2) + 'M';
  if (n >= 10_000) return (n / 1_000).toFixed(1) + 'K';
  return Math.floor(n).toLocaleString('ru-RU');
}

interface ResourceTickerProps {
  value: number;
  ratePerSec: number;
  cap?: number;
}

export function ResourceTicker({ value, ratePerSec, cap }: ResourceTickerProps) {
  const [cur, setCur] = useState(value);
  const ref = useRef({ value, ratePerSec, cap, last: Date.now() });

  useEffect(() => {
    ref.current = { ...ref.current, value, ratePerSec, cap };
    setCur(value);
  }, [value, ratePerSec, cap]);

  useEffect(() => {
    if (ratePerSec <= 0) return;
    let raf: number;
    const tick = () => {
      const now = Date.now();
      const dt = (now - ref.current.last) / 1000;
      ref.current.last = now;
      setCur((prev) => {
        const next = prev + ref.current.ratePerSec * dt;
        return ref.current.cap ? Math.min(next, ref.current.cap) : next;
      });
      raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
  }, [ratePerSec]);

  const atCap = cap != null && cur >= cap * 0.99;
  const nearCap = cap != null && cur >= cap * 0.9 && !atCap;
  const cls = atCap ? 'cap-full' : nearCap ? 'cap-warn' : '';

  return <span className={`val${cls ? ` ${cls}` : ''}`}>{fmtNum(cur)}</span>;
}
