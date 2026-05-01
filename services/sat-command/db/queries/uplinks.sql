-- name: NextUplinkSequence :one
INSERT INTO uplink_sequences (satellite_id, next_value)
VALUES ($1, 2)
ON CONFLICT (satellite_id) DO UPDATE
    SET next_value = uplink_sequences.next_value + 1
RETURNING (next_value - 1)::bigint AS sequence_number;

-- name: EnqueueUplink :one
INSERT INTO uplink_requests (
    id, tenant_id, satellite_id, command_def_id, parameters_json,
    scheduled_release, status, sequence_number, gateway_id,
    created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $10
)
RETURNING *;

-- name: GetUplink :one
SELECT * FROM uplink_requests WHERE id = $1;

-- name: CountUplinksForTenant :one
SELECT COUNT(*)::bigint AS total FROM uplink_requests
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('satellite_id')::uuid       IS NULL OR satellite_id    = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('status')::int              IS NULL OR status          = sqlc.narg('status')::int)
  AND (sqlc.narg('release_start')::timestamptz IS NULL OR scheduled_release >= sqlc.narg('release_start')::timestamptz)
  AND (sqlc.narg('release_end')::timestamptz   IS NULL OR scheduled_release <= sqlc.narg('release_end')::timestamptz);

-- name: ListUplinksForTenant :many
SELECT * FROM uplink_requests
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('satellite_id')::uuid       IS NULL OR satellite_id    = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('status')::int              IS NULL OR status          = sqlc.narg('status')::int)
  AND (sqlc.narg('release_start')::timestamptz IS NULL OR scheduled_release >= sqlc.narg('release_start')::timestamptz)
  AND (sqlc.narg('release_end')::timestamptz   IS NULL OR scheduled_release <= sqlc.narg('release_end')::timestamptz)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: UpdateUplinkStatus :one
UPDATE uplink_requests
SET
    status        = sqlc.arg('status')::int,
    error_message = COALESCE(NULLIF(sqlc.arg('error_message')::text, ''), error_message),
    released_at   = CASE WHEN sqlc.arg('status')::int = 2 THEN COALESCE(released_at, now())   ELSE released_at  END,
    acked_at      = CASE WHEN sqlc.arg('status')::int = 3 THEN COALESCE(acked_at, now())      ELSE acked_at     END,
    completed_at  = CASE WHEN sqlc.arg('status')::int IN (4, 5, 6) THEN COALESCE(completed_at, now()) ELSE completed_at END,
    updated_at    = now(),
    updated_by    = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
