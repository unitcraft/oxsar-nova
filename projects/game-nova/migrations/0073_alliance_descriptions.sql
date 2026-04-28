-- План 67 Ф.1: 3 описания альянса (D-041, U-015) + audit-метка передачи
-- лидерства (D-040, U-004 — поле для UI/истории; сам endpoint в Ф.3).
--
-- Все поля nullable: существующие записи uni01/uni02 не миграцируются
-- (R15 — без принудительной заливки). Legacy `alliances.description`
-- остаётся как есть; Ф.2 решит, оставлять ли его как
-- "description_external" по умолчанию или дропнуть.

-- +goose Up
ALTER TABLE alliances ADD COLUMN IF NOT EXISTS description_external text;
ALTER TABLE alliances ADD COLUMN IF NOT EXISTS description_internal text;
ALTER TABLE alliances ADD COLUMN IF NOT EXISTS description_apply    text;

-- Audit-метка последней передачи лидерства (для UI/истории; сам
-- handler — Ф.3 плана 67).
ALTER TABLE alliances ADD COLUMN IF NOT EXISTS leadership_transferred_at timestamptz;

-- +goose Down
ALTER TABLE alliances DROP COLUMN IF EXISTS leadership_transferred_at;
ALTER TABLE alliances DROP COLUMN IF EXISTS description_apply;
ALTER TABLE alliances DROP COLUMN IF EXISTS description_internal;
ALTER TABLE alliances DROP COLUMN IF EXISTS description_external;
