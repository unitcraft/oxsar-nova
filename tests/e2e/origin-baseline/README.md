# origin-baseline — эталонные скриншоты legacy-php

Часть [плана 73](../../../docs/plans/73-remaster-screenshot-diff-ci.md):
скриншоты legacy-php (`projects/game-legacy-php/`) — «золотой» baseline,
с которым потом сравнивается новый origin-фронт
(`projects/game-nova/frontends/origin/`) через pixel-diff.

Этот пакет реализует **Ф.1 + Ф.2** плана 73:

- Ф.1 — скрипт `take-screenshots.sh` + Playwright spec.
- Ф.2 — snapshot smoke-набора (7 экранов из 22) для верификации процесса.

Ф.3 (Playwright suite на новый origin-фронт + diff через pixelmatch),
Ф.4 (CI-job), Ф.5 (регламент release notes) — отдельные сессии.

## Что снимаем

Spring 1 (план 72 Ф.2) + Spring 2 (план 72 Ф.3) — 22 экрана.
Полный список — [screens.ts](screens.ts).

| Группа | Кол-во | План |
|---|---|---|
| Spring 1 (Main, Constructions, Research, Shipyard, Galaxy, Mission, Empire) | 7 | 72 Ф.2 |
| Spring 2 ч.1 (12 экранов Alliance) | 12 | 72 Ф.3 |
| Spring 2 ч.2 (Resource, Market, Repair, Battlestats) | 4 | 72 Ф.3 |

Spring 3+4+5 (планы 72 Ф.4-Ф.6) добавляются в `screens.ts` после того,
как соответствующие экраны будут реализованы в новом origin-фронте.

## Запуск

### Полный (все 22 экрана)

```bash
SMOKE=0 bash tests/e2e/origin-baseline/take-screenshots.sh
```

### Smoke (7 экранов из Ф.2 — для smoke процесса)

```bash
bash tests/e2e/origin-baseline/take-screenshots.sh
```

Скрипт:

1. Поднимает legacy-php docker-стек
   (`cd projects/game-legacy-php/docker && docker compose up -d`).
2. Ждёт пока nginx ответит на `/dev-login.php` (max 60s).
3. Устанавливает (если нужно) `node_modules` и Chromium-браузер.
4. Запускает Playwright spec — он логинится через `/dev-login.php`
   (dev-аккаунт `test` / userid=1 — см.
   [docs/legacy/game-legacy-access.md](../../../docs/legacy/game-legacy-access.md))
   и снимает PNG в `screenshots/`.
5. По умолчанию останавливает docker-стек (`docker compose stop`).
   Чтобы оставить запущенным: `KEEP_RUNNING=1 bash take-screenshots.sh`.

## Регламент обновления эталонов

**Технические правки** (refactor, переименования, чистка кода) НЕ должны
менять baseline. Если diff (Ф.3) показывает изменение пикселей — это
бажный refactor, чинить.

**Намеренные изменения** UI legacy (новые фичи / правки стилей в
`projects/game-legacy-php/`) обновляются через:

```bash
bash tests/e2e/origin-baseline/update-baselines.sh
```

Скрипт:

1. Спрашивает confirmation (`yes/no`).
2. Перезапускает полный snapshot.
3. Подсказывает добавить описание в коммит-сообщение и release notes.

В коммит-сообщении явно указать:

- Что именно изменилось в legacy-UI.
- Почему (issue / план / решение).
- Ссылку на release notes если применимо (план 73 §«Регламент»).

CI (Ф.4 — отдельная сессия) триггерит pixel-diff только на изменения
в `projects/game-nova/frontends/origin/`. Изменения в `screenshots/`
сами по себе НЕ запускают diff — diff = «новый origin vs текущий
эталон»; обновление эталона = намеренный snapshot.

## Игнор-зоны (план 73 §«Конвенции»)

Полная маскировка (через Playwright `mask:` опцию) применяется в
Ф.3 при сравнении, не на этапе снятия. Заведомо динамичные зоны:

- Текущее время в шапке.
- Прогресс-бары countdown (стройка, исследование, флот).
- Рекламные блоки legacy (если их нельзя отключить через config —
  см. [docs/plans/73-remaster-screenshot-diff-ci.md](../../../docs/plans/73-remaster-screenshot-diff-ci.md)).

## Структура

```
tests/e2e/origin-baseline/
├── README.md                 — этот файл
├── package.json              — Playwright + TypeScript dependencies
├── tsconfig.json             — strict TS-конфиг
├── playwright.config.ts      — Chromium-only, viewport 1440×900, ru-RU
├── screens.ts                — список 22 экранов + smoke-набор
├── baseline.spec.ts          — Playwright spec (login + screenshots)
├── take-screenshots.sh       — главный скрипт (поднимает Docker, снимает)
├── update-baselines.sh       — намеренное обновление с confirmation
├── .gitignore                — исключает node_modules/, отчёты
└── screenshots/              — PNG, коммитятся в репо
    ├── s-001-main.png
    ├── s-002-research.png
    └── …
```

## Зависимости

- Docker + Docker Compose (для подъёма legacy-php).
- Node 20+ (Playwright).
- ~500 MB места на диске для Chromium-браузера Playwright.

## Не делать

- НЕ редактировать legacy-PHP код (`projects/game-legacy-php/`) ради
  baseline'ов — снимаем что есть.
- НЕ снимать baseline для экранов которых ещё нет в новом origin-фронте.
- НЕ коммитить `node_modules/` или `playwright-report/`.
- НЕ удалять PNG без описания «почему» в коммит-сообщении.

## Связанные документы

- [docs/plans/73-remaster-screenshot-diff-ci.md](../../../docs/plans/73-remaster-screenshot-diff-ci.md)
  — основной план.
- [docs/research/origin-vs-nova/origin-ui-replication.md](../../../docs/research/origin-vs-nova/origin-ui-replication.md)
  — полный список экранов S-001..S-055.
- [docs/legacy/game-legacy-access.md](../../../docs/legacy/game-legacy-access.md)
  — как поднять legacy-php, dev-логин.
- [projects/game-nova/frontends/nova/e2e/](../../../projects/game-nova/frontends/nova/e2e/)
  — существующая Playwright-инфра nova-фронта (паттерны).
