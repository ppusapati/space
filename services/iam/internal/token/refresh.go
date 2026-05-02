// refresh.go — single-use refresh tokens with reuse-detection
// family invalidation.
//
// → REQ-FUNC-PLT-IAM-002 acceptance #1: refresh-token reuse
//   invalidates the entire session family.
// → design.md §4.1.1
//
// Storage shape (one row per token in refresh_tokens):
//
//   id          : ULID; printed value the client carries.
//   token_hash  : SHA-256 of the bearer token. We store the hash,
//                 not the token, so a DB read does not enable
//                 forgery.
//   family_id   : ULID grouping every token in the same lineage —
//                 first token issued at login starts the family;
//                 every subsequent refresh issues a new token in
//                 the same family.
//   parent_id   : the token this one replaced (NULL for the head).
//   user_id     : owning user.
//   tenant_id   : owning tenant.
//   session_id  : the IAM session this family belongs to.
//   issued_at   : creation timestamp.
//   expires_at  : expiration timestamp (issued_at + RefreshTokenTTL).
//   consumed_at : NULL until used; populated atomically by the
//                 Rotate call when this token is exchanged for a
//                 fresh one.
//   revoked     : bool. Set true on explicit logout OR on detection
//                 of a reuse attempt (in which case every row in
//                 the family is revoked at once).
//
// Reuse detection:
//
//   When a client presents a refresh token whose consumed_at is
//   NOT NULL (i.e. it has already been used), Rotate marks every
//   token in the same family as revoked, returns ErrReusedRefresh,
//   and fires an audit event. The session is dead; the user must
//   log in again.

package token

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

// RefreshTokenLength is the random-bytes length of a freshly minted
// refresh token. 32 bytes = 256 bits — well above the
// brute-force-search budget any attacker can mount in 7 days.
const RefreshTokenLength = 32

// RefreshIssue is the input shape for RefreshStore.Issue.
type RefreshIssue struct {
	UserID    string
	TenantID  string
	SessionID string
	FamilyID  string // empty → start a new family
	ParentID  string // empty → head of family
	IssuedAt  time.Time
	TTL       time.Duration // 0 → DefaultRefreshTokenTTL
}

// RefreshIssued is what RefreshStore.Issue returns.
type RefreshIssued struct {
	// Token is the bearer string handed to the client. Show the
	// caller exactly once — the database stores only the hash.
	Token string
	// ID is the row id (ULID-ish hex) used as the parent reference
	// of the next refresh in this family.
	ID string
	// FamilyID is the family this token belongs to. Returned for
	// the caller's audit + logging needs.
	FamilyID string
	// ExpiresAt mirrors the row's expires_at column.
	ExpiresAt time.Time
}

// RefreshRecord is the in-memory shape of one refresh_tokens row.
type RefreshRecord struct {
	ID         string
	UserID     string
	TenantID   string
	SessionID  string
	FamilyID   string
	ParentID   sql.NullString
	IssuedAt   time.Time
	ExpiresAt  time.Time
	ConsumedAt sql.NullTime
	Revoked    bool
}

// IsActiveAt reports whether the row should be treated as a valid
// refresh credential at instant t.
func (r RefreshRecord) IsActiveAt(t time.Time) bool {
	if r.Revoked {
		return false
	}
	if r.ConsumedAt.Valid {
		return false
	}
	if !t.Before(r.ExpiresAt) {
		return false
	}
	return true
}

// RefreshStore implements the persistence layer for refresh tokens.
type RefreshStore struct {
	pool  *pgxpool.Pool
	clock func() time.Time
}

// NewRefreshStore wraps a pgx pool. clock=nil → time.Now.
func NewRefreshStore(pool *pgxpool.Pool, clock func() time.Time) *RefreshStore {
	if clock == nil {
		clock = time.Now
	}
	return &RefreshStore{pool: pool, clock: clock}
}

