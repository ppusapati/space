// Package queue implements the chetana export-job queue.
//
// → REQ-FUNC-CMN-005; design.md §3.1, §5.2.
//
// Postgres-backed queue with lease-based worker checkout.
// `FOR UPDATE SKIP LOCKED` lets N workers pull jobs in parallel
// without a broker; the lease (leased_until) lets a crashed
// worker's job get re-picked within `lease_ttl + jitter`
// (acceptance #2).

package queue

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Status constants. Aligned with the migration's CHECK enum.
const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	StatusExpired   = "expired"
)

// Job is one row of export_jobs.
type Job struct {
	ID             string
	TenantID       string
	RequestedBy    string
	Kind           string
	Payload        json.RawMessage
	Status         string
	LeasedBy       string
	LeasedUntil    sql.NullTime
	Attempts       int
	MaxAttempts    int
	LastError      string
	S3Bucket       string
	S3Key          string
	PresignedURL   string
	PresignedUntil sql.NullTime
	BytesTotal     int64
	EnqueuedAt     time.Time
	StartedAt      sql.NullTime
	CompletedAt    sql.NullTime
	ExpiresAt      time.Time
}

// EnqueueInput is the per-call input for Enqueue.
type EnqueueInput struct {
	TenantID    string
	RequestedBy string
	Kind        string
	Payload     any
	MaxAttempts int           // 0 → 5
	RetainFor   time.Duration // 0 → 7 days (matches the column default)
}

// Store wraps a pgxpool.Pool with the queue helpers.
type Store struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewStore wraps a pool. clock=nil → time.Now.
func NewStore(pool *pgxpool.Pool, clock func() time.Time) (*Store, error) {
	if pool == nil {
		return nil, errors.New("queue: nil pool")
	}
	if clock == nil {
		clock = time.Now
	}
	return &Store{pool: pool, clk: clock}, nil
}

// Enqueue inserts a queued job. Returns the assigned id.
func (s *Store) Enqueue(ctx context.Context, in EnqueueInput) (string, error) {
	if in.TenantID == "" || in.Kind == "" {
		return "", errors.New("queue: TenantID + Kind are required")
	}
	maxAttempts := in.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	retain := in.RetainFor
	if retain <= 0 {
		retain = 7 * 24 * time.Hour
	}
	payloadBytes, err := json.Marshal(in.Payload)
	if err != nil {
		return "", fmt.Errorf("queue: marshal payload: %w", err)
	}
	if string(payloadBytes) == "null" {
		payloadBytes = []byte("{}")
	}
	now := s.clk().UTC()
	const q = `
INSERT INTO export_jobs
  (tenant_id, requested_by, kind, payload, max_attempts, expires_at, enqueued_at)
VALUES ($1, NULLIF($2,'')::uuid, $3, $4, $5, $6, $7)
RETURNING id
`
	var id string
	if err := s.pool.QueryRow(ctx, q,
		in.TenantID, in.RequestedBy, in.Kind, payloadBytes, maxAttempts,
		now.Add(retain), now,
	).Scan(&id); err != nil {
		return "", fmt.Errorf("queue: insert: %w", err)
	}
	return id, nil
}

