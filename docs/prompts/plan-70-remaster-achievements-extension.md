# Промпт: выполнить план 70 (achievements расширение)

**Дата создания**: 2026-04-28
**Статус плана**: ⏸ **Отложен** до пост-запуска origin (плана 74).
**План**: [docs/plans/70-remaster-achievements-extension.md](../plans/70-remaster-achievements-extension.md)
**Объём**: 1-2 нед.

⚠️ **НЕ запускать сейчас.** План 70 выведен из первой итерации
ремастера (см. roadmap-report «Часть V»). Возвращается к работе
после публичного запуска origin.

Промпт сохранён для будущего использования.

---

```
Задача: выполнить план 70 (ремастер) — расширение goal engine
game-nova под классические ачивки origin (~100 из na_achievement_datasheet).

ВАЖНОЕ:
- План реактивирован после публичного запуска origin (плана 74).
- R0: nova-ачивки uni01/uni02 НЕ пересматриваются. План касается
  ТОЛЬКО ачивок origin (через override-механизм).
- goal engine уже реализован в nova (план 17 + план 24).

ПЕРЕД НАЧАЛОМ:

1) git status --short.

2) Прочитай:
   - docs/plans/70-remaster-achievements-extension.md
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» +
     «Часть V» (отказ был временным; теперь активен)
   - docs/research/origin-vs-nova/divergence-log.md D-017
   - docs/adr/0009-currency-rebranding.md (награды в оксаритах)

ЧТО НУЖНО СДЕЛАТЬ:

1. Импорт-скрипт cmd/tools/import-origin-achievements/:
   - Читает na_achievement_datasheet из dump origin.
   - Пишет в configs/balance/origin.yaml секция achievements:
     ИЛИ отдельный файл configs/achievements/origin.yaml — решить.
   - Имена ачивок — snake_case (`first_metal_mine_lvl_5`, не
     `firstMetalMineLvl5`).

2. Расширение goal_defs (R1):
   - Условия: req_points, req_u_points (требования по очкам категорий).
   - Награды:
     · bonus_metal/silicon/hydrogen (полные слова, R1).
     · bonus_unit_<unit_name>: <count>.
     · bonus_oxsarit (только soft, ст. 1062 ГК).
     · НЕ bonus_credit, НЕ bonus_oxsar (hard нельзя как награду —
       юр-разделение, ADR-0009).

3. Frontend в origin-фронте:
   - projects/game-origin/frontend/src/features/achievements/
     с прогрессом и раскрытием полных условий (как в legacy-PHP —
     кликнул на ачивку, видишь точные требования).

ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА:

R0: nova-ачивки uni01/uni02 не пересматриваются. План — только origin.
R1: snake_case ключей, английский, полные слова.
R12: i18n — grep nova-bundle перед созданием новых ключей описаний.
   Многие ачивки могут переиспользовать существующие nova-ключи.
R15: без упрощений (golden-тесты на progression).

ВАЛЮТА (R1, ADR-0009):
- Награды только в оксаритах (bonus_oxsarit), НЕ оксары.

GIT-ИЗОЛЯЦИЯ:
- Свои пути: cmd/tools/import-origin-achievements/, configs/,
  internal/goals/ (расширение), origin-frontend, тесты,
  docs/plans/70-..., divergence-log.md (D-017).

УСПЕШНЫЙ ИСХОД:
- ~100 ачивок origin импортированы.
- goal engine расширен под legacy-условия.
- UI в origin-фронте работает.
- D-017 закрыт.

Стартуй.
```
