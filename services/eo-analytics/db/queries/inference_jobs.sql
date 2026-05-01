-- name: CreateInferenceJob :one
INSERT INTO inference_jobs (id, tenant_id, model_id, item_id, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $6)
RETURNING *;

-- name: GetInferenceJob :one
SELECT * FROM inference_jobs WHERE id = $1;

-- name: ListInferenceJobs :many
SELECT * FROM inference_jobs
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('status')::int IS NULL OR status = sqlc.narg('status')::int)
  AND (
        sqlc.narg('cursor_created_at')::timestamptz IS NULL
        OR (created_at, id) < (sqlc.narg('cursor_created_at')::timestamptz,
                               sqlc.arg('cursor_id')::uuid)
      )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('lim')::int;

-- name: UpdateInferenceJobStatus :one
UPDATE inference_jobs
SET
    status        = $2,
    output_uri    = COALESCE(NULLIF($3, ''), output_uri),
    error_message = COALESCE(NULLIF($4, ''), error_message),
    started_at    = CASE WHEN $2 = 2 THEN COALESCE(started_at, now()) ELSE started_at END,
    finished_at   = CASE WHEN $2 IN (3, 4) THEN now() ELSE finished_at END,
    updated_at    = now(),
    updated_by    = $5
WHERE id = $1
RETURNING *;
