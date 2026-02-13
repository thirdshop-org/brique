-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS sync_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    peer_id TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    items_received INTEGER DEFAULT 0,
    items_sent INTEGER DEFAULT 0,
    conflicts INTEGER DEFAULT 0,
    duration_ms INTEGER DEFAULT 0,
    error TEXT,
    FOREIGN KEY (peer_id) REFERENCES peers(id) ON DELETE CASCADE
);

CREATE INDEX idx_sync_logs_peer_id ON sync_logs(peer_id);
CREATE INDEX idx_sync_logs_timestamp ON sync_logs(timestamp DESC);

-- Add sync tracking columns to items table
ALTER TABLE items ADD COLUMN origin_peer_id TEXT;
ALTER TABLE items ADD COLUMN sync_version INTEGER DEFAULT 1;

CREATE INDEX idx_items_origin_peer ON items(origin_peer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_items_origin_peer;
DROP INDEX IF EXISTS idx_sync_logs_timestamp;
DROP INDEX IF EXISTS idx_sync_logs_peer_id;
DROP TABLE IF EXISTS sync_logs;

-- Note: SQLite doesn't support DROP COLUMN, so we can't cleanly remove the added columns
-- In a real rollback scenario, you'd need to recreate the items table without those columns
-- +goose StatementEnd
