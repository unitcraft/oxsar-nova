-- План 69 Ф.1: расширение domain-полей users для ремастера.
--
-- Дельта-аудит (Ф.0) сократил исходные 9 полей до 5 — остальные
-- уже закрыты другими миграциями (0046, 0050, 0051) или отказаны
-- (D-007 ui_theme YAGNI, D-019 home_planet_id R10/YAGNI, D-021
-- race мёртвое поле). См. docs/plans/69-...md «Ф.0».
--
-- Все поля nullable / с безопасным default — существующие fixtures
-- uni01/uni02 не миграцируются (R15).

-- +goose Up

-- D-001: исторический пик points (категории dm/be/of_points — отказ, YAGNI).
ALTER TABLE users ADD COLUMN IF NOT EXISTS max_points numeric(20, 4) NOT NULL DEFAULT 0;

-- D-004: защита новичков от атак (проверка в attack-handler — Ф.4).
ALTER TABLE users ADD COLUMN IF NOT EXISTS protected_until_at timestamptz;

-- D-005: домен-флаг наблюдателя. НЕ RBAC-роль (роли мигрировали в
-- identity-сервис, миграция 0070_drop_users_role.sql). is_observer —
-- это per-universe game-state (фильтрация из highscore и т.п.).
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_observer boolean NOT NULL DEFAULT false;

-- D-016: cooldown смены home/teleport-планеты (механика Ф.5).
-- НЕ дублирует stargate_cooldowns (миграция 0062): stargate — это
-- прыжок флота между лунами per planet_id; teleport — действие user.
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_planet_teleport_at timestamptz;

-- D-020: маркеры прочтения чата (chat_language — отказ, есть users.language).
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_global_chat_read_at timestamptz;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_ally_chat_read_at   timestamptz;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS last_ally_chat_read_at;
ALTER TABLE users DROP COLUMN IF EXISTS last_global_chat_read_at;
ALTER TABLE users DROP COLUMN IF EXISTS last_planet_teleport_at;
ALTER TABLE users DROP COLUMN IF EXISTS is_observer;
ALTER TABLE users DROP COLUMN IF EXISTS protected_until_at;
ALTER TABLE users DROP COLUMN IF EXISTS max_points;
