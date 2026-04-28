# План 78: Раскладка фронтов и переименование legacy-PHP

**Дата**: 2026-04-28
**Статус**: ✅ Завершён 2026-04-28 (Ф.1-Ф.6).
**Зависимости**: блокирует план 72 (новый origin-фронт).
**Связанные документы**:
- [adr/0010-universe-naming.md](../adr/0010-universe-naming.md) — обоснование именования.
- [плана 75](75-rename-game-origin-to-php.md) — образец safe-rename.

---

## Цель

Две связанные задачи раскладки:

1. **Подготовить раскладку фронтов под архитектуру «один backend
   (game-nova) обслуживает несколько фронтов»** — nova-стиль для
   uni01/uni02 и pixel-perfect клон визуала (origin) для вселенной
   origin.

2. **Переименовать `game-origin-php/` → `game-legacy-php/`.**
   Текущее имя путает: `game-origin-php` звучит как «origin-фронт на
   PHP», хотя это clean-room rewrite legacy oxsar2 для исторической
   совместимости и сравнения. После создания нового origin-фронта
   (план 72) путаница только усилится. Терминологически легаси
   ≠ origin (см. memory feedback_legacy_origin_terminology):
   legacy = `d:\Sources\oxsar2` + clean-room rewrite в репо;
   origin = новая вселенная на game-nova-backend + новый origin-фронт.

Сейчас:
```
projects/game-nova/
  api/, backend/, configs/, migrations/
  frontend/             ← единственный фронт, nova-стиль
projects/game-origin-php/  ← legacy PHP (на удаление после 72/74)
```

Зарезервировано (но не создано):
```
projects/game-origin/frontend/   ← должен был жить новый origin-фронт
```

После плана 78:
```
projects/game-nova/
  api/, backend/, configs/, migrations/
  frontends/
    nova/               ← бывший projects/game-nova/frontend/
    origin/             ← создаётся планом 72 после этой раскладки
projects/game-legacy-php/   ← бывший projects/game-origin-php/
```

Имя `game-origin/` полностью свободно — пригодится позже либо как
пустое (никогда не использовать, чтобы не путать), либо если в
архитектуре появится отдельный origin-сервис (маловероятно).

---

## Что НЕ делаем

- **Не создаём** `frontends/origin/` — это работа плана 72.
- **Не трогаем код** внутри `game-origin-php/`/`game-legacy-php/` —
  только переименование папки и ссылок на путь.
- **Не переименовываем** `nova` в коде/типах/i18n — это только раскладка папок.
- **Не реорганизуем** `portal/`, `admin-frontend/`, `identity/`, `billing/` —
  они остаются как есть. Только game-nova-фронт + legacy-PHP затрагиваются.
- **Не вычищаем** прилагательное `modern` из текстов/комментариев —
  это эпоха, не имя. Имя строго `nova`.

---

## Терминология

- **`nova`** — имя проекта/вселенной/фронта в путях, кодовых
  идентификаторах, конфигах, commit-сообщениях.
- **`modern`** — допустимо как прилагательное-эпоха в текстах
  («modern-числа», «modern-эпоха») для противопоставления legacy.
  Не использовать как имя файла/папки/идентификатора.
- **`origin`** — имя вселенной + имя origin-фронта.

---

## Этапы

### Ф.1. Переносы папок

- `git mv projects/game-nova/frontend projects/game-nova/frontends/nova`
- `git mv projects/game-origin-php projects/game-legacy-php`
- Проверить что git распознал оба mv как rename, не как delete+add
  (важно для blame-истории).

### Ф.2. Сборка / деплой / код

Замена 1: `projects/game-nova/frontend` → `projects/game-nova/frontends/nova`:
- `Makefile` — таргеты `frontend-run`, `frontend-build`, `frontend-test`.
- `.github/workflows/ci.yml` — paths/working-directory для frontend-job'ов.
- `deploy/docker-compose.yml` — context для frontend-сервиса.
- `deploy/docker-compose.e2e.yml` — context + volumes.
- `deploy/Dockerfile.playwright` — COPY/WORKDIR.
- `Dockerfile` внутри (переедет вместе с папкой, но проверить
  относительные пути).
