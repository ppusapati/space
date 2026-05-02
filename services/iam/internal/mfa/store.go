// store.go — Postgres persistence + in-memory replay cache for the
// MFA package.
//
// Two responsibilities:
//
//   1. Persist enrolled TOTP secrets + backup-code hashes;
//      verify-and-consume backup codes atomically.
//   2. Maintain an in-process replay cache keyed by
//      (user_id, step, code) so an attacker who steals a live TOTP
//      from the network cannot replay it within the same step
//      window. REQ-FUNC-PLT-IAM-004 acceptance #3.
//
// The replay cache is intentionally process-local. Cross-instance
// replay protection requires that the load balancer pin a user's
// session-establishment requests to a single replica during the
// 30-second TOTP window — the platform's IAM ingress already does
// session affinity for the login flow, so a process-local cache is
// sufficient. Instances bound to a shared Redis would only matter
// for active-active without affinity, which we don't run.

package mfa

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store wraps a pgxpool.Pool with the MFA-specific persistence
// helpers. Construct with NewStore.
type Store struct {
	pool *pgxpool.Pool
	clk  func() time.Time

	mu     sync.Mutex
	replay map[replayKey]struct{} // (user, step, code) → present
	// replayCleanupAfter governs how often the replay cache is
	// swept. 60s is more than enough — the longest window the cache
	// must retain an entry is StepSeconds × (2*ToleranceSteps + 1) =
	// 90 s.
	replayCleanupAfter time.Time
}

type replayKey struct {
	UserID string
	Step   uint64
	Code   string
}

// NewStore wraps a pool. clock=nil → time.Now.
func NewStore(pool *pgxpool.Pool, clock func() time.Time) *Store {
	if clock == nil {
		clock = time.Now
	}
	return &Store{
		pool:   pool,
		clk:    clock,
		replay: make(map[replayKey]struct{}),
	}
}

// EnrolledTOTP holds the row shape for mfa_totp_secrets.
type EnrolledTOTP struct {
	UserID    string
	Secret    []byte
	CreatedAt time.Time
	// VerifiedAt is non-zero once the user has proven possession of
	// the secret with a valid first code (REQ-FUNC-PLT-IAM-004
	// acceptance #1). Until verified the row exists but the secret
	// MUST NOT be treated as an active second factor.
	VerifiedAt sql.NullTime
}

