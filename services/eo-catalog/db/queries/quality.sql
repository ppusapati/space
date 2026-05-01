-- name: RecordQuality :one
INSERT INTO quality_results (id, item_id, cloud_cover, radiometric_rmse, geometric_accuracy_m, notes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListQualityForItem :many
SELECT * FROM quality_results
WHERE item_id = $1
ORDER BY computed_at DESC, id DESC
LIMIT 100;
