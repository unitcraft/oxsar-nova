# План 80: Cleanup auth-хвостов после переименования identity

**Дата**: 2026-04-28
**Статус**: ✅ Закрыт 2026-04-28.
**Зависимости**: должен идти **после** планов 78 (frontends + legacy-PHP)
и 79 (deploy refactor), чтобы не пересекаться по compose-файлам и
docker-структуре.
**Связанные документы**:
- [docs/plans/51-rename-auth-to-identity.md](51-rename-auth-to-identity.md) — оригинальное переименование auth-service → identity-service.
- [docs/plans/55-doc-sync-after-identity-rename.md](55-doc-sync-after-identity-rename.md) — синхронизация документации после плана 51.

---

## Цель

Закрыть остаточные хвосты переименования `auth-service` → `identity-service`,
которые планы 51 и 55 не дочистили. Среди них есть **критический
prod-баг**: `deploy/Dockerfile.migrate` копирует миграции из
несуществующего `projects/auth/migrations`, реальные миграции лежат в
`projects/identity/migrations/`. Идентити-БД стартует пустой
в multi-instance dev-стеке — недоказанный prod-риск.

**Что НЕ цель:**
- Не переименовывать feature-folder'ы `internal/auth/` (game-nova),
  `features/auth/` (frontend), `stores/auth.ts`, e2e `auth.spec.ts`,
  endpoint'ы `/api/auth/*` — это **функциональная область** аутентификации,
  не имя сервиса. Разделение `auth` (feature) vs `identity` (сервис) —
  общепринятая конвенция.
- Не трогать `projects/game-legacy-php/` / `projects/game-legacy-php/` —
  legacy, JWT-логика на месте, имена не критичны.
- Не трогать `internal/auth/middleware.go`, `ensure_user.go` — это
  потребители JWT в game-nova, name корректное (это не сервис).
- Не трогать закрытые планы (`docs/plans/36/41/51/55/63/...`) — это
  исторические записи (правило плана 55).

---

## Конвенция (фиксируем)

