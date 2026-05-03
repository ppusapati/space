// Package health implements the chetana platform-aggregator
// health surface.
//
// → REQ-FUNC-CMN-004; design.md §3.1, §4.3.
//
// Three pieces:
//
//   • Store      (this file) — Postgres persistence for the
//                              service_health roll-up + the
//                              health_incidents + transitions log.
//   • Aggregate  (aggregate.go) — periodic poll of every registered
//                                 service's /ready; populates the
//                                 store and triggers the alerter.
//   • Alerter    (alerter.go) — flap + sustained-failure detectors;
//                               routes to Slack / email / PagerDuty.

package health

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Status is the per-service current state. Aligned with the
// CHECK constraint in migration 0002.
const (
	StatusOK       = "ok"
	StatusDegraded = "degraded"
	StatusDown     = "down"
	StatusUnknown  = "unknown"
)

// Snapshot is one row of the service_health table.
type Snapshot struct {
	Service      string
	LastSeenAt   time.Time
	LastStatus   string
	LastError    string
	ErrorCount   int64
	SuccessCount int64
	UpdatedAt    time.Time
}

// IsHealthy reports whether the snapshot is in the OK state.
func (s Snapshot) IsHealthy() bool { return s.LastStatus == StatusOK }

// Store wraps a pgxpool.Pool with health persistence helpers.
type Store struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewStore wraps a pool. clock=nil → time.Now.
func NewStore(pool *pgxpool.Pool, clock func() time.Time) (*Store, error) {
	if pool == nil {
		return nil, errors.New("health: nil pool")
	}
	if clock == nil {
		clock = time.Now
	}
	return &Store{pool: pool, clk: clock}, nil
}

