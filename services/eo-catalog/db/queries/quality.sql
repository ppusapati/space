-- name: RecordQuality :one
INSERT INTO quality_results (id, item_id, cloud_cover, radiometric_rmse, geometric_accuracy_m, notes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListQualityForItem :many
SELECT * FROM quality_results
WHERE item_id = sqlc.arg('item_id')::uuid
  AND (
        sqlc.narg('cursor_computed_at')::timestamptz IS NULL
        OR (computed_at, id) < (sqlc.narg('cursor_computed_at')::timestamptz,
                                sqlc.arg('cursor_id')::uuid)
      )
ORDER BY computed_at DESC, id DESC
LIMIT sqlc.arg('lim')::int;
