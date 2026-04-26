-- +goose Up
-- План 17 D: Daily Quests.
-- Quest-определения — статические, hardcoded в коде (не в YAML), но
-- хранятся в БД для FK на quest_id в daily_quests.
CREATE TABLE daily_quest_defs (
    id              SERIAL PRIMARY KEY,
    key             TEXT NOT NULL UNIQUE,
    title           TEXT NOT NULL,
    condition_type  TEXT NOT NULL,   -- 'resource_earn'|'fleet_mission'|'research_done'|'spy_done'|'market_sell'
    condition_value JSONB NOT NULL,  -- {"target": 50000} или {"mission": 7} и т.п.
    target_progress INT  NOT NULL DEFAULT 1,  -- сколько надо набрать прогресса
    reward_credits  INT  NOT NULL DEFAULT 0,
    reward_metal    BIGINT NOT NULL DEFAULT 0,
    reward_silicon  BIGINT NOT NULL DEFAULT 0,
    reward_hydrogen BIGINT NOT NULL DEFAULT 0,
    weight          INT NOT NULL DEFAULT 100  -- вес для random выбора (выше = чаще)
);

-- Активные quest у игрока на день. PK = (user_id, def_id, date) гарантирует
-- невозможность дубля при race на lazy-генерации.
CREATE TABLE daily_quests (
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    def_id       INT  NOT NULL REFERENCES daily_quest_defs(id),
    date         DATE NOT NULL,
    progress     INT  NOT NULL DEFAULT 0,
    completed_at TIMESTAMPTZ,
    claimed_at   TIMESTAMPTZ,
    PRIMARY KEY (user_id, def_id, date)
);

CREATE INDEX ix_daily_quests_user_date ON daily_quests(user_id, date);

-- Seed-определения. resource_earn (ежедневная добыча) намеренно
-- исключён в MVP — требует трекинга delta, что добавит DB-нагрузку
-- на каждый planet tick. В v1.x можно вернуть через snapshot.
INSERT INTO daily_quest_defs (key, title, condition_type, condition_value, target_progress, reward_credits, reward_metal, reward_silicon, reward_hydrogen)
VALUES
  ('send_transport',     'Отправить 1 транспорт',    'fleet_mission', '{"mission":7}',  1, 15, 0,    0,    0),
  ('send_recycling',     'Отправить 1 ресайклер',    'fleet_mission', '{"mission":9}',  1, 15, 0,    0,    0),
  ('do_spy',             'Провести 1 шпионаж',       'fleet_mission', '{"mission":11}', 1, 15, 0,    0,    0),
  ('send_attack',        'Атаковать игрока',         'fleet_mission', '{"mission":10}', 1, 25, 0,    0,    0),
  ('send_position',      'Перебазировать флот',      'fleet_mission', '{"mission":6}',  1, 15, 0,    0,    0),
  ('send_expedition',    'Отправить 1 экспедицию',   'fleet_mission', '{"mission":15}', 1, 20, 0,    0,    0),
  ('do_research',        'Завершить исследование',   'research_done', '{}',             1, 30, 0,    0,    0),
  ('build_anything',     'Построить здание',         'building_done', '{}',             1, 20, 5000, 0,    0),
  ('build_3_buildings',  'Построить 3 здания',       'building_done', '{}',             3, 50, 15000, 5000, 0);

-- +goose Down
DROP TABLE daily_quests;
DROP TABLE daily_quest_defs;
