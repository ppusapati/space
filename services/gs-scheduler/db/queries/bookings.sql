-- name: RequestBooking :one
INSERT INTO bookings (
    id, tenant_id, pass_id, priority, status, purpose, notes,
    created_by, updated_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
RETURNING *;

-- name: GetBooking :one
SELECT * FROM bookings WHERE id = $1;

-- name: CountBookingsForTenant :one
SELECT COUNT(*)::bigint AS total FROM bookings
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('status')::int IS NULL OR status = sqlc.narg('status')::int);

-- name: ListBookingsForTenant :many
SELECT * FROM bookings
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (sqlc.narg('status')::int IS NULL OR status = sqlc.narg('status')::int)
ORDER BY created_at DESC, id DESC
OFFSET sqlc.arg('page_offset')::int
LIMIT  sqlc.arg('page_size')::int;

-- name: UpdateBookingStatus :one
UPDATE bookings
SET
    status        = sqlc.arg('status')::int,
    error_message = COALESCE(NULLIF(sqlc.arg('error_message')::text, ''), error_message),
    scheduled_at  = CASE WHEN sqlc.arg('status')::int = 3 THEN COALESCE(scheduled_at, now()) ELSE scheduled_at END,
    completed_at  = CASE WHEN sqlc.arg('status')::int IN (5, 6, 7) THEN COALESCE(completed_at, now()) ELSE completed_at END,
    updated_at    = now(),
    updated_by    = sqlc.arg('updated_by')::text
WHERE id = sqlc.arg('id')::uuid
RETURNING *;
