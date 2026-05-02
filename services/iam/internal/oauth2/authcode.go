// authcode.go — short-lived authorisation-code store.
//
// One row per outstanding auth code. Codes are SHA-256 hashed at
// rest so a database leak does not enable code-redemption forgery.
// Lifetimes are tight (10 minutes per OAuth 2.1 §4.1.2 guidance)
// and codes are single-use: redemption marks consumed_at and
// returns the binding so the token endpoint can mint the access +
// refresh pair.
//
// Re-presentation of an already-consumed code is a hard error
// (ErrAuthCodeReused). For now we surface the error to the client;
// a future hardening will mirror the refresh-token family-
// invalidation pattern (revoke any tokens minted from that code).

package oauth2

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

// AuthCodeTTL is the wall-clock lifetime of an outstanding auth
// code. OAuth 2.1 §4.1.2 says ≤ 10 minutes; we use exactly that.
const AuthCodeTTL = 10 * time.Minute

// AuthCodeBytes is the random-bytes length of a freshly minted
// code. 32 bytes = 256 bits — well above any feasible online
// brute-force budget.
const AuthCodeBytes = 32

// AuthCodeIssue is the input shape for AuthCodeStore.Issue.
type AuthCodeIssue struct {
	ClientID            string
	UserID              string
	TenantID            string
	SessionID           string
	RedirectURI         string
	Scopes              []string
	CodeChallenge       string
	CodeChallengeMethod string // always MethodS256 in v1; pinned at /authorize.
	Nonce               string // OIDC nonce, echoed into the id_token.
}

// AuthCodeIssued is what AuthCodeStore.Issue returns.
type AuthCodeIssued struct {
	Code      string
	ExpiresAt time.Time
}

// AuthCodeRecord is the in-memory shape of one oauth2_auth_codes
// row, returned by Redeem.
type AuthCodeRecord struct {
	ClientID            string
	UserID              string
	TenantID            string
	SessionID           string
	RedirectURI         string
	Scopes              []string
	CodeChallenge       string
	CodeChallengeMethod string
	Nonce               string
	IssuedAt            time.Time
	ExpiresAt           time.Time
}

// AuthCodeStore wraps a pgx pool with the auth-code persistence
// helpers. Construct with NewAuthCodeStore.
type AuthCodeStore struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewAuthCodeStore wraps a pool. clock=nil → time.Now.
func NewAuthCodeStore(pool *pgxpool.Pool, clock func() time.Time) *AuthCodeStore {
	if clock == nil {
		clock = time.Now
	}
	return &AuthCodeStore{pool: pool, clk: clock}
}

// Issue mints a fresh code for the given binding. Returns the
// opaque code string the authorisation-server's redirect handler
// hands to the user-agent and the wall-clock expiry.
func (s *AuthCodeStore) Issue(ctx context.Context, in AuthCodeIssue) (AuthCodeIssued, error) {
	if in.ClientID == "" || in.UserID == "" || in.RedirectURI == "" || in.CodeChallenge == "" {
		return AuthCodeIssued{}, errors.New("oauth2: invalid auth code issue inputs")
	}
	raw, err := newAuthCodeBytes()
	if err != nil {
		return AuthCodeIssued{}, fmt.Errorf("oauth2: random: %w", err)
	}
	rowID, err := newRowID()
	if err != nil {
		return AuthCodeIssued{}, fmt.Errorf("oauth2: row id: %w", err)
	}
	now := s.clk().UTC()
	expiresAt := now.Add(AuthCodeTTL)

	const q = `
INSERT INTO oauth2_auth_codes
  (id, code_hash, client_id, user_id, tenant_id, session_id,
   redirect_uri, scopes, code_challenge, code_challenge_method,
   nonce, issued_at, expires_at)
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
`
	if _, err := s.pool.Exec(ctx, q,
		rowID, hashAuthCode(raw),
		in.ClientID, in.UserID, in.TenantID, in.SessionID,
		in.RedirectURI, in.Scopes,
		in.CodeChallenge, in.CodeChallengeMethod,
		in.Nonce, now, expiresAt,
	); err != nil {
		return AuthCodeIssued{}, fmt.Errorf("oauth2: insert auth code: %w", err)
	}
	return AuthCodeIssued{
		Code:      encodeAuthCodeBearer(rowID, raw),
		ExpiresAt: expiresAt,
	}, nil
}

