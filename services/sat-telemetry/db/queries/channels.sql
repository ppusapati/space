-- name: DefineChannel :one
INSERT INTO channels (
    id, tenant_id, satellite_id, subsystem, name, units,
    value_type, min_value, max_value, sample_rate_hz, active,
    created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, true, $11, $11
)
RETURNING *;

-- name: GetChannel :one
SELECT * FROM channels WHERE id = $1;

-- name: ListChannels :many
SELECT * FROM channels
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('satellite_id')::uuid IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('subsystem')::text    IS NULL OR subsystem    = sqlc.narg('subsystem')::text)
  AND (
        sqlc.narg('cursor_created_at')::timestamptz IS NULL
        OR (created_at, id) < (sqlc.narg('cursor_created_at')::timestamptz,
                               sqlc.arg('cursor_id')::uuid)
      )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('lim')::int;

-- name: DeprecateChannel :one
UPDATE channels
SET
    active     = false,
    updated_at = now(),
    updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
