-- План 36 Ф.12: handoff-flow.
-- После переноса аутентификации в auth-service, в game-nova users password_hash
-- больше не источник истины. Юзеры, созданные через handoff (RSA-токен от
-- auth-service), не имеют пароля в game-db — он живёт в auth-db.
-- Старые юзеры (созданные через /api/auth/register до Ф.11) сохраняют свой хеш.

-- +goose Up
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;

-- +goose Down
-- Откат восстановит NOT NULL только если в таблице нет NULL-ов.
-- Если откат нужен — сначала придётся либо синхронизировать пароли из auth-db,
-- либо удалить юзеров с password_hash=NULL.
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
