-- name: CreateCollection :one
INSERT INTO collections (
    id, tenant_id, slug, title, description, license,
    bbox_lon_min, bbox_lat_min, bbox_lon_max, bbox_lat_max,
    temporal_start, temporal_end, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $13)
RETURNING *;

-- name: GetCollection :one
SELECT * FROM collections WHERE id = $1;

-- name: ListCollectionsForTenant :many
SELECT * FROM collections
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (
        sqlc.narg('cursor_created_at')::timestamptz IS NULL
        OR (created_at, id) < (sqlc.narg('cursor_created_at')::timestamptz,
                               sqlc.arg('cursor_id')::uuid)
      )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('lim')::int;
