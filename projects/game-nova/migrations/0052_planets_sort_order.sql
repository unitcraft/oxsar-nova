-- +goose Up
ALTER TABLE planets ADD COLUMN sort_order integer NOT NULL DEFAULT 0;
CREATE INDEX ix_planets_user_sort ON planets(user_id, sort_order);

-- Инициализация: для существующих планет sort_order = порядок создания.
WITH ordered AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY created_at) - 1 AS rn
    FROM planets
)
UPDATE planets p SET sort_order = o.rn FROM ordered o WHERE p.id = o.id;

-- +goose Down
DROP INDEX IF EXISTS ix_planets_user_sort;
ALTER TABLE planets DROP COLUMN IF EXISTS sort_order;
