-- +goose Up
CREATE TABLE expedition_visits (
    user_id TEXT NOT NULL,
    galaxy  INT  NOT NULL,
    system  INT  NOT NULL,
    visits  INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, galaxy, system)
);

-- +goose Down
DROP TABLE IF EXISTS expedition_visits;
