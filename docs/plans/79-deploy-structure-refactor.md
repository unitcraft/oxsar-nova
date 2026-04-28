# План 79: Рефакторинг раскладки deploy/

**Дата**: 2026-04-28
**Статус**: Открыт.
**Зависимости**: должен идти **после** плана 78 (frontends layout +
legacy-PHP rename), чтобы не пересекаться по `Makefile`/CI/compose.
Не блокирует план 72, 68, 76 — но компликается с любой
параллельной сессией, потому что меняет общие пути.
**Связанные документы**:
- [docs/ops/release-process.md](../ops/release-process.md) — потребитель compose-файлов в prod.
- [docs/ops/runbooks/docker-stacks.md](../ops/runbooks/docker-stacks.md) —
  9 ссылок, важные runbook-команды.
- [docs/plans/75-rename-game-origin-to-php.md](75-rename-game-origin-to-php.md),
  [docs/plans/78-frontends-layout-refactor.md](78-frontends-layout-refactor.md) —
  образцы safe-rename серий.

---

## Цель

Привести `deploy/` к современным конвенциям монорепо (Google Cloud
Build / AWS / Uber/Airbnb практики):

1. **Разделить compose-файлы** на `base.yml` + override'ы вместо 9
   плоских файлов с дублированием postgres/redis.
2. **Сервисные Dockerfile'ы** перенести из `deploy/Dockerfile.*` в
   `projects/<service>/Dockerfile` (рядом с кодом, который собирают).
3. **Инфра-Dockerfile'ы** оставить в `deploy/docker/` (сгруппировать).
4. **Конфиги** (nginx, prometheus, grafana) собрать в `deploy/configs/`.
5. **Скрипты** (`backup.sh`) переехать в `scripts/`.

---

## Текущая раскладка (плоская, ~25 файлов)

```
deploy/
  docker-compose.yml                  ← основной dev-стек (16KB)
  docker-compose.multiverse.yml       ← multi-instance dev (11KB)
  docker-compose.scaling.yml          ← N=2 smoke
  docker-compose.admin.yml            ← admin-bff + admin-frontend
  docker-compose.monitoring.yml       ← prometheus + grafana
  docker-compose.e2e.yml              ← Playwright stack
  docker-compose.e2e.ports.yml        ← override для портов
  docker-compose.prod.yml             ← prod
  docker-compose.yookassa-mock.yml    ← платёжный мок

  Dockerfile.admin-bff                ← сервисный, должен быть рядом с кодом
  Dockerfile.admin-frontend           ← сервисный
  Dockerfile.frontend-prod            ← prod-сборка nova-фронта (сервисный)
  Dockerfile.migrate                  ← инфра (запуск goose)
  Dockerfile.playwright               ← инфра (e2e runner)
  Dockerfile.prometheus               ← инфра
  Dockerfile.grafana                  ← инфра
  Dockerfile.testseed                 ← инфра (seed данных)

  nginx.admin.conf                    ← конфиг
  nginx.frontend.conf                 ← конфиг
  prometheus.yml                      ← конфиг
  grafana/provisioning/               ← конфиг

  backup.sh                           ← скрипт, должен быть в scripts/
  .env.multiverse.example             ← пример
  admin-ips.conf.example              ← пример
```

**Проблемы:**
- `postgres:16-alpine` определён в 4 compose'ах (yml/e2e/monitoring/
  multiverse/yookassa-mock).
- `redis:7-alpine` — в 2.
- Сервисные Dockerfile'ы оторваны от своего кода.
- Mix Dockerfile/compose/конфигов/скриптов на одном уровне.
- Нет ясного правила «где должен лежать новый Dockerfile».

---

## Целевая раскладка

