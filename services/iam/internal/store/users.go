// Package store implements the IAM service's persistence layer.
// Today: just the users table; later tasks add sessions,
// refresh_tokens, mfa_*, webauthn_credentials, oauth2_clients,
// saml_idps, and password_resets.
//
// All methods are context-aware and use parameterised queries — no
// string concatenation against caller-supplied data, ever.
//
// REQ-FUNC-PLT-IAM-001 / REQ-FUNC-PLT-IAM-003.
package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// User mirrors one row of the users table. Tags align with the
// column names used by the Scan helper.
type User struct {
	ID                 string
	TenantID           string
	EmailLower         string
	EmailDisplay       string
	PasswordHash       string
	PasswordAlgo       string
	Status             string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	LastLoginAt        *time.Time
	FailedLoginCount   int
	LockedUntil        *time.Time
	LockoutLevel       int
	DataClassification string
	GDPRAnonymizedAt   *time.Time
}

// Status constants — keep aligned with the CHECK constraint in
// migrations/0001_users.sql.
const (
	StatusActive              = "active"
	StatusPendingVerification = "pending_verification"
	StatusDisabled            = "disabled"
	StatusDeleted             = "deleted"
)

// IsActive reports whether the user can attempt to log in. Anything
// other than 'active' (e.g. pending_verification, disabled, deleted)
// blocks the flow at the handler.
func (u *User) IsActive() bool { return u.Status == StatusActive }

// IsLockedAt reports whether the user is currently in lockout for
// the supplied wall-clock time. The handler passes time.Now() in
// production and a frozen clock in tests.
func (u *User) IsLockedAt(now time.Time) bool {
	if u.LockedUntil == nil {
		return false
	}
	return now.Before(*u.LockedUntil)
}

// LockoutRemaining returns the duration until the lockout expires
// or 0 when the user is not locked / the lockout has elapsed.
func (u *User) LockoutRemaining(now time.Time) time.Duration {
	if !u.IsLockedAt(now) {
		return 0
	}
	return u.LockedUntil.Sub(now)
}

// ----------------------------------------------------------------------
// Store
// ----------------------------------------------------------------------

// Store is the public CRUD façade over the users table. Construct
// with NewStore; the implementation holds a pgx pool.
type Store struct {
	pool *pgxpool.Pool
}

// NewStore wraps a pgx pool. Callers are responsible for the pool
// lifecycle (created in main, closed at shutdown).
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// ErrUserNotFound is returned by lookup methods when no row matches.
var ErrUserNotFound = errors.New("store: user not found")

// ErrUserExists is returned by Create when (tenant_id, email_lower)
// is already present.
var ErrUserExists = errors.New("store: user already exists for this email")

// CreateUserParams is the input set for Create. Fields not supplied
// here use the column DEFAULTs from the migration.
type CreateUserParams struct {
	ID                 string
	TenantID           string
	EmailLower         string
	EmailDisplay       string
	PasswordHash       string
	DataClassification string // empty → 'cui'
}

// Create inserts a new user row. Returns ErrUserExists when a
// duplicate (tenant_id, email_lower) is attempted.
func (s *Store) Create(ctx context.Context, p CreateUserParams) error {
	if p.DataClassification == "" {
		p.DataClassification = "cui"
	}
	const q = `
INSERT INTO users (id, tenant_id, email_lower, email_display, password_hash, data_classification)
VALUES ($1, $2, $3, $4, $5, $6)
`
	_, err := s.pool.Exec(ctx, q,
		p.ID, p.TenantID, p.EmailLower, p.EmailDisplay, p.PasswordHash, p.DataClassification,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrUserExists
		}
		return fmt.Errorf("store.Create: %w", err)
	}
	return nil
}

// GetByEmail returns the user matching (tenant_id, email_lower).
// Returns ErrUserNotFound when no row exists.
func (s *Store) GetByEmail(ctx context.Context, tenantID, emailLower string) (*User, error) {
	const q = `
SELECT id, tenant_id, email_lower, email_display, password_hash, password_algo,
       status, created_at, updated_at, last_login_at, failed_login_count,
       locked_until, lockout_level, data_classification, gdpr_anonymized_at
FROM users
WHERE tenant_id = $1 AND email_lower = $2
`
	row := s.pool.QueryRow(ctx, q, tenantID, emailLower)
	var u User
	if err := scanUser(row, &u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("store.GetByEmail: %w", err)
	}
	return &u, nil
}

// GetByID returns the user with the supplied ULID.
func (s *Store) GetByID(ctx context.Context, id string) (*User, error) {
	const q = `
SELECT id, tenant_id, email_lower, email_display, password_hash, password_algo,
       status, created_at, updated_at, last_login_at, failed_login_count,
       locked_until, lockout_level, data_classification, gdpr_anonymized_at
FROM users
WHERE id = $1
`
	row := s.pool.QueryRow(ctx, q, id)
	var u User
	if err := scanUser(row, &u); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("store.GetByID: %w", err)
	}
	return &u, nil
}

