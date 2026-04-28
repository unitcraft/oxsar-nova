# Промпт: выполнить план 80 (cleanup auth-хвостов после identity-rename)

**Дата создания**: 2026-04-28
**План**: [docs/plans/80-auth-leftovers-cleanup.md](../plans/80-auth-leftovers-cleanup.md)
**Зависимости**: ✅ план 78 (frontends + legacy-PHP rename).
План 79 (deploy refactor) НЕ обязателен — если 79 не запущен,
работаешь с текущей раскладкой `deploy/Dockerfile.*` + плоской
`deploy/docker-compose*.yml`. Если 79 закрыт — пути уже в
`deploy/compose/` + `deploy/docker/`. Адаптируйся по факту.
**Объём**: ~1-2 часа, ~30-50 правок, 2-3 коммита.

---

```
Задача: выполнить план 80 — cleanup остаточных хвостов
переименования auth-service → identity-service после планов 51 и 55.

КОНТЕКСТ:

План 51 (закрыт) переименовал auth-service → identity-service
(Go-сервис теперь в projects/identity/backend/). План 55 (закрыт)
прошёлся по доке. Но РЕМОНТ НЕ ПОЛНЫЙ — остались хвосты:

КРИТИЧЕСКИЙ prod-баг (Ф.2):
- deploy/Dockerfile.migrate (или deploy/docker/migrate.Dockerfile
  после плана 79) копирует `projects/auth/migrations` —
  папки НЕТ. Реальные миграции лежат в `projects/identity/migrations/`.
  Идентити-БД стартует пустой в multi-instance dev-стеке (миграции
  не применяются). Docker COPY с несуществующего пути молча проходит
  в одних режимах buildx, в других падает — у нас сейчас проходит,
  потому и не нашли раньше.

Прочие хвосты:
- Dockerfile.auth (мёртвый, ссылается на несуществующий
  cmd/auth-service).
- Dockerfile.portal (мёртвый, ссылается на несуществующий
  cmd/portal в game-nova-backend).
- В compose: auth-db / authsvc / auth-migrate / auth-db-data —
  имена остались с эпохи auth-service, должны стать identity-*.
- Sentinel-ошибки в identity-сервисе с префиксом authsvc:
  (15+ штук в service.go, 2 в handler.go).
- Комментарии в коде: «Хеш пароля живёт в auth-db» (settings/handler.go,
  SettingsScreen.tsx) — должны указывать на identity-db.

ВАЖНО ПРО ТЕРМИНОЛОГИЮ (ЧТО НЕ ТРОГАТЬ):

- internal/auth/ (game-nova backend) — это middleware-потребитель JWT,
  имя ПРАВИЛЬНОЕ. НЕ трогать.
- features/auth/ (frontend), stores/auth.ts, e2e/critical/auth.spec.ts —
  feature-области аутентификации, НЕ имя сервиса. НЕ трогать.
- Endpoints /api/auth/login, /api/auth/register, /api/auth/refresh —
  REST API аутентификации, имя ПРАВИЛЬНОЕ. НЕ трогать.
- auth_rsa_key.pem — имя секрета зафиксировано в Docker secrets и
  prod-инфра, менять рискованно. НЕ трогать.
- Закрытые планы (docs/plans/36/41/51/55/63/...) — исторические записи
  (правило плана 55). НЕ трогать.
- projects/game-legacy-php/ — legacy, JWT-логика на месте. НЕ трогать.

ПРАВИЛО:
- `auth` = функциональная область (process, middleware, feature, endpoint).
- `identity` = имя микросервиса (его БД, его контейнер, его ENV).

ПЕРЕД НАЧАЛОМ:

ПЕРВЫМ ДЕЙСТВИЕМ (до любого чтения плана):

1) git status --short. cat docs/active-sessions.md.

2) ОБЯЗАТЕЛЬНО добавь свою строку в раздел «Активные сессии»:
   | <N> | План 80 auth-cleanup | deploy/, projects/identity/backend/internal/identitysvc/, projects/game-nova/backend/{Dockerfile.auth,Dockerfile.portal,internal/settings/handler.go}, projects/game-nova/frontends/nova/src/features/settings/SettingsScreen.tsx, .github/workflows/, docs/ | <дата-время> | fix(identity): cleanup auth-leftovers (план 80) |

3) Если в active-sessions.md есть ДРУГИЕ слоты (72 Ф.4, 73 Ф.1) —
   ОК, они работают в frontends/origin/ и tests/e2e/. Конфликта
   по файлам не будет: твоя территория — deploy/, identity-сервис,
   .github/workflows/, и точечные комментарии в game-nova-backend
   и frontends/nova.

   ВАЖНО: НЕ трогай:
   - frontends/origin/ (план 72 Ф.4 пишет туда).
   - tests/e2e/origin-baseline/ или tests/e2e/origin-screens/
     (план 73 Ф.1 пишет туда).

ТОЛЬКО ПОСЛЕ ШАГОВ 1-3 — переходи к чтению:

4) Прочитай ПОЛНОСТЬЮ:
   - docs/plans/80-auth-leftovers-cleanup.md (твоё ТЗ)
   - docs/research/origin-vs-nova/roadmap-report.md «Часть I.5» R0-R15
   - docs/plans/51-rename-auth-to-identity.md (контекст переименования)
   - deploy/Dockerfile.migrate (КРИТИЧНОЕ — ты будешь его править)
   - deploy/docker-compose.yml + deploy/docker-compose.multiverse.yml
     (или deploy/compose/dev.yml + multiverse.yml если план 79
     закрыт; адаптируйся)

5) Прочитай выборочно:
   - projects/identity/backend/internal/identitysvc/service.go
     (15 sentinel-ошибок)
   - projects/identity/backend/internal/identitysvc/handler.go
     (2 упоминания)
   - projects/game-nova/backend/internal/settings/handler.go:155
     (один комментарий)
   - projects/game-nova/frontends/nova/src/features/settings/SettingsScreen.tsx:64
     (один комментарий)

ЧТО НУЖНО СДЕЛАТЬ:

### Ф.1. Удалить мёртвые Dockerfile'ы

git rm projects/game-nova/backend/Dockerfile.auth
git rm projects/game-nova/backend/Dockerfile.portal

ПЕРЕД удалением — проверка:
grep -rn "Dockerfile\.auth\|Dockerfile\.portal" deploy/ Makefile .github/

Если найдены ссылки — это ошибка ревью, разберись (возможно где-то
build:dockerfile: deploy/Dockerfile.auth — тогда сначала удалить
ссылки в compose/CI, потом git rm).

### Ф.2. Починить migrate (CRITICAL)

ФАЙЛ: deploy/Dockerfile.migrate (если план 79 не закрыт)
   ИЛИ deploy/docker/migrate.Dockerfile (если 79 закрыт).

Замени строку:
   COPY projects/auth/migrations /migrations/auth
на:
   COPY projects/identity/migrations /migrations/identity

Все compose-команды/ENV что упоминают /migrations/auth —
заменить на /migrations/identity:
- deploy/docker-compose.yml: `command: ["goose", "-dir",
  "/migrations/auth", ...]` → "/migrations/identity".
- deploy/docker-compose.multiverse.yml: то же.
- deploy/.env.multiverse.example: MIGRATE_DIR=/migrations/auth →
  /migrations/identity.

Smoke ОБЯЗАТЕЛЬНО:
1. docker compose -f deploy/docker-compose.yml build identity-migrate
   (или migrate если имя ещё не переименовано — это в Ф.3).
2. docker compose -f deploy/docker-compose.yml up identity-migrate
   (или migrate). Логи: миграции 0001+0002+... применяются,
   exit code 0.
3. psql к identity-БД: \dt — должны быть таблицы users,
   user_consents, refresh_tokens и пр. Если таблиц нет — Ф.2
   ещё не починена.

### Ф.3. Compose-rename (auth-* → identity-*)

В deploy/docker-compose.yml, deploy/docker-compose.multiverse.yml,
deploy/.env.multiverse.example (или их аналогах в deploy/compose/
после плана 79):

- service `auth-db` → `identity-db`
- service `auth-migrate` → `identity-migrate`
- volume `auth-db-data` → `identity-db-data`
- POSTGRES_USER: authsvc → identitysvc
- POSTGRES_PASSWORD: authsvc → identitysvc (dev-only пароль)
- POSTGRES_DB: authsvc → identitysvc
- pg_isready -U authsvc → -U identitysvc
- DB_URL: postgres://authsvc:authsvc@auth-db:5432/authsvc →
  postgres://identitysvc:identitysvc@identity-db:5432/identitysvc

Все depends_on / volumes ссылки на старые имена — обновить.

В .env.multiverse.example:
- IDENTITY_DB_URL=postgres://authsvc:change_me_auth@auth-db:... →
  postgres://identitysvc:change_me_identity@identity-db:...

Smoke:
- docker compose ... down -v (удалить старые volumes; ВАЖНО предупредить
  пользователя что БД пересоздастся пустой — у него локально были
  тестовые юзеры).
- docker compose ... up — стек поднимается, identity-db healthy,
  identity-migrate exit 0, миграции применились.

ВАЖНО: down -v — DESTRUCTIVE для dev-БД. ПЕРЕД ЗАПУСКОМ — спроси
пользователя «Готов к down -v локально (потеряешь dev-данные)?
Альтернатива: попробовать ALTER USER/RENAME DATABASE без down,
но это сложно с goose, проще пересоздать.»

### Ф.4. Sentinel-ошибки в identity-сервисе

В projects/identity/backend/internal/identitysvc/service.go и
handler.go:

Замени `errors.New("authsvc: ...")` → `errors.New("identitysvc: ...")`.

Sed-подобная замена по этим двум файлам:
   authsvc: → identitysvc:

ВАЖНО: проверь что нигде в коде нет matchers вида
strings.Contains(err.Error(), "authsvc:") — если есть, перевести
вместе. Унификация: errors.Is(err, ErrXxx) уже не зависит от текста.

go test ./internal/identitysvc/... — должно остаться зелёным
(тексты ошибок не специфичны в тестах, обычно проверяют sentinel
через errors.Is).

### Ф.5. Прочистить остальные потребители (точечно)

- projects/portal/backend/cmd/server/main.go (1 упоминание) — посмотри
  контекст, обнови если ссылается на auth-service / authsvc.
- projects/portal/backend/internal/portalsvc/{handler,service,credits}.go
  (3 упоминания) — то же.
- projects/billing/backend/internal/billing/webhook.go (1) — то же.
- projects/portal/frontend/vite.config.ts (1) — proxy → /api/auth/* OK,
  proxy → auth-service host = поправить.
- projects/portal/frontend/src/pages/ProfilePage.tsx (1) — комментарий?
- projects/game-nova/frontends/nova/vite.config.ts (1) — то же что
  portal/frontend.
- projects/game-nova/backend/Dockerfile (1 упоминание) — комментарий
  или COPY?
- projects/game-nova/backend/cmd/server/main.go (7 упоминаний) —
  большинство = auth-middleware (НЕ трогать), но проверь не указывает
  ли на auth-service host где-то.
- projects/game-nova/backend/internal/config/config.go (4) —
  ENV-переменные. Если AUTH_DB_URL — переименовать в IDENTITY_DB_URL
  (но проверь backwards compat, скорее всего уже IDENTITY_*).

### Ф.6. Комментарии в коде

projects/game-nova/backend/internal/settings/handler.go:155 —
`// Хеш пароля живёт в auth-db` → `// Хеш пароля живёт в identity-db`.

