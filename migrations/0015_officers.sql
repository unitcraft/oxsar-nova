-- +goose Up
-- Officers — временные подписки, модифицирующие фактор-поля
-- (users.* / planets.*) на период действия. Работают симметрично
-- артефактам: при активации Apply → UPDATE factor, при истечении
-- Revert → возврат.
--
-- effect JSONB формат:
--   {"scope":"user|all_planets","field":"exchange_rate|produce_factor|...",
--    "op":"add","delta":-0.1}
-- op="set" для absolute-значений; эффекты officer'ов сейчас только
-- multiplicative ('add' к коэффициенту).

CREATE TABLE officer_defs (
    key           text PRIMARY KEY,
    title         text    NOT NULL,
    description   text    NOT NULL,
    duration_days integer NOT NULL,
    cost_credit   bigint  NOT NULL,
    effect        jsonb   NOT NULL
);

CREATE TABLE officer_active (
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    officer_key  text NOT NULL REFERENCES officer_defs(key) ON DELETE RESTRICT,
    activated_at timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL,
    PRIMARY KEY (user_id, officer_key)
);
CREATE INDEX ix_officer_active_expire ON officer_active(expires_at);

-- Seed начального набора. Все эффекты — "add" delta к factor-полям
-- (default factor = 1.0 → +0.1 = 1.1 = +10%). exchange_rate — чем
-- меньше, тем выгоднее (default 1.2 → -0.2 = 1.0 = честный паритет).
INSERT INTO officer_defs (key, title, description, duration_days, cost_credit, effect) VALUES
    ('ADMIRAL',   'Адмирал',
     'Ускоряет постройку кораблей в верфи на 10%.',
     7, 500,
     '{"scope":"user","field":"build_factor","op":"add","delta":0.1}'::jsonb),
    ('GEOLOGIST', 'Геолог',
     'Увеличивает добычу ресурсов на всех планетах на 10%.',
     7, 500,
     '{"scope":"all_planets","field":"produce_factor","op":"add","delta":0.1}'::jsonb),
    ('ENGINEER',  'Инженер',
     'Ускоряет строительство зданий на 25%.',
     7, 500,
     '{"scope":"all_planets","field":"build_factor","op":"add","delta":0.25}'::jsonb),
    ('MERCHANT',  'Торговец',
     'Улучшает обменный курс (exchange_rate → 1.0, паритет).',
     7, 300,
     '{"scope":"user","field":"exchange_rate","op":"add","delta":-0.2}'::jsonb);

-- +goose Down
DROP TABLE IF EXISTS officer_active;
DROP TABLE IF EXISTS officer_defs;
