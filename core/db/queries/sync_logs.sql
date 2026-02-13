-- name: CreateSyncLog :one
INSERT INTO sync_logs (peer_id, timestamp, items_received, items_sent, conflicts, duration_ms, error)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSyncLogsByPeer :many
SELECT * FROM sync_logs WHERE peer_id = ? ORDER BY timestamp DESC LIMIT ?;

-- name: GetRecentSyncLogs :many
SELECT * FROM sync_logs ORDER BY timestamp DESC LIMIT ?;

-- name: GetSyncLog :one
SELECT * FROM sync_logs WHERE id = ?;

-- name: DeleteOldSyncLogs :exec
DELETE FROM sync_logs WHERE timestamp < ?;
