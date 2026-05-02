// Package reset implements the chetana password-reset flow.
//
// → REQ-FUNC-PLT-IAM-010 (256-bit token, 1h TTL, hashed at rest,
//                          single-use, rate-limited, constant-time
//                          response).
// → design.md §4.1.1.
//
// Two-phase flow:
//
//   1. Request(email) — handler returns 202 unconditionally
//      (constant-time + email-existence non-disclosure). When the
//      email maps to an active user we INSERT a row in
//      password_resets and emit a notify event with the bearer
//      token; otherwise we silently no-op.
//
//   2. Confirm(token, newPassword) — handler hashes the token,
//      finds the matching active row, marks it consumed, and
//      replaces the user's password hash. All under one
//      transaction. Single-use is enforced by the consumed_at
//      check; reuse returns ErrTokenAlreadyUsed.
//
// Bearer format: <rowID>.<base64url-unpadded(secret)>. Same shape
// as the refresh-token + auth-code bearers so the parser is
// uniform across the IAM service.

package reset

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Defaults per REQ-FUNC-PLT-IAM-010.
const (
	// TokenBytes is the random-bytes length of a freshly minted
	// reset secret. 32 bytes = 256 bits — well above any feasible
	// online brute-force budget over the 1h TTL.
	TokenBytes = 32

	// DefaultTTL is the wall-clock lifetime of a token.
	DefaultTTL = time.Hour
)

// Issued is what Store.Issue returns. Token is shown to the user
// EXACTLY ONCE — the database stores only the SHA-256 hash.
type Issued struct {
	Token     string
	ExpiresAt time.Time
}

// Record is the in-memory shape of one password_resets row,
// returned from Redeem.
type Record struct {
	ID        string
	UserID    string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// Store wraps a pgx pool with the reset-token persistence
// helpers. Construct with NewStore.
type Store struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewStore wraps a pool. clock=nil → time.Now.
func NewStore(pool *pgxpool.Pool, clock func() time.Time) *Store {
	if clock == nil {
		clock = time.Now
	}
	return &Store{pool: pool, clk: clock}
}

// Issue mints a fresh reset token for the supplied user. TTL=0
// uses DefaultTTL.
func (s *Store) Issue(ctx context.Context, userID string, ttl time.Duration) (Issued, error) {
	if userID == "" {
		return Issued{}, errors.New("reset: empty user_id")
	}
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	rowID, err := newRowID()
	if err != nil {
		return Issued{}, fmt.Errorf("reset: row id: %w", err)
	}
	secret, err := newTokenBytes()
	if err != nil {
		return Issued{}, fmt.Errorf("reset: secret: %w", err)
	}
	now := s.clk().UTC()
	expiresAt := now.Add(ttl)

	const q = `
INSERT INTO password_resets
  (id, token_hash, user_id, issued_at, expires_at)
VALUES ($1, $2, $3, $4, $5)
`
	if _, err := s.pool.Exec(ctx, q,
		rowID, hashToken(secret), userID, now, expiresAt,
	); err != nil {
		return Issued{}, fmt.Errorf("reset: insert: %w", err)
	}
	return Issued{
		Token:     encodeBearer(rowID, secret),
		ExpiresAt: expiresAt,
	}, nil
}

// Redeem atomically:
//   1. Looks up the row by id, verifies the hash.
//   2. Asserts the row is not consumed and not expired.
//   3. Marks consumed_at = now.
//   4. Returns the binding so the caller can update the user.
//
// Re-presentation of an already-consumed token returns
// ErrTokenAlreadyUsed.
func (s *Store) Redeem(ctx context.Context, presented string) (*Record, error) {
	rowID, secret, err := decodeBearer(presented)
	if err != nil {
		return nil, ErrTokenNotFound
	}
	now := s.clk().UTC()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("reset: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
SELECT id, token_hash, user_id, issued_at, expires_at, consumed_at
FROM password_resets
WHERE id = $1
FOR UPDATE
`
	var (
		rec        Record
		hash       string
		consumedAt sql.NullTime
	)
	err = tx.QueryRow(ctx, q, rowID).Scan(
		&rec.ID, &hash, &rec.UserID,
		&rec.IssuedAt, &rec.ExpiresAt, &consumedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("reset: lookup: %w", err)
	}
	if hashToken(secret) != hash {
		return nil, ErrTokenNotFound
	}
	if consumedAt.Valid {
		return nil, ErrTokenAlreadyUsed
	}
	if !now.Before(rec.ExpiresAt) {
		return nil, ErrTokenExpired
	}
	if _, err := tx.Exec(ctx,
		`UPDATE password_resets SET consumed_at = $1 WHERE id = $2`,
		now, rowID,
	); err != nil {
		return nil, fmt.Errorf("reset: mark consumed: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("reset: commit: %w", err)
	}
	return &rec, nil
}

// CountRecentForUser returns the number of tokens issued for the
// user within the supplied window. Used by the request handler to
// enforce the 3-per-hour rate cap.
func (s *Store) CountRecentForUser(ctx context.Context, userID string, window time.Duration) (int, error) {
	if userID == "" {
		return 0, errors.New("reset: empty user_id")
	}
	if window <= 0 {
		window = time.Hour
	}
	cutoff := s.clk().UTC().Add(-window)
	var n int
	if err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM password_resets WHERE user_id = $1 AND issued_at >= $2`,
		userID, cutoff,
	).Scan(&n); err != nil {
		return 0, fmt.Errorf("reset: count: %w", err)
	}
	return n, nil
}

// ----------------------------------------------------------------------
// Internal helpers
// ----------------------------------------------------------------------

func newRowID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func newTokenBytes() ([]byte, error) {
	b := make([]byte, TokenBytes)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// hashToken returns the lowercase hex SHA-256 of the secret.
func hashToken(secret []byte) string {
	sum := sha256.Sum256(secret)
	return hex.EncodeToString(sum[:])
}

// encodeBearer combines the row id and the random secret into the
// bearer string handed to the user. Format: <rowID>.<b64url(secret)>.
func encodeBearer(rowID string, secret []byte) string {
	return rowID + "." + base64.RawURLEncoding.EncodeToString(secret)
}

// decodeBearer is the inverse.
func decodeBearer(s string) (string, []byte, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			rowID := s[:i]
			b64 := s[i+1:]
			if rowID == "" || b64 == "" {
				return "", nil, ErrTokenNotFound
			}
			secret, err := base64.RawURLEncoding.DecodeString(b64)
			if err != nil {
				return "", nil, ErrTokenNotFound
			}
			return rowID, secret, nil
		}
	}
	return "", nil, ErrTokenNotFound
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrTokenNotFound is returned when no row matches the presented
// token (or the bearer string is malformed).
var ErrTokenNotFound = errors.New("reset: token not found")

// ErrTokenExpired is returned when expires_at has elapsed.
var ErrTokenExpired = errors.New("reset: token expired")

// ErrTokenAlreadyUsed is returned when the matching row is already
// marked consumed.
var ErrTokenAlreadyUsed = errors.New("reset: token already used")
