-- +goose Up
-- Группа офицеров: одновременно может быть активен только один офицер
-- из одной group_key. NULL означает отсутствие ограничений (офицер-одиночка).
ALTER TABLE officer_defs
    ADD COLUMN group_key text;

-- Адмирал и Инженер оба влияют на build_factor — взаимоисключают друг друга.
UPDATE officer_defs SET group_key = 'build' WHERE key IN ('ADMIRAL', 'ENGINEER');

-- +goose Down
ALTER TABLE officer_defs DROP COLUMN group_key;
