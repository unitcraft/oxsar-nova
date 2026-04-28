# План 70: Achievements расширение (legacy + общий движок)

**Дата**: 2026-04-28
**Статус**: Скелет (детали допишет агент-реализатор при старте)
**Зависимости**: goal engine уже реализован в nova (план 17 / план 24).
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md)
- [docs/research/origin-vs-nova/divergence-log.md](../research/origin-vs-nova/divergence-log.md) D-017
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R1-R5 + раздел плана 70

---

## Цель

Расширить goal engine game-nova под legacy-ачивки origin
(~100 ачивок из `na_achievement_datasheet`). Применимо для всех
вселенных — это **общий знаменатель** (R1).

---

## Что делаем

- Импорт ~100 ачивок из `projects/game-origin-php/migrations/002_data.sql`
  таблицы `na_achievement_datasheet` в `configs/goals.yml`.
- Расширение `goal_defs` под условия:
  - `req_points`, `req_u_points` (требования по очкам категорий)
  - `bonus_metal`, `bonus_silicon`, `bonus_hydrogen` (награды
    ресурсами)
  - `bonus_*_unit` (награда юнитами)
  - `bonus_oxsarit` (награда оксаритами — по R1 «Особый случай:
    валюта»; НЕ `bonus_credit`)
- UI: `frontend/src/features/achievements/` с прогрессом и
  раскрытием полных условий (как в origin — кликнул на ачивку,
  видишь точные требования).

---

## Что НЕ делаем

- Не оставляем legacy-имя `na_achievement_datasheet` — в nova
  будет таблица `achievements` или конфиг `configs/goals.yml`
  (по R1).
- Не дублируем механику в game-nova-backend — расширяем
  существующий goal engine, не пишем новый.

## Этапы (детали — при старте)

- Ф.1. Импорт-скрипт `cmd/tools/import-legacy-achievements/`
  (читает из dump origin → пишет в configs/goals.yml).
- Ф.2. Расширение `goal_defs` в Go.
- Ф.3. UI-frontend компонент.
- Ф.4. Тесты + golden на progression.
- Ф.5. Финализация.

## Конвенции (R1-R5)

- Имена ачивок в YAML — snake_case
  (`first_metal_mine_lvl_5`, не `firstMetalMineLvl5`).
- Награды:
  - Ресурсы — `bonus_metal/silicon/hydrogen` (полные слова, R1).
  - Валюта — **только оксариты** (`bonus_oxsarit`), НЕ `bonus_credit`,
    НЕ `bonus_oxsar` (hard-валюту нельзя выдавать как награду —
    юр-разделение, ADR-0009).
- Юниты — `bonus_unit_<unit_name>: <count>`.

## Объём

1-2 недели. Импорт + расширение goal engine + UI.

## References

- D-017 в divergence-log.md.
- План 17 (gameplay-improvements) — daily quests / achievements в nova.
- План 24 (ai-players) — goal engine.
- ADR-0009 — почему награда в оксаритах, а не «кредитах»/оксарах.