// Redeem atomically:
//   1. Looks up the row by id, verifies the code hash.
//   2. Asserts the row is not consumed and not expired.
//   3. Marks consumed_at = now.
//   4. Returns the binding so the caller can mint tokens.
//
// Whole sequence runs in a single transaction so two concurrent
// redemption attempts can't both succeed.
func (s *AuthCodeStore) Redeem(ctx context.Context, presented string) (*AuthCodeRecord, error) {
	rowID, secret, err := decodeAuthCodeBearer(presented)
	if err != nil {
		return nil, ErrAuthCodeNotFound
	}
	now := s.clk().UTC()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("oauth2: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
SELECT code_hash, client_id, user_id, tenant_id, session_id,
       redirect_uri, scopes, code_challenge, code_challenge_method,
       nonce, issued_at, expires_at, consumed_at
FROM oauth2_auth_codes
WHERE id = $1
FOR UPDATE
`
	var (
		rec        AuthCodeRecord
		hash       string
		consumedAt sql.NullTime
	)
	err = tx.QueryRow(ctx, q, rowID).Scan(
		&hash, &rec.ClientID, &rec.UserID, &rec.TenantID, &rec.SessionID,
		&rec.RedirectURI, &rec.Scopes,
		&rec.CodeChallenge, &rec.CodeChallengeMethod,
		&rec.Nonce, &rec.IssuedAt, &rec.ExpiresAt, &consumedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrAuthCodeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("oauth2: lookup: %w", err)
	}
	if hashAuthCode(secret) != hash {
		return nil, ErrAuthCodeNotFound
	}
	if consumedAt.Valid {
		return nil, ErrAuthCodeReused
	}
	if !now.Before(rec.ExpiresAt) {
		return nil, ErrAuthCodeExpired
	}
	if _, err := tx.Exec(ctx,
		`UPDATE oauth2_auth_codes SET consumed_at = $1 WHERE id = $2`,
		now, rowID,
	); err != nil {
		return nil, fmt.Errorf("oauth2: mark consumed: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("oauth2: commit: %w", err)
	}
	return &rec, nil
}

// ----------------------------------------------------------------------
// Internal helpers
// ----------------------------------------------------------------------

func newAuthCodeBytes() ([]byte, error) {
	b := make([]byte, AuthCodeBytes)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func newRowID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashAuthCode(secret []byte) string {
	sum := sha256.Sum256(secret)
	return hex.EncodeToString(sum[:])
}

// encodeAuthCodeBearer combines the row id and the random secret
// into the bearer string handed to the user-agent. Format:
//
//	<rowID>.<base64url-unpadded(secret)>
func encodeAuthCodeBearer(rowID string, secret []byte) string {
	return rowID + "." + base64.RawURLEncoding.EncodeToString(secret)
}

// decodeAuthCodeBearer is the inverse.
func decodeAuthCodeBearer(s string) (string, []byte, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			rowID := s[:i]
			b64 := s[i+1:]
			if rowID == "" || b64 == "" {
				return "", nil, ErrAuthCodeNotFound
			}
			secret, err := base64.RawURLEncoding.DecodeString(b64)
			if err != nil {
				return "", nil, ErrAuthCodeNotFound
			}
			return rowID, secret, nil
		}
	}
	return "", nil, ErrAuthCodeNotFound
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrAuthCodeNotFound is returned when the presented code does
// not match any active row OR the bearer encoding is malformed.
var ErrAuthCodeNotFound = errors.New("oauth2: authorization code not found")

// ErrAuthCodeExpired is returned when the code's expires_at has
// elapsed.
var ErrAuthCodeExpired = errors.New("oauth2: authorization code expired")

// ErrAuthCodeReused is returned when the matching row is already
// consumed.
var ErrAuthCodeReused = errors.New("oauth2: authorization code already consumed")
