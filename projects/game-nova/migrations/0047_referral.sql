-- +goose Up
ALTER TABLE users ADD COLUMN referred_by uuid REFERENCES users(id) ON DELETE SET NULL;
CREATE INDEX idx_users_referred_by ON users(referred_by) WHERE referred_by IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_users_referred_by;
ALTER TABLE users DROP COLUMN IF EXISTS referred_by;
