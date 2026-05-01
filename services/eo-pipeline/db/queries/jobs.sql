-- name: CreateJob :one
INSERT INTO jobs (
    id, tenant_id, item_id, stage, status, parameters_json, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs WHERE id = $1;

-- name: CountJobsForTenant :one
SELECT COUNT(*)::bigint AS total FROM jobs
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('status')::int IS NULL OR status = sqlc.narg('status')::int)
  AND (sqlc.narg('stage')::int  IS NULL OR stage  = sqlc.narg('stage')::int);

-- name: ListJobsForTenant :many
SELECT * FROM jobs
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('status')::int IS NULL OR status = sqlc.narg('status')::int)
  AND (sqlc.narg('stage')::int  IS NULL OR stage  = sqlc.narg('stage')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: UpdateJobStatus :one
UPDATE jobs
SET
    status        = sqlc.arg('status')::int,
    output_uri    = COALESCE(NULLIF(sqlc.arg('output_uri')::text, ''), output_uri),
    error_message = COALESCE(NULLIF(sqlc.arg('error_message')::text, ''), error_message),
    started_at    = CASE WHEN sqlc.arg('status')::int = 2 THEN COALESCE(started_at, now()) ELSE started_at END,
    finished_at   = CASE WHEN sqlc.arg('status')::int IN (3, 4, 5) THEN now() ELSE finished_at END,
    updated_at    = now(),
    updated_by    = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
