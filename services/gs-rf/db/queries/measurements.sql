-- name: RecordMeasurement :one
INSERT INTO link_measurements (
    id, tenant_id, pass_id, station_id, antenna_id, sampled_at,
    rssi_dbm, snr_db, ber, fer, frequency_hz, doppler_shift_hz, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetMeasurement :one
SELECT * FROM link_measurements WHERE id = $1;

-- name: CountMeasurementsForTenant :one
SELECT COUNT(*)::bigint AS total FROM link_measurements
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('pass_id')::uuid    IS NULL OR pass_id    = sqlc.narg('pass_id')::uuid)
  AND (sqlc.narg('station_id')::uuid IS NULL OR station_id = sqlc.narg('station_id')::uuid)
  AND (sqlc.narg('time_start')::timestamptz IS NULL OR sampled_at >= sqlc.narg('time_start')::timestamptz)
  AND (sqlc.narg('time_end')::timestamptz   IS NULL OR sampled_at <= sqlc.narg('time_end')::timestamptz);

-- name: ListMeasurementsForTenant :many
SELECT * FROM link_measurements
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('pass_id')::uuid    IS NULL OR pass_id    = sqlc.narg('pass_id')::uuid)
  AND (sqlc.narg('station_id')::uuid IS NULL OR station_id = sqlc.narg('station_id')::uuid)
  AND (sqlc.narg('time_start')::timestamptz IS NULL OR sampled_at >= sqlc.narg('time_start')::timestamptz)
  AND (sqlc.narg('time_end')::timestamptz   IS NULL OR sampled_at <= sqlc.narg('time_end')::timestamptz)
ORDER BY sampled_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;
