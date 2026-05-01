-- name: CreateLinkBudget :one
INSERT INTO link_budgets (
    id, tenant_id, pass_id, station_id, antenna_id, satellite_id,
    carrier_freq_hz, tx_power_dbm, tx_gain_dbi, rx_gain_dbi, rx_noise_temp_k,
    bandwidth_hz, slant_range_km, free_space_loss_db, atmospheric_loss_db,
    polarization_loss_db, pointing_loss_db, predicted_eb_n0_db, predicted_snr_db,
    link_margin_db, notes, created_by, updated_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17,
    $18, $19, $20, $21, $22, $22
)
RETURNING *;

-- name: GetLinkBudget :one
SELECT * FROM link_budgets WHERE id = $1;

-- name: CountLinkBudgetsForTenant :one
SELECT COUNT(*)::bigint AS total FROM link_budgets
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('pass_id')::uuid       IS NULL OR pass_id      = sqlc.narg('pass_id')::uuid)
  AND (sqlc.narg('station_id')::uuid    IS NULL OR station_id   = sqlc.narg('station_id')::uuid)
  AND (sqlc.narg('satellite_id')::uuid  IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid);

-- name: ListLinkBudgetsForTenant :many
SELECT * FROM link_budgets
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('pass_id')::uuid       IS NULL OR pass_id      = sqlc.narg('pass_id')::uuid)
  AND (sqlc.narg('station_id')::uuid    IS NULL OR station_id   = sqlc.narg('station_id')::uuid)
  AND (sqlc.narg('satellite_id')::uuid  IS NULL OR satellite_id = sqlc.narg('satellite_id')::uuid)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;
