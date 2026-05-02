-- План 72.1.55.E (effects подплан): исправление типов из 0097.
--
-- Verify-handsweep после 0097 показал:
--   - esps в legacy — int 1..99 (количество espionage_sensor по
--     умолчанию в форме отправки шпионажа), а не bool.
--     Legacy `Preferences.class.php:407-414` clamp-ит [1..99].
--   - ipcheck в legacy — default true (включена защита смены IP),
--     юзер может opt-out. В 0097 default стоял false.
--
-- Делаем безопасную миграцию: переопределяем тип/default. Существующие
-- значения боев боев boolean мапятся на: false → 5 (legacy default
-- ESPIONAGE_DRONES_DEFAULT), true → 5 — потеря не критична, фича
-- свеженая, прод-юзеров с прежним значением нет.

-- +goose Up

-- esps boolean → smallint
ALTER TABLE users DROP COLUMN IF EXISTS esps;
ALTER TABLE users ADD  COLUMN esps smallint NOT NULL DEFAULT 5
    CHECK (esps >= 1 AND esps <= 99);

-- ipcheck default false → true (legacy IPCHECK_ENABLED)
ALTER TABLE users ALTER COLUMN ipcheck SET DEFAULT true;
UPDATE users SET ipcheck = true WHERE ipcheck = false;

-- last_seen_ip — для ipcheck effect (план 72.1.55.E):
-- middleware сравнивает X-Forwarded-For/RemoteAddr с этим значением;
-- при расхождении и ipcheck=true → AutoMsg-уведомление + UPDATE.
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_seen_ip text;

-- +goose Down

ALTER TABLE users DROP COLUMN IF EXISTS last_seen_ip;
ALTER TABLE users DROP COLUMN IF EXISTS esps;
ALTER TABLE users ADD  COLUMN esps boolean NOT NULL DEFAULT true;

ALTER TABLE users ALTER COLUMN ipcheck SET DEFAULT false;
