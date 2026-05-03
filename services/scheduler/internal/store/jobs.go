// Package store implements the scheduler's persistence: jobs +
// job_runs CRUD.

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/scheduler/internal/cron"
)

// RetryPolicy is the JSONB shape stored on jobs.retry_policy.
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	BackoffS    int           `json:"backoff_s"`
}

// DefaultRetryPolicy is the chetana-wide fallback.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{MaxAttempts: 1, BackoffS: 0}
}

// Backoff returns the duration the runner should wait before the
// next attempt at `attempt` (1-indexed). Linear backoff in v1;
// exponential variants land later via a `Strategy` field.
func (p RetryPolicy) Backoff(attempt int) time.Duration {
	if attempt <= 1 || p.BackoffS <= 0 {
		return 0
	}
	return time.Duration(p.BackoffS*(attempt-1)) * time.Second
}

// Job is the in-memory shape of one jobs row.
type Job struct {
	ID          string
	TenantID    string
	Name        string
	Description string
	Schedule    string // cron expression
	Timezone    string
	Enabled     bool
	TimeoutS    int
	RetryPolicy RetryPolicy
	Payload     json.RawMessage
	LastRunAt   sql.NullTime
	NextRunAt   sql.NullTime
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Run is the in-memory shape of one job_runs row.
type Run struct {
	ID           int64
	JobID        string
	TenantID     string
	RunnerID     string
	StartedAt    time.Time
	FinishedAt   sql.NullTime
	Status       string
	ExitCode     int
	Output       string
	ErrorExcerpt string
	Attempt      int
	Trigger      string // "cron" | "manual"
}

// Status constants. Aligned with the migration's CHECK enum.
const (
	StatusRunning   = "running"
	StatusSucceeded = "succeeded"
	StatusFailed    = "failed"
	StatusTimeout   = "timeout"
	StatusSkipped   = "skipped"
)

// Trigger constants.
const (
	TriggerCron   = "cron"
	TriggerManual = "manual"
)

// JobStore wraps a pgxpool.Pool with the job CRUD helpers.
type JobStore struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewJobStore wraps a pool. clock=nil → time.Now.
func NewJobStore(pool *pgxpool.Pool, clock func() time.Time) (*JobStore, error) {
	if pool == nil {
		return nil, errors.New("store: nil pool")
	}
	if clock == nil {
		clock = time.Now
	}
	return &JobStore{pool: pool, clk: clock}, nil
}

// CreateInput is the per-call input for Create.
type CreateInput struct {
	TenantID    string
	Name        string
	Description string
	Schedule    string // cron; "" → manual-only
	Timezone    string // "" → UTC
	TimeoutS    int    // 0 → 60
	RetryPolicy RetryPolicy
	Payload     any
}

// Create inserts a new job. Validates the cron expression so
// users can't ship a schedule that the runner can't parse later.
func (s *JobStore) Create(ctx context.Context, in CreateInput) (string, error) {
	if in.TenantID == "" || in.Name == "" {
		return "", errors.New("store: TenantID + Name required")
	}
	if in.TimeoutS <= 0 {
		in.TimeoutS = 60
	}
	if in.RetryPolicy == (RetryPolicy{}) {
		in.RetryPolicy = DefaultRetryPolicy()
	}
	var nextRunAt sql.NullTime
	if in.Schedule != "" {
		sched, err := cron.Parse(in.Schedule, in.Timezone)
		if err != nil {
			return "", err
		}
		nextRunAt = sql.NullTime{Time: sched.Next(s.clk().UTC()), Valid: true}
	}
	policyBytes, err := json.Marshal(in.RetryPolicy)
	if err != nil {
		return "", fmt.Errorf("store: marshal policy: %w", err)
	}
	payloadBytes, err := json.Marshal(in.Payload)
	if err != nil {
		return "", fmt.Errorf("store: marshal payload: %w", err)
	}
	if string(payloadBytes) == "null" {
		payloadBytes = []byte("{}")
	}
	const q = `
INSERT INTO jobs
  (tenant_id, name, description, schedule, timezone, timeout_s,
   retry_policy, payload, next_run_at)
VALUES ($1, $2, $3, $4, COALESCE(NULLIF($5,''),'UTC'), $6, $7, $8, $9)
RETURNING id
`
	var id string
	if err := s.pool.QueryRow(ctx, q,
		in.TenantID, in.Name, in.Description,
		in.Schedule, in.Timezone, in.TimeoutS,
		policyBytes, payloadBytes, nextRunAt,
	).Scan(&id); err != nil {
		return "", fmt.Errorf("store: insert: %w", err)
	}
	return id, nil
}

// SetEnabled flips the enabled bit. Used by the admin "pause"
// endpoint (acceptance #3 — toggles take effect immediately).
func (s *JobStore) SetEnabled(ctx context.Context, jobID string, enabled bool) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE jobs SET enabled = $2, updated_at = $3 WHERE id = $1
`, jobID, enabled, s.clk().UTC())
	if err != nil {
		return fmt.Errorf("store: set enabled: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrJobNotFound
	}
	return nil
}

// Get returns a single job.
func (s *JobStore) Get(ctx context.Context, jobID string) (*Job, error) {
	const q = `
SELECT id, tenant_id, name, description, schedule, timezone, enabled,
       timeout_s, retry_policy, payload, last_run_at, next_run_at,
       created_at, updated_at
FROM jobs WHERE id = $1
`
	var (
		j           Job
		policyRaw   []byte
	)
	err := s.pool.QueryRow(ctx, q, jobID).Scan(
		&j.ID, &j.TenantID, &j.Name, &j.Description, &j.Schedule, &j.Timezone, &j.Enabled,
		&j.TimeoutS, &policyRaw, &j.Payload, &j.LastRunAt, &j.NextRunAt,
		&j.CreatedAt, &j.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrJobNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("store: get: %w", err)
	}
	if err := json.Unmarshal(policyRaw, &j.RetryPolicy); err != nil {
		return nil, fmt.Errorf("store: parse policy: %w", err)
	}
	return &j, nil
}

// DueBefore returns the set of enabled jobs whose next_run_at is
// at or before `cutoff`. Workers iterate this list, attempting the
// per-job lock — only one wins per tick.
func (s *JobStore) DueBefore(ctx context.Context, cutoff time.Time, limit int) ([]Job, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
SELECT id, tenant_id, name, description, schedule, timezone, enabled,
       timeout_s, retry_policy, payload, last_run_at, next_run_at,
       created_at, updated_at
FROM jobs
WHERE enabled = true
  AND schedule <> ''
  AND next_run_at IS NOT NULL
  AND next_run_at <= $1
ORDER BY next_run_at ASC
LIMIT $2
`, cutoff.UTC(), limit)
	if err != nil {
		return nil, fmt.Errorf("store: due: %w", err)
	}
	defer rows.Close()
	var out []Job
	for rows.Next() {
		var (
			j         Job
			policyRaw []byte
		)
		if err := rows.Scan(
			&j.ID, &j.TenantID, &j.Name, &j.Description, &j.Schedule, &j.Timezone, &j.Enabled,
			&j.TimeoutS, &policyRaw, &j.Payload, &j.LastRunAt, &j.NextRunAt,
			&j.CreatedAt, &j.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("store: scan: %w", err)
		}
		_ = json.Unmarshal(policyRaw, &j.RetryPolicy)
		out = append(out, j)
	}
	return out, rows.Err()
}

