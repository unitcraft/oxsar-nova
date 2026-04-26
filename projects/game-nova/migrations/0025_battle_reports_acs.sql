-- +goose Up
-- Add acs_participants jsonb to battle_reports.
-- Format: [{"user_id":"...","fleet_id":"..."}] — all ACS attacker fleets.
-- Empty array or NULL for solo attacks.
ALTER TABLE battle_reports ADD COLUMN acs_participants jsonb;

-- +goose Down
ALTER TABLE battle_reports DROP COLUMN IF EXISTS acs_participants;
