-- +goose Up
CREATE TABLE IF NOT EXISTS tokens (
    hash TEXT PRIMARY KEY,
    user_id UUID NOT NULL  REFERENCES users ON DELETE CASCADE,
    expiry TIMESTAMPTZ  NOT NULL,
    scope TEXT NOT NULL
);


-- +goose Down
DROP TABLE IF EXISTS tokens;
