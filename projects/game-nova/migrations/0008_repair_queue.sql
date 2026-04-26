-- +goose Up
-- repair_queue — очередь ремонтной фабрики (§10 ТЗ, ex ext/ExtRepair).
-- В M2-MVP поддерживается только DISASSEMBLE (batch): игрок выбирает
-- здоровые корабли/оборону, они списываются со стока немедленно,
-- заранее списывается required-стоимость; при завершении события
-- начисляется return-стоимость (итоговая прибыль = return - required).
--
-- kind в events: 51 = EVENT_DISASSEMBLE (совпадает с legacy).
-- REPAIR (kind=50) добавим после порта боя, когда появится ships.damaged_count.

CREATE TABLE repair_queue (
    id              uuid PRIMARY KEY,
    planet_id       uuid    NOT NULL REFERENCES planets(id) ON DELETE CASCADE,
    user_id         uuid    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    unit_id         integer NOT NULL,
    is_defense      boolean NOT NULL DEFAULT false,
    mode            text    NOT NULL,            -- 'disassemble' | 'repair'
    count           bigint  NOT NULL,
    return_metal    numeric(20, 0) NOT NULL,     -- будет зачислено на целой очереди
    return_silicon  numeric(20, 0) NOT NULL,
    return_hydrogen numeric(20, 0) NOT NULL,
    per_unit_seconds integer NOT NULL,
    start_at        timestamptz NOT NULL,
    end_at          timestamptz NOT NULL,
    status          queue_status NOT NULL DEFAULT 'running',
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ix_repair_queue_planet ON repair_queue(planet_id) WHERE status IN ('queued','running');

-- +goose Down
DROP TABLE IF EXISTS repair_queue;
