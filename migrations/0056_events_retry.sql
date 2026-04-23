-- +goose Up
-- Поля для retry-механики и диагностики.
ALTER TABLE events ADD COLUMN IF NOT EXISTS attempt      integer NOT NULL DEFAULT 0;
ALTER TABLE events ADD COLUMN IF NOT EXISTS last_error   text;
ALTER TABLE events ADD COLUMN IF NOT EXISTS next_retry_at timestamptz;

-- Tie-breaker: стабильный порядок при одинаковых fire_at.
CREATE INDEX IF NOT EXISTS ix_events_fire_at_id
    ON events(fire_at, id) WHERE state = 'wait';

-- +goose Down
DROP INDEX IF EXISTS ix_events_fire_at_id;
ALTER TABLE events DROP COLUMN IF EXISTS next_retry_at;
ALTER TABLE events DROP COLUMN IF EXISTS last_error;
ALTER TABLE events DROP COLUMN IF EXISTS attempt;
