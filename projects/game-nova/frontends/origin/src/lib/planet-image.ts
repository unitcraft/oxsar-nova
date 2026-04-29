// Маппинг planet_type -> картинка планеты (тема `standard` legacy-PHP).
//
// План 72.1 ч.17: legacy показывает `<planet_type>NN.jpg` (см.
// projects/game-legacy-php/src/templates/standard/main.tpl §111 и
// na_planet.picture). Origin делает то же, но детерминированно по
// planet.id чтобы у каждой планеты была стабильная картинка между
// перезагрузками страницы (ваниант 01..NN не меняется).
//
// Соответствие количеству вариантов в legacy:
//   dschjungelplanet: 10, eisplanet: 10, gasplanet: 8,
//   normaltempplanet: 7, trockenplanet: 10, wasserplanet: 9,
//   wuestenplanet: 4.
// Ассеты — `public/assets/origin/images/planets/<file>.jpg` (большие)
// и `public/assets/origin/images/planets/small/s_<file>.jpg` (sidebar).
//
// Если planet_type неизвестен (старые/импортные планеты) — fallback
// на `unformed.jpg`.

const VARIANTS: Record<string, number> = {
  dschjungelplanet: 10,
  eisplanet: 10,
  gasplanet: 8,
  normaltempplanet: 7,
  trockenplanet: 10,
  wasserplanet: 9,
  wuestenplanet: 4,
};

const PLANETS_BASE = '/assets/origin/images/planets';

// Стабильный 32-битный хеш строки (FNV-1a). Используется только для
// выбора варианта картинки — не криптография.
function hash32(s: string): number {
  let h = 0x811c9dc5;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 0x01000193);
  }
  return h >>> 0;
}

function variantIndex(seed: string, count: number): number {
  if (count <= 0) return 1;
  return (hash32(seed) % count) + 1;
}

function pad2(n: number): string {
  return n < 10 ? `0${n}` : String(n);
}

export function planetImageUrl(
  planetType: string | null | undefined,
  seed: string,
): string {
  if (!planetType || !(planetType in VARIANTS)) {
    return `${PLANETS_BASE}/unformed.jpg`;
  }
  const idx = variantIndex(seed, VARIANTS[planetType]!);
  return `${PLANETS_BASE}/${planetType}${pad2(idx)}.jpg`;
}

export function planetImageSmallUrl(
  planetType: string | null | undefined,
  seed: string,
): string {
  if (!planetType || !(planetType in VARIANTS)) {
    return `${PLANETS_BASE}/small/s_unformed.jpg`;
  }
  const idx = variantIndex(seed, VARIANTS[planetType]!);
  return `${PLANETS_BASE}/small/s_${planetType}${pad2(idx)}.jpg`;
}

export function moonImageUrl(): string {
  return `${PLANETS_BASE}/mond.jpg`;
}

export function moonImageSmallUrl(): string {
  return `${PLANETS_BASE}/small/s_mond.jpg`;
}