projects/game-nova/frontends/nova/src/features/settings/SettingsScreen.tsx:64 —
то же.

### Ф.7. Smoke + финализация

- make compose-up или docker compose up — стек встаёт чисто.
- make backend-test — Go-тесты identity зелёные.
- make e2e — auth-flow работает (опционально, если Docker
  доступен).
- Шапка плана 80 ✅.
- Запись итерации в docs/project-creation.txt.

ПРОВЕРКА чистоты grep'ом:
   grep -rn "auth-db\|authsvc\|projects/auth/migrations\|Dockerfile\.auth\|Dockerfile\.portal" \
     --exclude-dir=node_modules --exclude-dir=.git \
     deploy/ projects/identity/ projects/game-nova/ projects/portal/ \
     projects/billing/ Makefile .github/ \
     | grep -v "auth_rsa_key" | grep -v "/api/auth/" | grep -v "internal/auth/" \
     | grep -v "features/auth/" | grep -v "stores/auth.ts" \
     | grep -v "e2e.*auth.spec.ts" | grep -v "game-legacy-php"

Должен вернуть 0 результатов (исключения отфильтрованы grep -v).

══════════════════════════════════════════════════════════════════
ОБЯЗАТЕЛЬНЫЕ ПРАВИЛА R0-R15 — см. блок rules-r0-r15.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/rules-r0-r15.md

