-- name: InsertTelemetryFrame :one
INSERT INTO telemetry_frames (
    id, tenant_id, satellite_id, apid, virtual_channel, sequence_count,
    sat_time, payload_size_bytes, payload_sha256, frame_type, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetFrame :one
SELECT * FROM telemetry_frames WHERE id = $1;

-- name: ListFrames :many
SELECT * FROM telemetry_frames
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('satellite_id')::uuid IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
  AND (sqlc.narg('frame_type')::text   IS NULL OR frame_type   = sqlc.narg('frame_type')::text)
  AND (sqlc.narg('time_start')::timestamptz IS NULL OR ground_time >= sqlc.narg('time_start')::timestamptz)
  AND (sqlc.narg('time_end')::timestamptz   IS NULL OR ground_time <= sqlc.narg('time_end')::timestamptz)
  AND (
        sqlc.narg('cursor_ground_time')::timestamptz IS NULL
        OR (ground_time, id) < (sqlc.narg('cursor_ground_time')::timestamptz,
                                sqlc.arg('cursor_id')::uuid)
      )
ORDER BY ground_time DESC, id DESC
LIMIT sqlc.arg('lim')::int;
