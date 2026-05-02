-- План 72.1.55 Task I (P72.S4.SETTINGS subset 1:1): legacy
-- preferences.tpl поля, реализуемые на nova-стеке (без legacy-PHP
-- тем skin/template — те остаются в P72.S4.SETTINGS как
-- архитектурное упрощение).
--
-- Поля:
--   - show_all_constructions/research/shipyard/defense (4 bool):
--     показывать недоступные юниты в info-таблицах (default true,
--     legacy default тоже true).
--   - planetorder smallint (0=date join, 1=name, 2=coords): сортировка
--     planet-switcher и sidebar.
--   - esps bool: уровень детализации espionage report (default true).
--   - ipcheck bool: предупреждать при логине с другого IP
--     (default false — security-feature opt-in).
--
-- Effects (применение этих preferences на UI/backend) — backlog
-- 72.1.55.* подплана. Этот патч закрывает только storage + settings UI.

-- +goose Up

ALTER TABLE users ADD COLUMN IF NOT EXISTS show_all_constructions boolean NOT NULL DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS show_all_research      boolean NOT NULL DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS show_all_shipyard      boolean NOT NULL DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS show_all_defense       boolean NOT NULL DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS planetorder smallint NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS esps    boolean NOT NULL DEFAULT true;
ALTER TABLE users ADD COLUMN IF NOT EXISTS ipcheck boolean NOT NULL DEFAULT false;

-- +goose Down

ALTER TABLE users DROP COLUMN IF EXISTS ipcheck;
ALTER TABLE users DROP COLUMN IF EXISTS esps;
ALTER TABLE users DROP COLUMN IF EXISTS planetorder;
ALTER TABLE users DROP COLUMN IF EXISTS show_all_defense;
ALTER TABLE users DROP COLUMN IF EXISTS show_all_shipyard;
ALTER TABLE users DROP COLUMN IF EXISTS show_all_research;
ALTER TABLE users DROP COLUMN IF EXISTS show_all_constructions;
