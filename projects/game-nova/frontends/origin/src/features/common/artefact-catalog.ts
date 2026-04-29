// Локальный справочник имён артефактов для карточек инвентаря
// (план 72 Ф.4 Spring 3, S-013 ArtefactsScreen).
//
// Полный catalog с эффектом / lifetime / delay живёт на backend и
// доступен через `GET /api/artefacts/catalog/{type}` (см. S-014
// ArtefactInfoScreen). На S-013 для каждой карточки делать отдельный
// catalog-запрос неэффективно (N+1), поэтому здесь — минимальная
// карта unit_id → i18n-ключи имени/описания.
//
// Если unit_id не найден — UI показывает «Артефакт #{id}» (фолбэк).

export interface ArtefactCatalogEntry {
  id: number;
  /** namespace.key с базовым именем (info.{name}) */
  i18nName: string;
  /** namespace.key с коротким описанием (info.{name}Desc) — может отсутствовать */
  i18nDesc?: string;
  /** namespace.key с длинным описанием (info.{name}FullDesc) — может отсутствовать */
  i18nFullDesc?: string;
  /** Можно ли активировать (toggle on/off) или только использовать одноразово. */
  activatable?: boolean;
  /**
   * Имя файла-картинки в `/assets/origin/images/buildings/<image>` (без
   * расширения; расширение угадывается через IMAGE_EXTS — legacy кладёт
   * gif/png рядом). Совпадает с `key:` в configs/artefacts.yml. План 72.1
   * ч.17: pixel-perfect рендер артефактов (legacy
   * `templates/standard/artefacts.tpl` показывает картинку для каждого).
   */
  image?: string;
}

// Минимальный набор: соответствия id ↔ i18n взяты из info.* в
// configs/i18n/ru.yml. Расширяется по мере появления новых артефактов
// в configs/balance/{default,origin}.yaml.
export const ARTEFACT_CATALOG: ArtefactCatalogEntry[] = [
  {
    id: 3001,
    i18nName: 'info.catalyst',
    i18nDesc: 'info.catalystDesc',
    i18nFullDesc: 'info.catalystFullDesc',
    activatable: true,
    image: 'catalyst',
  },
  {
    id: 3002,
    i18nName: 'info.atomicDensifier',
    i18nDesc: 'info.atomicDensifierDesc',
    i18nFullDesc: 'info.atomicDensifierFullDesc',
    activatable: true,
    image: 'atomic_densifier',
  },
  {
    id: 3110,
    i18nName: 'info.assemblyModule3110',
    i18nDesc: 'info.assemblyModule3110Desc',
    i18nFullDesc: 'info.assemblyModule3110FullDesc',
    image: 'assembly_module',
  },
  {
    id: 421,
    i18nName: 'info.assemblyModule421',
    i18nDesc: 'info.assemblyModule421Desc',
    i18nFullDesc: 'info.assemblyModule421FullDesc',
    image: 'assembly_module',
  },
];

export function findArtefactCatalog(
  unitId: number,
): ArtefactCatalogEntry | undefined {
  return ARTEFACT_CATALOG.find((e) => e.id === unitId);
}

// План 72.1 ч.17: возвращает URL картинки артефакта или null.
// В legacy картинки лежат в нескольких форматах (gif/png), поэтому
// проверяем оба расширения через onerror-fallback в JSX.
export function artefactImageUrl(unitId: number): string | null {
  const entry = findArtefactCatalog(unitId);
  if (!entry?.image) return null;
  return `/assets/origin/images/buildings/${entry.image}.gif`;
}

// Fallback URL — если основной gif не нашёлся, пробуем png.
export function artefactImageUrlFallback(unitId: number): string | null {
  const entry = findArtefactCatalog(unitId);
  if (!entry?.image) return null;
  return `/assets/origin/images/buildings/${entry.image}.png`;
}

// Заглушка-картинка артефакта (legacy `usable_artefact.gif`).
export const ARTEFACT_FALLBACK_IMAGE =
  '/assets/origin/images/buildings/usable_artefact.gif';
