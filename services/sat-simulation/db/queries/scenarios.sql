-- name: CreateScenario :one
INSERT INTO scenarios (
    id, tenant_id, slug, title, description, spec_json, created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $7
)
RETURNING *;

-- name: GetScenario :one
SELECT * FROM scenarios WHERE id = $1;

-- name: CountScenariosForTenant :one
SELECT COUNT(*)::bigint AS total FROM scenarios
WHERE tenant_id = sqlc.arg('tenant_id')::uuid;

-- name: ListScenariosForTenant :many
SELECT * FROM scenarios
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: DeprecateScenario :one
UPDATE scenarios
SET
    active     = false,
    updated_at = now(),
    updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
