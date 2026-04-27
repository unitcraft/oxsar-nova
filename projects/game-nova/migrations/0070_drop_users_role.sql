-- План 52 (RBAC unification): identity-service владеет ролями, game-nova
-- больше не хранит users.role локально. Middleware admin.RequireRole
-- читает роли/permissions из JWT-claims (план 52 Ф.3).
--
-- До миграции существующие admin-юзеры должны быть провязаны через
-- identity user_roles (миграция 0006_migrate_user_roles в identity-БД).
-- ВАЖНО: эта миграция запускается ПОСЛЕ identity-миграций 0005/0006,
-- иначе в момент проката кто-то с ролью admin может потерять доступ.
--
-- Rollback: возвращаем колонку и enum (но данные при rollback'е
-- не восстанавливаются — придётся delegate к identity-service).

-- +goose Up
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS role;
DROP TYPE IF EXISTS user_role;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM ('player', 'support', 'admin', 'superadmin');
ALTER TABLE users ADD COLUMN role user_role NOT NULL DEFAULT 'player';
-- Внимание: данные ролей не восстанавливаются — после down-migration
-- все юзеры сбрасываются в 'player'. Восстановление через identity API
-- (planon 52 Ф.2 endpoints).
-- +goose StatementEnd
