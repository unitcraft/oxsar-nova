-- +goose Up
-- План 72.1 ч.20.11: симулятор сохраняет результаты в battle_reports
-- с флагом is_simulation=true.
--
-- attacker_user_id/defender_user_id уже NULLable (миграция 0009).
-- Добавляем is_simulation для фильтрации.

ALTER TABLE battle_reports ADD COLUMN IF NOT EXISTS is_simulation boolean NOT NULL DEFAULT false;
CREATE INDEX IF NOT EXISTS ix_battle_reports_is_simulation ON battle_reports(is_simulation, at DESC);

-- +goose Down
DROP INDEX IF EXISTS ix_battle_reports_is_simulation;
ALTER TABLE battle_reports DROP COLUMN IF EXISTS is_simulation;
