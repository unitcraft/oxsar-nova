-- План 72.1.54 (P72.S2.ALLIANCE_PREFS 1:1): legacy `updateAllyPrefs`
-- (Alliance.class.php:979) принимает 10 полей. Реализовано в nova:
-- open, description_external, description_internal, description_apply.
-- Здесь добавляются оставшиеся 6: logo, homepage, foundername,
-- show_member, show_homepage, memberlist_sort.
--
-- Поля nullable / с дефолтами — backfill автоматический (legacy при
-- регистрации тоже не требует их заполнения).
--
-- Ограничения по длине дублируют legacy: logo/homepage ≤ 128 символов
-- (legacy Str::length($logo) > 128); foundername ≤ MAX_CHARS_ALLY_NAME=64
-- (legacy `Core::getOptions()->get("MAX_CHARS_ALLY_NAME")`); memberlist_sort
-- — small int (legacy ENUM из 4-5 значений; интерпретация на FE).

-- +goose Up

ALTER TABLE alliances ADD COLUMN IF NOT EXISTS logo            text;
ALTER TABLE alliances ADD COLUMN IF NOT EXISTS homepage        text;
ALTER TABLE alliances ADD COLUMN IF NOT EXISTS foundername     text;
ALTER TABLE alliances ADD COLUMN IF NOT EXISTS show_member     boolean NOT NULL DEFAULT true;
ALTER TABLE alliances ADD COLUMN IF NOT EXISTS show_homepage   boolean NOT NULL DEFAULT true;
ALTER TABLE alliances ADD COLUMN IF NOT EXISTS memberlist_sort smallint NOT NULL DEFAULT 0;

-- Длина проверяется в backend service.go перед UPDATE; CHECK в БД
-- избыточен, потому что валидация исходит из legacy options
-- (`MAX_CHARS_ALLY_NAME`, etc.), а constraint в БД сложнее менять.

-- +goose Down

ALTER TABLE alliances DROP COLUMN IF EXISTS memberlist_sort;
ALTER TABLE alliances DROP COLUMN IF EXISTS show_homepage;
ALTER TABLE alliances DROP COLUMN IF EXISTS show_member;
ALTER TABLE alliances DROP COLUMN IF EXISTS foundername;
ALTER TABLE alliances DROP COLUMN IF EXISTS homepage;
ALTER TABLE alliances DROP COLUMN IF EXISTS logo;
