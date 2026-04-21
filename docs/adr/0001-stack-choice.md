# ADR-0001: Выбор технологического стека

- Status: Accepted
- Date: 2026-04-21

## Context

Legacy-проект oxsar2 написан на PHP/Yii 1.1 + MySQL + Memcache. Yii 1.x снят
с поддержки, миграций нет, тестов нет. Нужно переписать с нуля.

## Decision

- Backend — Go 1.23+ (конкурентность, статическая типизация, одиночный
  бинарник, стандартная библиотека закрывает 80% нужд).
- БД — PostgreSQL 16 (JSONB, CITEXT, range types, лучше блокировки).
- Кеш/pub-sub — Redis 7.
- Frontend — React 18 + TypeScript + Vite.
- Контракты — OpenAPI 3.1 + JSON Schema (кодогенерация обязательна).

## Consequences

- Разработка требует знаний Go у команды.
- SQL-миграции (goose) вместо unconstrained dump-файлов.
- Нет ORM с магией — только sqlc.
- Общая типизация бэк ↔ фронт через OpenAPI-codegen.
