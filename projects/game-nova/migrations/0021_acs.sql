-- +goose Up
-- ACS (Allied Combat System) — группировка нескольких флотов для атаки
-- одной цели. Все флоты в группе прибывают одновременно (по arrive_at
-- ведущего флота) и участвуют в одном бою.

CREATE TABLE IF NOT EXISTS acs_groups (
    id          uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    target_galaxy  int     NOT NULL,
    target_system  int     NOT NULL,
    target_position int    NOT NULL,
    target_is_moon boolean NOT NULL DEFAULT false,
    arrive_at   timestamptz NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

-- Каждый флот в группе регистрируется отдельной строкой.
ALTER TABLE fleets ADD COLUMN IF NOT EXISTS acs_group_id uuid REFERENCES acs_groups(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS fleets_acs_group_id_idx ON fleets(acs_group_id) WHERE acs_group_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS fleets_acs_group_id_idx;
ALTER TABLE fleets DROP COLUMN IF EXISTS acs_group_id;
DROP TABLE IF EXISTS acs_groups;
