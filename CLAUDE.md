# CLAUDE.md

Руководство для разработчиков (людей и AI-ассистентов) по работе с кодовой базой
oxsar-nova. Цель — чтобы новый участник мог продуктивно войти в проект за час.

## Что это такое

Порт legacy-игры oxsar2 (PHP/Yii 1.1 + MySQL) на современный стек: Go + TypeScript +
PostgreSQL + Redis. Боевой движок — порт отдельного java-проекта
`d:\Sources\oxsar2-java` на Go. Полное ТЗ — [oxsar-spec.txt](docs/oxsar-spec.txt).

## Onboarding: одноразовые настройки

**В начале каждой новой сессии Claude Code** (или каждой первой
рабочей сессии разработчика-человека) выполнить:

```bash
git config core.hooksPath scripts/git-hooks
```

Это активирует расшаренный git-hook `commit-msg`, который автоматически
удаляет из коммит-сообщений `Co-Authored-By: Claude <noreply@anthropic.com>`
(техническая метка AI, не подходит как git-стандарт «соавторство» —
см. [план 41 §6](docs/plans/41-origin-rights.md) и
[docs/origin-rights.md](docs/origin-rights.md) §6).

**Проверка:** `git config --get core.hooksPath` должно вернуть
`scripts/git-hooks`. Команда идемпотентна — повторный запуск не
ломает ничего, можно просто всегда выполнять при старте.

Опционально, для полной автоматизации, скопировать
[docs/ops/claude-code-attribution.md](docs/ops/claude-code-attribution.md)
шаблон в локальный `.claude/settings.json` (он в `.gitignore`).

Подробности про hook — [scripts/git-hooks/README.md](scripts/git-hooks/README.md).

## Запуск

```bash
make dev-up        # docker-compose up postgres + redis + миграции
make backend-run   # API-сервер (порт 8080)
make worker-run    # event-loop воркер
make frontend-run  # vite dev-сервер (порт 5173)
make test          # все тесты
make lint          # все линтеры
```

## Структура

- `projects/game-nova/backend/cmd/server`   — HTTP/WS entry point
- `projects/game-nova/backend/cmd/worker`   — фоновый обработчик events
- `projects/game-nova/backend/cmd/tools`    — CLI-утилиты (reseed, ресинк артефактов)
- `projects/game-nova/backend/internal/*`   — домены (auth, planet, fleet, battle, …)
- `projects/game-nova/backend/pkg/*`        — общие утилиты (rng, proto, ids)
- `projects/game-nova/frontends/nova/src/api`     — сгенерированный клиент из OpenAPI (фронт nova-стиля для uni01/uni02)
- `projects/game-nova/frontends/nova/src/features/<domain>` — вертикальные срезы UI
- `projects/game-nova/frontends/origin/` (план 72) — зарезервировано под новый React-фронт origin (ремастер, planы 72-74)
- `projects/game-legacy-php/`          — legacy PHP реализация oxsar2 (clean-room rewrite, на удаление после готовности нового origin-фронта; план 75)
- `projects/portal/frontend/`          — портал oxsar-nova.ru (Vite/React)
- `projects/portal/backend/` (plan 36) — portal API (новости, предложения)
- `projects/identity/backend/` (plan 36, переименован в плане 51) — identity-service (JWT/JWKS, OAuth, users, global credits)
- `migrations/`               — goose SQL-миграции
- `configs/`                  — YAML-справочники (источник истины для юнитов/зданий)
- `projects/game-nova/api/openapi.yaml`          — источник истины для REST-контрактов

## Правила кода (обязательно)

Подробно — в §17 [oxsar-spec.txt](docs/oxsar-spec.txt). Ключевое:

### Go

- Go 1.23+, `gofmt`, `goimports`, `golangci-lint` (конфиг в `.golangci.yml`).
- `ctx context.Context` — первый параметр у всех IO/service-методов.
- Ошибки заворачиваются через `fmt.Errorf("context: %w", err)`; между слоями —
  типизированные sentinel-ошибки (`battle.ErrInvalidInput`).
