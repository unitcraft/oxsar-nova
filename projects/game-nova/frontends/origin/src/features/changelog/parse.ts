// Парсер CHANGELOG.md → Release[] (план 72 Ф.5 Spring 4 ч.2 — S-044).
//
// Lightweight: парсит только `## version — date` заголовки и body
// между ними. Без полноценного MD-renderer'а — формат CHANGELOG.md
// фиксированный, мы его и пишем.

export interface Release {
  version: string;
  changes: string;
}

export function parseChangelog(md: string): Release[] {
  const lines = md.split(/\r?\n/);
  const releases: Release[] = [];
  let current: Release | null = null;
  let buf: string[] = [];

  const flush = () => {
    if (current) {
      current.changes = buf.join('\n').trim();
      releases.push(current);
    }
  };

  for (const line of lines) {
    const m = /^##\s+(.+?)\s*$/.exec(line);
    if (m) {
      flush();
      current = { version: m[1] ?? '', changes: '' };
      buf = [];
      continue;
    }
    if (current) {
      buf.push(line);
    }
  }
  flush();
  return releases;
}
