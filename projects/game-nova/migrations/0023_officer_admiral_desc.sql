-- +goose Up
-- ADMIRAL description was "attack +10%" but the effect is build_factor +0.1
-- (fleet ship build speed). Correcting the description.
UPDATE officer_defs
SET description = 'Ускоряет постройку кораблей в верфи на 10%.'
WHERE key = 'ADMIRAL';

-- +goose Down
UPDATE officer_defs
SET description = 'Увеличивает атаку всех ship-stack''ов флота на 10%.'
WHERE key = 'ADMIRAL';
