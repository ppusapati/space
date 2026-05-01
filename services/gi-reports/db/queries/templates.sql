-- name: CreateTemplate :one
INSERT INTO report_templates (
    id, tenant_id, slug, name, description, template_uri, format,
    parameters_schema, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
RETURNING *;

-- name: GetTemplate :one
SELECT * FROM report_templates WHERE id = $1;

-- name: CountTemplatesForTenant :one
SELECT COUNT(*)::bigint AS total FROM report_templates
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('format')::int IS NULL OR format = sqlc.narg('format')::int);

-- name: ListTemplatesForTenant :many
SELECT * FROM report_templates
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('format')::int IS NULL OR format = sqlc.narg('format')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: DeprecateTemplate :one
UPDATE report_templates
SET
    active     = false,
    updated_at = now(),
    updated_by = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