// Issue mints a new refresh token. When in.FamilyID is empty the
// caller is starting a new family (post-login); otherwise the new
// token joins the existing family + records its parent.
func (s *RefreshStore) Issue(ctx context.Context, in RefreshIssue) (RefreshIssued, error) {
	if in.UserID == "" || in.TenantID == "" || in.SessionID == "" {
		return RefreshIssued{}, errors.New("refresh: empty user/tenant/session")
	}
	if in.IssuedAt.IsZero() {
		in.IssuedAt = s.clock().UTC()
	}
	ttl := in.TTL
	if ttl <= 0 {
		ttl = DefaultRefreshTokenTTL
	}
	expiresAt := in.IssuedAt.Add(ttl)

	rawToken, err := newRefreshSecret()
	if err != nil {
		return RefreshIssued{}, fmt.Errorf("refresh: random: %w", err)
	}
	rowID, err := newRandomHex(16)
	if err != nil {
		return RefreshIssued{}, fmt.Errorf("refresh: row id: %w", err)
	}
	familyID := in.FamilyID
	if familyID == "" {
		familyID, err = newRandomHex(16)
		if err != nil {
			return RefreshIssued{}, fmt.Errorf("refresh: family id: %w", err)
		}
	}

	hash := hashRefresh(rawToken)
	const q = `
INSERT INTO refresh_tokens
  (id, token_hash, family_id, parent_id, user_id, tenant_id, session_id,
   issued_at, expires_at)
VALUES
  ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`
	var parent sql.NullString
	if in.ParentID != "" {
		parent = sql.NullString{String: in.ParentID, Valid: true}
	}
	if _, err := s.pool.Exec(ctx, q,
		rowID, hash, familyID, parent, in.UserID, in.TenantID, in.SessionID,
		in.IssuedAt, expiresAt,
	); err != nil {
		return RefreshIssued{}, fmt.Errorf("refresh: insert: %w", err)
	}
	return RefreshIssued{
		Token:     encodeRefreshBearer(rowID, rawToken),
		ID:        rowID,
		FamilyID:  familyID,
		ExpiresAt: expiresAt,
	}, nil
}

