-- +goose Up
-- Отображаемый ранг участника альянса (произвольный текст от owner'а).
-- Пустая строка = не задан (отображается как 'member' или 'owner').
ALTER TABLE alliance_members
    ADD COLUMN rank_name text NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE alliance_members DROP COLUMN rank_name;
