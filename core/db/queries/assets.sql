-- name: CreateAsset :one
INSERT INTO assets (
    item_id, type, name, file_path, file_size, file_hash, created_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetAssetByID :one
SELECT * FROM assets
WHERE id = ?;

-- name: GetAssetsByItemID :many
SELECT * FROM assets
WHERE item_id = ?
ORDER BY created_at DESC;

-- name: DeleteAsset :exec
DELETE FROM assets
WHERE id = ?;

-- name: CountAssetsByItemID :one
SELECT COUNT(*) FROM assets
WHERE item_id = ?;

-- name: CountAssetsByItemIDAndType :one
SELECT COUNT(*) FROM assets
WHERE item_id = ? AND type = ?;
