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
  },
  {
    id: 3002,
    i18nName: 'info.atomicDensifier',
    i18nDesc: 'info.atomicDensifierDesc',
    i18nFullDesc: 'info.atomicDensifierFullDesc',
    activatable: true,
  },
  {
    id: 3110,
    i18nName: 'info.assemblyModule3110',
    i18nDesc: 'info.assemblyModule3110Desc',
    i18nFullDesc: 'info.assemblyModule3110FullDesc',
  },
  {
    id: 421,
    i18nName: 'info.assemblyModule421',
    i18nDesc: 'info.assemblyModule421Desc',
    i18nFullDesc: 'info.assemblyModule421FullDesc',
  },
];

export function findArtefactCatalog(
  unitId: number,
): ArtefactCatalogEntry | undefined {
  return ARTEFACT_CATALOG.find((e) => e.id === unitId);
}
