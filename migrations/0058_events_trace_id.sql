-- +goose Up
ALTER TABLE events ADD COLUMN IF NOT EXISTS trace_id uuid;
CREATE INDEX IF NOT EXISTS ix_events_trace_id ON events(trace_id) WHERE trace_id IS NOT NULL;

ALTER TABLE events_dead ADD COLUMN IF NOT EXISTS trace_id uuid;

-- +goose Down
DROP INDEX IF EXISTS ix_events_trace_id;
ALTER TABLE events_dead DROP COLUMN IF EXISTS trace_id;
ALTER TABLE events DROP COLUMN IF EXISTS trace_id;
