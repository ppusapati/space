-- name: RegisterModel :one
INSERT INTO models (id, tenant_id, name, version, task, framework, artefact_uri, metadata_json, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
RETURNING *;

-- name: GetModel :one
SELECT * FROM models WHERE id = $1;

-- name: ListModels :many
SELECT * FROM models
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('task')::int IS NULL OR task = sqlc.narg('task')::int)
  AND (
        sqlc.narg('cursor_created_at')::timestamptz IS NULL
        OR (created_at, id) < (sqlc.narg('cursor_created_at')::timestamptz,
                               sqlc.arg('cursor_id')::uuid)
      )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('lim')::int;

-- name: DeactivateModel :one
UPDATE models SET active = false, updated_at = now(), updated_by = $2 WHERE id = $1
RETURNING *;
