-- name: StartIngestSession :one
INSERT INTO ingest_sessions (
    id, tenant_id, booking_id, pass_id, station_id, satellite_id,
    status, started_at, created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, now(), $8, $8)
RETURNING *;

-- name: GetIngestSession :one
SELECT * FROM ingest_sessions WHERE id = $1;

-- name: CountIngestSessionsForTenant :one
SELECT COUNT(*)::bigint AS total FROM ingest_sessions
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('station_id')::uuid    IS NULL OR station_id   = sqlc.narg('station_id')::uuid)
  AND (sqlc.narg('satellite_id')::uuid  IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('status')::int         IS NULL OR status       = sqlc.narg('status')::int);

-- name: ListIngestSessionsForTenant :many
SELECT * FROM ingest_sessions
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('station_id')::uuid    IS NULL OR station_id   = sqlc.narg('station_id')::uuid)
  AND (sqlc.narg('satellite_id')::uuid  IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('status')::int         IS NULL OR status       = sqlc.narg('status')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: UpdateIngestStatus :one
UPDATE ingest_sessions
SET
    status        = sqlc.arg('status')::int,
    error_message = COALESCE(NULLIF(sqlc.arg('error_message')::text, ''), error_message),
    completed_at  = CASE WHEN sqlc.arg('status')::int IN (3, 4, 5) THEN COALESCE(completed_at, now()) ELSE completed_at END,
    updated_at    = now(),
    updated_by    = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;

-- name: BumpIngestCounters :exec
UPDATE ingest_sessions
SET
    frames_received = frames_received + 1,
    bytes_received  = bytes_received  + sqlc.arg('bytes')::bigint,
    updated_at      = now()
WHERE id = sqlc.arg('id')::uuid;
