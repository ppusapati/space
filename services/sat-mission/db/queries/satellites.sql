-- name: RegisterSatellite :one
INSERT INTO satellites (
    id, tenant_id, name, norad_id, international_designator, config_json, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
RETURNING *;

-- name: GetSatellite :one
SELECT * FROM satellites WHERE id = $1;

-- name: ListSatellites :many
SELECT * FROM satellites
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (
        sqlc.narg('cursor_created_at')::timestamptz IS NULL
        OR (created_at, id) < (sqlc.narg('cursor_created_at')::timestamptz,
                               sqlc.arg('cursor_id')::uuid)
      )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('lim')::int;

-- name: UpdateTLE :one
UPDATE satellites
SET tle_line1 = $2, tle_line2 = $3, updated_at = now(), updated_by = $4
WHERE id = $1
RETURNING *;

-- name: UpdateOrbitalState :one
UPDATE satellites
SET
    last_state_rx_km   = $2,
    last_state_ry_km   = $3,
    last_state_rz_km   = $4,
    last_state_vx_km_s = $5,
    last_state_vy_km_s = $6,
    last_state_vz_km_s = $7,
    last_state_epoch   = $8,
    updated_at         = now(),
    updated_by         = $9
WHERE id = $1
RETURNING *;

-- name: SetMode :one
UPDATE satellites
SET current_mode = $2, updated_at = now(), updated_by = $3
WHERE id = $1
RETURNING *;
