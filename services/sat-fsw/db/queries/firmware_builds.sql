-- name: RegisterFirmwareBuild :one
INSERT INTO firmware_builds (
    id, tenant_id, target_platform, subsystem, version, git_sha,
    artefact_uri, artefact_size_bytes, artefact_sha256, status, notes,
    created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $12
)
RETURNING *;

-- name: GetFirmwareBuild :one
SELECT * FROM firmware_builds WHERE id = $1;

-- name: CountFirmwareBuildsForTenant :one
SELECT COUNT(*)::bigint AS total FROM firmware_builds
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('subsystem')::text IS NULL OR subsystem = sqlc.narg('subsystem')::text)
  AND (sqlc.narg('status')::int    IS NULL OR status    = sqlc.narg('status')::int);

-- name: ListFirmwareBuildsForTenant :many
SELECT * FROM firmware_builds
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('subsystem')::text IS NULL OR subsystem = sqlc.narg('subsystem')::text)
  AND (sqlc.narg('status')::int    IS NULL OR status    = sqlc.narg('status')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: UpdateFirmwareBuildStatus :one
UPDATE firmware_builds
SET
    status     = sqlc.arg('status')::int,
    notes      = COALESCE(NULLIF(sqlc.arg('notes')::text, ''), notes),
    updated_at = now(),
    updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
