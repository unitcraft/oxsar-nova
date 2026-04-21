-- +goose Up
-- Добавляем статус к отношениям: pending (предложение ждёт подтверждения)
-- или active (обе стороны согласились). WAR активно сразу.
ALTER TABLE alliance_relationships
    ADD COLUMN status TEXT NOT NULL DEFAULT 'active'
        CHECK (status IN ('pending', 'active'));

-- +goose Down
ALTER TABLE alliance_relationships DROP COLUMN status;
