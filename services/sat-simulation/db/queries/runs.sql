-- name: StartRun :one
INSERT INTO simulation_runs (
    id, tenant_id, satellite_id, scenario_id, mode, status,
    parameters_json, created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $8
)
RETURNING *;

-- name: GetRun :one
SELECT * FROM simulation_runs WHERE id = $1;

-- name: CountRunsForTenant :one
SELECT COUNT(*)::bigint AS total FROM simulation_runs
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('satellite_id')::uuid IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('scenario_id')::uuid  IS NULL OR scenario_id  = sqlc.narg('scenario_id')::uuid)
  AND (sqlc.narg('status')::int        IS NULL OR status       = sqlc.narg('status')::int)
  AND (sqlc.narg('mode')::int          IS NULL OR mode         = sqlc.narg('mode')::int);

-- name: ListRunsForTenant :many
SELECT * FROM simulation_runs
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('satellite_id')::uuid IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('scenario_id')::uuid  IS NULL OR scenario_id  = sqlc.narg('scenario_id')::uuid)
  AND (sqlc.narg('status')::int        IS NULL OR status       = sqlc.narg('status')::int)
  AND (sqlc.narg('mode')::int          IS NULL OR mode         = sqlc.narg('mode')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: UpdateRunStatus :one
UPDATE simulation_runs
SET
    status        = sqlc.arg('status')::int,
    log_uri       = COALESCE(NULLIF(sqlc.arg('log_uri')::text, ''), log_uri),
    telemetry_uri = COALESCE(NULLIF(sqlc.arg('telemetry_uri')::text, ''), telemetry_uri),
    results_json  = CASE
                      WHEN sqlc.arg('results_json')::text = '' THEN results_json
                      ELSE sqlc.arg('results_json')::jsonb
                    END,
    score         = CASE WHEN sqlc.arg('score')::double precision = 0 THEN score
                         ELSE sqlc.arg('score')::double precision END,
    error_message = COALESCE(NULLIF(sqlc.arg('error_message')::text, ''), error_message),
    started_at    = CASE WHEN sqlc.arg('status')::int = 2 THEN COALESCE(started_at, now()) ELSE started_at END,
    finished_at   = CASE WHEN sqlc.arg('status')::int IN (3, 4, 5) THEN now() ELSE finished_at END,
    updated_at    = now(),
    updated_by    = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
