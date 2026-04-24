.DEFAULT_GOAL := help
SHELL := /bin/bash

ROOT := $(shell pwd)
BACKEND := $(ROOT)/backend
FRONTEND := $(ROOT)/frontend
MIGRATIONS := $(ROOT)/migrations

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
	@echo "  frontend-run    - run vite dev server"
	@echo "  test            - run all tests"
	@echo "  lint            - run all linters"
	@echo "  gen             - regenerate sqlc + openapi clients"
	@echo "  import-datasheets - generate configs/construction.yml from legacy dump (OXSAR2_DUMP)"

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
# Тестовые логины: admin/alice/bob/eve/charlie, пароль test-password-123.
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
	@echo "  пароль:  test-password-123"
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
lint: backend-lint frontend-lint

.PHONY: gen
gen:
	cd $(BACKEND) && go generate ./...
	cd $(FRONTEND) && npm run gen:api

# import-datasheets генерирует configs/construction.yml из legacy SQL-дампа.
# Запускать вручную после получения дампа или при обновлении баланса.
# Требует: OXSAR2_DUMP указывает на каталог с na_construction.sql и др.
.PHONY: import-datasheets
import-datasheets:
	cd $(BACKEND) && go run ./cmd/tools/import-datasheets \
		--input="$(OXSAR2_DUMP)" \
		--output="$(ROOT)/configs"
