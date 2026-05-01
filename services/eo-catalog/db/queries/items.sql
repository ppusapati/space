-- name: CreateItem :one
INSERT INTO items (
    id, tenant_id, collection_id, mission, platform, instrument, datetime,
    bbox_lon_min, bbox_lat_min, bbox_lon_max, bbox_lat_max,
    geometry_geojson, cloud_cover, properties_json, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $15)
RETURNING *;

-- name: GetItem :one
SELECT * FROM items WHERE id = $1;

-- name: CountItemsForTenant :one
SELECT COUNT(*)::bigint AS total FROM items
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('collection_id')::uuid IS NULL OR collection_id = sqlc.narg('collection_id')::uuid);

-- name: ListItemsForTenant :many
SELECT * FROM items
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('collection_id')::uuid IS NULL OR collection_id = sqlc.narg('collection_id')::uuid)
ORDER BY datetime DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;