// AdvanceNext recomputes next_run_at + stamps last_run_at after a
// successful tick.
func (s *JobStore) AdvanceNext(ctx context.Context, jobID string, ranAt time.Time, nextAt time.Time) error {
	tag, err := s.pool.Exec(ctx, `
UPDATE jobs SET last_run_at = $2, next_run_at = $3, updated_at = $4
WHERE id = $1
`, jobID, ranAt.UTC(), nextAt.UTC(), s.clk().UTC())
	if err != nil {
		return fmt.Errorf("store: advance: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrJobNotFound
	}
	return nil
}

// StartRun inserts a job_runs row in the running state. Returns
// the assigned id.
func (s *JobStore) StartRun(ctx context.Context, jobID, runnerID, trigger string, attempt int) (int64, error) {
	if trigger == "" {
		trigger = TriggerCron
	}
	var (
		runID    int64
		tenantID string
	)
	if err := s.pool.QueryRow(ctx,
		`SELECT tenant_id FROM jobs WHERE id = $1`, jobID,
	).Scan(&tenantID); err != nil {
		return 0, fmt.Errorf("store: tenant lookup: %w", err)
	}
	if err := s.pool.QueryRow(ctx, `
INSERT INTO job_runs (job_id, tenant_id, runner_id, attempt, trigger)
VALUES ($1, $2, $3, $4, $5)
RETURNING id
`, jobID, tenantID, runnerID, attempt, trigger).Scan(&runID); err != nil {
		return 0, fmt.Errorf("store: start run: %w", err)
	}
	return runID, nil
}

// FinishRunInput is the per-call payload for FinishRun.
type FinishRunInput struct {
	Status       string
	ExitCode     int
	Output       string
	ErrorExcerpt string
}

// FinishRun stamps the run as completed.
func (s *JobStore) FinishRun(ctx context.Context, runID int64, in FinishRunInput) error {
	if _, err := s.pool.Exec(ctx, `
UPDATE job_runs
SET finished_at = $2, status = $3, exit_code = $4, output = $5, error_excerpt = $6
WHERE id = $1
`, runID, s.clk().UTC(), in.Status, in.ExitCode, in.Output, in.ErrorExcerpt); err != nil {
		return fmt.Errorf("store: finish run: %w", err)
	}
	return nil
}

// History returns the most recent N runs for a job.
func (s *JobStore) History(ctx context.Context, jobID string, limit int) ([]Run, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.pool.Query(ctx, `
SELECT id, job_id, tenant_id, runner_id, started_at, finished_at,
       status, exit_code, output, error_excerpt, attempt, trigger
FROM job_runs
WHERE job_id = $1
ORDER BY started_at DESC
LIMIT $2
`, jobID, limit)
	if err != nil {
		return nil, fmt.Errorf("store: history: %w", err)
	}
	defer rows.Close()
	var out []Run
	for rows.Next() {
		var r Run
		if err := rows.Scan(
			&r.ID, &r.JobID, &r.TenantID, &r.RunnerID, &r.StartedAt, &r.FinishedAt,
			&r.Status, &r.ExitCode, &r.Output, &r.ErrorExcerpt, &r.Attempt, &r.Trigger,
		); err != nil {
			return nil, fmt.Errorf("store: scan run: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrJobNotFound is returned by Get/SetEnabled/AdvanceNext when no row matches.
var ErrJobNotFound = errors.New("store: job not found")