// Rotate exchanges a presented refresh token for a fresh one.
// Atomic semantics:
//   • Look up the row by id and verify the hash.
//   • If the row is already consumed (reuse attempt), revoke the
//     entire family and return ErrReusedRefresh.
//   • If the row is revoked or expired, return ErrInvalidRefresh.
//   • Otherwise mark this row consumed_at = now and INSERT a new
//     row in the same family with parent_id = this.id.
//
// The whole sequence runs in a single transaction so two concurrent
// reuse attempts can't both squeak through the consumed-at check.
func (s *RefreshStore) Rotate(ctx context.Context, presented string) (RefreshIssued, error) {
	rowID, secret, err := decodeRefreshBearer(presented)
	if err != nil {
		return RefreshIssued{}, ErrInvalidRefresh
	}

	now := s.clock().UTC()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return RefreshIssued{}, fmt.Errorf("refresh: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	rec, err := lookupForUpdate(ctx, tx, rowID)
	if err != nil {
		return RefreshIssued{}, err
	}
	if hashRefresh(secret) != rec.tokenHash {
		return RefreshIssued{}, ErrInvalidRefresh
	}

	if rec.consumedAt.Valid {
		// Reuse attempt — revoke the entire family AND commit so
		// the revocation persists even when we return an error.
		if err := revokeFamily(ctx, tx, rec.familyID); err != nil {
			return RefreshIssued{}, fmt.Errorf("refresh: revoke family: %w", err)
		}
		if err := tx.Commit(ctx); err != nil {
			return RefreshIssued{}, fmt.Errorf("refresh: commit revoke: %w", err)
		}
		return RefreshIssued{}, ErrReusedRefresh
	}
	if rec.revoked {
		return RefreshIssued{}, ErrInvalidRefresh
	}
	if !now.Before(rec.expiresAt) {
		return RefreshIssued{}, ErrExpiredRefresh
	}

	// Mark consumed.
	if _, err := tx.Exec(ctx,
		`UPDATE refresh_tokens SET consumed_at = $1 WHERE id = $2`,
		now, rowID,
	); err != nil {
		return RefreshIssued{}, fmt.Errorf("refresh: mark consumed: %w", err)
	}

	// Issue successor inside the same transaction.
	rawToken, err := newRefreshSecret()
	if err != nil {
		return RefreshIssued{}, fmt.Errorf("refresh: new secret: %w", err)
	}
	newID, err := newRandomHex(16)
	if err != nil {
		return RefreshIssued{}, fmt.Errorf("refresh: new id: %w", err)
	}
	expiresAt := now.Add(DefaultRefreshTokenTTL)
	if _, err := tx.Exec(ctx,
		`INSERT INTO refresh_tokens
		  (id, token_hash, family_id, parent_id, user_id, tenant_id, session_id,
		   issued_at, expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		newID, hashRefresh(rawToken), rec.familyID,
		sql.NullString{String: rowID, Valid: true},
		rec.userID, rec.tenantID, rec.sessionID,
		now, expiresAt,
	); err != nil {
		return RefreshIssued{}, fmt.Errorf("refresh: insert successor: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return RefreshIssued{}, fmt.Errorf("refresh: commit: %w", err)
	}
	return RefreshIssued{
		Token:     encodeRefreshBearer(newID, rawToken),
		ID:        newID,
		FamilyID:  rec.familyID,
		ExpiresAt: expiresAt,
	}, nil
}

// RevokeFamily marks every token in a family revoked. Called by
// Logout.
func (s *RefreshStore) RevokeFamily(ctx context.Context, familyID string) error {
	if _, err := s.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE family_id = $1`,
		familyID,
	); err != nil {
		return fmt.Errorf("refresh: revoke family: %w", err)
	}
	return nil
}

// RevokeSession marks every token in every family belonging to the
// session as revoked. Used by session-revocation flows.
func (s *RefreshStore) RevokeSession(ctx context.Context, sessionID string) error {
	if _, err := s.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE session_id = $1`,
		sessionID,
	); err != nil {
		return fmt.Errorf("refresh: revoke session: %w", err)
	}
	return nil
}

// ----------------------------------------------------------------------
// Internal helpers
// ----------------------------------------------------------------------

type forUpdateRow struct {
	id         string
	tokenHash  string
	familyID   string
	userID     string
	tenantID   string
	sessionID  string
	expiresAt  time.Time
	consumedAt sql.NullTime
	revoked    bool
}

func lookupForUpdate(ctx context.Context, tx pgx.Tx, id string) (forUpdateRow, error) {
	const q = `
SELECT id, token_hash, family_id, user_id, tenant_id, session_id,
       expires_at, consumed_at, revoked
FROM refresh_tokens
WHERE id = $1
FOR UPDATE
`
	var r forUpdateRow
	err := tx.QueryRow(ctx, q, id).Scan(
		&r.id, &r.tokenHash, &r.familyID, &r.userID, &r.tenantID, &r.sessionID,
		&r.expiresAt, &r.consumedAt, &r.revoked,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return forUpdateRow{}, ErrInvalidRefresh
	}
	if err != nil {
		return forUpdateRow{}, fmt.Errorf("refresh: lookup: %w", err)
	}
	return r, nil
}

func revokeFamily(ctx context.Context, tx pgx.Tx, familyID string) error {
	_, err := tx.Exec(ctx,
		`UPDATE refresh_tokens SET revoked = true WHERE family_id = $1`,
		familyID,
	)
	return err
}

// newRefreshSecret generates RefreshTokenLength random bytes.
func newRefreshSecret() ([]byte, error) {
	b := make([]byte, RefreshTokenLength)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// hashRefresh returns the lowercase hex SHA-256 of the secret.
func hashRefresh(secret []byte) string {
	sum := sha256.Sum256(secret)
	return hex.EncodeToString(sum[:])
}

// encodeRefreshBearer combines the row id and the random secret
// into the bearer string handed to clients. Format:
//
//	<rowID>.<base64url-unpadded(secret)>
//
// The dot is a forbidden character inside the row id (it's hex)
// and the base64url alphabet, so the split is unambiguous.
func encodeRefreshBearer(rowID string, secret []byte) string {
	return rowID + "." + base64.RawURLEncoding.EncodeToString(secret)
}

// decodeRefreshBearer is the inverse. Returns ErrInvalidRefresh
// when the input doesn't match the expected shape.
func decodeRefreshBearer(s string) (string, []byte, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			rowID := s[:i]
			b64 := s[i+1:]
			if rowID == "" || b64 == "" {
				return "", nil, ErrInvalidRefresh
			}
			secret, err := base64.RawURLEncoding.DecodeString(b64)
			if err != nil {
				return "", nil, ErrInvalidRefresh
			}
			return rowID, secret, nil
		}
	}
	return "", nil, ErrInvalidRefresh
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrInvalidRefresh is returned when the presented token does not
// match any active row OR the bearer encoding is malformed. The
// caller treats this as 401 Unauthorized.
var ErrInvalidRefresh = errors.New("refresh: invalid token")

// ErrExpiredRefresh is returned when the presented token's
// expires_at has elapsed.
var ErrExpiredRefresh = errors.New("refresh: expired token")

// ErrReusedRefresh is returned when the presented token has
// already been consumed. Receiving this error means the caller
// MUST treat the surrounding session as compromised; the IAM
// service has revoked every token in the family.
var ErrReusedRefresh = errors.New("refresh: token reused — family invalidated")
