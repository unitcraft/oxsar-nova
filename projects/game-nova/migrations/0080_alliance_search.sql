-- План 67 Ф.4: полнотекстовый поиск альянсов (U-012).
--
-- Добавляем GIN-индекс на expression to_tsvector(simple, name||tag) —
-- без хранимой колонки tsvector. Преимущества:
--   - Не нужен trigger для синхронизации tsvector с name/tag.
--   - Postgres сам пересчитывает выражение при INSERT/UPDATE/SELECT.
--   - Конфигурация simple (а не russian/english) — потому что name/tag
--     обычно латинские и пользователи ожидают prefix-match без
--     морфологии (TAG → TAG, не TAG-ы → ТЭГ).
--
-- Использование (в service.List):
--   WHERE to_tsvector('simple', a.name || ' ' || a.tag) @@ to_tsquery('simple', $1)
--   или (для prefix-match): @@ to_tsquery('simple', $1 || ':*')
--
-- B-tree индексы по member_count и created_at — для сортировки и
-- range-фильтров (min/max_members).

-- +goose Up
CREATE INDEX IF NOT EXISTS ix_alliances_search
    ON alliances
    USING GIN (to_tsvector('simple', name || ' ' || tag));

-- Для фильтра is_open (binary, partial-index — короче).
CREATE INDEX IF NOT EXISTS ix_alliances_is_open
    ON alliances(is_open) WHERE is_open = true;

-- +goose Down
DROP INDEX IF EXISTS ix_alliances_is_open;
DROP INDEX IF EXISTS ix_alliances_search;