// SaveEnrollment persists an unverified TOTP secret. Calling this
// twice for the same user replaces the prior pending enrollment —
// the previous row is removed in the same transaction so a stale
// QR code can't haunt the user later.
func (s *Store) SaveEnrollment(ctx context.Context, userID string, secret []byte) error {
	if userID == "" {
		return errors.New("mfa: empty user_id")
	}
	if len(secret) == 0 {
		return errors.New("mfa: empty secret")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("mfa: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`DELETE FROM mfa_totp_secrets WHERE user_id = $1 AND verified_at IS NULL`,
		userID,
	); err != nil {
		return fmt.Errorf("mfa: clear pending: %w", err)
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO mfa_totp_secrets (user_id, secret, created_at)
		 VALUES ($1, $2, $3)`,
		userID, secret, s.clk().UTC(),
	); err != nil {
		return fmt.Errorf("mfa: insert: %w", err)
	}
	return tx.Commit(ctx)
}

// MarkVerified flips verified_at on the user's pending TOTP row.
// Called after the user submits a code that Verify accepts. Idempotent.
func (s *Store) MarkVerified(ctx context.Context, userID string) error {
	if _, err := s.pool.Exec(ctx,
		`UPDATE mfa_totp_secrets
		 SET verified_at = COALESCE(verified_at, $2)
		 WHERE user_id = $1`,
		userID, s.clk().UTC(),
	); err != nil {
		return fmt.Errorf("mfa: mark verified: %w", err)
	}
	return nil
}

// LoadActive returns the user's verified TOTP row (or nil + nil when
// the user has no enrollment). Verified means: row present AND
// verified_at NOT NULL.
func (s *Store) LoadActive(ctx context.Context, userID string) (*EnrolledTOTP, error) {
	const q = `
SELECT user_id, secret, created_at, verified_at
FROM mfa_totp_secrets
WHERE user_id = $1 AND verified_at IS NOT NULL
`
	var row EnrolledTOTP
	err := s.pool.QueryRow(ctx, q, userID).Scan(
		&row.UserID, &row.Secret, &row.CreatedAt, &row.VerifiedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("mfa: load: %w", err)
	}
	return &row, nil
}

// LoadPending returns the user's NOT-yet-verified TOTP row (used
// during the enrollment finalisation step). Returns nil + nil when
// no pending row exists.
func (s *Store) LoadPending(ctx context.Context, userID string) (*EnrolledTOTP, error) {
	const q = `
SELECT user_id, secret, created_at, verified_at
FROM mfa_totp_secrets
WHERE user_id = $1 AND verified_at IS NULL
`
	var row EnrolledTOTP
	err := s.pool.QueryRow(ctx, q, userID).Scan(
		&row.UserID, &row.Secret, &row.CreatedAt, &row.VerifiedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("mfa: load pending: %w", err)
	}
	return &row, nil
}

// DeleteEnrollment removes the user's TOTP row entirely. Used when
// the user disables MFA from the settings UI.
func (s *Store) DeleteEnrollment(ctx context.Context, userID string) error {
	if _, err := s.pool.Exec(ctx,
		`DELETE FROM mfa_totp_secrets WHERE user_id = $1`,
		userID,
	); err != nil {
		return fmt.Errorf("mfa: delete totp: %w", err)
	}
	return nil
}

// SaveBackupCodes replaces the user's backup-code book with a fresh
// set. The old codes are deleted in the same transaction so a user
// who regenerates can never accidentally satisfy auth with a printed
// page from the previous round.
func (s *Store) SaveBackupCodes(ctx context.Context, userID string, codes []BackupCodeIssued) error {
	if userID == "" {
		return errors.New("mfa: empty user_id")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("mfa: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx,
		`DELETE FROM mfa_backup_codes WHERE user_id = $1`,
		userID,
	); err != nil {
		return fmt.Errorf("mfa: clear codes: %w", err)
	}
	now := s.clk().UTC()
	for _, c := range codes {
		if _, err := tx.Exec(ctx,
			`INSERT INTO mfa_backup_codes
			   (user_id, prefix, code_hash, created_at)
			 VALUES ($1, $2, $3, $4)`,
			userID, c.PrefixIdx, c.Hash, now,
		); err != nil {
			return fmt.Errorf("mfa: insert code: %w", err)
		}
	}
	return tx.Commit(ctx)
}

// ConsumeBackupCode atomically:
//   1. Loads every active row for (userID, prefixOf(presented)).
//   2. bcrypt-compares each hash against `presented`.
//   3. Marks the matching row consumed_at=now and commits.
//
// Returns nil on success; ErrBackupCodeNotFound when no active row
// matches; ErrBackupCodeReused when the matching row is already
// consumed (this is treated as a soft-fail rather than triggering
// the family-revocation pattern used for refresh tokens — backup
// codes are explicitly multi-shot resources, just one-shot per code).
func (s *Store) ConsumeBackupCode(ctx context.Context, userID, presented string) error {
	if userID == "" || presented == "" {
		return ErrBackupCodeNotFound
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("mfa: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
SELECT id, code_hash, consumed_at
FROM mfa_backup_codes
WHERE user_id = $1 AND prefix = $2
FOR UPDATE
`
	rows, err := tx.Query(ctx, q, userID, PrefixOf(presented))
	if err != nil {
		return fmt.Errorf("mfa: query: %w", err)
	}
	type cand struct {
		id         int64
		hash       []byte
		consumedAt sql.NullTime
	}
	var candidates []cand
	for rows.Next() {
		var c cand
		if err := rows.Scan(&c.id, &c.hash, &c.consumedAt); err != nil {
			rows.Close()
			return fmt.Errorf("mfa: scan: %w", err)
		}
		candidates = append(candidates, c)
	}
	rows.Close()
	if len(candidates) == 0 {
		return ErrBackupCodeNotFound
	}

	hashes := make([][]byte, len(candidates))
	for i, c := range candidates {
		hashes[i] = c.hash
	}
	idx, err := VerifyBackupCode(presented, hashes)
	if err != nil {
		return ErrBackupCodeNotFound
	}
	matched := candidates[idx]
	if matched.consumedAt.Valid {
		return ErrBackupCodeReused
	}
	if _, err := tx.Exec(ctx,
		`UPDATE mfa_backup_codes SET consumed_at = $1 WHERE id = $2`,
		s.clk().UTC(), matched.id,
	); err != nil {
		return fmt.Errorf("mfa: mark consumed: %w", err)
	}
	return tx.Commit(ctx)
}

// CountActiveBackupCodes returns the number of UN-consumed backup
// codes still on the user's book. Surfaced in the settings UI so
// users know when to regenerate.
func (s *Store) CountActiveBackupCodes(ctx context.Context, userID string) (int, error) {
	var n int
	if err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM mfa_backup_codes
		 WHERE user_id = $1 AND consumed_at IS NULL`,
		userID,
	).Scan(&n); err != nil {
		return 0, fmt.Errorf("mfa: count: %w", err)
	}
	return n, nil
}

// ConsumeReplayWindow records that (userID, step, code) has been
// successfully verified. Returns true if this is the first
// presentation; false if it has already been seen within the
// active window (a replay attempt). Callers MUST treat a false
// return as authentication failure.
//
// Cache entries are swept every minute. The longest a single entry
// must live is StepSeconds × (2*ToleranceSteps + 1) = 90 s; the
// 60-second sweep + the entry-time horizon below cover that.
func (s *Store) ConsumeReplayWindow(userID string, step uint64, code string) bool {
	now := s.clk()
	s.mu.Lock()
	defer s.mu.Unlock()

	if now.After(s.replayCleanupAfter) {
		s.gcReplayLocked(now)
		s.replayCleanupAfter = now.Add(time.Minute)
	}

	key := replayKey{UserID: userID, Step: step, Code: code}
	if _, seen := s.replay[key]; seen {
		return false
	}
	s.replay[key] = struct{}{}
	return true
}

// gcReplayLocked drops cache entries whose step is older than the
// tolerance window. Cheap because the cache is tiny (worst case:
// active_users × 3 entries during the window).
func (s *Store) gcReplayLocked(now time.Time) {
	cutoff := uint64(now.Unix()/StepSeconds) - ToleranceSteps - 1
	for k := range s.replay {
		if k.Step < cutoff {
			delete(s.replay, k)
		}
	}
}

// ErrBackupCodeReused is returned when the matching backup-code row
// is already marked consumed.
var ErrBackupCodeReused = errors.New("mfa: backup code already used")