- `.gitignore` — записи про `node_modules` и `dist`.
- `projects/game-nova/configs/i18n/` — путь не меняется, но если
  где-то в frontend есть импорт через относительный `../../configs/`,
  нужна правка глубины.
- `projects/game-nova/frontend/src/components/feedback/feedback.ts`
  (упоминание пути в комментарии).

Замена 2: `projects/game-origin-php` → `projects/game-legacy-php`
(355 упоминаний в 67 файлах, из них критичны для сборки/CI):
- `Makefile` — таргеты для legacy-PHP сборки/докера, если есть.
- `deploy/docker-compose*.yml` — context legacy-PHP сервиса.
- `.github/workflows/ci.yml` — paths/working-directory legacy-job'ов.
- `.gitignore` — записи про legacy-PHP node_modules/vendor.
- **Go-код в game-nova-backend** (упоминают путь в комментариях/
  тестовых данных):
  - `projects/game-nova/backend/internal/event/handlers.go`
  - `projects/game-nova/backend/internal/origin/alien/doc.go`
  - `projects/game-nova/backend/internal/origin/alien/golden_test.go`
  - `projects/game-nova/backend/internal/origin/economy/golden_test.go`
  - `projects/game-nova/backend/cmd/tools/import-legacy-balance/main.go`
- **Скрипты внутри переехавшей папки**:
  - `projects/game-legacy-php/tools/*.sh` (пути в bash-скриптах
    относительные — проверить нет ли абсолютных или GIT_ROOT-
    зависимых).
  - `projects/game-legacy-php/tools/dump-alien-ai.php`,
    `dump-balance-formulas.php` — проверить include-пути.
  - `projects/game-legacy-php/migrations/fixtures/README.md`.
  - `projects/game-legacy-php/src/core/util/Moderation.util.class.php`.

### Ф.3. Документация (поиск-замена)

Две глобальные замены:

**Замена 1**: `projects/game-nova/frontend` → `projects/game-nova/frontends/nova`
по:
- `CLAUDE.md`
- `docs/plans/*.md` (~12 файлов)
- `docs/prompts/**/*.md` (~10 файлов)
- `docs/research/origin-vs-nova/*.md` (~2 файла)
- `docs/adr/*.md` (1 файл)
- `docs/ops/*.md` (1 файл)

**Замена 2**: `game-origin-php` → `game-legacy-php`
по:
- `CLAUDE.md`
- `docs/plans/*.md` (~30 файлов)
- `docs/prompts/**/*.md` (~15 файлов)
- `docs/research/origin-vs-nova/*.md` (~6 файлов)
- `docs/adr/0010-universe-domain-naming.md`
- `docs/legacy/game-origin-access.md` — переименовать файл в
  `game-legacy-access.md`, обновить ссылки.
- `docs/ops/legal-compliance-audit.md`
- `docs/simplifications.md`
- `docs/ai-debug-examples/*.md` (2 файла)
- `docs/ui/dev-log.md`

**Не трогаем** исторические записи в `docs/project-creation.txt`,
которые описывают прошлые итерации — там путь зафиксирован как
факт момента написания (зеркало правила плана 55).

