// Package session implements the chetana IAM service's
// session-state machine on top of the `sessions` table created by
// migration 0002.
//
// → REQ-FUNC-PLT-IAM-009 (sessions, idle/absolute timeouts,
//                          concurrency cap, revocation).
// → design.md §4.1.1.
//
// A session is the unit of "I am still logged in". A session_id
// is stamped onto every issued JWT (login.handler + the OAuth
// auth-code → token redemption both call session.Manager.Create).
// Every protected RPC consults Manager.Touch, which:
//
//   • Looks up the row, refusing revoked / idle-expired /
//     absolute-expired sessions.
//   • Bumps last_seen_at to defer the idle-timeout horizon.
//
// Concurrency cap: a successful Create that would push the user
// past MaxConcurrent active sessions evicts the oldest by
// issued_at (revoking it) so the new login wins. This matches the
// "6th login bumps the 1st" expectation in REQ-FUNC-PLT-IAM-009
// acceptance #1.
//
// Revocation is immediate: every protected RPC's Touch runs the
// revoked_at check, so a Revoke call on /admin/revoke takes effect
// on the very next request the affected user makes — there's no
// access-token cache to invalidate. The 15m access-token TTL
// remains the upper bound when a verifier hasn't seen the new
// session state yet, but Touch closes that window for any service
// that does call Manager (which is every service-side interceptor
// in TASK-P1-AUTHZ-001).

package session

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Defaults per REQ-FUNC-PLT-IAM-009.
const (
	// DefaultIdleTimeout is the rolling no-activity window. Bumped
	// every Touch; once exceeded the session is treated as expired.
	DefaultIdleTimeout = time.Hour

	// DefaultAbsoluteLifetime is the hard ceiling on session age,
	// regardless of activity. After this point the user must log
	// in again.
	DefaultAbsoluteLifetime = 24 * time.Hour

	// DefaultMaxConcurrent caps the number of simultaneously-active
	// sessions a user can hold. A 6th concurrent login evicts the
	// oldest (by issued_at).
	DefaultMaxConcurrent = 5
)

// Config configures the Manager's policy knobs.
type Config struct {
	// IdleTimeout overrides DefaultIdleTimeout.
	IdleTimeout time.Duration

	// AbsoluteLifetime overrides DefaultAbsoluteLifetime.
	AbsoluteLifetime time.Duration

	// MaxConcurrent overrides DefaultMaxConcurrent. Zero or
	// negative values fall back to the default.
	MaxConcurrent int

	// Now injects a clock for tests. nil → time.Now.
	Now func() time.Time
}

// Manager is the session-state engine.
type Manager struct {
	pool *pgxpool.Pool
	cfg  Config
}

// NewManager wraps a pool. Config zero values are filled with
// DefaultIdleTimeout / DefaultAbsoluteLifetime / DefaultMaxConcurrent /
// time.Now.
func NewManager(pool *pgxpool.Pool, cfg Config) (*Manager, error) {
	if pool == nil {
		return nil, errors.New("session: nil pool")
	}
	if cfg.IdleTimeout <= 0 {
		cfg.IdleTimeout = DefaultIdleTimeout
	}
	if cfg.AbsoluteLifetime <= 0 {
		cfg.AbsoluteLifetime = DefaultAbsoluteLifetime
	}
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = DefaultMaxConcurrent
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Manager{pool: pool, cfg: cfg}, nil
}

// CreateInput is the handler-supplied per-login data.
type CreateInput struct {
	UserID             string
	TenantID           string
	ClientIP           string
	UserAgent          string
	AMR                []string
	DataClassification string // empty → "cui"
}

// Created is the result of a successful Create.
type Created struct {
	SessionID         string
	IssuedAt          time.Time
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	// EvictedSessionIDs lists the session ids the concurrency-cap
	// enforcement evicted to make room for this one. Empty when
	// the user was already under the cap.
	EvictedSessionIDs []string
}

