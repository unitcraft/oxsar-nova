# План 70: Achievements расширение (отложен после старта origin)

**Дата**: 2026-04-28
**Статус**: ⏸ **Отложен** (выведен из первой итерации ремастера —
см. roadmap-report.md «Часть V. Что НЕ делать»). Возвращается к
работе после публичного запуска origin (план 74).
**Зависимости**: goal engine уже реализован в nova (план 17 / план 24).
**Связанные документы**:
- [62-origin-on-nova-feasibility.md](62-origin-on-nova-feasibility.md)
- [docs/research/origin-vs-nova/divergence-log.md](../research/origin-vs-nova/divergence-log.md) D-017
- [docs/research/origin-vs-nova/roadmap-report.md](../research/origin-vs-nova/roadmap-report.md) —
  R0-R7 + раздел плана 70

---

## Почему отложен

Ачивки **в nova уже реализованы** (план 17 + goal engine). Чтобы
сократить scope ремастера и быстрее довести pixel-perfect-клон до
старта, классические ачивки origin **выводятся из первой итерации**:

- В origin-фронте (план 72) экран Achievements **не реализуется**
  на старте. Пункт меню скрыт ИЛИ ведёт на заглушку «Скоро».
- ~100 ачивок из `na_achievement_datasheet` **не импортируются**
  в `configs/goals.yml` сейчас.
- Расширение `goal_defs` под legacy-условия (`req_u_points`,
  `bonus_*_unit`) — откладывается.

**Когда возвращаемся:** после публичного запуска origin (план 74),
когда нужно будет вернуть классическую игроковую прогрессию
oxsar2 в виде ачивок origin.

**По R0:** nova-ачивки (uni01/uni02) **не трогаем** в любом
случае — план 70 касается только импорта классики oxsar2 для
вселенной origin.

---

## Цель (когда план будет реактивирован)

Расширить goal engine game-nova под классические ачивки origin
(~100 из `na_achievement_datasheet`). Применимо **только к
вселенной origin** через override-механизм; nova-ачивки для
modern-вселенных не пересматриваются (R0).

---

## Что делаем (когда план будет реактивирован)

- Импорт ~100 ачивок из `projects/game-origin-php/migrations/002_data.sql`
  таблицы `na_achievement_datasheet` в
  `configs/balance/origin.yaml` (секция `achievements:`) или
  отдельный override `configs/achievements/origin.yaml` — решение
  при реактивации.
- Расширение `goal_defs` под условия:
  - `req_points`, `req_u_points` (требования по очкам категорий)
  - `bonus_metal`, `bonus_silicon`, `bonus_hydrogen` (награды
    ресурсами)
  - `bonus_*_unit` (награда юнитами)
  - `bonus_oxsarit` (награда оксаритами — по R1 «Особый случай:
    валюта»; НЕ `bonus_credit`)
- UI в origin-фронте: `projects/game-origin/frontend/src/features/achievements/`
  с прогрессом и раскрытием полных условий (как в legacy-PHP —
  кликнул на ачивку, видишь точные требования).

---

## Что НЕ делаем

- Не трогаем nova-ачивки (план 17) — они работают для modern-вселенных
  как есть (R0).
- Не оставляем legacy-имя `na_achievement_datasheet` — в nova
  будет конфиг (по R1).
- Не дублируем механику в game-nova-backend — расширяем
  существующий goal engine, не пишем новый.

## Этапы (детали — при реактивации)

- Ф.1. Импорт-скрипт `cmd/tools/import-origin-achievements/`
  (читает из legacy-PHP-дампа → пишет конфиг origin).
- Ф.2. Расширение `goal_defs` в Go под классические условия.
- Ф.3. UI-фронт в origin-фронте.
- Ф.4. Тесты + golden на progression.
- Ф.5. Финализация.

## Конвенции (R0-R7)

- R0: nova-ачивки uni01/uni02 не пересматриваются. План касается
  **только origin**.
- R1: имена ачивок в YAML — snake_case
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
- roadmap-report.md «Часть V» — обоснование отложения.