// RecordCheck UPSERTs the latest probe outcome. Called by the
// aggregator after each /ready poll. Returns the previous status
// so the alerter can detect transitions.
func (s *Store) RecordCheck(ctx context.Context, service, status, errMsg string) (prevStatus string, err error) {
	if service == "" {
		return "", errors.New("health: empty service")
	}
	if status == "" {
		status = StatusUnknown
	}
	now := s.clk().UTC()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("health: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Pre-read the previous row (under lock so a concurrent
	// aggregator tick can't race us).
	const lockQ = `
SELECT last_status FROM service_health WHERE service = $1 FOR UPDATE
`
	err = tx.QueryRow(ctx, lockQ, service).Scan(&prevStatus)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		prevStatus = "" // first-ever check
	case err != nil:
		return "", fmt.Errorf("health: read previous: %w", err)
	}

	// UPSERT the latest snapshot. Counters increment per branch.
	const upsertQ = `
INSERT INTO service_health
  (service, last_seen_at, last_status, last_error, error_count, success_count, updated_at)
VALUES
  ($1, $2, $3, $4,
   CASE WHEN $3 = 'ok' THEN 0 ELSE 1 END,
   CASE WHEN $3 = 'ok' THEN 1 ELSE 0 END,
   $2)
ON CONFLICT (service) DO UPDATE SET
   last_seen_at  = EXCLUDED.last_seen_at,
   last_status   = EXCLUDED.last_status,
   last_error    = EXCLUDED.last_error,
   error_count   = service_health.error_count + CASE WHEN $3 = 'ok' THEN 0 ELSE 1 END,
   success_count = service_health.success_count + CASE WHEN $3 = 'ok' THEN 1 ELSE 0 END,
   updated_at    = EXCLUDED.updated_at
`
	if _, err := tx.Exec(ctx, upsertQ, service, now, status, errMsg); err != nil {
		return "", fmt.Errorf("health: upsert: %w", err)
	}

	// Log the transition when the status changed (or first
	// observation). The flap detector reads this table.
	if prevStatus != status {
		if _, err := tx.Exec(ctx, `
INSERT INTO service_transitions (service, from_status, to_status, transitioned_at)
VALUES ($1, $2, $3, $4)
`, service, fallbackStatus(prevStatus), status, now); err != nil {
			return "", fmt.Errorf("health: log transition: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("health: commit: %w", err)
	}
	return prevStatus, nil
}

// Roll returns one row per registered service for /v1/health/services.
func (s *Store) Roll(ctx context.Context) ([]Snapshot, error) {
	rows, err := s.pool.Query(ctx, `
SELECT service, last_seen_at, last_status, last_error,
       error_count, success_count, updated_at
FROM service_health
ORDER BY service ASC
`)
	if err != nil {
		return nil, fmt.Errorf("health: roll: %w", err)
	}
	defer rows.Close()
	var out []Snapshot
	for rows.Next() {
		var sn Snapshot
		if err := rows.Scan(
			&sn.Service, &sn.LastSeenAt, &sn.LastStatus, &sn.LastError,
			&sn.ErrorCount, &sn.SuccessCount, &sn.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("health: scan: %w", err)
		}
		out = append(out, sn)
	}
	return out, rows.Err()
}

// CountTransitionsSince returns the number of state transitions
// for `service` in the last `window`. Used by the flap detector.
func (s *Store) CountTransitionsSince(ctx context.Context, service string, window time.Duration) (int, error) {
	cutoff := s.clk().UTC().Add(-window)
	var n int
	if err := s.pool.QueryRow(ctx, `
SELECT count(*) FROM service_transitions
WHERE service = $1 AND transitioned_at >= $2
`, service, cutoff).Scan(&n); err != nil {
		return 0, fmt.Errorf("health: count transitions: %w", err)
	}
	return n, nil
}

// SustainedSince returns the duration the service has been in
// its current non-OK status. Returns (0, false) when the service
// is currently OK or has no row yet.
func (s *Store) SustainedSince(ctx context.Context, service string) (time.Duration, bool, error) {
	const q = `
SELECT to_status, transitioned_at
FROM service_transitions
WHERE service = $1
ORDER BY transitioned_at DESC
LIMIT 1
`
	var (
		toStatus string
		at       time.Time
	)
	err := s.pool.QueryRow(ctx, q, service).Scan(&toStatus, &at)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("health: sustained: %w", err)
	}
	if toStatus == StatusOK {
		return 0, false, nil
	}
	return s.clk().UTC().Sub(at), true, nil
}

// Incident is one row of health_incidents.
type Incident struct {
	ID          int64
	Service     string
	State       string
	Severity    string
	OpenedAt    time.Time
	ResolvedAt  sql.NullTime
	Transitions int
	Note        string
}

// IncidentState constants.
const (
	StateFlap              = "flap"
	StateSustainedFailure  = "sustained_failure"
)

// Severity constants.
const (
	SeverityWarn = "warn"
	SeverityPage = "page"
)

// OpenIncident creates an open incident OR returns the existing
// open one (per the (service, state) UNIQUE WHERE resolved_at
// IS NULL index). Idempotent — repeated detection ticks DO NOT
// create duplicate rows; they update transitions + note.
func (s *Store) OpenIncident(ctx context.Context, service, state, severity, note string, transitions int) (*Incident, error) {
	if service == "" || state == "" {
		return nil, errors.New("health: empty service / state")
	}
	if severity == "" {
		severity = SeverityWarn
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("health: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Look for an existing open incident.
	const lookupQ = `
SELECT id, opened_at, severity, transitions, note
FROM health_incidents
WHERE service = $1 AND state = $2 AND resolved_at IS NULL
FOR UPDATE
`
	var inc Incident
	inc.Service = service
	inc.State = state
	err = tx.QueryRow(ctx, lookupQ, service, state).Scan(
		&inc.ID, &inc.OpenedAt, &inc.Severity, &inc.Transitions, &inc.Note,
	)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		// New incident.
		now := s.clk().UTC()
		if err := tx.QueryRow(ctx, `
INSERT INTO health_incidents (service, state, severity, opened_at, transitions, note)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id
`, service, state, severity, now, transitions, note).Scan(&inc.ID); err != nil {
			return nil, fmt.Errorf("health: insert incident: %w", err)
		}
		inc.OpenedAt = now
		inc.Severity = severity
		inc.Transitions = transitions
		inc.Note = note
	case err != nil:
		return nil, fmt.Errorf("health: lookup incident: %w", err)
	default:
		// Existing open incident — bump transitions + note.
		if _, err := tx.Exec(ctx, `
UPDATE health_incidents
SET transitions = $2, note = $3
WHERE id = $1
`, inc.ID, transitions, note); err != nil {
			return nil, fmt.Errorf("health: update incident: %w", err)
		}
		inc.Transitions = transitions
		inc.Note = note
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("health: commit: %w", err)
	}
	return &inc, nil
}

// ResolveOpenIncidents closes every open incident for `service`.
// Called by the aggregator when a previously-failing service
// returns to OK.
func (s *Store) ResolveOpenIncidents(ctx context.Context, service string) (int64, error) {
	if service == "" {
		return 0, errors.New("health: empty service")
	}
	tag, err := s.pool.Exec(ctx, `
UPDATE health_incidents
SET resolved_at = $2
WHERE service = $1 AND resolved_at IS NULL
`, service, s.clk().UTC())
	if err != nil {
		return 0, fmt.Errorf("health: resolve: %w", err)
	}
	return tag.RowsAffected(), nil
}

// PruneTransitions deletes transition log entries older than
// `keep`. Called periodically by the aggregator's sweep so the
// flap detector's count query stays cheap.
func (s *Store) PruneTransitions(ctx context.Context, keep time.Duration) (int64, error) {
	cutoff := s.clk().UTC().Add(-keep)
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM service_transitions WHERE transitioned_at < $1`, cutoff)
	if err != nil {
		return 0, fmt.Errorf("health: prune: %w", err)
	}
	return tag.RowsAffected(), nil
}

// fallbackStatus collapses an empty (first-ever) prevStatus to a
// known marker so the transition log column stays NOT-NULL safe.
func fallbackStatus(s string) string {
	if s == "" {
		return StatusUnknown
	}
	return s
}
