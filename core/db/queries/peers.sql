-- name: CreatePeer :one
INSERT INTO peers (id, name, address, last_seen, is_trusted)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetPeer :one
SELECT * FROM peers WHERE id = ?;

-- name: GetAllPeers :many
SELECT * FROM peers ORDER BY last_seen DESC;

-- name: GetTrustedPeers :many
SELECT * FROM peers WHERE is_trusted = 1 ORDER BY last_seen DESC;

-- name: UpdatePeerLastSeen :exec
UPDATE peers SET last_seen = ? WHERE id = ?;

-- name: UpdatePeerLastSync :exec
UPDATE peers SET last_sync = ? WHERE id = ?;

-- name: UpdatePeerTrust :exec
UPDATE peers SET is_trusted = ? WHERE id = ?;

-- name: DeletePeer :exec
DELETE FROM peers WHERE id = ?;

-- name: GetPeerByAddress :one
SELECT * FROM peers WHERE address = ?;
