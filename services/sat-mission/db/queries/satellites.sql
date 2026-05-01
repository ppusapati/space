-- name: RegisterSatellite :one
INSERT INTO satellites (
    id, tenant_id, name, norad_id, international_designator, config_json, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
RETURNING *;

-- name: GetSatellite :one
SELECT * FROM satellites WHERE id = $1;

-- name: CountSatellitesForTenant :one
SELECT COUNT(*)::bigint AS total FROM satellites
WHERE tenant_id = sqlc.arg('tenant_id')::uuid;

-- name: ListSatellitesForTenant :many
SELECT * FROM satellites
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: UpdateTLE :one
UPDATE satellites
SET tle_line1 = sqlc.arg('tle_line1')::text,
    tle_line2 = sqlc.arg('tle_line2')::text,
    updated_at = now(),
    updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;

-- name: UpdateOrbitalState :one
UPDATE satellites
SET
    last_state_rx_km   = sqlc.arg('rx_km')::double precision,
    last_state_ry_km   = sqlc.arg('ry_km')::double precision,
    last_state_rz_km   = sqlc.arg('rz_km')::double precision,
    last_state_vx_km_s = sqlc.arg('vx_km_s')::double precision,
    last_state_vy_km_s = sqlc.arg('vy_km_s')::double precision,
    last_state_vz_km_s = sqlc.arg('vz_km_s')::double precision,
    last_state_epoch   = sqlc.arg('epoch')::timestamptz,
    updated_at         = now(),
    updated_by         = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;

-- name: SetMode :one
UPDATE satellites
SET current_mode = sqlc.arg('mode')::int,
    updated_at = now(),
    updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
