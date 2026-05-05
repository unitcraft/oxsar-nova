-- +goose Up
-- Бонусные уровни исследований (от артефактов, суперкомпьютера и т.п.).
-- Legacy: research2user.added (NS::getAddedResearch).
ALTER TABLE research ADD COLUMN IF NOT EXISTS added integer NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE research DROP COLUMN IF EXISTS added;
