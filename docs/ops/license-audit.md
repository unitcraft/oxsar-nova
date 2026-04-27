# Аудит лицензий зависимостей

**Дата последнего аудита:** 2026-04-27
**Связанные документы:** [LICENSE](../../LICENSE) (PolyForm Noncommercial 1.0.0),
[../plans/40-license-audit.md](../plans/40-license-audit.md),
[.github/workflows/ci.yml](../../.github/workflows/ci.yml).

## Зачем

PolyForm Noncommercial 1.0.0 несовместим с copyleft-семействами лицензий
(GPL/AGPL/LGPL): попадание такой зависимости в граф автоматически
переводит всю производную работу под условия copyleft, что аннулирует
защитную функцию PolyForm. Аудит фиксирует список разрешённых лицензий
и подключает CI-проверку, которая блокирует появление несовместимой
зависимости в будущем.

## Whitelist (разрешённые лицензии)

Лицензии, считающиеся совместимыми с PolyForm Noncommercial 1.0.0:

- `MIT`
- `Apache-2.0`
- `BSD-2-Clause`
- `BSD-3-Clause`
- `ISC`
- `MPL-2.0`
- `Unlicense`
- `CC0-1.0`
- `Zlib`

Любая лицензия вне этого списка — рассматривается индивидуально.
По умолчанию CI отклоняет такую зависимость.

## Запрещённые лицензии

- `GPL-*` (любые версии)
- `AGPL-*`
- `LGPL-*`
- `SSPL-*`, `Commons Clause`, `BUSL-*` — отклоняются по умолчанию.
- `CDDL`, `EPL` — отклоняются по умолчанию (рассматриваются индивидуально).

## Сводка по проектам (на 2026-04-27)

### Go-стек (4 модуля)

Сканер: `go-licenses` (Google). Команда: `go-licenses check ./... --disallowed_types=forbidden,restricted`.

| Модуль                        | Внешних зависимостей | MIT | Apache-2.0 | BSD-3-Clause | BSD-2-Clause | ISC | Forbidden |
|-------------------------------|---------------------:|----:|-----------:|-------------:|-------------:|----:|----------:|
| `projects/game-nova/backend`  | 26                   | 12  | 4          | 8            | 1            | 1   | 0         |
| `projects/portal/backend`     | 25                   | 12  | 4          | 8            | 1            | 0   | 0         |
| `projects/identity/backend`   | 24                   | 11  | 4          | 8            | 1            | 0   | 0         |
| `projects/billing/backend`    | 24                   | 11  | 4          | 8            | 1            | 0   | 0         |

**Итог:** во всех 4 модулях отсутствуют GPL/AGPL/LGPL.
Распределение: ~46% MIT, ~33% BSD-3-Clause, ~16% Apache-2.0, остальное —
BSD-2-Clause и ISC. `go-licenses check` возвращает exit 0 для всех модулей.

### Frontend (2 проекта)

Сканер: `license-checker` (npm). Команда:
`npx license-checker --production --onlyAllow "<whitelist>"`.

| Проект                            | Production-зависимостей | MIT | Unlicense |
|-----------------------------------|------------------------:|----:|----------:|
| `projects/game-nova/frontend`     | 22                      | 21  | 1 (`isbot`)|
| `projects/portal/frontend`        | 0                       | 0   | 0         |

`projects/portal/frontend` на момент аудита не имеет внешних
production-зависимостей (используются только сам пакет и его собственный
код). `isbot` распространяется под Unlicense (public domain) — в whitelist.

**Итог:** во всех 2 проектах в production-бандле отсутствуют GPL/AGPL/LGPL.
Dev-зависимости (TypeScript, vitest, ESLint и т.п.) не сканируются —
они не попадают в production-артефакт.

### PHP/game-origin

Аудит PHP/Composer-зависимостей **отложен** до закрытия плана 43
(Recipe → Composer). На момент этого аудита PHP-стек содержит ~28
GPL-файлов legacy-фреймворка Recipe (Sebastian Noll, GPL-2.0+),
которые удаляются в плане 43. CI license-check на PHP не запускается
до завершения плана 43. После закрытия плана 43 — добавить отдельный
шаг в CI с `composer licenses` или эквивалентом.

## CI-проверка

Job `license-check` в [.github/workflows/ci.yml](../../.github/workflows/ci.yml)
запускается на каждый push в `main` и каждый pull request параллельно
с `lint` и `test`. Шаги:

1. `go-licenses check ./... --disallowed_types=forbidden,restricted`
   — для каждого из 4 backend-модулей.
2. `npx license-checker --production --onlyAllow "<whitelist>"`
   — для каждого из 2 frontend-проектов.

При обнаружении несовместимой лицензии job падает, PR не сливается.

## Что делать, если CI обнаружил GPL/AGPL/LGPL

1. **Идентифицировать пакет** — в логах job `license-check`
   указан конкретный пакет и его лицензия.
2. **Найти MIT/Apache-альтернативу** — проверить экосистему,
   часто популярные пакеты имеют permissive-форки.
3. **Удалить зависимость** — если она транзитивная и не нужна
   напрямую, обновить непосредственного предка до версии без
   неё.
4. **Согласовать исключение** — если альтернативы нет, обсудить
   с автором проекта в issue. Whitelist расширяется только через
   изменение этого документа и CI-конфига; не глушить через
   `--ignorePackages` без согласования.

При появлении пакета под dual-license (например, `(MIT OR Apache-2.0)`)
расширить whitelist в CI и в этом документе соответствующим
SPDX-выражением.
