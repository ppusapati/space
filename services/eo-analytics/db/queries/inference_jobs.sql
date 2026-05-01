-- name: CreateInferenceJob :one
INSERT INTO inference_jobs (id, tenant_id, model_id, item_id, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $6)
RETURNING *;

-- name: GetInferenceJob :one
SELECT * FROM inference_jobs WHERE id = $1;

-- name: CountInferenceJobsForTenant :one
SELECT COUNT(*)::bigint AS total FROM inference_jobs
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('status')::int IS NULL OR status = sqlc.narg('status')::int);

-- name: ListInferenceJobsForTenant :many
SELECT * FROM inference_jobs
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('status')::int IS NULL OR status = sqlc.narg('status')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: UpdateInferenceJobStatus :one
UPDATE inference_jobs
SET
    status        = sqlc.arg('status')::int,
    output_uri    = COALESCE(NULLIF(sqlc.arg('output_uri')::text, ''), output_uri),
    error_message = COALESCE(NULLIF(sqlc.arg('error_message')::text, ''), error_message),
    started_at    = CASE WHEN sqlc.arg('status')::int = 2 THEN COALESCE(started_at, now()) ELSE started_at END,
    finished_at   = CASE WHEN sqlc.arg('status')::int IN (3, 4) THEN now() ELSE finished_at END,
    updated_at    = now(),
    updated_by    = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
