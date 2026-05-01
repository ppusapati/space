-- name: RegisterModel :one
INSERT INTO models (id, tenant_id, name, version, task, framework, artefact_uri, metadata_json, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
RETURNING *;

-- name: GetModel :one
SELECT * FROM models WHERE id = $1;

-- name: CountModelsForTenant :one
SELECT COUNT(*)::bigint AS total FROM models
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('task')::int IS NULL OR task = sqlc.narg('task')::int);

-- name: ListModelsForTenant :many
SELECT * FROM models
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('task')::int IS NULL OR task = sqlc.narg('task')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: DeactivateModel :one
UPDATE models
SET active = false, updated_at = now(), updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
