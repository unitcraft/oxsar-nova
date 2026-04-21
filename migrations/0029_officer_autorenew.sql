-- +goose Up
ALTER TABLE officer_active ADD COLUMN auto_renew boolean NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE officer_active DROP COLUMN auto_renew;