// Checkout claims one queued job (or one whose lease has elapsed)
// for the supplied worker id. Returns nil + nil when no job is
// available — the worker should sleep and retry.
//
// `leaseTTL` is how long the worker holds the lease before another
// worker can re-claim it. The processor MUST call ExtendLease for
// long-running jobs that would otherwise blow past the TTL.
func (s *Store) Checkout(ctx context.Context, workerID string, leaseTTL time.Duration) (*Job, error) {
	if workerID == "" {
		return nil, errors.New("queue: empty workerID")
	}
	if leaseTTL <= 0 {
		leaseTTL = 60 * time.Second
	}
	now := s.clk().UTC()
	leasedUntil := now.Add(leaseTTL)

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("queue: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const lockQ = `
SELECT id FROM export_jobs
WHERE (status = 'queued')
   OR (status = 'running' AND leased_until < $1)
ORDER BY enqueued_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`
	var jobID string
	err = tx.QueryRow(ctx, lockQ, now).Scan(&jobID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("queue: lock: %w", err)
	}

	const claimQ = `
UPDATE export_jobs
SET status = 'running',
    leased_by = $2,
    leased_until = $3,
    attempts = attempts + 1,
    started_at = COALESCE(started_at, $4)
WHERE id = $1
RETURNING id, tenant_id, COALESCE(requested_by::text, ''), kind, payload,
          status, leased_by, leased_until,
          attempts, max_attempts, last_error,
          s3_bucket, s3_key, presigned_url, presigned_until,
          bytes_total, enqueued_at, started_at, completed_at, expires_at
`
	var job Job
	if err := tx.QueryRow(ctx, claimQ, jobID, workerID, leasedUntil, now).Scan(
		&job.ID, &job.TenantID, &job.RequestedBy, &job.Kind, &job.Payload,
		&job.Status, &job.LeasedBy, &job.LeasedUntil,
		&job.Attempts, &job.MaxAttempts, &job.LastError,
		&job.S3Bucket, &job.S3Key, &job.PresignedURL, &job.PresignedUntil,
		&job.BytesTotal, &job.EnqueuedAt, &job.StartedAt, &job.CompletedAt, &job.ExpiresAt,
	); err != nil {
		return nil, fmt.Errorf("queue: claim: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("queue: commit: %w", err)
	}
	return &job, nil
}

// ExtendLease pushes leased_until forward for a still-running job.
// Long-running processors call this every (leaseTTL/2) so a
// crashed worker is detected within `leaseTTL + jitter`.
func (s *Store) ExtendLease(ctx context.Context, jobID, workerID string, leaseTTL time.Duration) error {
	if jobID == "" || workerID == "" {
		return errors.New("queue: empty jobID / workerID")
	}
	if leaseTTL <= 0 {
		leaseTTL = 60 * time.Second
	}
	tag, err := s.pool.Exec(ctx, `
UPDATE export_jobs
SET leased_until = $3
WHERE id = $1 AND leased_by = $2 AND status = 'running'
`, jobID, workerID, s.clk().UTC().Add(leaseTTL))
	if err != nil {
		return fmt.Errorf("queue: extend lease: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrLeaseLost
	}
	return nil
}

// CompleteOutput is the per-call payload for Complete.
type CompleteOutput struct {
	S3Bucket       string
	S3Key          string
	PresignedURL   string
	PresignedUntil time.Time
	BytesTotal     int64
}

// Complete marks a job succeeded and stores the output pointer.
func (s *Store) Complete(ctx context.Context, jobID, workerID string, out CompleteOutput) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE export_jobs
SET status = 'succeeded',
    leased_by = '',
    leased_until = NULL,
    completed_at = $3,
    s3_bucket = $4,
    s3_key = $5,
    presigned_url = $6,
    presigned_until = $7,
    bytes_total = $8,
    last_error = ''
WHERE id = $1 AND leased_by = $2 AND status = 'running'
`, jobID, workerID, s.clk().UTC(),
		out.S3Bucket, out.S3Key, out.PresignedURL, out.PresignedUntil, out.BytesTotal)
	if err != nil {
		return fmt.Errorf("queue: complete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrLeaseLost
	}
	return nil
}

// Fail marks a job failed (or queued for retry when attempts <
// max_attempts).
func (s *Store) Fail(ctx context.Context, jobID, workerID, errMsg string) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE export_jobs
SET status = CASE
        WHEN attempts >= max_attempts THEN 'failed'
        ELSE 'queued'
    END,
    leased_by = '',
    leased_until = NULL,
    last_error = $3,
    completed_at = CASE WHEN attempts >= max_attempts THEN $4 ELSE NULL END
WHERE id = $1 AND leased_by = $2 AND status = 'running'
`, jobID, workerID, errMsg, s.clk().UTC())
	if err != nil {
		return fmt.Errorf("queue: fail: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrLeaseLost
	}
	return nil
}

// Get returns a single job by id. Used by the read endpoint
// (`GET /v1/export/jobs/{id}`).
func (s *Store) Get(ctx context.Context, jobID string) (*Job, error) {
	const q = `
SELECT id, tenant_id, COALESCE(requested_by::text, ''), kind, payload,
       status, leased_by, leased_until,
       attempts, max_attempts, last_error,
       s3_bucket, s3_key, presigned_url, presigned_until,
       bytes_total, enqueued_at, started_at, completed_at, expires_at
FROM export_jobs WHERE id = $1
`
	var job Job
	err := s.pool.QueryRow(ctx, q, jobID).Scan(
		&job.ID, &job.TenantID, &job.RequestedBy, &job.Kind, &job.Payload,
		&job.Status, &job.LeasedBy, &job.LeasedUntil,
		&job.Attempts, &job.MaxAttempts, &job.LastError,
		&job.S3Bucket, &job.S3Key, &job.PresignedURL, &job.PresignedUntil,
		&job.BytesTotal, &job.EnqueuedAt, &job.StartedAt, &job.CompletedAt, &job.ExpiresAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrJobNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("queue: get: %w", err)
	}
	return &job, nil
}

// ListExpired returns the set of jobs past their `expires_at`
// horizon. Used by the cleanup sweep.
func (s *Store) ListExpired(ctx context.Context, limit int) ([]Job, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
SELECT id, tenant_id, kind, s3_bucket, s3_key, expires_at
FROM export_jobs
WHERE expires_at < $1 AND status <> 'expired'
ORDER BY expires_at ASC
LIMIT $2
`, s.clk().UTC(), limit)
	if err != nil {
		return nil, fmt.Errorf("queue: list expired: %w", err)
	}
	defer rows.Close()

	var out []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(&j.ID, &j.TenantID, &j.Kind, &j.S3Bucket, &j.S3Key, &j.ExpiresAt); err != nil {
			return nil, fmt.Errorf("queue: scan: %w", err)
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// MarkExpired flips the status column on a job whose S3 object
// has been deleted by the cleanup sweep.
func (s *Store) MarkExpired(ctx context.Context, jobID string) error {
	if _, err := s.pool.Exec(ctx, `
UPDATE export_jobs SET status = 'expired',
                       leased_by = '',
                       leased_until = NULL
WHERE id = $1
`, jobID); err != nil {
		return fmt.Errorf("queue: mark expired: %w", err)
	}
	return nil
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrLeaseLost is returned by ExtendLease/Complete/Fail when the
// supplied worker no longer holds the job's lease (typically
// because another worker re-claimed it after the previous lease
// elapsed).
var ErrLeaseLost = errors.New("queue: lease lost")

// ErrJobNotFound is returned by Get when no row matches.
var ErrJobNotFound = errors.New("queue: job not found")
