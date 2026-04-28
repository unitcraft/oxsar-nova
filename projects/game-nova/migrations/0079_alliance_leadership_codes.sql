-- План 67 Ф.3: коды подтверждения для transfer-leadership.
--
-- По аналогии с account_deletion_codes (миграция 0051), но scope —
-- per-alliance. Один активный код на альянс: текущий owner запрашивает
-- код, получает его системным сообщением, затем подтверждает передачу
-- передавая (new_owner_id, code).
--
-- Запись удаляется при успешном transfer'е, expire'нувшие чистятся
-- лениво при попытке использовать (как в settings/delete.go).

-- +goose Up
CREATE TABLE alliance_leadership_codes (
    alliance_id      uuid PRIMARY KEY REFERENCES alliances(id) ON DELETE CASCADE,
    -- Текущий owner на момент запроса кода. Если за TTL он сменился —
    -- код инвалидируется (см. handler).
    requester_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- Целевой получатель лидерства. Хранится здесь чтобы код был
    -- bound к конкретной паре (current_owner, new_owner) — нельзя
    -- запросить код для одного и подтвердить для другого.
    new_owner_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash        text NOT NULL,
    issued_at        timestamptz NOT NULL DEFAULT now(),
    expires_at       timestamptz NOT NULL,
    attempts         integer NOT NULL DEFAULT 0,
    CHECK (attempts >= 0)
);

-- Очистка expired-записей: используется handler'ом для лениво-
-- инвалидированных кодов.
CREATE INDEX ix_alliance_leadership_codes_expires
    ON alliance_leadership_codes(expires_at);

-- +goose Down
DROP INDEX IF EXISTS ix_alliance_leadership_codes_expires;
DROP TABLE IF EXISTS alliance_leadership_codes;
