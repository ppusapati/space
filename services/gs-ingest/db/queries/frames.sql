-- name: RecordDownlinkFrame :one
INSERT INTO downlink_frames (
    id, tenant_id, session_id, apid, virtual_channel, sequence_count,
    ground_time, payload_size_bytes, payload_sha256, payload_uri, frame_type,
    created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: GetDownlinkFrame :one
SELECT * FROM downlink_frames WHERE id = $1;

-- name: CountDownlinkFramesForTenant :one
SELECT COUNT(*)::bigint AS total FROM downlink_frames
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('session_id')::uuid IS NULL OR session_id = sqlc.narg('session_id')::uuid)
  AND (sqlc.narg('frame_type')::text IS NULL OR frame_type = sqlc.narg('frame_type')::text)
  AND (sqlc.narg('time_start')::timestamptz IS NULL OR ground_time >= sqlc.narg('time_start')::timestamptz)
  AND (sqlc.narg('time_end')::timestamptz   IS NULL OR ground_time <= sqlc.narg('time_end')::timestamptz);

-- name: ListDownlinkFramesForTenant :many
SELECT * FROM downlink_frames
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('session_id')::uuid IS NULL OR session_id = sqlc.narg('session_id')::uuid)
  AND (sqlc.narg('frame_type')::text IS NULL OR frame_type = sqlc.narg('frame_type')::text)
  AND (sqlc.narg('time_start')::timestamptz IS NULL OR ground_time >= sqlc.narg('time_start')::timestamptz)
  AND (sqlc.narg('time_end')::timestamptz   IS NULL OR ground_time <= sqlc.narg('time_end')::timestamptz)
ORDER BY ground_time DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;
