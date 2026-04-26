-- +goose Up
-- Добавить категорию достижений для объединения пассивных и стартовых.

ALTER TABLE achievement_defs ADD COLUMN category text NOT NULL DEFAULT 'passive';

-- Обновить существующие ачивки (все пока пассивные):
UPDATE achievement_defs SET category = 'passive'
WHERE key IN ('FIRST_METAL', 'FIRST_SILICON', 'FIRST_ARTEFACT', 'FIRST_WIN', 'FIRST_COLONY');

-- Добавить стартовые ачивки (Tutorial-цепочка):
INSERT INTO achievement_defs (key, title, description, points, category) VALUES
    ('STARTER_BUILD_METALMINE',    'Первая шахта',                'Постройте шахту на своей планете',                10, 'starter'),
    ('STARTER_BUILD_SOLARPLANT',   'Солнечная энергия',           'Постройте солнечный растений на своей планете',  10, 'starter'),
    ('STARTER_BUILD_METALLURGY',   'Металлургический завод',      'Постройте металлургический завод',                10, 'starter'),
    ('STARTER_BUILD_SHIPYARD',     'Верфь готова',                'Постройте верфь на своей планете',                10, 'starter'),
    ('STARTER_BUILD_LAB',          'Лаборатория',                 'Постройте лабораторию на своей планете',          10, 'starter'),
    ('STARTER_RESEARCH_TECH',      'Первое исследование',         'Проведите своё первое исследование',              10, 'starter'),
    ('STARTER_BUILD_SHIP',         'Первый корабль',              'Постройте первый корабль',                        10, 'starter'),
    ('STARTER_SEND_MISSION',       'Первая миссия',               'Отправьте первую боевую миссию',                  10, 'starter');

-- +goose Down
ALTER TABLE achievement_defs DROP COLUMN category;
DELETE FROM achievement_defs WHERE category = 'starter';
