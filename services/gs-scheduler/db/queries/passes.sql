-- name: InsertContactPass :one
INSERT INTO contact_passes (
    id, tenant_id, station_id, satellite_id, aos_time, tca_time, los_time,
    max_elevation_deg, aos_azimuth_deg, los_azimuth_deg, source,
    created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
RETURNING *;

-- name: GetContactPass :one
SELECT * FROM contact_passes WHERE id = $1;

-- name: CountContactPassesForTenant :one
SELECT COUNT(*)::bigint AS total FROM contact_passes
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('station_id')::uuid    IS NULL OR station_id   = sqlc.narg('station_id')::uuid)
  AND (sqlc.narg('satellite_id')::uuid  IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('aos_start')::timestamptz IS NULL OR aos_time >= sqlc.narg('aos_start')::timestamptz)
  AND (sqlc.narg('aos_end')::timestamptz   IS NULL OR aos_time <= sqlc.narg('aos_end')::timestamptz)
  AND (sqlc.narg('min_elevation_deg')::double precision IS NULL
       OR max_elevation_deg >= sqlc.narg('min_elevation_deg')::double precision);

-- name: ListContactPassesForTenant :many
SELECT * FROM contact_passes
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('station_id')::uuid    IS NULL OR station_id   = sqlc.narg('station_id')::uuid)
  AND (sqlc.narg('satellite_id')::uuid  IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('aos_start')::timestamptz IS NULL OR aos_time >= sqlc.narg('aos_start')::timestamptz)
  AND (sqlc.narg('aos_end')::timestamptz   IS NULL OR aos_time <= sqlc.narg('aos_end')::timestamptz)
  AND (sqlc.narg('min_elevation_deg')::double precision IS NULL
       OR max_elevation_deg >= sqlc.narg('min_elevation_deg')::double precision)
ORDER BY aos_time ASC, id ASC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;