```
deploy/
  compose/
    base.yml                          ← postgres + redis + общие сети + volumes
    dev.yml                           ← override: backend + workers + frontend для local
    prod.yml                          ← override: prod-конфиг
    e2e.yml                           ← override: playwright + testseed
    e2e.ports.yml                     ← как сейчас
    multiverse.yml                    ← override: multi-instance
    scaling.yml                       ← override: N=2 smoke
    admin.yml                         ← override: admin-bff + admin-frontend
    monitoring.yml                    ← override: prometheus + grafana
    yookassa-mock.yml                 ← override: платёжный мок

  docker/
    migrate.Dockerfile                ← инфра-only (тулзы, не сервисный код)
    playwright.Dockerfile
    testseed.Dockerfile
    prometheus.Dockerfile
    grafana.Dockerfile

  configs/
    nginx.admin.conf
    nginx.frontend.conf
    prometheus.yml
    grafana/provisioning/

  examples/
    .env.multiverse.example
    admin-ips.conf.example

scripts/
  git-hooks/                          ← как сейчас
  backup.sh                           ← переехал из deploy/

projects/
  game-nova/backend/Dockerfile        ← как сейчас
  game-nova/frontends/nova/Dockerfile ← как сейчас (после плана 78)
  admin-bff/Dockerfile                ← переехал из deploy/Dockerfile.admin-bff
  admin-frontend/Dockerfile           ← переехал
  game-nova/frontends/nova/Dockerfile.prod  ← бывший deploy/Dockerfile.frontend-prod
  ...
```

**Запуск compose:**
- Local dev: `docker compose -f deploy/compose/base.yml -f deploy/compose/dev.yml up`
- Local + monitoring: `... -f base.yml -f dev.yml -f monitoring.yml up`
- E2E: `... -f base.yml -f e2e.yml -f e2e.ports.yml up`
- Prod: `... -f base.yml -f prod.yml up`

Через Makefile-таргеты, чтобы не запоминать комбинации.

---

## Что НЕ делаем

- **Не переписываем** содержимое compose-сервисов — только перетасовка
  и вынос общих определений в `base.yml`.
- **Не трогаем** k8s/helm — у нас этого нет, не вводим.
- **Не вводим** terraform/ansible/pulumi — отложено до пост-запуска.
- **Не переименовываем** существующие сервисы внутри compose — только
  раскладка файлов.
- **Не делаем** этот план параллельно с другими сессиями — слишком много
  общих путей. Запускать в одиночку, после плана 78.

---

## Этапы

### Ф.1. Сборка `base.yml`

- Создать `deploy/compose/base.yml` с общими определениями:
  - postgres (extracting from `docker-compose.yml`)
  - redis
  - networks (default, internal-если-есть)
  - volumes (postgres-data, redis-data, etc.)
- Все остальные compose-файлы будут override'ить или extend'ить
  через `extends:` (Compose v2 spec).

### Ф.2. Перенос compose-файлов

- `git mv deploy/docker-compose.yml deploy/compose/dev.yml`
- `git mv deploy/docker-compose.prod.yml deploy/compose/prod.yml`
- `git mv deploy/docker-compose.e2e.yml deploy/compose/e2e.yml`
- `git mv deploy/docker-compose.e2e.ports.yml deploy/compose/e2e.ports.yml`
- `git mv deploy/docker-compose.multiverse.yml deploy/compose/multiverse.yml`
- `git mv deploy/docker-compose.scaling.yml deploy/compose/scaling.yml`
- `git mv deploy/docker-compose.admin.yml deploy/compose/admin.yml`
- `git mv deploy/docker-compose.monitoring.yml deploy/compose/monitoring.yml`
- `git mv deploy/docker-compose.yookassa-mock.yml deploy/compose/yookassa-mock.yml`

Внутри каждого override'а:
- Убрать определения postgres/redis (они в `base.yml`).
- Заменить relative-пути в `build:` / `volumes:` под новую глубину
  (поднялись на один уровень).

### Ф.3. Перенос Dockerfile'ов

