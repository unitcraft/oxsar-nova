# План 82: Record holders aggregation (per-unit/research/ship topN)

**Дата**: 2026-04-28
**Статус**: Черновик. Запуск **после плана 74** (публичный запуск)
либо раньше, если кто-то жалуется на пустой Records-экран.
**Зависимости**: ✅ план 64 (override-схема), ✅ план 78 (раскладка).
Не блокирует план 72 — origin-фронт получает endpoint-skeleton в
Ф.4 Spring 3 (план 72), реальные данные подключатся когда план 82
закроется.
**Связанные документы**:
- [план 72 §Backend-расширения по требованию](72-remaster-origin-frontend-pixel-perfect.md) — записан как trade-off P72.S3.X.
- [docs/research/origin-vs-nova/origin-ui-replication.md](../research/origin-vs-nova/origin-ui-replication.md) S-031.
- [docs/simplifications.md](../simplifications.md) запись P72.S3.X.

---

## Зачем

В legacy oxsar2 был экран Records — top-1/top-N по разным критериям
(самая высокая шахта, самая большая исследовательская лаборатория,
самый большой флот определённого типа и т.д.). Это **категориальные
рекорды** на per-unit basis, не общий highscore (общий есть в nova
через `/records` который сейчас отдаёт top по очкам).

Origin-фронт (план 72 Ф.4) рендерит экран `S-031 RecordsScreen` со
вкладками по категориям. Без агрегаций все вкладки пусты. В Ф.4
Spring 3 реализован endpoint-skeleton + empty-state UI; план 82
наполняет данными.

Это применимо ко всем вселенным (nova/origin) — **общий знаменатель**
по R0-исключению (как alliance, биржа артефактов и т.д.).

---

## Что НЕ цель

- Не дублирует общий highscore (`/api/scores` или подобный, который
  уже работает по очкам и реализован).
- Не агрегирует **в реальном времени** — это сильно нагружает
  postgres. Используем materialized view + cron-refresh (15-30 мин
  периодичность).
- Не делаем UI — он уже есть (план 72 Ф.4 Spring 3, RecordsScreen).
- Не трогаем оригинальный endpoint `/api/records` если он покрывает
  общий highscore — добавляем фильтры или новый endpoint без
  пересечения.

---

## Категории рекордов (по legacy)

- **Здания**: top-1 игрок по уровню каждого здания (металл-шахта,
  кристалл-шахта, водород-шахта, солнечная станция, ангар, лаб,
  верфь, оборона-фабрика, ракетная шахта, террафабрика, ...).
- **Исследования**: top-1 по уровню каждого исследования.
- **Юниты**: top-1 по количеству каждого типа корабля (small/large/
  fighter/cruiser/battle/ds/...).
- **Оборона**: top-1 по количеству каждого типа обороны.
- **Альянсы**: top-1 альянс по числу членов / суммарным очкам / и т.д.
  (опционально — может оказаться в плане 67 расширении).

Конкретный список — извлечь из legacy `templates/records.tpl` +
`Records.class.php` при старте.

---

## Архитектура

### Materialized view

`record_holders_mv` (materialized view) — одна строка на (category,
unit_type), содержит:
- top_user_id
- top_user_name
- top_value (BIGINT — уровень здания / количество юнитов)
- universe_id (R10 — рекорды per-universe)
- updated_at

Расчёт через CTE по соответствующим таблицам:
- buildings: SELECT user_id, MAX(level) FROM building_levels GROUP
  BY user_id, unit_id; затем top per unit_id per universe.
- research: то же по research_levels.
- units/defense: SELECT FROM planet_units / planet_defense (подсчёт).

### Cron refresh

`internal/records/refresher.go` — singleton-задача, плановая через
existing scheduler (план 32):
- Раз в 15-30 мин (параметр в configs/schedule.yaml).
- `REFRESH MATERIALIZED VIEW CONCURRENTLY record_holders_mv` —
  CONCURRENTLY чтобы не блокировать чтение.
- Lock через advisory lock (план 32 паттерн).

### Endpoint

`GET /api/records?type=building&unit_id=1` — вернуть top-1 (или
top-N если limit передан) по категории + unit_id.
- Существующий `/records` handler (records/handler.go) расширить
  методом `topByCategory(ctx, category, unitID, universeID)`.
- Pretty-name резолв через i18n (имя здания, имя юнита).
- R8 Prometheus, R10 universe_id, R12 i18n.

### Frontend

UI уже есть (план 72 Ф.4 Spring 3 `RecordsScreen`). После
закрытия плана 82 endpoint начнёт возвращать данные, frontend
не меняется.

---

## Этапы

- Ф.1. Миграция `record_holders_mv` + индексы.
- Ф.2. Refresher-задача в scheduler + tests.
- Ф.3. Расширение records-handler'а методом topByCategory + tests.
- Ф.4. Property-based tests (rapid): инвариант «top_value ≥ всех
  остальных user'ов в категории».
- Ф.5. Финализация (P72.S3.X в simplifications.md → ✅ закрыт,
  шапка плана 82 ✅).

---

## Объём

~400-600 строк Go + миграция + тесты. 1-2 недели.

3-4 коммита.

---

## Триггеры запуска

- ✅ план 74 публичный запуск закрыт, ИЛИ
- 🟡 жалоба от dev/тестера «пустой Records-экран в origin-фронте» (10+ запросов).

До этого план остаётся в очереди.

---

## Связанные

- План 72 Ф.4 (Spring 3) — endpoint-skeleton для S-031.
- План 32 (multi-instance scheduler) — refresher как scheduled task.
- План 17 (gameplay improvements) — может содержать category records
  как часть geyser, проверить пересечение при старте.