**Memory** (`C:\Users\Евгений\.claude\projects\d--Sources-oxsar-nova\memory\`)
— пересмотреть 3 записи: `feedback_legacy_origin_terminology.md`,
`reference_legacy_docker.md`, `project_origin_vs_nova.md` — заменить
`game-origin-php` на `game-legacy-php`. Эта правка вне репо.

### Ф.3.5. Зомби-Dockerfile'ы и сломанные пути после переименований

Найдено при ревью (касается путей фронта/legacy, поэтому здесь):

**Мёртвые Dockerfile'ы (удалить):**
- `projects/game-nova/backend/Dockerfile.auth` — ссылается на
  `./cmd/auth-service`, которой нет с плана 51 (auth → identity).
  Никем не используется. `git rm`.
- `projects/game-nova/backend/Dockerfile.portal` — ссылается на
  `cmd/portal`, которой нет с плана 36 (portal вынесен в
  `projects/portal/`). `git rm`.
  Перед удалением — `grep -rn "Dockerfile\.auth\|Dockerfile\.portal"`
  по active-коду (deploy/, Makefile, .github/) — должно быть 0
  ссылок.

**Сломанные пути в живом Dockerfile (починить):**
- `deploy/Dockerfile.frontend-prod` (используется в prod-сборке):
  ```
  COPY frontend/package.json frontend/package-lock.json* ./
  COPY frontend ./
  ```
  Папка `frontend/` в корне репо НЕ существует и никогда не
  существовала после расщепления projects/. Должно быть:
  ```
  COPY projects/game-nova/frontends/nova/package.json projects/game-nova/frontends/nova/package-lock.json* ./
  COPY projects/game-nova/frontends/nova ./
  ```
  И `WORKDIR` / итоговый `COPY --from=builder /app/dist ...` без
  изменений (внутри builder-стейджа путь /app тот же).
- `deploy/nginx.frontend.conf` ссылка в этом Dockerfile (`COPY
  deploy/nginx.frontend.conf ...`) — корректна, не трогаем (в плане
  79 этот конфиг переедет в `deploy/configs/`, но 79 идёт после 78).

### Ф.4. Smoke

- `make frontend-run` стартует Vite на 5173 из новой папки.
- `make frontend-build` собирает прод-бандл.
- `make backend-test` — Go-тесты зелёные (golden_test.go
  читают legacy-PHP-tools по новым путям).
- `make e2e` — Playwright проходит (базовый smoke).
- `docker compose -f deploy/docker-compose.yml up frontend` собирает
  и стартует контейнер из новой раскладки.
- Если есть docker-compose сервис legacy-PHP — стартует с
  context'ом `projects/game-legacy-php`.

### Ф.5. Финализация

- Шапка плана 78 ✅.
- Запись итерации в `docs/project-creation.txt`.
- Обновление [docs/plans/72-remaster-origin-frontend-pixel-perfect.md](72-remaster-origin-frontend-pixel-perfect.md)
  и его промпта: путь `projects/game-origin/frontend/` → `projects/game-nova/frontends/origin/`.
- Обновление [docs/plans/76-remaster-nova-frontend-exchange-ui.md](76-remaster-nova-frontend-exchange-ui.md)
  и его промпта.
- Обновление CLAUDE.md описания структуры репозитория (frontends/
  раскладка + переименование legacy-PHP).

---

## Конвенции (R1)

- Папка `frontends/` (мн.ч.) — потому что их несколько (nova, origin,
  возможно позже mobile/experimental).
- Имена внутри — `nova`, `origin`, без префиксов.
- Файлы внутри каждого фронта структурируются как раньше
  (Vite-проект: `src/`, `public/`, `package.json`, `Dockerfile`,
  `vite.config.ts`, `tsconfig.json`).

---

## Объём

~1-2 часа работы агента. Замена 1 (frontend→frontends/nova): ~15-25
правок путей в коде/конфигах + ~85 правок в документации. Замена 2
(game-origin-php → game-legacy-php): ~355 упоминаний в 67 файлах,
большинство автоматизируется через `git mv` + sed-replace в docs.

**Два коммита** (изоляция blame):
1. `refactor(game-nova): frontend → frontends/nova (план 78 Ф.1.1)`
2. `refactor(legacy): game-origin-php → game-legacy-php (план 78 Ф.1.2)`

Либо один коммит если агенту удобнее — но в commit-message явно
перечислить **обе** замены.

---

## Что разблокирует

- **План 72** — теперь origin-фронт создаётся в
  `projects/game-nova/frontends/origin/`.
- **Будущий план** «mobile-фронт» / «experimental-фронт» — кладётся
  рядом без новых решений.
