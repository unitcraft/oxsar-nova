-- Добавляем global_user_id для связки с auth-service (plan-36).
-- UUID пользователя из auth_db.users.id

ALTER TABLE `na_users`
  ADD COLUMN `global_user_id` VARCHAR(36) NULL DEFAULT NULL
    COMMENT 'UUID из auth-service (plan-36)'
    AFTER `userid`,
  ADD UNIQUE INDEX `uq_global_user_id` (`global_user_id`);
