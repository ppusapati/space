-- name: CreateItem :one
INSERT INTO items (
    id, tenant_id, collection_id, mission, platform, instrument, datetime,
    bbox_lon_min, bbox_lat_min, bbox_lon_max, bbox_lat_max,
    geometry_geojson, cloud_cover, properties_json, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $15)
RETURNING *;

-- name: GetItem :one
SELECT * FROM items WHERE id = $1;

-- name: SearchItems :many
SELECT * FROM items
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (
        sqlc.narg('collection_id')::uuid IS NULL
        OR collection_id = sqlc.narg('collection_id')::uuid
      )
  AND datetime BETWEEN sqlc.arg('datetime_start')::timestamptz
                   AND sqlc.arg('datetime_end')::timestamptz
  AND (
        sqlc.narg('max_cloud_cover')::double precision IS NULL
        OR cloud_cover <= sqlc.narg('max_cloud_cover')::double precision
      )
  AND (
        sqlc.narg('bbox_lon_min')::double precision IS NULL
        OR (
              bbox_lon_max >= sqlc.narg('bbox_lon_min')::double precision
          AND bbox_lon_min <= sqlc.narg('bbox_lon_max')::double precision
          AND bbox_lat_max >= sqlc.narg('bbox_lat_min')::double precision
          AND bbox_lat_min <= sqlc.narg('bbox_lat_max')::double precision
        )
      )
  AND (
        sqlc.narg('cursor_datetime')::timestamptz IS NULL
        OR (datetime, id) < (sqlc.narg('cursor_datetime')::timestamptz,
                             sqlc.arg('cursor_id')::uuid)
      )
ORDER BY datetime DESC, id DESC
LIMIT sqlc.arg('lim')::int;
