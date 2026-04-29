# oxsar-nova

Браузерная MMO-стратегия в стиле OGame. Современная реализация классического oxsar2
на Go (backend) и TypeScript/React (frontend).

Полное техническое задание: [docs/oxsar-spec.txt](docs/oxsar-spec.txt).

## Статус

Проект находится на ранней стадии — собран каркас M0 и частично M1 из roadmap ТЗ
(см. §16). Подробная разбивка по модулям — в [docs/status.md](docs/status.md).

## Быстрый старт

Требования: Docker, Docker Compose, Go 1.23+, Node 20+, make.

```bash
make dev-up                  # поднять postgres + redis
cd backend && go mod download  # первый запуск: скачать Go-зависимости
cd frontend && npm install     # первый запуск: скачать npm-зависимости
make migrate-up              # применить миграции БД

make backend-run             # запустить API-сервер
make worker-run              # event-loop воркер (отдельный терминал)
make frontend-run            # vite dev-сервер
```

API по умолчанию слушает `:8080`, фронт — `:5173`.

Другие compose-стеки (e2e, preview) — [docs/ops/runbooks/docker-stacks.md](docs/ops/runbooks/docker-stacks.md).

## Структура репозитория

```
backend/       Go-сервис (API + воркер + CLI-утилиты)
frontend/      React + TypeScript SPA
migrations/    goose SQL-миграции (append-only)
configs/       YAML-справочники (здания, корабли, rapidfire, артефакты)
api/           OpenAPI 3.1 спецификация (источник истины для REST)
testdata/      фикстуры для тестов (бой, экономика)
docs/          документация, ADR, планы
deploy/        docker-образы, k8s-манифесты
```

## Навигация по документации

### Обязательное к чтению

- [CLAUDE.md](CLAUDE.md) — как работать с репозиторием (для разработчиков и AI-ассистентов).
- [docs/oxsar-spec.txt](docs/oxsar-spec.txt) — полное ТЗ.
- [docs/status.md](docs/status.md) — матрица готовности модулей.
- [docs/release-roadmap.md](docs/release-roadmap.md) — приоритетный список задач до запуска в прод.
- [docs/simplifications.md](docs/simplifications.md) — реестр всех принятых упрощений (trade-offs).

### Архитектура и планы

- [docs/adr/](docs/adr/) — принятые архитектурные решения.
- [docs/plans/](docs/plans/) — планы итераций (нумерованные, 01–24).
- [docs/project-creation.txt](docs/project-creation.txt) — дневник итераций.
- [docs/design/](docs/design/) — мокапы UI.

### Баланс

- [docs/balance/analysis.md](docs/balance/analysis.md) — разбор формул (производство, стоимости, бой).
- [docs/balance/audit.md](docs/balance/audit.md) — реестр игровых дыр и эксплойтов с расчётами.

### UI

- [docs/ui/design-spec.md](docs/ui/design-spec.md) — ТЗ на дизайн UI.
- [docs/ui/ogame-comparison.md](docs/ui/ogame-comparison.md) — сравнение с OGame UI.
- [docs/ui/test-matrix.md](docs/ui/test-matrix.md) — матрица e2e-тестов.
- [docs/ui/dev-log.md](docs/ui/dev-log.md) — дневник UI-доработок.

### Эксплуатация

- [docs/ops/vps-sizing.md](docs/ops/vps-sizing.md) — VPS-конфигурации под разный DAU.
- [docs/ops/payment-integration.md](docs/ops/payment-integration.md) — подключение платёжных шлюзов.
- [docs/ops/runbooks/](docs/ops/runbooks/) — операционные runbooks.

### Справочное

- [api/openapi.yaml](api/openapi.yaml) — контракт REST API.
- [docs/code-stats.md](docs/code-stats.md) — статистика кода.
- [docs/legacy/game-reference.md](docs/legacy/game-reference.md) — доступ к legacy oxsar2 для сверки.

## Лицензия

Код распространяется под [PolyForm Noncommercial 1.0.0](LICENSE):
бесплатное использование в личных, исследовательских и образовательных
целях. Любое коммерческое использование требует отдельной лицензии от
автора — см. [COMMERCIAL-LICENSE.md](COMMERCIAL-LICENSE.md).

Применимое право и юрисдикция для коммерческих лицензий и CLA — Российская
Федерация (см. [COMMERCIAL-LICENSE.md](COMMERCIAL-LICENSE.md) §Governing Law,
[CLA.md](CLA.md) §8).

Автор кода — Evgeniy Golovin. При разработке использовался AI-ассистент
Claude Code как технический инструмент; творческий отбор, проверка,
интеграция и компоновка кода выполнены автором. Подробности правового
статуса — [docs/origin-rights.md](docs/origin-rights.md).

### Участие в разработке

Pull request'ы приветствуются — процесс описан в
[CONTRIBUTING.md](CONTRIBUTING.md). Перед тем как ваш PR будет принят,
нужно подписать [Contributor License Agreement](CLA.md): это делается
одной строкой в комментарии через бот
[cla-assistant.io](https://cla-assistant.io/), и нужно один раз на всю
историю ваших контрибьюций. CLA требуется, чтобы автор мог включать
ваш вклад и в публичные noncommercial-релизы, и в коммерческие
лицензии.

test CLA