- Логирование: `log/slog` с полями `user_id`, `planet_id`, `event_id`, `trace_id`.
- БД: pgx напрямую (`db.Pool().Exec`, `QueryRow`, `Query`) с
  параметризованными запросами; SQL пишется inline в handler/service,
  без code-generation. Транзакции — `repo.InTx(ctx, fn)` или
  `pgx.BeginTx(...)`. (Замечание 2026-04-28: ранее CLAUDE.md упоминал
  sqlc, но фактически он не используется — `sqlc.yaml`, директории
  `queries/`, импортов `sqlc` нет в репозитории. Если в будущем
  введём sqlc — это будет отдельный план миграции.)
- Запрещено: `init()` с побочными эффектами, глобальные изменяемые переменные
  (кроме конфига, загруженного при старте), `panic` в прод-коде (кроме bootstrap),
  `any`/`interface{}` в публичных API без причины.
- Контекст: любой внешний IO имеет deadline. `context.Background()` — только
  в main и воркерах.
- Конкурентность: `errgroup` + `context`; мьютексы — только для явного shared
  state; goroutine-ы имеют понятный конец жизни.

### TypeScript

- `tsconfig`: strict, `noUncheckedIndexedAccess`, `exactOptionalPropertyTypes`.
- ESLint + Prettier, `jsx-a11y`, `eslint-plugin-import`.
- Запрещено: `any`, `@ts-ignore` (только `@ts-expect-error` с описанием),
  default exports (кроме route-модулей), `console.log` в прод-коде.
- Сервер-стейт — TanStack Query, UI-стейт — Zustand.
- DTO — только из сгенерированного OpenAPI-клиента, ручные типы запрещены.

### Коммиты и PR

- Conventional Commits: `feat:`, `fix:`, `refactor:`, `test:`, `docs:`, `chore:`, …
- Trunk-based, short-lived branches, rebase, squash-merge.
- PR ≤ 400 строк diff; большие фичи — серия PR за feature-флагом.
- Минимум 1 approve (2 — для `battle`, `economy`, `auth`, `market`).
- CI должен быть зелёным: lint, test, openapi-check, security.

## Балансировочные формулы

Формулы производства, стоимостей, времени, боя берутся **один-в-один** из oxsar2
и oxsar2-java. Менять баланс без явного ADR и согласования с геймдизайном —
запрещено. Что можно менять сразу — перечислено в §18 ТЗ.

## Тестирование

- Покрытие изменённых строк: ≥ 70% для обычного кода, ≥ 85% для
  `battle`/`economy`/`event`.
- Бой: property-based (rapid) + golden-файлы в `projects/game-nova/backend/internal/battle/testdata/*.json` +
  сравнение с `oxsar2-java/assault/dist/oxsar2-java.jar` (см. §14.4 ТЗ).
- Флаки-тесты запрещены. Нестабильный тест — карантин + тикет.

## Ссылки

- [oxsar-spec.txt](docs/oxsar-spec.txt) — полное ТЗ.
- [docs/status.md](docs/status.md) — матрица готовности модулей.
- [docs/release-roadmap.md](docs/release-roadmap.md) — приоритеты до запуска в прод.
- [docs/simplifications.md](docs/simplifications.md) — tracker всех
  принятых упрощений с планом возврата. **Любой новый trade-off
  записывается сюда в момент принятия, не "потом".**
- [docs/balance/audit.md](docs/balance/audit.md) — реестр найденных
  игровых дыр и эксплойтов с расчётами. Новая дыра → сначала запись
  в audit, потом план исправления.
- [docs/balance/analysis.md](docs/balance/analysis.md) — разбор
  балансных формул (производство, стоимости, бой).
- [docs/plans/](docs/plans/) — планы итераций (нумерованные).
- [docs/adr/](docs/adr/) — архитектурные решения.
- [docs/ui/](docs/ui/) — UI-спеки, сравнение с OGame, матрица e2e, dev-log.
- [docs/ops/](docs/ops/) — эксплуатация: VPS-sizing, платежи, runbooks, паттерны.
- [docs/ops/event-audit-pattern.md](docs/ops/event-audit-pattern.md) — почему не удалять записи events при отменах.
- [docs/ops/release-process.md](docs/ops/release-process.md) — релиз-процесс (теги, GHCR, откат, хотфиксы).
- [docs/legacy/](docs/legacy/) — доступ к oxsar2 для сверки.
- [projects/game-nova/api/openapi.yaml](projects/game-nova/api/openapi.yaml) — контракт REST API.
