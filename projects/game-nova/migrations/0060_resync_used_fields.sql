-- +goose Up
-- План 23: синхронизация used_fields на существующих планетах.
-- Раньше used_fields никогда не увеличивался при постройке зданий,
-- из-за чего все планеты после M4 имели used_fields=0 независимо
-- от количества зданий. Теперь handler события build-complete
-- инкрементирует used_fields — но для планет, построенных до этого
-- коммита, нужен bulk-resync:
--
--   used_fields = COUNT(*) FROM buildings WHERE planet_id = p.id
--
-- Solar satellites лежат в ships (они временно-учитываются как
-- энергетические спутники, а не здания), поэтому COUNT на buildings
-- корректен.

UPDATE planets p
SET used_fields = COALESCE(
    (SELECT COUNT(*)::integer FROM buildings b WHERE b.planet_id = p.id),
    0
);

-- +goose Down
-- Откат: обнуляем used_fields. Логика «инкремент при постройке»
-- останется в коде, так что если опять накатим вверх — значение
-- снова восстановится через resync.
UPDATE planets SET used_fields = 0;