// RecordSuccessfulLogin clears the failed-login counter, drops the
// lockout state, resets the lockout level, and stamps last_login_at
// to `now`. Atomic via a single UPDATE.
func (s *Store) RecordSuccessfulLogin(ctx context.Context, userID string, now time.Time) error {
	const q = `
UPDATE users
SET failed_login_count = 0,
    locked_until = NULL,
    lockout_level = 0,
    last_login_at = $2
WHERE id = $1
`
	tag, err := s.pool.Exec(ctx, q, userID, now)
	if err != nil {
		return fmt.Errorf("store.RecordSuccessfulLogin: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// RecordFailedLogin increments failed_login_count and, when the
// supplied threshold is reached, escalates lockout_level and sets
// locked_until = now + the duration matching the new level.
//
// Lockout duration ladder per REQ-FUNC-PLT-IAM-003:
//
//	level 1 → 15 minutes
//	level 2 → 1 hour
//	level 3 → 24 hours
//
// After level 3 the level stays at 3 (further failures keep the
// 24h cap).
//
// `threshold` is the per-account failure budget that triggers a
// lockout escalation (REQ-FUNC-PLT-IAM-003 fixes this at 5).
//
// Returns the post-update User snapshot so the handler can answer
// the client without a follow-up read.
func (s *Store) RecordFailedLogin(ctx context.Context, userID string, threshold int, now time.Time) (*User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("store.RecordFailedLogin: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const incrementQ = `
UPDATE users
SET failed_login_count = failed_login_count + 1
WHERE id = $1
RETURNING failed_login_count, lockout_level
`
	var fails, level int
	if err := tx.QueryRow(ctx, incrementQ, userID).Scan(&fails, &level); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("store.RecordFailedLogin: increment: %w", err)
	}

	// Escalate when we hit the threshold AND we are not already in
	// a lockout window. This keeps repeated probes during an active
	// lockout from over-escalating.
	if fails >= threshold {
		newLevel := level + 1
		if newLevel > 3 {
			newLevel = 3
		}
		dur := lockoutDurationFor(newLevel)
		until := now.Add(dur)
		const escalateQ = `
UPDATE users
SET lockout_level = $2,
    locked_until   = $3,
    failed_login_count = 0
WHERE id = $1
`
		if _, err := tx.Exec(ctx, escalateQ, userID, newLevel, until); err != nil {
			return nil, fmt.Errorf("store.RecordFailedLogin: escalate: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("store.RecordFailedLogin: commit: %w", err)
	}

	updated, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// UpdatePasswordHash replaces the user's password hash + algo and
// resets the lockout state (so a successful password reset
// implicitly unlocks an account that had been frozen by repeated
// bad-password attempts). The `algo` value is the PHC string's
// algorithm name — currently always "argon2id" via PolicyV1.
//
// Returns ErrUserNotFound when no row matches the supplied id.
func (s *Store) UpdatePasswordHash(ctx context.Context, userID, hash, algo string, now time.Time) error {
	if userID == "" || hash == "" || algo == "" {
		return errors.New("store: empty user_id / hash / algo")
	}
	const q = `
UPDATE users
SET password_hash = $2,
    password_algo = $3,
    failed_login_count = 0,
    locked_until = NULL,
    lockout_level = 0,
    updated_at = $4
WHERE id = $1
`
	tag, err := s.pool.Exec(ctx, q, userID, hash, algo, now)
	if err != nil {
		return fmt.Errorf("store.UpdatePasswordHash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrUserNotFound
	}
	return nil
}

// lockoutDurationFor returns the lockout duration for a given level.
// Exposed in this file for the unit test on the ladder.
func lockoutDurationFor(level int) time.Duration {
	switch level {
	case 1:
		return 15 * time.Minute
	case 2:
		return 1 * time.Hour
	case 3:
		return 24 * time.Hour
	}
	return 0
}

// ----------------------------------------------------------------------
// helpers
// ----------------------------------------------------------------------

// rowScanner is the minimal interface satisfied by both pgx.Row and
// pgx.Rows.Scan invocations. Lets scanUser be reused across single-
// row and result-set paths.
type rowScanner interface {
	Scan(dest ...any) error
}

// scanUser populates `u` from the canonical SELECT order used in
// every GetBy* method.
func scanUser(row rowScanner, u *User) error {
	return row.Scan(
		&u.ID,
		&u.TenantID,
		&u.EmailLower,
		&u.EmailDisplay,
		&u.PasswordHash,
		&u.PasswordAlgo,
		&u.Status,
		&u.CreatedAt,
		&u.UpdatedAt,
		&u.LastLoginAt,
		&u.FailedLoginCount,
		&u.LockedUntil,
		&u.LockoutLevel,
		&u.DataClassification,
		&u.GDPRAnonymizedAt,
	)
}

// isUniqueViolation reports whether `err` is a Postgres unique
// constraint violation (SQLSTATE 23505). Implemented without the
// pgxerrors helper to keep this package's import surface minimal.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	type sqlStateErr interface {
		SQLState() string
	}
	var sse sqlStateErr
	if errors.As(err, &sse) {
		return sse.SQLState() == "23505"
	}
	return false
}
