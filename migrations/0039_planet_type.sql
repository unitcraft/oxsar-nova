-- +goose Up
ALTER TABLE planets ADD COLUMN IF NOT EXISTS planet_type TEXT NOT NULL DEFAULT '';

-- Проставить тип существующим планетам детерминированно из позиции.
-- Используем простую логику: средняя позиция → normaltempplanet, крайние → eisplanet/trockenplanet.
-- Точный алгоритм (с rng) применится только при создании новых планет.
UPDATE planets SET planet_type = CASE
    WHEN is_moon THEN 'moon'
    WHEN position <= 2 THEN 'trockenplanet'
    WHEN position <= 3 THEN 'wuestenplanet'
    WHEN position <= 7 THEN 'dschjungelplanet'
    WHEN position <= 10 THEN 'normaltempplanet'
    WHEN position <= 13 THEN 'wasserplanet'
    WHEN position <= 14 THEN 'eisplanet'
    ELSE 'gasplanet'
END
WHERE planet_type = '';

-- +goose Down
ALTER TABLE planets DROP COLUMN IF EXISTS planet_type;
