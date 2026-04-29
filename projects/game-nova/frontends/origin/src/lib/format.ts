// Форматтеры origin-фронта (план 72.1 ч.20).
//
// formatNumber: округляет до целого, разделяет тысячи точкой.
// Pixel-perfect клон legacy ru-локали: PHP fNumber() →
// number_format с THOUSANDS_SEPERATOR='.' (Functions.inc.php).
// На скрине legacy ru видно «409.600», «4.500.000».
//
// formatDuration: формат HH:MM:SS как в legacy (на скрине Время «00:06:23»).

const THOUSANDS_SEP = '.';

export function formatNumber(n: number): string {
  if (!Number.isFinite(n)) return '—';
  const sign = n < 0 ? '-' : '';
  const abs = Math.abs(Math.floor(n));
  return sign + abs.toString().replace(/\B(?=(\d{3})+(?!\d))/g, THOUSANDS_SEP);
}

export function formatCoords(g: number, s: number, p: number): string {
  return `[${g}:${s}:${p}]`;
}

// formatDuration в формате HH:MM:SS (как legacy resource.tpl и required_res_table).
export function formatDuration(totalSeconds: number): string {
  if (!Number.isFinite(totalSeconds) || totalSeconds < 0) return '—';
  const sec = Math.floor(totalSeconds);
  const days = Math.floor(sec / 86400);
  const hours = Math.floor((sec % 86400) / 3600);
  const minutes = Math.floor((sec % 3600) / 60);
  const seconds = sec % 60;
  if (days > 0) {
    return `${days}д ${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
  }
  return `${pad(hours)}:${pad(minutes)}:${pad(seconds)}`;
}

function pad(n: number): string {
  return n < 10 ? `0${n}` : String(n);
}

export function secondsUntil(iso: string): number {
  const t = Date.parse(iso);
  if (Number.isNaN(t)) return 0;
  return Math.max(0, Math.round((t - Date.now()) / 1000));
}