Особо важно:
- R0: nova-баланс не меняем (этот план — refactor инфры).
- R7: backward compat технических интерфейсов не требуется до плана 74.
- R15: без MVP-сокращений. Перенеси аккуратно, smoke проверь.
══════════════════════════════════════════════════════════════════

══════════════════════════════════════════════════════════════════
GIT-ИЗОЛЯЦИЯ ОТ ПАРАЛЛЕЛЬНЫХ СЕССИЙ — см. блок git-isolation.md
СКОПИРУЙ СЮДА полностью раздел из docs/prompts/_blocks/git-isolation.md

Свои пути для CC_AGENT_PATHS:
- deploy/Dockerfile.migrate (или deploy/docker/migrate.Dockerfile)
- deploy/docker-compose.yml (или deploy/compose/dev.yml)
- deploy/docker-compose.multiverse.yml (или deploy/compose/multiverse.yml)
- deploy/.env.multiverse.example (или deploy/examples/.env.multiverse.example)
- projects/game-nova/backend/Dockerfile.auth (удалить)
- projects/game-nova/backend/Dockerfile.portal (удалить)
- projects/identity/backend/internal/identitysvc/service.go
- projects/identity/backend/internal/identitysvc/handler.go
- projects/game-nova/backend/internal/settings/handler.go
- projects/game-nova/frontends/nova/src/features/settings/SettingsScreen.tsx
- projects/portal/backend/cmd/server/main.go
- projects/portal/backend/internal/portalsvc/
- projects/billing/backend/internal/billing/webhook.go
- projects/portal/frontend/vite.config.ts
- projects/portal/frontend/src/pages/ProfilePage.tsx
- projects/game-nova/frontends/nova/vite.config.ts
- projects/game-nova/backend/Dockerfile (если будут правки)
- projects/game-nova/backend/cmd/server/main.go (точечно, если AUTH_*)
- projects/game-nova/backend/internal/config/config.go (если AUTH_DB_URL)
- .github/workflows/ci.yml (если есть auth-* job)
- .github/workflows/admin-console.yml (если есть auth-* job)
- docs/plans/80-auth-leftovers-cleanup.md
- docs/active-sessions.md
- docs/project-creation.txt (запись итерации)

