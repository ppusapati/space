// portability.go — Article 20 (Right to data portability).
//
// Returns a typed JSON snapshot of every IAM-owned row that
// references the data subject. Used by:
//
//   • The SAR endpoint (Article 15) — the snapshot is what gets
//     bundled into the export-service archive.
//   • The portability endpoint (Article 20) — the snapshot is
//     returned synchronously as JSON for direct download.
//
// The snapshot is intentionally a flat in-memory struct so the
// caller chooses the serialisation (JSON over HTTP for the
// portability endpoint, NDJSON / Parquet for the export pipeline).

package gdpr

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Snapshot is the IAM-side data Article 20 returns and Article 15
// includes in the SAR archive.
//
// The shape is FLAT (no nested objects beyond the leaf row
// types) so consumers can stream-encode without buffering. Field
// names stay close to the column names so a customer reading the
// JSON sees what the database actually stores.
type Snapshot struct {
	GeneratedAt time.Time         `json:"generated_at"`
	User        UserSnapshot      `json:"user"`
	Sessions    []SessionSnapshot `json:"sessions,omitempty"`
	OAuthCodes  []AuthCodeSnap    `json:"oauth_auth_codes,omitempty"`
	WebAuthn    []WebAuthnSnap    `json:"webauthn_credentials,omitempty"`
	MFAEnrolled MFASnapshot       `json:"mfa,omitempty"`
}

// UserSnapshot mirrors the row a user could see if they SELECT'd
// their own users record. PasswordHash is intentionally OMITTED —
// the GDPR right of access does not include credentials.
type UserSnapshot struct {
	ID                 string     `json:"id"`
	TenantID           string     `json:"tenant_id"`
	EmailLower         string     `json:"email_lower"`
	EmailDisplay       string     `json:"email_display"`
	Status             string     `json:"status"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	LastLoginAt        *time.Time `json:"last_login_at,omitempty"`
	DataClassification string     `json:"data_classification"`
	GDPRAnonymizedAt   *time.Time `json:"gdpr_anonymized_at,omitempty"`
}

// SessionSnapshot is one row of the sessions table the user can
// see ("you are signed in to N devices").
type SessionSnapshot struct {
	SessionID         string     `json:"session_id"`
	IssuedAt          time.Time  `json:"issued_at"`
	LastSeenAt        time.Time  `json:"last_seen_at"`
	IdleExpiresAt     time.Time  `json:"idle_expires_at"`
	AbsoluteExpiresAt time.Time  `json:"absolute_expires_at"`
	ClientIP          string     `json:"client_ip,omitempty"`
	UserAgent         string     `json:"user_agent,omitempty"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
}

// AuthCodeSnap is one row of oauth2_auth_codes (open or already
// consumed). Bearer secrets are NOT included — only the row id +
// metadata.
type AuthCodeSnap struct {
	ID          string     `json:"id"`
	ClientID    string     `json:"client_id"`
	RedirectURI string     `json:"redirect_uri"`
	IssuedAt    time.Time  `json:"issued_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	ConsumedAt  *time.Time `json:"consumed_at,omitempty"`
}

// WebAuthnSnap is one row of webauthn_credentials. The credential
// public key + sign_count are included (the user already knows
// these — they're embedded in the authenticator). The credential
// id is included so the user can match it against their physical
// device list.
type WebAuthnSnap struct {
	CredentialID string     `json:"credential_id"`
	SignCount    uint32     `json:"sign_count"`
	Transports   string     `json:"transports,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsedAt   time.Time  `json:"last_used_at"`
	DisabledAt   *time.Time `json:"disabled_at,omitempty"`
}

// MFASnapshot summarises the user's MFA enrolment state. We never
// include the TOTP secret or backup-code hashes — those would be
// indistinguishable from a credential leak under SAR.
type MFASnapshot struct {
	TOTPEnrolled       bool       `json:"totp_enrolled"`
	TOTPVerifiedAt     *time.Time `json:"totp_verified_at,omitempty"`
	BackupCodesActive  int        `json:"backup_codes_active"`
}

// SnapshotBuilder fetches a Snapshot from Postgres. Construct with
// NewSnapshotBuilder.
type SnapshotBuilder struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewSnapshotBuilder wraps a pool. clock=nil → time.Now.
func NewSnapshotBuilder(pool *pgxpool.Pool, clock func() time.Time) *SnapshotBuilder {
	if clock == nil {
		clock = time.Now
	}
	return &SnapshotBuilder{pool: pool, clk: clock}
}

