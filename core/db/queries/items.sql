-- name: CreateItem :one
INSERT INTO items (
    name, category, brand, model, serial_number,
    purchase_date, photo_path, notes, created_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
)
RETURNING *;

-- name: GetItemByID :one
SELECT * FROM items
WHERE id = ?;

-- name: GetAllItems :many
SELECT * FROM items
ORDER BY updated_at DESC;

-- name: UpdateItem :exec
UPDATE items
SET
    name = ?,
    category = ?,
    brand = ?,
    model = ?,
    serial_number = ?,
    purchase_date = ?,
    photo_path = ?,
    notes = ?,
    updated_at = ?
WHERE id = ?;

-- name: DeleteItem :exec
DELETE FROM items
WHERE id = ?;

-- name: SearchItems :many
SELECT * FROM items
WHERE name LIKE ? OR brand LIKE ? OR category LIKE ?
ORDER BY updated_at DESC;

-- name: GetItemsModifiedSince :many
SELECT * FROM items
WHERE updated_at > ?
ORDER BY updated_at DESC;

-- name: CountItems :one
SELECT COUNT(*) FROM items;
