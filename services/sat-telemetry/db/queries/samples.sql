-- name: InsertSample :exec
INSERT INTO telemetry_samples (
    id, tenant_id, satellite_id, frame_id, channel_id,
    sample_time, value_double, value_int, value_bool, value_text
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: QuerySamples :many
SELECT id, tenant_id, satellite_id, frame_id, channel_id,
       sample_time, value_double, value_int, value_bool, value_text, ingested_at
FROM telemetry_samples
WHERE tenant_id  = sqlc.arg('tenant_id')::uuid
  AND channel_id = sqlc.arg('channel_id')::uuid
  AND (sqlc.narg('time_start')::timestamptz IS NULL OR sample_time >= sqlc.narg('time_start')::timestamptz)
  AND (sqlc.narg('time_end')::timestamptz   IS NULL OR sample_time <= sqlc.narg('time_end')::timestamptz)
ORDER BY sample_time ASC, id ASC
LIMIT sqlc.arg('lim')::int;