// Create opens a new session and enforces the per-user concurrency
// cap. Eviction is atomic: the new INSERT and any required
// revocations of older sessions run in the same transaction, so
// two concurrent logins can't both squeak past the cap.
func (m *Manager) Create(ctx context.Context, in CreateInput) (*Created, error) {
	if in.UserID == "" || in.TenantID == "" {
		return nil, errors.New("session: empty user_id / tenant_id")
	}
	classification := in.DataClassification
	if classification == "" {
		classification = "cui"
	}
	now := m.cfg.Now().UTC()

	id, err := newSessionID()
	if err != nil {
		return nil, fmt.Errorf("session: id: %w", err)
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("session: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// 1. Evict the oldest sessions (if any) to make room. We
	// FOR UPDATE the active row set so a parallel transaction
	// trying the same eviction is forced to serialise.
	const activeQ = `
SELECT id FROM sessions
WHERE user_id = $1
  AND revoked_at IS NULL
  AND $2 < absolute_expires_at
  AND $2 < idle_expires_at
ORDER BY issued_at ASC
FOR UPDATE
`
	rows, err := tx.Query(ctx, activeQ, in.UserID, now)
	if err != nil {
		return nil, fmt.Errorf("session: list active: %w", err)
	}
	var activeIDs []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			rows.Close()
			return nil, fmt.Errorf("session: scan: %w", err)
		}
		activeIDs = append(activeIDs, sid)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("session: rows: %w", err)
	}

	// We're about to insert a new active session, so the cap fires
	// when the existing active count is >= MaxConcurrent.
	var evicted []string
	if surplus := len(activeIDs) - (m.cfg.MaxConcurrent - 1); surplus > 0 {
		// Revoke the surplus oldest entries.
		toRevoke := activeIDs[:surplus]
		for _, sid := range toRevoke {
			if _, err := tx.Exec(ctx, `
UPDATE sessions
SET revoked_at = $2, revoked_by = $3
WHERE id = $1 AND revoked_at IS NULL
`, sid, now, "concurrency_cap"); err != nil {
				return nil, fmt.Errorf("session: evict: %w", err)
			}
			evicted = append(evicted, sid)
		}
	}

	// 2. Insert the new session.
	idleExpiresAt := now.Add(m.cfg.IdleTimeout)
	absoluteExpiresAt := now.Add(m.cfg.AbsoluteLifetime)
	const insertQ = `
INSERT INTO sessions
  (id, user_id, tenant_id, issued_at, last_seen_at,
   absolute_expires_at, idle_expires_at,
   client_ip, user_agent, amr, data_classification)
VALUES ($1, $2, $3, $4, $4, $5, $6, $7, $8, $9, $10)
`
	if _, err := tx.Exec(ctx, insertQ,
		id, in.UserID, in.TenantID, now,
		absoluteExpiresAt, idleExpiresAt,
		in.ClientIP, in.UserAgent,
		amrSlice(in.AMR),
		classification,
	); err != nil {
		return nil, fmt.Errorf("session: insert: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("session: commit: %w", err)
	}
	return &Created{
		SessionID:         id,
		IssuedAt:          now,
		IdleExpiresAt:     idleExpiresAt,
		AbsoluteExpiresAt: absoluteExpiresAt,
		EvictedSessionIDs: evicted,
	}, nil
}

// Status is the in-memory shape of one sessions row.
type Status struct {
	SessionID         string
	UserID            string
	TenantID          string
	IssuedAt          time.Time
	LastSeenAt        time.Time
	IdleExpiresAt     time.Time
	AbsoluteExpiresAt time.Time
	RevokedAt         sql.NullTime
	RevokedBy         string
}

// IsActiveAt reports whether the session is a valid credential at
// time t.
func (s *Status) IsActiveAt(t time.Time) bool {
	if s.RevokedAt.Valid {
		return false
	}
	if !t.Before(s.AbsoluteExpiresAt) {
		return false
	}
	if !t.Before(s.IdleExpiresAt) {
		return false
	}
	return true
}

// Touch validates a session and bumps last_seen_at.
//
// Returns the (possibly-updated) Status on success. On failure,
// callers MUST treat the response as "session no longer valid" —
// the principal's access token is effectively dead even before
// its 15-minute exp.
//
// Errors:
//   • ErrSessionNotFound       — no row matches the session_id.
//   • ErrSessionRevoked        — revoked_at is not null.
//   • ErrSessionIdleTimeout    — last_seen_at + IdleTimeout < now.
//   • ErrSessionAbsoluteExpired — issued_at + AbsoluteLifetime < now.
func (m *Manager) Touch(ctx context.Context, sessionID string) (*Status, error) {
	if sessionID == "" {
		return nil, ErrSessionNotFound
	}
	now := m.cfg.Now().UTC()

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("session: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
SELECT id, user_id, tenant_id, issued_at, last_seen_at,
       absolute_expires_at, idle_expires_at, revoked_at,
       COALESCE(revoked_by, '')
FROM sessions
WHERE id = $1
FOR UPDATE
`
	var s Status
	err = tx.QueryRow(ctx, q, sessionID).Scan(
		&s.SessionID, &s.UserID, &s.TenantID,
		&s.IssuedAt, &s.LastSeenAt,
		&s.AbsoluteExpiresAt, &s.IdleExpiresAt,
		&s.RevokedAt, &s.RevokedBy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("session: lookup: %w", err)
	}

	if s.RevokedAt.Valid {
		return &s, ErrSessionRevoked
	}
	if !now.Before(s.AbsoluteExpiresAt) {
		return &s, ErrSessionAbsoluteExpired
	}
	if !now.Before(s.IdleExpiresAt) {
		return &s, ErrSessionIdleTimeout
	}

	// Bump last_seen_at + push the idle horizon forward. Absolute
	// expiry is never bumped (that's what makes it "absolute").
	idleExpiresAt := now.Add(m.cfg.IdleTimeout)
	if _, err := tx.Exec(ctx, `
UPDATE sessions
SET last_seen_at = $2, idle_expires_at = $3
WHERE id = $1
`, sessionID, now, idleExpiresAt); err != nil {
		return nil, fmt.Errorf("session: touch: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("session: commit: %w", err)
	}
	s.LastSeenAt = now
	s.IdleExpiresAt = idleExpiresAt
	return &s, nil
}

// Revoke marks the session revoked. `by` is recorded in
// revoked_by for the audit chain ("user_logout" / "admin_revoke" /
// "concurrency_cap"). Idempotent: revoking an already-revoked
// session is a no-op.
func (m *Manager) Revoke(ctx context.Context, sessionID, by string) error {
	if sessionID == "" {
		return ErrSessionNotFound
	}
	if by == "" {
		by = "explicit"
	}
	tag, err := m.pool.Exec(ctx, `
UPDATE sessions
SET revoked_at = $2, revoked_by = $3
WHERE id = $1 AND revoked_at IS NULL
`, sessionID, m.cfg.Now().UTC(), by)
	if err != nil {
		return fmt.Errorf("session: revoke: %w", err)
	}
	if tag.RowsAffected() == 0 {
		// Either the session doesn't exist or was already revoked.
		// We treat both as success for idempotency, but surface
		// not-found to the caller via a follow-up Status check if
		// they need to distinguish.
		return nil
	}
	return nil
}

// RevokeAllForUser revokes every active session for a user. Used
// by /admin/revoke-user and the password-reset flow.
func (m *Manager) RevokeAllForUser(ctx context.Context, userID, by string) (int64, error) {
	if userID == "" {
		return 0, errors.New("session: empty user_id")
	}
	if by == "" {
		by = "explicit"
	}
	tag, err := m.pool.Exec(ctx, `
UPDATE sessions
SET revoked_at = $2, revoked_by = $3
WHERE user_id = $1 AND revoked_at IS NULL
`, userID, m.cfg.Now().UTC(), by)
	if err != nil {
		return 0, fmt.Errorf("session: revoke user: %w", err)
	}
	return tag.RowsAffected(), nil
}

// CountActiveForUser returns the number of currently-active
// sessions for the user. Surfaced in the settings UI ("you are
// signed in to N devices").
func (m *Manager) CountActiveForUser(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, errors.New("session: empty user_id")
	}
	now := m.cfg.Now().UTC()
	var n int
	if err := m.pool.QueryRow(ctx, `
SELECT count(*) FROM sessions
WHERE user_id = $1
  AND revoked_at IS NULL
  AND $2 < absolute_expires_at
  AND $2 < idle_expires_at
`, userID, now).Scan(&n); err != nil {
		return 0, fmt.Errorf("session: count: %w", err)
	}
	return n, nil
}

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------

// newSessionID returns 32 hex characters of secure randomness.
func newSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// amrSlice returns a non-nil amr slice; pgx encodes nil as NULL,
// but the column is NOT NULL with default ARRAY[].
func amrSlice(amr []string) []string {
	if len(amr) == 0 {
		return []string{}
	}
	out := make([]string, len(amr))
	copy(out, amr)
	return out
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrSessionNotFound is returned when no row matches the supplied
// session id.
var ErrSessionNotFound = errors.New("session: not found")

// ErrSessionRevoked is returned by Touch when revoked_at is set.
// Callers translate this into 401 Unauthorized with reason
// "session_revoked".
var ErrSessionRevoked = errors.New("session: revoked")

// ErrSessionIdleTimeout is returned by Touch when last_seen_at +
// IdleTimeout has elapsed. REQ-FUNC-PLT-IAM-009 acceptance #2.
var ErrSessionIdleTimeout = errors.New("session: idle timeout")

// ErrSessionAbsoluteExpired is returned by Touch when issued_at +
// AbsoluteLifetime has elapsed.
var ErrSessionAbsoluteExpired = errors.New("session: absolute expired")
