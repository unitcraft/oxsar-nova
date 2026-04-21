-- +goose Up
-- Поддержка ремонта оборонительных установок: damaged_count и shell_percent
-- хранятся аналогично таблице ships.
ALTER TABLE defense
    ADD COLUMN damaged_count bigint NOT NULL DEFAULT 0,
    ADD COLUMN shell_percent double precision NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE defense
    DROP COLUMN damaged_count,
    DROP COLUMN shell_percent;