ВНИМАНИЕ: НЕ трогай:
- projects/game-nova/frontends/origin/ (план 72 Ф.4 параллельно).
- tests/e2e/origin-baseline/ или origin-screens/ (план 73 Ф.1 параллельно).
- internal/auth/, features/auth/, stores/auth.ts, e2e/auth.spec.ts —
  feature-области, имя правильное.
- /api/auth/* endpoint'ы — REST имена правильные.
- auth_rsa_key.pem — имя секрета.
- Закрытые планы в docs/plans/.
- projects/game-legacy-php/.
══════════════════════════════════════════════════════════════════

КОММИТЫ:

2-3 коммита для blame-изоляции:

1) fix(deploy): миграция identity (CRITICAL fix Dockerfile.migrate) +
   удаление мёртвых Dockerfile'ов (план 80 Ф.1+Ф.2)
2) refactor(deploy): auth-db → identity-db rename + sentinel-ошибки
   identitysvc: (план 80 Ф.3+Ф.4)
3) chore(identity): cleanup комментариев + остальные потребители
   (план 80 Ф.5+Ф.6+Ф.7)

Trailer: Generated-with: Claude Code
ВСЕГДА с двойным тире:
git commit -m "..." -- $CC_AGENT_PATHS

ЧЕГО НЕ ДЕЛАТЬ:

- НЕ ломать backwards-compat внутренних интерфейсов (R7) —
  но он не требуется до плана 74, ОК.
- НЕ запускать docker compose down -v БЕЗ согласия пользователя
  (destructive для dev-данных).
- НЕ трогай НИЧЕГО из ЧТО НЕ ТРОГАТЬ списка выше.
- НЕ забывай про -- в git commit.

УСПЕШНЫЙ ИСХОД:

- Dockerfile.auth + Dockerfile.portal удалены.
- Dockerfile.migrate копирует projects/identity/migrations/.
- compose: auth-db→identity-db, auth-migrate→identity-migrate,
  authsvc→identitysvc.
- Sentinel-ошибки identitysvc: вместо authsvc:.
- Smoke `make compose-up`: identity-db healthy, миграции применились,
  таблицы существуют.
- grep-проверка чистоты возвращает 0 результатов.
- Шапка плана 80 ✅.
- Запись в docs/project-creation.txt.
- Удалена строка из docs/active-sessions.md.

Стартуй.
```
