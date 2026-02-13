-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS peers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    address TEXT NOT NULL,
    last_seen TIMESTAMP,
    last_sync TIMESTAMP,
    is_trusted BOOLEAN DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_peers_is_trusted ON peers(is_trusted);
CREATE INDEX idx_peers_last_seen ON peers(last_seen);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_peers_last_seen;
DROP INDEX IF EXISTS idx_peers_is_trusted;
DROP TABLE IF EXISTS peers;
-- +goose StatementEnd
