# План 51: Переименование auth-service → identity-service

**Дата**: 2026-04-27
**Статус**: Активный
**Зависимости**: нет блокирующих; должен быть выполнен **до** плана 52
(RBAC-unification зафиксирует имя в новых endpoints, JWT issuer claim,
документации — после плана 52 цена переименования вырастет в 3-5 раз).
**Связанные документы**: [52-rbac-unification.md](52-rbac-unification.md),
[53-admin-frontend.md](53-admin-frontend.md), [54-billing-limits.md](54-billing-limits.md).

---

## Зачем

`projects/auth/` сейчас содержит больше чем authentication:
- Authentication (логин/JWT) ✅ — соответствует имени
- Authorization (выпуск ролей, проверка) — уже не auth
- User management (lazy-create, consent, delete)
- Global credits — кросс-вселенский кошелёк
- Moderation (UGC blacklist)
- Rate limiting

В индустрии такой набор называется **Identity Service** (Identity Provider
по терминологии OIDC). Имя `auth` ассоциируется с library-stack 2010-х
(passport.js, devise) и не отражает реального scope сервиса.

Современные референсы: ORY Kratos, Keycloak, Auth0, AWS Cognito —
все они «Identity Provider», но называются нейтральным именем продукта,
не «auth».

В JWT issuer claim будет `iss: https://identity.oxsar-nova.ru` — это
автоматически делает сервис IdP по OIDC-стандарту.

## Что переименовываем

| Слой | Сейчас | Станет |
|---|---|---|
| Каталог в монорепо | `projects/auth/` | `projects/identity/` |
| Go module | `oxsar/auth` | `oxsar/identity` |
| Импорты пакетов | `oxsar/auth/internal/...`, `oxsar/auth/pkg/jwtrs`, `oxsar/auth/pkg/metrics` | `oxsar/identity/...` |
| Внутренний пакет | `internal/authsvc/` | `internal/identitysvc/` |
| Docker-compose service | `auth` | `identity` |
| ENV-vars (внутри identity) | `AUTH_*` (`AUTH_DB_URL`, `AUTH_JWT_PRIVATE_KEY`, `AUTH_PORT`) | `IDENTITY_*` |
| ENV-vars (в клиентах: game-nova/billing/portal/game-origin) | `AUTH_JWKS_URL`, `AUTH_BASE_URL` | `IDENTITY_JWKS_URL`, `IDENTITY_BASE_URL` |
| HTTP route prefix | `/auth/*` | `/identity/*` (с redirect 301 со старого) |
| JWT issuer claim | `https://auth.oxsar-nova.ru` (либо что сейчас) | `https://identity.oxsar-nova.ru` |
| БД (если есть отдельная) | `auth_db` | `identity_db` |
| Документация | все упоминания "auth-service" | "identity-service" |
| План-файлы | ссылки в 36/41/42/44 и др. | обновить ссылки на 51 |

**Что НЕ переименовываем**:
- `pkg/jwtrs` остаётся в identity (тот же путь). Это библиотека JWT-RS256
  signing/verifying, имя описывает **что делает**, не привязано к
  «auth».
- Слово `auth` в общеязыковом значении («authentication», «authorize»,
  HTTP `Authorization` header) — это термины OIDC/RFC, не наш сервис.
- БД-таблица `users` остаётся `users`.
- JWT-токен называется JWT, не «auth-token».

## Этапы

### Ф.1. Подготовка и инвентаризация

- `git grep -l "oxsar/auth" projects/` — список файлов с импортами.
- `git grep -l "auth-service\|AUTH_" projects/ docs/` — конфиги и упоминания.
- `git grep -E "/auth/(login|register|refresh|logout|jwks)" projects/` — клиенты.
- Сохранить список в `docs/plans/51-followup-checklist.md` (для ручной
  проверки в Ф.7).

### Ф.2. Каталог и Go module

- `git mv projects/auth projects/identity` (history сохраняется).
- `projects/identity/backend/go.mod`: `module oxsar/identity`.
- `git mv projects/identity/backend/internal/authsvc projects/identity/backend/internal/identitysvc`.
- `git grep -lE "oxsar/auth/" | xargs sed -i 's|oxsar/auth/|oxsar/identity/|g'`.
- `git grep -lE "internal/authsvc" | xargs sed -i 's|internal/authsvc|internal/identitysvc|g'`.
- Проверка: `cd projects/identity/backend && go mod tidy && go build ./...`
  должны пройти без ошибок.

### Ф.3. Docker, nginx, CI

- `deploy/docker-compose*.yml`: rename service `auth` → `identity`,
  обновить depends_on в зависимых сервисах.
- nginx upstream: `auth` → `identity` (если есть).
- `deploy/scaling.yml` overlay (если есть): обновить.
- `.github/workflows/ci.yml`: пути к тестам/билдам identity.
- `Makefile` targets: `make identity-run`, `make identity-test`.

### Ф.4. ENV-vars

- В `projects/identity/`: переименовать все ENV `AUTH_*` → `IDENTITY_*`
  через `sed -i` по `cmd/server/main.go` и `internal/`.
- В клиентах (`projects/game-nova/`, `projects/billing/`, `projects/portal/`):
  `AUTH_JWKS_URL` → `IDENTITY_JWKS_URL`, `AUTH_BASE_URL` →
  `IDENTITY_BASE_URL`.