**Сервисные → рядом с кодом:**
- `deploy/Dockerfile.admin-bff` → `projects/admin-bff/Dockerfile`
- `deploy/Dockerfile.admin-frontend` → `projects/admin-frontend/Dockerfile`
- `deploy/Dockerfile.frontend-prod` →
  `projects/game-nova/frontends/nova/Dockerfile.prod`
  + **починить сломанные пути внутри**: сейчас он содержит
  `COPY frontend/package.json ...` и `COPY frontend ./`, но папки
  `frontend/` в корне репо **не существует** (никогда не существовала
  после расщепления projects/). Заменить на:
  ```
  COPY projects/game-nova/frontends/nova/package.json projects/game-nova/frontends/nova/package-lock.json* ./
  COPY projects/game-nova/frontends/nova ./
  ```
  И обновить ссылку на nginx.frontend.conf под новый путь после Ф.4
  (`deploy/configs/nginx.frontend.conf`).

**Инфра → в `deploy/docker/`** (с переименованием в стиле
`<name>.Dockerfile`):
- `deploy/Dockerfile.migrate` → `deploy/docker/migrate.Dockerfile`
- `deploy/Dockerfile.playwright` → `deploy/docker/playwright.Dockerfile`
- `deploy/Dockerfile.testseed` → `deploy/docker/testseed.Dockerfile`
- `deploy/Dockerfile.prometheus` → `deploy/docker/prometheus.Dockerfile`
- `deploy/Dockerfile.grafana` → `deploy/docker/grafana.Dockerfile`

Внутри Dockerfile'ов — обновить относительные пути `COPY ...` под
новый build-context.

**Build-context для prometheus и grafana — особый случай:**

Сейчас `Dockerfile.prometheus` имеет `COPY prometheus.yml ...` и
`Dockerfile.grafana` имеет `COPY grafana/provisioning ...` —
это работает только если build-context = `deploy/`. После Ф.3+Ф.4:
- prometheus.yml переехал в `deploy/configs/prometheus.yml`,
- grafana/ переехал в `deploy/configs/grafana/`.

