-- +goose Up
ALTER TABLE users ADD COLUMN vacation_since TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN vacation_last_end TIMESTAMPTZ;

-- +goose Down
ALTER TABLE users DROP COLUMN vacation_since;
ALTER TABLE users DROP COLUMN vacation_last_end;
