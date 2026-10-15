-- +goose Up
CREATE TABLE visits
(
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    link_id    BIGINT REFERENCES links (id) NOT NULL,
    ip         varchar(45)                  NOT NULL,
    user_agent TEXT                         NOT NULL,
    status     INT                          NOT NULL,
    created_at TIMESTAMPTZ                  NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE visits;