Унифицируем: build-context = **корень репо** (как у всех остальных
Dockerfile'ов). В compose:
```yaml
prometheus:
  build:
    context: ..                                # корень репо (от deploy/compose/)
    dockerfile: deploy/docker/prometheus.Dockerfile
grafana:
  build:
    context: ..
    dockerfile: deploy/docker/grafana.Dockerfile
```

Внутри Dockerfile'ов:
```dockerfile
# prometheus.Dockerfile
COPY deploy/configs/prometheus.yml /etc/prometheus/prometheus.yml

# grafana.Dockerfile
COPY deploy/configs/grafana/provisioning /etc/grafana/provisioning
```

Это **современнее** — единый context для всего стека, легче
ориентироваться, не нужны `..`-ссылки внутри Dockerfile'ов.

### Ф.4. Перенос конфигов и примеров

- `git mv deploy/nginx.admin.conf deploy/configs/nginx.admin.conf`
- `git mv deploy/nginx.frontend.conf deploy/configs/nginx.frontend.conf`
- `git mv deploy/prometheus.yml deploy/configs/prometheus.yml`
- `git mv deploy/grafana deploy/configs/grafana`
- `git mv deploy/.env.multiverse.example deploy/examples/.env.multiverse.example`
- `git mv deploy/admin-ips.conf.example deploy/examples/admin-ips.conf.example`

### Ф.5. Перенос скриптов

- `git mv deploy/backup.sh scripts/backup.sh`

### Ф.6. Обновление потребителей

**Makefile (10 ссылок):**
- Все `docker compose -f deploy/docker-compose.X.yml` → новые пути.
- Завести таргеты-обёртки чтобы инкапсулировать комбинации:
  - `make compose-up` = base + dev
  - `make compose-up-monitoring` = base + dev + monitoring
  - `make compose-up-e2e` = base + e2e + e2e.ports
  - `make compose-up-prod` = base + prod
  - `make compose-up-multiverse` = base + multiverse
  - `make compose-up-admin` = base + admin

**CI (`.github/workflows/ci.yml`, `admin-console.yml` — 14 ссылок):**
- Обновить `working-directory` / `-f deploy/...` пути.

**Документация (~140 ссылок):**
- `docs/ops/runbooks/docker-stacks.md` (9 ссылок) — критично, это
  оперативный runbook.
- `docs/ops/runbooks/backup-and-monitoring.md` (8 ссылок).
- `docs/ops/release-process.md` (6 ссылок).
- `docs/ops/admin-access.md` (7 ссылок).
- `docs/plans/*.md` (~30 файлов, ~80 ссылок) — заменить find-replace.
- `docs/prompts/*.md` (~5 файлов).

**Не трогаем** исторические записи в `docs/project-creation.txt`,
`docs/simplifications.md` (правило плана 55).

### Ф.7. Smoke

- `make compose-up` — стартует postgres+redis+backend+frontend.
- `make compose-up-monitoring` — добавляется prometheus+grafana,
  все targets живые.
- `make compose-up-e2e` — Playwright прогоняет smoke-тест.
- `make compose-up-multiverse` — uni01+uni02 встают.
- `make compose-up-admin` — admin-bff отвечает.
- `docker compose ... build` — все сервисные образы собираются с
  новых путей Dockerfile'а.
- Backup-скрипт: `bash scripts/backup.sh --dry-run` (если поддерживает)
  или ручной smoke шага 1.

### Ф.8. Финализация

- Шапка плана 79 ✅.
- Запись итерации в `docs/project-creation.txt`.
- Обновление CLAUDE.md секции «Запуск» под новые таргеты.
- Обновление [docs/ops/runbooks/docker-stacks.md](../ops/runbooks/docker-stacks.md)
  под новую раскладку (это **главный потребитель**).

---

## Конвенции (R1)

- **Compose**: `<профиль>.yml`, без префикса `docker-compose-`.
  Запуск всегда через комбинацию `-f base.yml -f <profile>.yml`.
- **Dockerfile (инфра)**: `<name>.Dockerfile` (Docker BuildKit-friendly).
- **Dockerfile (сервис)**: `Dockerfile` рядом с `package.json`/`go.mod`/
  `composer.json`. Если несколько вариантов — `Dockerfile.prod`,
  `Dockerfile.dev`.
- **Конфиги**: `deploy/configs/<name>` — конфиги ИНФРА-сервисов
  (nginx/prometheus/grafana). Конфиги нашего кода живут в
  `projects/<service>/configs/`.
- **Скрипты**: `scripts/<name>.sh` для оперативных скриптов
  (backup, deploy, smoke). `scripts/git-hooks/` для git-hooks.

---

## Объём

~3-4 часа агента. ~180 правок путей (на основании grep `deploy/...` =
180 occurrences в 49 файлах) + создание `base.yml` (~80 строк) +
обновление Makefile (~10 таргетов).

**3-4 коммита** (изоляция blame):
1. `refactor(deploy): extract base.yml + перенос compose в compose/ (план 79 Ф.1+Ф.2)`
2. `refactor(deploy): сервисные Dockerfile'ы → projects/<svc>/, инфра → deploy/docker/ (Ф.3)`
3. `refactor(deploy): configs/, examples/, scripts/backup.sh (Ф.4+Ф.5)`
4. `refactor(deploy): обновить потребителей (Makefile, CI, runbooks) (Ф.6+Ф.8)`

---

## Что разблокирует

- **Будущие планы** добавления нового compose-профиля (e.g. staging,
  preview-environments) — кладут override рядом с остальными в
  `compose/`.
- **Будущие сервисы** — Dockerfile рядом с кодом по правилу.
- **Чище CI-конфиги** — меньше дублирования в matrix-сборках.
- **Меньше ошибок** при правках postgres/redis версий — в одном месте.
