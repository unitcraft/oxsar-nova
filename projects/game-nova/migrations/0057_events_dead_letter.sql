-- +goose Up
-- Dead-letter таблица: хранилище для error-events старше порога.
-- Схема повторяет events + failed_at. Cron переносит сюда строки из
-- events, где state='error' AND processed_at < now() - 7 days.
CREATE TABLE events_dead (
    id            uuid PRIMARY KEY,
    user_id       uuid,
    planet_id     uuid,
    kind          integer NOT NULL,
    fire_at       timestamptz NOT NULL,
    payload       jsonb NOT NULL,
    created_at    timestamptz NOT NULL,
    processed_at  timestamptz,
    attempt       integer NOT NULL DEFAULT 0,
    last_error    text,
    failed_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ix_events_dead_kind ON events_dead(kind, failed_at DESC);
CREATE INDEX ix_events_dead_user ON events_dead(user_id, failed_at DESC);

-- +goose Down
DROP TABLE IF EXISTS events_dead;
