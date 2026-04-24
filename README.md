# oxsar-nova

Браузерная MMO-стратегия в стиле OGame. Современная реализация классического oxsar2
на Go (backend) и TypeScript/React (frontend).

Полное техническое задание: [oxsar-spec.txt](oxsar-spec.txt).

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

## Структура репозитория

```
backend/       Go-сервис (API + воркер + CLI-утилиты)
frontend/      React + TypeScript SPA
migrations/    goose SQL-миграции (append-only)
configs/       YAML-справочники (здания, корабли, rapidfire, артефакты)
api/           OpenAPI 3.1 спецификация (источник истины для REST)
testdata/      фикстуры для тестов (бой, экономика)
docs/          документация, ADR
deploy/        docker-образы, k8s-манифесты
```

## Контакты и документация

- [CLAUDE.md](CLAUDE.md) — как работать с репозиторием для разработчиков и AI-ассистентов.
- [docs/adr/](docs/adr/) — принятые архитектурные решения.
- [oxsar-spec.txt](oxsar-spec.txt) — полное ТЗ проекта.

## Лицензия

Код распространяется под [PolyForm Noncommercial 1.0.0](LICENSE):
бесплатное использование в личных, исследовательских и образовательных
целях. Любое коммерческое использование требует отдельной лицензии от
автора — см. [COMMERCIAL-LICENSE.md](COMMERCIAL-LICENSE.md).

Весь код в репозитории написан Evgeniy Golovin совместно с Claude Code.

### Участие в разработке

Pull request'ы приветствуются. Отправляя PR, вы соглашаетесь с тем, что
ваш вклад публикуется на условиях той же PolyForm Noncommercial 1.0.0 и
может быть включён в коммерческие релизы автора. Для сколько-нибудь
объёмных вкладов планируется подключить формальный CLA
(см. [cla-assistant.io](https://cla-assistant.io/)) — до этого момента
крупные изменения лучше согласовывать заранее через issue.
