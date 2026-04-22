-- +goose Up
-- Добавить фактор производства для управления % производства каждого здания (0-100%).

ALTER TABLE buildings ADD COLUMN production_factor integer NOT NULL DEFAULT 100;

-- +goose Down
ALTER TABLE buildings DROP COLUMN production_factor;
