-- +goose Up
-- Базовые расширения и типы.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "citext";

-- +goose StatementBegin
CREATE TYPE event_state AS ENUM ('wait', 'start', 'ok', 'error');
CREATE TYPE queue_status AS ENUM ('queued', 'running', 'done', 'cancelled');
CREATE TYPE user_role AS ENUM ('player', 'support', 'admin', 'superadmin');
-- +goose StatementEnd

CREATE TABLE users (
    id              uuid PRIMARY KEY,
    username        citext NOT NULL UNIQUE,
    email           citext NOT NULL UNIQUE,
    password_hash   text   NOT NULL,
    role            user_role NOT NULL DEFAULT 'player',
    language        text   NOT NULL DEFAULT 'ru',
    timezone        text   NOT NULL DEFAULT 'UTC',
    cur_planet_id   uuid,
    points          numeric(20, 4) NOT NULL DEFAULT 0,
    u_points        numeric(20, 4) NOT NULL DEFAULT 0,
    r_points        numeric(20, 4) NOT NULL DEFAULT 0,
    b_points        numeric(20, 4) NOT NULL DEFAULT 0,
    a_points        numeric(20, 4) NOT NULL DEFAULT 0,
    e_points        numeric(20, 4) NOT NULL DEFAULT 0,
    battles         integer NOT NULL DEFAULT 0,
    credit          numeric(15, 2) NOT NULL DEFAULT 5.00,
    research_factor real NOT NULL DEFAULT 1,
    ipcheck         boolean NOT NULL DEFAULT false,
    umode           boolean NOT NULL DEFAULT false,
    umode_until     timestamptz,
    tutorial_state  integer NOT NULL DEFAULT 0,
    regtime         timestamptz NOT NULL DEFAULT now(),
    last_seen       timestamptz NOT NULL DEFAULT now(),
    created_at      timestamptz NOT NULL DEFAULT now(),
    deleted_at      timestamptz
);

CREATE INDEX ix_users_last_seen ON users(last_seen);
CREATE INDEX ix_users_points    ON users(points DESC);

-- +goose Down
DROP TABLE IF EXISTS users;

-- +goose StatementBegin
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS queue_status;
DROP TYPE IF EXISTS event_state;
-- +goose StatementEnd
