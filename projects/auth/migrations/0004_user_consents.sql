-- План 44 (152-ФЗ): фиксируем согласие пользователя на обработку
-- персональных данных. На случай спора — IP, User-Agent, версия документа
-- и timestamp принятия. user_id — UUID (см. 0001_init.sql).

-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_consents (
    id                   BIGSERIAL PRIMARY KEY,
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    consent_type         TEXT NOT NULL,
    consent_text_version TEXT NOT NULL,
    accepted_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    accepted_ip          INET,
    accepted_user_agent  TEXT
);

CREATE INDEX user_consents_user_type_idx
    ON user_consents (user_id, consent_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_consents;
-- +goose StatementEnd
