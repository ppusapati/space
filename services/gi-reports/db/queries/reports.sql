-- name: GenerateReport :one
INSERT INTO reports (
    id, tenant_id, template_id, status, parameters_json, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $6)
RETURNING *;

-- name: GetReport :one
SELECT * FROM reports WHERE id = $1;

-- name: CountReportsForTenant :one
SELECT COUNT(*)::bigint AS total FROM reports
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('template_id')::uuid IS NULL OR template_id = sqlc.narg('template_id')::uuid)
  AND (sqlc.narg('status')::int       IS NULL OR status      = sqlc.narg('status')::int);

-- name: ListReportsForTenant :many
SELECT * FROM reports
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('template_id')::uuid IS NULL OR template_id = sqlc.narg('template_id')::uuid)
  AND (sqlc.narg('status')::int       IS NULL OR status      = sqlc.narg('status')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: UpdateReportStatus :one
UPDATE reports
SET
    status        = sqlc.arg('status')::int,
    output_uri    = COALESCE(NULLIF(sqlc.arg('output_uri')::text, ''), output_uri),
    error_message = COALESCE(NULLIF(sqlc.arg('error_message')::text, ''), error_message),
    started_at    = CASE WHEN sqlc.arg('status')::int = 2 THEN COALESCE(started_at, now()) ELSE started_at END,
    finished_at   = CASE WHEN sqlc.arg('status')::int IN (3, 4, 5) THEN COALESCE(finished_at, now()) ELSE finished_at END,
    updated_at    = now(),
    updated_by    = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
