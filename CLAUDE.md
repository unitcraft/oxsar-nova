# CLAUDE.md

Руководство для разработчиков (людей и AI-ассистентов) по работе с кодовой базой
oxsar-nova. Цель — чтобы новый участник мог продуктивно войти в проект за час.

## Что это такое

Порт legacy-игры oxsar2 (PHP/Yii 1.1 + MySQL) на современный стек: Go + TypeScript +
PostgreSQL + Redis. Боевой движок — порт отдельного java-проекта
`d:\Sources\oxsar2-java` на Go. Полное ТЗ — [oxsar-spec.txt](oxsar-spec.txt).

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

- `backend/cmd/server`   — HTTP/WS entry point
- `backend/cmd/worker`   — фоновый обработчик events
- `backend/cmd/tools`    — CLI-утилиты (reseed, ресинк артефактов)
- `backend/internal/*`   — домены (auth, planet, fleet, battle, …)
- `backend/pkg/*`        — общие утилиты (rng, proto, ids)
- `frontend/src/api`     — сгенерированный клиент из OpenAPI
- `frontend/src/features/<domain>` — вертикальные срезы UI
- `migrations/`          — goose SQL-миграции
- `configs/`             — YAML-справочники (источник истины для юнитов/зданий)
- `api/openapi.yaml`     — источник истины для REST-контрактов

## Правила кода (обязательно)

Подробно — в §17 [oxsar-spec.txt](oxsar-spec.txt). Ключевое:

### Go

- Go 1.23+, `gofmt`, `goimports`, `golangci-lint` (конфиг в `.golangci.yml`).
- `ctx context.Context` — первый параметр у всех IO/service-методов.
- Ошибки заворачиваются через `fmt.Errorf("context: %w", err)`; между слоями —
  типизированные sentinel-ошибки (`battle.ErrInvalidInput`).
- Логирование: `log/slog` с полями `user_id`, `planet_id`, `event_id`, `trace_id`.
- БД: только через `sqlc`-сгенерированные методы + сырой SQL в `backend/queries/`
  для сложной агрегации. Транзакции — `repo.InTx(ctx, fn)`.
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
- Бой: property-based (rapid) + golden-файлы в `testdata/battle/*.json` +
  сравнение с `oxsar2-java/assault/dist/oxsar2-java.jar` (см. §14.4 ТЗ).
- Флаки-тесты запрещены. Нестабильный тест — карантин + тикет.

## Ссылки

- [oxsar-spec.txt](oxsar-spec.txt) — полное ТЗ.
- [docs/status.md](docs/status.md) — матрица готовности модулей.
- [docs/adr/](docs/adr/) — архитектурные решения.
- [api/openapi.yaml](api/openapi.yaml) — контракт REST API.
