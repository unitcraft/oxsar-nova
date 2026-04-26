-- +goose Up
-- Добавляем is_moon в PK debris_fields, чтобы поле над планетой и над
-- луной считались независимо (как в OGame).
ALTER TABLE debris_fields DROP CONSTRAINT debris_fields_pkey;
ALTER TABLE debris_fields ADD COLUMN is_moon boolean NOT NULL DEFAULT false;
ALTER TABLE debris_fields ADD PRIMARY KEY (galaxy, system, position, is_moon);

-- +goose Down
ALTER TABLE debris_fields DROP CONSTRAINT debris_fields_pkey;
ALTER TABLE debris_fields DROP COLUMN is_moon;
ALTER TABLE debris_fields ADD PRIMARY KEY (galaxy, system, position);
