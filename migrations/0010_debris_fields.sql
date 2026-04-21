-- +goose Up
-- debris_fields — обломки ship'ов на орбите после боя.
--
-- Legacy-эквивалент: game/Galaxy.class.php хранит debris как
-- «фиктивные планеты» на position=16 (orbit). В современном
-- стеке просто отдельная таблица, привязанная к координатам
-- (независимо от того, есть ли живая планета на них).
--
-- После ATTACK 30% metal+silicon от стоимости уничтоженных ship'ов
-- попадает сюда (defense в debris НЕ идёт, как в OGame). RECYCLING
-- (kind=9) собирает debris в carry.
--
-- Один field на координаты (UPSERT при втором бое туда же).

CREATE TABLE debris_fields (
    galaxy       integer NOT NULL,
    system       integer NOT NULL,
    position     integer NOT NULL,
    metal        numeric(20, 0) NOT NULL DEFAULT 0,
    silicon      numeric(20, 0) NOT NULL DEFAULT 0,
    last_update  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (galaxy, system, position)
);

-- +goose Down
DROP TABLE IF EXISTS debris_fields;
