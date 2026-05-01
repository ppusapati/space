-- name: CreateAsset :one
INSERT INTO assets (id, item_id, key, href, media_type, title, roles)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAssetsForItem :many
SELECT * FROM assets WHERE item_id = $1 ORDER BY key;
