-- +goose Up
-- AutoMsg — шаблоны системных сообщений, отправляемых игроку при
-- определённых событиях (регистрация, первая планета и т.п.).
--
-- Подстановка переменных — через {{name}} в body_template.
-- Backend заменяет их значениями из vars-мапы вызова Send(userID, key, vars).
--
-- automsg_sent фиксирует «отправлено»: повторный Send с тем же (user,key)
-- ничего не делает. Это важно для идемпотентности (многократная
-- регистрация не должна спамить одним и тем же welcome).

CREATE TABLE automsg_defs (
    key            text PRIMARY KEY,
    title          text NOT NULL,
    body_template  text NOT NULL,
    folder         integer NOT NULL DEFAULT 2
);

CREATE TABLE automsg_sent (
    user_id  uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key      text NOT NULL REFERENCES automsg_defs(key) ON DELETE CASCADE,
    sent_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, key)
);

-- Seed — 3 базовых автомесседжа.
INSERT INTO automsg_defs (key, title, body_template, folder) VALUES
    ('WELCOME',
     'Добро пожаловать в oxsar-nova',
     'Приветствую, {{username}}! Ваше приключение начинается на планете {{planet_name}} в координатах {{coords}}. '
     || 'Загляните во вкладку Постройки — metal_mine и silicon_lab начнут добычу сразу, solar_plant даст им энергию. '
     || 'Удачи в галактике.',
     2),
    ('STARTER_GUIDE',
     'Краткая инструкция',
     'Основные вкладки:'
     || E'\n- Постройки: шахты, верфь, ремонтная фабрика.'
     || E'\n- Исследования: открывают новые корабли и улучшают бой.'
     || E'\n- Верфь: стройте корабли и оборону.'
     || E'\n- Галактика: разведка пустых и чужих систем.'
     || E'\n- Флот: отправка транспортов, атак, шпионов, колоний.'
     || E'\n- Ракеты: удары по обороне противника.'
     || E'\n- Рынок: обмен ресурсов. Рынок артефактов: артефакты за credit.'
     || E'\n- Офицеры: временные бонусы к факторам.'
     || E'\n- Сообщения: боевые отчёты и шпионаж.'
     || E'\n- Достижения: прогресс игрока.',
     2),
    ('FIRST_ATTACK_RECEIVED',
     'Вас атакуют!',
     'Игрок {{attacker}} направил флот на вашу планету {{planet_name}}. Прибытие: {{arrive_at}}. '
     || 'Рекомендуется укрепить оборону или отправить транспортом ресурсы в безопасное место.',
     2);

-- +goose Down
DROP TABLE IF EXISTS automsg_sent;
DROP TABLE IF EXISTS automsg_defs;
