// Форматтеры origin-фронта (план 72 Ф.2 Spring 1).
//
// formatNumber: округляет до целого, разделяет тысячи U+00A0 NO-BREAK
// SPACE (как в legacy: PHP number_format даёт обычный пробел, шаблон
// затем меняет на &nbsp; — на SPA можно сразу класть NBSP, чтобы число
// не переносилось при wrap).
//
// formatDuration / secondsUntil: для таймеров миссий, очередей строек.

const NBSP = ' ';

export function formatNumber(n: number): string {
  if (!Number.isFinite(n)) return '—';
  const sign = n < 0 ? '-' : '';
  const abs = Math.abs(Math.floor(n));
  return sign + abs.toString().replace(/\B(?=(\d{3})+(?!\d))/g, NBSP);
}

export function formatCoords(g: number, s: number, p: number): string {
  return `[${g}:${s}:${p}]`;
}

export function formatDuration(totalSeconds: number): string {
  if (!Number.isFinite(totalSeconds) || totalSeconds < 0) return '—';
  const sec = Math.floor(totalSeconds);
  const days = Math.floor(sec / 86400);
  const hours = Math.floor((sec % 86400) / 3600);
  const minutes = Math.floor((sec % 3600) / 60);
  const seconds = sec % 60;
  const parts: string[] = [];
  if (days > 0) parts.push(`${days}д`);
  if (hours > 0 || days > 0) parts.push(`${hours}ч`);
  if (minutes > 0 || hours > 0 || days > 0) parts.push(`${minutes}м`);
  parts.push(`${seconds}с`);
  return parts.join(' ');
}

export function secondsUntil(iso: string): number {
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return 0;
  return Math.max(0, Math.round((t - Date.now()) / 1000));
}
