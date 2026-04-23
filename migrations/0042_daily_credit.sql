-- +goose Up
-- Отслеживание последнего начисления ежедневного бонуса кредитов.
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_daily_credit_at timestamptz;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS last_daily_credit_at;
