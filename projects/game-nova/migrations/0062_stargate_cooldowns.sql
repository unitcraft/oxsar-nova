-- +goose Up
-- План 20 Ф.5: Stargate Jump.
-- Cooldown между прыжками: 3600 * 0.7^(jump_gate_level - 1) секунд.
-- Хранится last_jump_at на луне-источнике (по planet_id).
CREATE TABLE stargate_cooldowns (
    planet_id    uuid PRIMARY KEY REFERENCES planets(id) ON DELETE CASCADE,
    last_jump_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE stargate_cooldowns;
