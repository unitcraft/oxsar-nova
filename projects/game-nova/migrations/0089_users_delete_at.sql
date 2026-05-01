-- План 72.1.30: 7-day grace на удалении аккаунта (legacy
-- `Preferences::updateDeletion` ставит `users.delete = time() + 604800`).
--
-- Закрывает хвост P72.1.5.B.DELETION_NO_GRACE: вместо немедленного
-- soft-delete после ConfirmDeletion ставится `delete_at = now() + 7 days`,
-- юзер может отменить через POST /api/me/deletion/cancel в grace-period.
-- Физическое soft-delete выполняет event-handler по таймеру.

-- +goose Up

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS delete_at timestamptz;

-- Partial-индекс для воркера (поиск pending удалений).
CREATE INDEX IF NOT EXISTS users_delete_at_idx
    ON users (delete_at)
    WHERE delete_at IS NOT NULL AND deleted_at IS NULL;

-- +goose Down

DROP INDEX IF EXISTS users_delete_at_idx;
ALTER TABLE users DROP COLUMN IF EXISTS delete_at;
