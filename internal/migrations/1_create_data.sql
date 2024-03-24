-- +goose Up
CREATE TABLE data
(
    id         uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    fqdn       text        NOT NULL,
    ip         text        NOT NULL,
    is_active  bool        NOT NULL DEFAULT true,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ON data (fqdn);
CREATE INDEX ON data (ip);
CREATE INDEX ON data (expires_at);
CREATE INDEX ON data (created_at);

-- +goose Down
DROP TABLE data;