- `.env.example` файлы — обновить.
- `docker-compose.yml` (env): обновить keys и значения.

### Ф.5. HTTP routes + JWT issuer

- В identity: route prefix `/auth/*` → `/identity/*`.
- В identity: добавить **301 redirect** для каждого старого пути:
  `/auth/login` → `/identity/login` и т.д. (на 30 дней, потом снести).
- JWT issuer (`iss` claim в новых токенах): `https://identity.oxsar-nova.ru`.
- Старые токены (с `iss=auth.oxsar-nova.ru`) принимаем ещё 1 неделю
  для grace period — клиенты успеют обновиться. После — отвергаем.
- В JWKS-валидаторах клиентов: temporarily allow both issuers, потом
  убрать old.

### Ф.6. Документация

- `git grep -l "auth-service\|projects/auth" docs/` → массовый sed
  с ручной проверкой каждого документа.
- Особое внимание:
  - `docs/architecture/*.md` (если есть).
  - `docs/ops/admin-access.md` (создаётся в плане 53).
  - `docs/plans/36-portal-multiverse.md` (упоминает auth-service).
  - `docs/plans/41-origin-rights.md`.
  - `docs/plans/44-personal-data-152fz.md`.
  - `README.md` главный.
  - `CLAUDE.md` (memory-instructions).
- В `docs/project-creation.txt` итерации, упоминающие auth-service —
  не правим (исторические записи).

### Ф.7. Smoke и верификация

- `docker compose up identity` — поднимается без ошибок.
- `curl http://localhost:PORT/identity/.well-known/jwks.json` —
  возвращает JWKS.
- `curl -X POST http://localhost:PORT/identity/register` — регистрация
  работает.
- Логин в game-nova через identity → JWT валидируется. Старый JWT
  (если есть в браузере) принимается grace period.
- `make test` — все тесты зелёные (важно: jwtrs_test.go в каждом из
  4 модулей — auth/portal/game-nova/billing — должны успешно
  обновиться).
- `make lint` — golangci-lint без ошибок.
- Чек-лист из `docs/plans/51-followup-checklist.md` — каждый пункт
  отмечен.

### Ф.8. Финализация

- Удалить чек-лист followup-файла после полной верификации.
- Обновить `docs/project-creation.txt` записью «итерация 51».
- В `docs/simplifications.md` — НЕ записываем (это не упрощение, а
  нейминг-рефакторинг, никаких компромиссов).
- Один большой коммит (или 3-4 логических: каталог/импорты, ENV,
  routes/issuer, docs).

## Тестирование

- Backend: `go test ./...` в каждом из 4 модулей (identity,
  game-nova, billing, portal) — все зелёные.
- E2E: Playwright-сценарий «register → login → access protected
  endpoint» в game-nova и portal.
- Регрессии: smoke 41 страниц game-origin (шаблон из
  `docs/prompts/compare-screens.md`) — login через JWT не сломан.
- Manual: реально открыть admin-route в браузере, проверить JWT
  в DevTools.

## Риски и митигация

1. **Grace-period для issuer**: старые токены могут гулять у юзеров в
   браузерах. 7 дней grace — компромис между быстрым закрытием
   и UX. После можно делать force-logout.
2. **Параллельная сессия с соседним агентом**: возможен конфликт по
   импортам (он тоже может править файлы из identity/auth). Делать
   когда никто не работает в репо, либо в worktree.
3. **CI/CD pipeline**: переименование сервиса в pipeline = риск
   простоя deploy'я. Делать в feature-branch, мержить когда CI зелёный.
4. **Историческая привязка**: коммиты «auth-service» останутся в логе
   — это нормально. Не переписываем git history.

## Возможный rollback

1. `git revert` коммита переименования каталога — Go-импорты
   ломаются, нужно перезапускать сервисы.
2. Лучше иметь feature-flag `IDENTITY_RENAMED=false` и держать
   двойной alias на уровне docker-compose (alias service `auth` →
   `identity`) на 1-2 дня после деплоя.

## Альтернативы (отвергнуты)

- **Оставить `auth`**: технически работает, но не соответствует scope
  сервиса. После плана 52 (RBAC unification) и 53 (admin frontend) имя
  застрянет в JWT issuer, OpenAPI specs, документации — переименование
  будет в 3-5 раз дороже.
- **Имя `idp`**: аббревиатура, ассоциация с ролью в OIDC-flow, не с
  именем сервиса. Никто в индустрии так не называет (Google = accounts,
  ORY = kratos, Auth0 = auth0). Затрудняет чтение в коде.
- **Имя `identity-service`**: явнее, но длиннее. В монорепо
  `projects/identity/` смотрится лучше короткое имя.
- **Имя `iam`** (Identity & Access Management): корпоративный термин
  AWS/GCP, обычно когда есть сложные policies/groups. Для нас оверкилл.
- **Имя `accounts`** (стиль Google): акценты на user-profile, у нас же
  fokus на identity+roles.

## Итог

Один большой rename-рефакторинг ради будущей чистоты архитектуры.
Стоимость сейчас ≈ 4-8 часов работы. Цена откладывания (после
выполнения 52-54) ≈ 16-24 часа. Результат — сервис identity с
правильным OIDC-naming, готовый к плану 52 (RBAC unification).
