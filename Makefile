.DEFAULT_GOAL := help
SHELL := /bin/bash

ROOT := $(shell pwd)
BACKEND := $(ROOT)/projects/game-nova/backend
FRONTEND := $(ROOT)/projects/game-nova/frontends/nova
PORTAL_FRONTEND := $(ROOT)/projects/portal/frontend
MIGRATIONS := $(ROOT)/projects/game-nova/migrations

GOOSE_DRIVER := postgres
GOOSE_DBSTRING ?= postgres://oxsar:oxsar@localhost:5432/oxsar?sslmode=disable

# Путь к дампу legacy-таблиц oxsar2 (для import-datasheets → construction.yml).
# Переопределите если дамп лежит в другом месте.
OXSAR2_DUMP ?= d:/Sources/oxsar2/sql/table_dump

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  dev-up          - start postgres + redis via docker-compose"
	@echo "  dev-down        - stop local infra"
	@echo "  migrate-up      - apply all pending migrations"
	@echo "  migrate-down    - roll back one migration"
	@echo "  backend-run     - run HTTP/WS server"
	@echo "  worker-run      - run event-loop worker"
	@echo "  frontend-run         - run game vite dev server"
	@echo "  portal-frontend-run  - run portal vite dev server"
	@echo "  test            - run all tests"
	@echo "  lint            - run all linters"
	@echo "  gen             - regenerate sqlc + openapi clients"
	@echo "  import-datasheets - generate projects/game-nova/configs/construction.yml from legacy dump (OXSAR2_DUMP)"
	@echo "  i18n-audit        - scan codebase for hardcoded Cyrillic strings, write report"
	@echo "  i18n-rename       - rename i18n keys SCREAMING_SNAKE→lowerCamelCase (one-time, point of no return)"
	@echo "  i18n-check        - run all i18n CI checks (no-printf + consistency ru↔en)"

.PHONY: dev-up
dev-up:
	docker compose -f deploy/docker-compose.yml up -d

.PHONY: dev-down
dev-down:
	docker compose -f deploy/docker-compose.yml down

.PHONY: migrate-up
migrate-up:
	cd $(MIGRATIONS) && GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose up

.PHONY: migrate-down
migrate-down:
	cd $(MIGRATIONS) && GOOSE_DRIVER=$(GOOSE_DRIVER) GOOSE_DBSTRING="$(GOOSE_DBSTRING)" goose down

.PHONY: backend-run
backend-run:
	cd $(BACKEND) && go run ./cmd/server

.PHONY: worker-run
worker-run:
	cd $(BACKEND) && go run ./cmd/worker

.PHONY: frontend-run
frontend-run:
	cd $(FRONTEND) && npm run dev

.PHONY: portal-frontend-run
portal-frontend-run:
	cd $(PORTAL_FRONTEND) && npm run dev

# testseed — детерминированные данные для E2E (5 игроков, альянс, сообщения).
# Использует DB_URL из окружения (или GOOSE_DBSTRING как fallback).
# Флаг --reset сначала чистит игровые таблицы.
.PHONY: test-seed
test-seed:
	cd $(BACKEND) && DB_URL="$${DB_URL:-$(GOOSE_DBSTRING)}" go run ./cmd/tools/testseed --reset

# E2E — Playwright, поднимает backend+frontend и прогоняет спеки из frontend/e2e.
.PHONY: test-e2e
test-e2e:
	cd $(FRONTEND) && npm run test:e2e

# E2E в Docker — полный стек (pg+redis+migrate+backend+worker+testseed+frontend+playwright).
# Используется в CI и локально, когда нужен воспроизводимый прогон без
# ручного запуска терминалов.
.PHONY: test-e2e-docker
test-e2e-docker:
	# Не используем --exit-code-from: он неявно включает
	# --abort-on-container-exit и валит всё при успешном exit одноразового
	# сервиса (migrate/testseed). Подход: поднимаем стек в detached-режиме,
	# логи рядом, ждём playwright отдельной командой, забираем его код.
	COMPOSE_BAKE=true docker compose -f deploy/docker-compose.e2e.yml up -d --build
	docker compose -f deploy/docker-compose.e2e.yml logs -f --no-log-prefix playwright & \
		LOGS=$$!; \
		STATUS=$$(docker wait deploy-playwright-1 2>/dev/null || echo 1); \
		kill $$LOGS 2>/dev/null; \
		docker compose -f deploy/docker-compose.e2e.yml down -v --remove-orphans; \
		exit $$STATUS

# Очистка E2E-стека (останавливает все контейнеры, удаляет анонимные
# volumes — TMPFS для pg всё равно не персистится).
.PHONY: test-e2e-docker-down
test-e2e-docker-down:
	docker compose -f deploy/docker-compose.e2e.yml down -v --remove-orphans

