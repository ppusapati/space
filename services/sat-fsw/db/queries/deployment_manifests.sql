-- name: CreateDeploymentManifest :one
INSERT INTO deployment_manifests (
    id, tenant_id, satellite_id, manifest_version, status,
    assignments_json, notes, created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $8
)
RETURNING *;

-- name: GetDeploymentManifest :one
SELECT * FROM deployment_manifests WHERE id = $1;

-- name: CountDeploymentManifestsForTenant :one
SELECT COUNT(*)::bigint AS total FROM deployment_manifests
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('satellite_id')::uuid IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('status')::int        IS NULL OR status       = sqlc.narg('status')::int);

-- name: ListDeploymentManifestsForTenant :many
SELECT * FROM deployment_manifests
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('satellite_id')::uuid IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('status')::int        IS NULL OR status       = sqlc.narg('status')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: UpdateDeploymentManifestStatus :one
UPDATE deployment_manifests
SET
    status     = sqlc.arg('status')::int,
    notes      = COALESCE(NULLIF(sqlc.arg('notes')::text, ''), notes),
    updated_at = now(),
    updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
