-- name: DefineCommand :one
INSERT INTO command_defs (
    id, tenant_id, satellite_id, subsystem, name, opcode,
    parameters_schema, description, created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $9
)
RETURNING *;

-- name: GetCommand :one
SELECT * FROM command_defs WHERE id = $1;

-- name: CountCommandsForTenant :one
SELECT COUNT(*)::bigint AS total FROM command_defs
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('satellite_id')::uuid IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('subsystem')::text    IS NULL OR subsystem    = sqlc.narg('subsystem')::text);

-- name: ListCommandsForTenant :many
SELECT * FROM command_defs
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('satellite_id')::uuid IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('subsystem')::text    IS NULL OR subsystem    = sqlc.narg('subsystem')::text)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: DeprecateCommand :one
UPDATE command_defs
SET
    active     = false,
    updated_at = now(),
    updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
