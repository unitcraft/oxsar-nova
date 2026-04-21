.DEFAULT_GOAL := help
SHELL := /bin/bash

ROOT := $(shell pwd)
BACKEND := $(ROOT)/backend
FRONTEND := $(ROOT)/frontend
MIGRATIONS := $(ROOT)/migrations

GOOSE_DRIVER := postgres
GOOSE_DBSTRING ?= postgres://oxsar:oxsar@localhost:5432/oxsar?sslmode=disable

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
