-- name: SubmitAnalysisJob :one
INSERT INTO analysis_jobs (
    id, tenant_id, type, status, input_uris, parameters_json,
    created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
RETURNING *;

-- name: GetAnalysisJob :one
SELECT * FROM analysis_jobs WHERE id = $1;

-- name: CountAnalysisJobsForTenant :one
SELECT COUNT(*)::bigint AS total FROM analysis_jobs
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('status')::int IS NULL OR status = sqlc.narg('status')::int)
  AND (sqlc.narg('type')::int   IS NULL OR type   = sqlc.narg('type')::int);

-- name: ListAnalysisJobsForTenant :many
SELECT * FROM analysis_jobs
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('status')::int IS NULL OR status = sqlc.narg('status')::int)
  AND (sqlc.narg('type')::int   IS NULL OR type   = sqlc.narg('type')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: UpdateAnalysisJobStatus :one
UPDATE analysis_jobs
SET
    status               = sqlc.arg('status')::int,
    output_uri           = COALESCE(NULLIF(sqlc.arg('output_uri')::text, ''), output_uri),
    results_summary_json = CASE WHEN sqlc.arg('results_summary_json')::text = ''
                                THEN results_summary_json
                                ELSE sqlc.arg('results_summary_json')::jsonb
                           END,
    error_message        = COALESCE(NULLIF(sqlc.arg('error_message')::text, ''), error_message),
    started_at           = CASE WHEN sqlc.arg('status')::int = 2 THEN COALESCE(started_at, now()) ELSE started_at END,
    finished_at          = CASE WHEN sqlc.arg('status')::int IN (3, 4, 5) THEN COALESCE(finished_at, now()) ELSE finished_at END,
    updated_at           = now(),
    updated_by           = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
