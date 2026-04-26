-- +goose Up
-- Achievements — пассивные награды за прогресс (legacy na_phrases
-- содержит 20+ ключей для MENU_ACHIEVEMENTS). Стартуем с 5 базовых;
-- дополнения — правкой seed в последующих миграциях.

CREATE TABLE achievement_defs (
    key          text PRIMARY KEY,             -- FIRST_METAL, FIRST_WIN, ...
    title        text NOT NULL,
    description  text NOT NULL,
    points       integer NOT NULL DEFAULT 1    -- условные «очки достижений»
);

CREATE TABLE achievements_user (
    user_id       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement   text NOT NULL REFERENCES achievement_defs(key) ON DELETE CASCADE,
    unlocked_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, achievement)
);
CREATE INDEX ix_achievements_user ON achievements_user(user_id, unlocked_at DESC);

-- Seed начального каталога. title/description отражают legacy
-- na_phrases-ключи (FIRST_METAL и т.п.); UI будет брать локализацию
-- из i18n по convention ACHIEVEMENT_<KEY>.
INSERT INTO achievement_defs (key, title, description, points) VALUES
    ('FIRST_METAL',    'Metal — основа всего', 'Построен первый metal_mine.', 1),
    ('FIRST_SILICON',  'Кремниевое вещество',  'Построен первый silicon_lab.', 1),
    ('FIRST_ARTEFACT', 'Охотник за артефактами', 'Получен первый артефакт.', 2),
    ('FIRST_WIN',      'Первая победа',        'Выигран первый бой mission=ATTACK.', 3),
    ('FIRST_COLONY',   'Колонизатор',          'Основана первая колония.', 3);

-- +goose Down
DROP TABLE IF EXISTS achievements_user;
DROP TABLE IF EXISTS achievement_defs;