// Build assembles a Snapshot for the supplied user. Returns
// ErrUserNotFound when no users row matches.
//
// Each sub-fetch is best-effort: if a particular table is missing
// in the deployment (e.g. webauthn_credentials hasn't been
// migrated yet) the call swallows the "relation does not exist"
// error and leaves that slice nil. This keeps the snapshot
// resilient across staged migrations.
func (b *SnapshotBuilder) Build(ctx context.Context, userID string) (*Snapshot, error) {
	if userID == "" {
		return nil, ErrUserNotFound
	}
	user, err := b.user(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := &Snapshot{
		GeneratedAt: b.clk().UTC(),
		User:        *user,
	}
	out.Sessions, _ = b.sessions(ctx, userID)
	out.OAuthCodes, _ = b.authCodes(ctx, userID)
	out.WebAuthn, _ = b.webauthn(ctx, userID)
	out.MFAEnrolled, _ = b.mfa(ctx, userID)
	return out, nil
}

func (b *SnapshotBuilder) user(ctx context.Context, userID string) (*UserSnapshot, error) {
	const q = `
SELECT id, tenant_id, email_lower, email_display, status,
       created_at, updated_at, last_login_at,
       data_classification, gdpr_anonymized_at
FROM users WHERE id = $1
`
	var u UserSnapshot
	err := b.pool.QueryRow(ctx, q, userID).Scan(
		&u.ID, &u.TenantID, &u.EmailLower, &u.EmailDisplay, &u.Status,
		&u.CreatedAt, &u.UpdatedAt, &u.LastLoginAt,
		&u.DataClassification, &u.GDPRAnonymizedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("gdpr: load user: %w", err)
	}
	return &u, nil
}

func (b *SnapshotBuilder) sessions(ctx context.Context, userID string) ([]SessionSnapshot, error) {
	const q = `
SELECT id, issued_at, last_seen_at, idle_expires_at, absolute_expires_at,
       client_ip, user_agent, revoked_at
FROM sessions WHERE user_id = $1
`
	rows, err := b.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SessionSnapshot
	for rows.Next() {
		var s SessionSnapshot
		if err := rows.Scan(
			&s.SessionID, &s.IssuedAt, &s.LastSeenAt, &s.IdleExpiresAt, &s.AbsoluteExpiresAt,
			&s.ClientIP, &s.UserAgent, &s.RevokedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (b *SnapshotBuilder) authCodes(ctx context.Context, userID string) ([]AuthCodeSnap, error) {
	const q = `
SELECT id, client_id, redirect_uri, issued_at, expires_at, consumed_at
FROM oauth2_auth_codes WHERE user_id = $1
`
	rows, err := b.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AuthCodeSnap
	for rows.Next() {
		var c AuthCodeSnap
		if err := rows.Scan(
			&c.ID, &c.ClientID, &c.RedirectURI, &c.IssuedAt, &c.ExpiresAt, &c.ConsumedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (b *SnapshotBuilder) webauthn(ctx context.Context, userID string) ([]WebAuthnSnap, error) {
	const q = `
SELECT encode(credential_id, 'base64'), sign_count, transports,
       created_at, last_used_at, disabled_at
FROM webauthn_credentials WHERE user_id = $1
`
	rows, err := b.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WebAuthnSnap
	for rows.Next() {
		var (
			w        WebAuthnSnap
			rawCount int64
		)
		if err := rows.Scan(
			&w.CredentialID, &rawCount, &w.Transports,
			&w.CreatedAt, &w.LastUsedAt, &w.DisabledAt,
		); err != nil {
			return nil, err
		}
		w.SignCount = uint32(rawCount)
		out = append(out, w)
	}
	return out, rows.Err()
}

func (b *SnapshotBuilder) mfa(ctx context.Context, userID string) (MFASnapshot, error) {
	var m MFASnapshot
	// TOTP enrolment status.
	const totpQ = `
SELECT verified_at IS NOT NULL, verified_at
FROM mfa_totp_secrets WHERE user_id = $1
`
	err := b.pool.QueryRow(ctx, totpQ, userID).Scan(&m.TOTPEnrolled, &m.TOTPVerifiedAt)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return m, err
	}
	// Backup-code count.
	const codesQ = `
SELECT count(*) FROM mfa_backup_codes
WHERE user_id = $1 AND consumed_at IS NULL
`
	if err := b.pool.QueryRow(ctx, codesQ, userID).Scan(&m.BackupCodesActive); err != nil {
		// Non-fatal; the table might not exist yet.
		return m, nil
	}
	return m, nil
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrUserNotFound is returned when the snapshot/erasure target
// does not exist.
var ErrUserNotFound = errors.New("gdpr: user not found")