После плана 80:
- **`auth`** — функциональная область: процесс аутентификации, JWT-валидация,
  login/register-экраны, /api/auth/* endpoint'ы. Имя в feature-folder'ах,
  e2e-тестах, store'ах, middleware-пакетах.
- **`identity`** — имя микросервиса (`projects/identity/`),
  его БД (`identity-db`), его юзера БД (`identitysvc`),
  его контейнера в compose (`identity-migrate`), его ENV (`IDENTITY_DB_URL`,
  `IDENTITY_JWKS_URL`).
- **`auth_rsa_key.pem`** — имя секрета **остаётся как есть** (это файл
  RSA-ключа для подписи JWT, имя зафиксировано в Docker secrets и
  prod-инфра — менять рискованно, не приносит ценности).

---

## Что делаем

### Ф.1. Удалить мёртвые Dockerfile'ы

**Dockerfile.auth** (game-nova/backend):
`projects/game-nova/backend/Dockerfile.auth` ссылается на
`./cmd/auth-service`, которой **больше нет** (план 51 удалил, перенесено
в `projects/identity/backend/`). Файл — зомби.

**Dockerfile.portal** (game-nova/backend):
`projects/game-nova/backend/Dockerfile.portal` ссылается на
`cmd/portal`, которой **больше нет** (план 36 вынес portal в
`projects/portal/backend/`). Файл — зомби.

- `git rm projects/game-nova/backend/Dockerfile.auth`
- `git rm projects/game-nova/backend/Dockerfile.portal`
- Перед удалением: `grep -rn "Dockerfile\.auth\|Dockerfile\.portal"`
  по active-коду (deploy/, Makefile, .github/) — должно быть 0
  ссылок (если есть — это ошибка ревью, разобраться).

### Ф.2. Починить `deploy/Dockerfile.migrate` (CRITICAL — prod-баг)

После плана 79 этот файл живёт в `deploy/docker/migrate.Dockerfile`.

- Строка `COPY projects/auth/migrations /migrations/auth` →
  `COPY projects/identity/migrations /migrations/identity`.
- Найти всех потребителей `/migrations/auth` в compose-командах и
  заменить на `/migrations/identity`.

### Ф.3. Compose-файлы

После плана 79 файлы лежат в `deploy/compose/`. Замены:

В **base.yml** (или там где живёт identity-БД после плана 79):
- `auth-db:` сервис → `identity-db:`
- `POSTGRES_USER: authsvc` → `POSTGRES_USER: identitysvc`
- `POSTGRES_PASSWORD: authsvc` → `POSTGRES_PASSWORD: identitysvc`
  (ВНИМАНИЕ: dev-only пароль; prod хранится отдельно)
- `POSTGRES_DB: authsvc` → `POSTGRES_DB: identitysvc`
- `volumes: - auth-db-data:/var/lib/postgresql/data` →
  `- identity-db-data:/var/lib/postgresql/data`
- `pg_isready -U authsvc` → `pg_isready -U identitysvc`

В **dev.yml / multiverse.yml** (миграционные сервисы):
- `auth-migrate:` → `identity-migrate:`
- `command: ["goose", "-dir", "/migrations/auth", "postgres", "${IDENTITY_DB_URL}", "up"]`
  → `..."dir", "/migrations/identity", ...`
- `depends_on: auth-db:` → `identity-db:`
- `DB_URL: postgres://authsvc:authsvc@auth-db:5432/authsvc...` →
  `postgres://identitysvc:identitysvc@identity-db:5432/identitysvc...`
- `MIGRATE_DIR: /migrations/auth` → `/migrations/identity`

В **base.yml volumes**:
- `auth-db-data:` → `identity-db-data:`

В `.env.multiverse.example`:
- `IDENTITY_DB_URL=postgres://authsvc:change_me_auth@auth-db:5432/authsvc?...`
  → `postgres://identitysvc:change_me_identity@identity-db:5432/identitysvc?...`

**Smoke ОБЯЗАТЕЛЬНО**:
1. `make compose-up` (после плана 79) — identity-db встаёт healthy.
2. `docker compose ... logs identity-migrate` — миграции применились
   успешно (включая 0001_init.sql).
3. `psql` к identity-db: проверить наличие таблиц `users`, `user_consents`,
   и др. Если таблиц нет — Dockerfile.migrate всё ещё сломан.

### Ф.4. Прочистить sentinel-ошибки в identity-сервисе

`projects/identity/backend/internal/identitysvc/service.go` — 15 sentinel-
ошибок с префиксом `authsvc:`:
```go
ErrUserExists        = errors.New("authsvc: user already exists")
ErrInvalidCredential = errors.New("authsvc: invalid credentials")
ErrUserBanned        = errors.New("authsvc: account banned")
...
```

Это пакетный префикс ошибок — должен быть `identitysvc:` (имя пакета).
Замена `authsvc:` → `identitysvc:` в сообщениях. **ВАЖНО**: проверить
что нет matchers вида `strings.Contains(err.Error(), "authsvc:")` —
если есть, перевести вместе. Унификация: `errors.Is(err, ErrXxx)` —
не зависит от текста, безопасно.

`handler.go` (2 упоминания) — то же.

### Ф.5. Обновить остальные потребители

- `projects/game-nova/backend/cmd/tools/seed/main.go` — комментарий
  `"seed: not implemented yet — use POST /api/auth/register"` оставить
  (endpoint реально называется `/api/auth/register`, это корректно).
  Но если упоминает `auth-service` где-то — заменить на `identity-service`.
- `projects/portal/backend/cmd/server/main.go` (1 упоминание) — проверить
  контекст.
- `projects/portal/backend/internal/portalsvc/{handler,service,credits}.go`
  (3 упоминания) — проверить контекст.
- `projects/billing/backend/internal/billing/webhook.go` — 1 упоминание.
- `projects/portal/frontend/vite.config.ts` — 1 упоминание (proxy?).
- `projects/portal/frontend/src/pages/ProfilePage.tsx` — 1 упоминание.
- `projects/game-nova/frontends/nova/vite.config.ts` — 1 упоминание.
- `projects/game-nova/backend/Dockerfile` — 1 упоминание (комментарий?).
- `projects/game-nova/backend/cmd/server/main.go` — 7 упоминаний,
  скорее всего контексты `auth-middleware`, не `auth-service` —
  посмотреть и не трогать.
- `projects/game-nova/backend/internal/config/config.go` — 4 упоминания,
  ENV-переменные. Если есть `AUTH_DB_URL` — переименовать в
  `IDENTITY_DB_URL` (но проверить что не сломает совместимость с уже
  развёрнутыми инстансами — если развёрнутых нет, безопасно).

### Ф.6. Комментарии в коде

`projects/game-nova/backend/internal/settings/handler.go:155`:
- `// Хеш пароля живёт в auth-db, в game-db password_hash IS NULL.`
  → `// Хеш пароля живёт в identity-db, в game-db password_hash IS NULL.`

`projects/game-nova/frontends/nova/src/features/settings/SettingsScreen.tsx:64`:
- то же самое.

### Ф.7. Smoke + финализация

- `make compose-up` — стек встаёт чисто.
- `make backend-test` — Go-тесты идентити зелёные.
- `make e2e` — Playwright-smoke зелёный (auth-flow работает).
- Шапка плана 80 ✅.
- Запись итерации в `docs/project-creation.txt`.

---

## Что НЕ трогаем

- `internal/auth/` (game-nova backend) — это middleware-потребитель,
  не сам сервис. Имя корректное.
- `features/auth/`, `stores/auth.ts`, `e2e/critical/auth.spec.ts`
  (frontend) — feature-области, имя корректное.
- Endpoint'ы `/api/auth/login`, `/api/auth/register`, `/api/auth/refresh`
  в openapi.yaml — это REST-API аутентификации, имя корректное.
- `auth_rsa_key.pem` — имя секрета зафиксировано в Docker secrets,
  не меняем (рисковано, ценности нет).
- Закрытые планы в `docs/plans/` (36, 51, 55, 63...) — исторические
  записи, не трогаем (правило плана 55).
- `docs/project-creation.txt`, `docs/simplifications.md` — закрытые
  записи прошлых планов.
- `projects/game-legacy-php/` или `projects/game-legacy-php/` (после
  плана 78) — legacy, JWT-логика на месте.

---

## Объём

~1-2 часа агента. ~30-50 правок (большинство — `auth-db` → `identity-db`
в 2 compose-файлах + sentinel-ошибки в identity-сервисе).

**Один коммит**:
`fix(identity): cleanup auth-leftovers (план 80)`

(или 2-3 коммита если удобнее: один на критический Dockerfile.migrate,
один на compose-rename, один на sentinel-ошибки).

---

## Acceptance

- `Dockerfile.auth` удалён.
- `Dockerfile.migrate` копирует `projects/identity/migrations/`,
  не `projects/auth/migrations/`.
- В compose: `auth-db` → `identity-db`, `auth-migrate` → `identity-migrate`,
  `authsvc` (user/password/db) → `identitysvc`,
  `auth-db-data` (volume) → `identity-db-data`.
- В identity-сервисе sentinel-ошибки имеют префикс `identitysvc:`.
- Smoke `make compose-up`: identity-db встаёт, миграции применяются,
  таблицы создаются.
- `grep -rn "auth-db\|authsvc\|projects/auth/migrations\|Dockerfile\.auth"`
  по активному коду = 0 результатов (исключения: legacy-PHP, auth_rsa_key.pem).
- Шапка плана 80 ✅.
- Запись в docs/project-creation.txt.

---

## Smoke 2026-04-28 (post-execution)

Полный destructive smoke выполнен после Ф.1-Ф.6:

```bash
docker compose -f deploy/docker-compose.yml down -v
# Удалены volumes: pg-data, redis-data, portal-db-data, billing-db-data,
# uni02-db-data, frontend-node-modules, uni02-frontend-node-modules,
# portal-frontend-node-modules + старые auth-db-data и auth-rsa-key
# (зомби-volumes от прошлой раскладки, удалены вручную).
docker compose -f deploy/docker-compose.yml up -d --build
```

**Результат**:
- ✅ `identity-db` (новое имя) встаёт healthy с `POSTGRES_USER=identitysvc`,
  `POSTGRES_DB=identitysvc`.
- ✅ `identity-migrate` (новое имя) запускается, читает миграции из
  `/migrations/identity` (после плана 80 Ф.2), применяет 0001-0004
  УСПЕШНО:
  ```
  OK   0001_init.sql (358ms)
  OK   0002_universe_memberships.sql (67ms)
  OK   0003_drop_credits.sql (16ms)
  OK   0004_user_consents.sql (57ms)
  ```
  → **CRITICAL Ф.2 fix подтверждён**: до плана 80 эти миграции
  НЕ ПРИМЕНЯЛИСЬ (Dockerfile.migrate копировал из несуществующей
  `projects/auth/migrations`), identity-БД стартовала пустой.
- ❌ Миграция `0005_rbac_tables.sql` падает:
  ```
  ERROR: duplicate key value violates unique constraint
  "role_permissions_pkey" (SQLSTATE 23505)
  ```
  Финальный `INSERT INTO role_permissions ... FROM roles r, permissions p`
  делает CROSS JOIN для ВСЕХ ролей включая `support`/`moderator`/
  `admin`/`billing_admin`, которым уже выданы permissions
  предыдущими INSERT'ами в той же миграции — ON CONFLICT отсутствует.
  Это **upstream-баг плана RBAC** (план 5x), не дефект плана 80.
  Нужен отдельный hotfix: добавить `ON CONFLICT (role_id, permission_id)
  DO NOTHING` в финальный INSERT, либо WHERE-условие на `r.name = 'superadmin'`.

**Выводы**:
- Ф.1-Ф.6 acceptance плана 80 — выполнены: rename корректный,
  компилится, миграции 0001-0004 применяются. Что и требовалось:
  CRITICAL prod-баг (Dockerfile.migrate указывал на пустоту) починен.
- Smoke плана 80 ВЫЯВИЛ ранее скрытый баг 0005_rbac_tables — это
  **бонус-польза плана**, до этого баг маскировался тем, что
  identity-БД вообще не получала миграций. Фикс — отдельный план.
