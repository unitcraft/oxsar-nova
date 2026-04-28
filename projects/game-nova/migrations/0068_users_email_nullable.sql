-- План 36 Nice-10: email больше не источник истины в game-db.
-- Источник — identity-db (identity-service). Lazy-create в EnsureUserMiddleware пишет
-- NULL (email удалён из RSA-claims как PII).
--
-- UNIQUE constraint оставляем — он просто не действует на NULL-ы (PostgreSQL
-- считает NULL'ы разными).
--
-- Старые юзеры (созданные через legacy /api/auth/register) сохраняют свой email.

-- +goose Up
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

-- +goose Down
-- Откат восстановит NOT NULL только если в таблице нет NULL-ов.
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
