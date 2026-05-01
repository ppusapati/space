-- name: CreateAntenna :one
INSERT INTO antennas (
    id, tenant_id, station_id, slug, name, band, min_freq_hz, max_freq_hz,
    polarization, gain_dbi, slew_rate_deg_per_s, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12)
RETURNING *;

-- name: GetAntenna :one
SELECT * FROM antennas WHERE id = $1;

-- name: CountAntennasForTenant :one
SELECT COUNT(*)::bigint AS total FROM antennas
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('station_id')::uuid IS NULL OR station_id = sqlc.narg('station_id')::uuid)
  AND (sqlc.narg('band')::int        IS NULL OR band       = sqlc.narg('band')::int);

-- name: ListAntennasForTenant :many
SELECT * FROM antennas
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('station_id')::uuid IS NULL OR station_id = sqlc.narg('station_id')::uuid)
  AND (sqlc.narg('band')::int        IS NULL OR band       = sqlc.narg('band')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: DeprecateAntenna :one
UPDATE antennas
SET
    active     = false,
    updated_at = now(),
    updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
