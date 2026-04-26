-- +goose Up
-- План 20 Ф.7 + ADR-0005: Astrophysics (id=112).
-- Даём всем существующим игрокам стартовый astro_level=2,
-- чтобы breaking change (лимит колоний и слотов экспедиций) не
-- ломал их текущее состояние.
--
-- Стартовый astro=2 даёт:
--   colony_limit       = floor(2/2) + 1 = 2 → но лимит у нас MAX(astro, computer+1),
--                        так что игроки с computer_tech>=1 ничего не теряют.
--   expedition_slots   = max(1, floor(sqrt(2))) = 1
--
-- Новые регистрации тоже получают astro=2 (см. starter в auth/service.go).
INSERT INTO research (user_id, unit_id, level)
SELECT u.id, 112, 2
FROM users u
WHERE NOT EXISTS (
    SELECT 1 FROM research r WHERE r.user_id = u.id AND r.unit_id = 112
);

-- +goose Down
DELETE FROM research WHERE unit_id = 112;
