-- План 67 Ф.1: журнал активности альянса (U-013).
--
-- По образцу admin_audit_log (миграция 0059): запись на каждое
-- значимое действие — invite/kick/promote/relation_proposed/
-- description_changed/leadership_transferred и т.д.
--
-- R10: nova однобазная (universe = отдельный инстанс БД), поэтому
-- universe_id не добавляется. Если в будущем nova станет
-- мультитенантной в одной БД — добавится отдельной миграцией.
--
-- actor_id: пользователь, выполнивший действие (NULL = системное,
-- например авто-выход после inactivity).
-- target_kind/target_id: на кого/что направлено (user|alliance|
-- relation|rank|''). Пустая строка = действие без явного объекта.

-- +goose Up
CREATE TABLE alliance_audit_log (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    alliance_id  uuid NOT NULL REFERENCES alliances(id) ON DELETE CASCADE,
    actor_id     uuid REFERENCES users(id) ON DELETE SET NULL,
    action       text NOT NULL,
    target_kind  text NOT NULL DEFAULT '',
    target_id    text NOT NULL DEFAULT '',
    payload      jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at   timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ix_alliance_audit_alliance_created
    ON alliance_audit_log(alliance_id, created_at DESC);
CREATE INDEX ix_alliance_audit_actor
    ON alliance_audit_log(actor_id, created_at DESC)
    WHERE actor_id IS NOT NULL;
CREATE INDEX ix_alliance_audit_action
    ON alliance_audit_log(alliance_id, action, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS alliance_audit_log;