# ui-preview — ручной осмотр UI на e2e-стеке с пробросом портов.
# Поднимает всё (kroмe playwright) с mock-платежами и детерминированными
# seed-данными, UI доступен на http://localhost:5173, API на :8081.
# Тестовые логины: admin/alice/bob/eve/charlie, пароль DevPass123.
.PHONY: ui-preview
ui-preview:
	docker compose \
		-f deploy/docker-compose.e2e.yml \
		-f deploy/docker-compose.e2e.ports.yml \
		up -d --build \
		postgres redis migrate backend worker testseed frontend
	@echo ""
	@echo "  UI:      http://localhost:5173"
	@echo "  API:     http://localhost:8081"
	@echo "  логины:  admin / alice / bob / eve / charlie"
	@echo "  пароль:  DevPass123"
	@echo ""
	@echo "  Остановить: make ui-preview-down"

.PHONY: ui-preview-down
ui-preview-down:
	docker compose \
		-f deploy/docker-compose.e2e.yml \
		-f deploy/docker-compose.e2e.ports.yml \
		down -v --remove-orphans

# Аудит покрытия API → UI (все эндпоинты OpenAPI должны вызываться из фронта).
.PHONY: api-coverage
api-coverage:
	cd $(FRONTEND) && npm run api:coverage

.PHONY: backend-test
backend-test:
	cd $(BACKEND) && go test ./...

.PHONY: frontend-test
frontend-test:
	cd $(FRONTEND) && npm test

.PHONY: test
test: backend-test frontend-test

.PHONY: backend-lint
backend-lint:
	cd $(BACKEND) && golangci-lint run ./...

.PHONY: frontend-lint
frontend-lint:
	cd $(FRONTEND) && npm run lint

.PHONY: lint
lint: backend-lint frontend-lint check-duplicates

# check-duplicates: сверяет содержимое DUPLICATE-файлов между Go-модулями
# (game-nova/identity/portal/billing). При drift'е — печатает unified-diff
# и exit 1. План 85.
.PHONY: check-duplicates
check-duplicates:
	cd $(BACKEND) && go run ./cmd/tools/check-duplicates -root="$(ROOT)"

# sync-duplicates: тиражирует тело эталона (первый путь в DUPLICATE-маркере)
# по копиям с заменой per-module import-prefix. Использовать вручную после
# сознательной правки эталона. Не запускать в CI. План 85.
.PHONY: sync-duplicates
sync-duplicates:
	cd $(BACKEND) && go run ./cmd/tools/check-duplicates -fix -root="$(ROOT)"

.PHONY: gen
gen:
	cd $(BACKEND) && go generate ./...
	cd $(FRONTEND) && npm run gen:api

# import-datasheets генерирует projects/game-nova/configs/construction.yml из legacy SQL-дампа.
# Запускать вручную после получения дампа или при обновлении баланса.
# Требует: OXSAR2_DUMP указывает на каталог с na_construction.sql и др.
.PHONY: import-datasheets
import-datasheets:
	cd $(BACKEND) && go run ./cmd/tools/import-datasheets \
		--input="$(OXSAR2_DUMP)" \
		--output="$(ROOT)/projects/game-nova/configs"

# i18n-audit: сканирует frontend/src и backend на хардкод-кириллицу.
# Выход: docs/plans/33-i18n-audit-report.md
.PHONY: i18n-audit
i18n-audit:
	cd $(BACKEND) && go run ./cmd/tools/i18n-audit \
		--root="$(ROOT)" \
		--dict="$(ROOT)/projects/game-nova/configs/i18n/ru.yml" \
		--out="$(ROOT)/docs/plans/33-i18n-audit-report.md"

# i18n-rename: переименовывает ключи/группы в projects/game-nova/configs/i18n/*.yml
# SCREAMING_SNAKE_CASE → lowerCamelCase и %s/%d → {{name}}.
# Запускать ОДИН РАЗ, после — закоммитить и поднять LOCALE_VERSION.
.PHONY: i18n-rename
i18n-rename:
	cd $(BACKEND) && go run ./cmd/tools/i18n-rename \
		--dir="$(ROOT)/projects/game-nova/configs/i18n" \
		--glossary="$(ROOT)/docs/plans/33-i18n-placeholder-glossary.yml" \
		--map-out="$(ROOT)/projects/game-nova/configs/i18n/i18n-rename-map.json"

# i18n-check: CI-проверки i18n (нет %s/%d в YAML, консистентность ключей ru↔en).
.PHONY: i18n-check
i18n-check:
	cd $(BACKEND) && go test ./internal/i18n/... -run 'TestNoPrintf|TestI18nConsistency' -v

# wiki-gen: генерирует docs/wiki/ru/{buildings,ships,defense,research}/*.md
# из configs/. План 19 (game-wiki). Запускать после каждого изменения
# балансовых YAML.
.PHONY: wiki-gen
wiki-gen:
	cd $(BACKEND) && go run ./cmd/tools/wiki-gen \
		--configs="$(ROOT)/projects/game-nova/configs" \
		--out="$(ROOT)/docs/wiki/ru"
