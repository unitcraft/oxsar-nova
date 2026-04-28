/**
 * Список экранов origin-фронта для baseline-скриншотов (план 73 Ф.1+Ф.2).
 *
 * Покрытие — Spring 1 (план 72 Ф.2) + Spring 2 (план 72 Ф.3): 22 экрана.
 * Spring 3-5 (планы 72 Ф.4+Ф.5+Ф.6) добавляются отдельной сессией после
 * того, как соответствующие экраны будут реализованы в новом origin-фронте.
 *
 * URL — только формат `?go=Page&action=...`. PATH_INFO в legacy-php
 * не парсится (memory `reference_game_origin_routing.md`).
 *
 * `id` — стабильный код экрана из docs/research/origin-vs-nova/origin-ui-replication.md.
 * `name` — kebab-case, используется в имени PNG.
 * `path` — относительный URL от LEGACY_URL.
 */

export interface Screen {
  readonly id: string;
  readonly name: string;
  readonly path: string;
  /** Описание для дев-логов / отчётов (русский, как в origin-ui-replication.md). */
  readonly description: string;
}

export const SCREENS: ReadonlyArray<Screen> = [
  // Spring 1 — план 72 Ф.2 (коммит 47d1f0ef65)
  { id: 'S-001', name: 'main', path: '/game.php?go=Main', description: 'Главный экран' },
  { id: 'S-003', name: 'constructions', path: '/game.php?go=Constructions', description: 'Здания' },
  { id: 'S-002', name: 'research', path: '/game.php?go=Research', description: 'Исследования' },
  { id: 'S-004', name: 'shipyard', path: '/game.php?go=Shipyard', description: 'Верфь' },
  { id: 'S-005', name: 'galaxy', path: '/game.php?go=Galaxy&galaxy=1&system=1', description: 'Галактика' },
  { id: 'S-006', name: 'mission', path: '/game.php?go=Mission', description: 'Миссии флота' },
  { id: 'S-042', name: 'empire', path: '/game.php?go=Empire', description: 'Империя' },

  // Spring 2 ч.1 — план 72 Ф.3 (коммит 48ef07cf19), 12 alliance-экранов.
  // Альянсные actions: используем основные представления, на которых рендерится разный layout.
  // Игрок test (userid=1) должен быть в альянсе для большинства этих экранов;
  // если не в альянсе — будут показаны фронтальные «без альянса» состояния.
  { id: 'S-012-overview', name: 'alliance-overview', path: '/game.php?go=Alliance', description: 'Альянс — обзор' },
  { id: 'S-012-members', name: 'alliance-members', path: '/game.php?go=Alliance&action=memberlist', description: 'Альянс — члены' },
  { id: 'S-012-diplomacy', name: 'alliance-diplomacy', path: '/game.php?go=Alliance&action=diplomacy', description: 'Альянс — дипломатия' },
  { id: 'S-012-ranks', name: 'alliance-ranks', path: '/game.php?go=Alliance&action=manageRanks', description: 'Альянс — ранги' },
  { id: 'S-012-globalmail', name: 'alliance-globalmail', path: '/game.php?go=Alliance&action=globalMail', description: 'Альянс — global mail' },
  { id: 'S-012-search', name: 'alliance-search', path: '/game.php?go=Alliance&action=allySearch', description: 'Альянс — поиск' },
  { id: 'S-012-applications', name: 'alliance-applications', path: '/game.php?go=Alliance&action=applications', description: 'Альянс — заявки' },
  { id: 'S-012-relations', name: 'alliance-relations', path: '/game.php?go=Alliance&action=acceptRelation', description: 'Альянс — отношения' },
  { id: 'S-012-found', name: 'alliance-found', path: '/game.php?go=Alliance&action=foundAlliance', description: 'Альянс — основать' },
  { id: 'S-012-manage', name: 'alliance-manage', path: '/game.php?go=Alliance&action=manageAlly', description: 'Альянс — управление' },
  { id: 'S-012-apply', name: 'alliance-apply', path: '/game.php?go=Alliance&action=apply', description: 'Альянс — подать заявку' },
  { id: 'S-012-candidates', name: 'alliance-candidates', path: '/game.php?go=Alliance&action=candidates', description: 'Альянс — кандидаты' },

  // Spring 2 ч.2 — план 72 Ф.3 (коммит 590a68b428), 5 экранов
  { id: 'S-025', name: 'resource', path: '/game.php?go=Resource', description: 'Ресурсы' },
  { id: 'S-026', name: 'market', path: '/game.php?go=Market', description: 'Рынок ресурсов' },
  { id: 'S-048', name: 'repair', path: '/game.php?go=Repair', description: 'Ремонт' },
  { id: 'S-017', name: 'battlestats', path: '/game.php?go=Battlestats', description: 'Боевая статистика' },
  // FleetOperations — отдельного контроллера в legacy нет, это nova-агрегатор
  // (объединяет Mission history + холдинг + ACS). В origin-фронте — собственный
  // экран; в legacy ближайший аналог = ?go=Mission, который уже снят выше под
  // S-006. Не включаем повторно.
];

/** Smoke-набор для Ф.2 — 7 экранов покрывают основные layout-паттерны. */
export const SMOKE_SCREEN_IDS: ReadonlyArray<string> = [
  'S-001',
  'S-003',
  'S-002',
  'S-012-overview',
  'S-025',
  'S-026',
  'S-048',
];
