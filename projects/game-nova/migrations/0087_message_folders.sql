-- План 72.1.17: legacy паритет message folders.
--
-- Legacy `Folder` таблица (config/consts.php:508-518) определяет 11
-- константных папок (1..11). Origin до сих пор использовал folder
-- произвольно: welcome/starter/inactivity/firstAttack клались в
-- folder=2, что в legacy = MSG_FOLDER_SENT (исходящие). Это
-- семантический баг.
--
-- Этот план:
-- 1. Создаёт справочник message_folders с seed legacy-констант.
-- 2. Вводит новую константу 12 = SYSTEM для наших служебных
--    шаблонов welcome/inactivity/firstAttack/account-delete-code,
--    которых в legacy нет (это расширение oxsar-nova).
-- 3. Бэкфилл: переводит существующие записи folder=2 (наши system)
--    в folder=12 (SYSTEM); для других значений ничего не меняем.
-- 4. Бэкфилл: alliance/transfer.go и settings/delete.go писали
--    folder=13 — переводим их в folder=12 (SYSTEM) тоже.
-- 5. ALTER messages.folder NOT NULL DEFAULT 1 (на всякий случай) +
--    FK на message_folders для целостности.

-- +goose Up

CREATE TABLE IF NOT EXISTS message_folders (
    folder_id     smallint PRIMARY KEY,
    label_key     text NOT NULL,           -- i18n-ключ группы msgFolder.<key>
    is_standard   boolean NOT NULL DEFAULT true,
    display_order smallint NOT NULL DEFAULT 0
);

-- Seed legacy MSG_FOLDER_* (config/consts.php:508-518)
-- + наша SYSTEM=12 для welcome/inactivity (в legacy этих msg нет).
INSERT INTO message_folders (folder_id, label_key, display_order) VALUES
    (1,  'inbox',        10),
    (2,  'sent',         20),
    (3,  'fleet',        30),
    (4,  'spy',          40),
    (5,  'battleReports', 50),
    (6,  'alliance',     60),
    (7,  'artefacts',    70),
    (8,  'credit',       80),
    (9,  'expedition',   90),
    (10, 'recycler',    100),
    (11, 'surveillance', 110),
    (12, 'system',      120)
ON CONFLICT (folder_id) DO NOTHING;

-- Бэкфилл: наши welcome/starter/inactivity/firstAttack писались в
-- folder=2, что в legacy = SENT. Переводим в folder=12 (SYSTEM).
-- Критерий: from_user_id IS NULL (системные сообщения без отправителя).
UPDATE messages
SET folder = 12
WHERE folder = 2 AND from_user_id IS NULL;

-- Бэкфилл: alliance.transfer и settings.delete писали folder=13
-- (отсутствует в legacy const). Переводим в folder=12 (SYSTEM).
UPDATE messages
SET folder = 12
WHERE folder = 13;

-- FK после бэкфилла, чтобы не падать на legacy-данных.
ALTER TABLE messages
    ADD CONSTRAINT messages_folder_fk
    FOREIGN KEY (folder) REFERENCES message_folders(folder_id);

CREATE INDEX IF NOT EXISTS ix_messages_folder
    ON messages (to_user_id, folder, created_at DESC)
    WHERE deleted_at IS NULL;

-- +goose Down

DROP INDEX IF EXISTS ix_messages_folder;
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_folder_fk;
DROP TABLE IF EXISTS message_folders;
