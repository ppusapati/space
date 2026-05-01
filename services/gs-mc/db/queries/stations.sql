-- name: CreateGroundStation :one
INSERT INTO ground_stations (
    id, tenant_id, slug, name, country_code, latitude_deg, longitude_deg, altitude_m,
    created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
RETURNING *;

-- name: GetGroundStation :one
SELECT * FROM ground_stations WHERE id = $1;

-- name: CountGroundStationsForTenant :one
SELECT COUNT(*)::bigint AS total FROM ground_stations
WHERE tenant_id = sqlc.arg('tenant_id')::uuid;

-- name: ListGroundStationsForTenant :many
SELECT * FROM ground_stations
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: DeprecateGroundStation :one
UPDATE ground_stations
SET
    active     = false,
    updated_at = now(),
    updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
