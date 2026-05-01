-- name: CreateTileSet :one
INSERT INTO tile_sets (
    id, tenant_id, slug, name, description, format, projection,
    min_zoom, max_zoom, source_uri, attribution, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
RETURNING *;

-- name: GetTileSet :one
SELECT * FROM tile_sets WHERE id = $1;

-- name: CountTileSetsForTenant :one
SELECT COUNT(*)::bigint AS total FROM tile_sets
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('format')::int IS NULL OR format = sqlc.narg('format')::int);

-- name: ListTileSetsForTenant :many
SELECT * FROM tile_sets
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('format')::int IS NULL OR format = sqlc.narg('format')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: DeprecateTileSet :one
UPDATE tile_sets
SET
    active     = false,
    updated_at = now(),
    updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
