-- +goose Up
-- Расширение каталога достижений: добавляем 10 новых достижений
-- сверх базовых 5 (migration 0014).

INSERT INTO achievement_defs (key, title, description, points) VALUES
    ('FIRST_FLEET',       'Флотоводец',          'Отправлен первый флот.', 2),
    ('FIRST_EXPEDITION',  'Первооткрыватель',    'Завершена первая экспедиция.', 2),
    ('FIRST_RESEARCH',    'Учёный',              'Получен первый уровень исследования.', 1),
    ('BATTLE_10',         'Ветеран',             'Выиграно 10 боёв.', 5),
    ('FLEET_50',          'Адмирал',             'Построено суммарно 50 кораблей.', 5),
    ('ARTEFACT_MARKET',   'Торговец',            'Куплен артефакт на рынке.', 3),
    ('SPY_SUCCESS',       'Разведчик',           'Успешно проведена шпионская миссия.', 2),
    ('RECYCLING',         'Мусорщик',            'Собраны обломки с поля боя.', 2),
    ('ROCKET_LAUNCH',     'Ракетчик',            'Запущена первая ракета.', 2),
    ('SCORE_1000',        'Тысячник',            'Набрано 1000 очков рейтинга.', 4)
ON CONFLICT (key) DO NOTHING;

-- +goose Down
DELETE FROM achievement_defs WHERE key IN (
    'FIRST_FLEET', 'FIRST_EXPEDITION', 'FIRST_RESEARCH', 'BATTLE_10',
    'FLEET_50', 'ARTEFACT_MARKET', 'SPY_SUCCESS', 'RECYCLING',
    'ROCKET_LAUNCH', 'SCORE_1000'
);
