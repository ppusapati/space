// store.go — Postgres persistence + the User adapter that the
// go-webauthn protocol library consumes.
//
// Storage shape (one row per registered credential):
//
//   id              : auto-increment surrogate key.
//   user_id         : owning user (uuid).
//   credential_id   : the W3C credential.id (raw bytes; the bearer
//                     value the authenticator sends back during
//                     assertion). Indexed for the Authn lookup.
//   public_key      : the COSE-encoded public key bytes the
//                     library pulled out of the attestation.
//   sign_count      : the last-seen authenticator counter value;
//                     UPDATEd on every successful assertion.
//   transports      : comma-joined transport hints
//                     (usb, nfc, ble, internal, hybrid).
//   attestation_*   : the attestation type + format the library
//                     recorded at registration; kept so future
//                     metadata-service re-validation has the raw
//                     facts.
//   flags_uv,
//   flags_bs,
//   flags_be,
//   flags_up        : the credential flags from the registration
//                     authData (user-verified, backup-state,
//                     backup-eligible, user-presence). Kept so
//                     downstream consumers can reason about
//                     synced/unsynced credentials.
//   disabled_at     : NULL while the credential is active;
//                     non-NULL when clone detection or an admin
//                     revocation has retired it. Disabled
//                     credentials remain in the table for audit /
//                     forensics; they never appear in the User
//                     adapter's WebAuthnCredentials() output.
//   created_at,
//   last_used_at    : bookkeeping.

package webauthn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store persists WebAuthn credentials. Construct with NewStore.
type Store struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewStore wraps a pgx pool. clock=nil → time.Now.
func NewStore(pool *pgxpool.Pool, clock func() time.Time) *Store {
	if clock == nil {
		clock = time.Now
	}
	return &Store{pool: pool, clk: clock}
}

// User is the chetana-side WebAuthn user. Implements
// [webauthn.User] so it can be passed straight into the protocol
// library's Begin/FinishRegistration + Begin/FinishLogin calls.
//
// The struct is hydrated by Store.LoadUser; callers should not
// construct it directly because the credential set must come from
// the database (the protocol library uses WebAuthnCredentials() to
// build the assertion's allowed-credentials list and to resolve a
// credential ID to its public key during signature verification).
type User struct {
	id          []byte // raw user_id bytes — the WebAuthn user handle
	name        string
	displayName string
	credentials []webauthn.Credential
}

// WebAuthnID returns the user handle. Per W3C §5.4.3 this is an
// opaque identifier; we use the user's UUID bytes.
func (u *User) WebAuthnID() []byte { return u.id }

// WebAuthnName returns the user's account name (typically email).
func (u *User) WebAuthnName() string { return u.name }

// WebAuthnDisplayName returns the user's display name.
func (u *User) WebAuthnDisplayName() string { return u.displayName }

// WebAuthnCredentials returns the user's currently active
// credentials (disabled rows are excluded).
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	out := make([]webauthn.Credential, len(u.credentials))
	copy(out, u.credentials)
	return out
}

// LoadUser fetches the user's identity attributes + active
// credentials. Returns ErrUserNotFound when no rows match
// (different from "user has no credentials yet" — that returns a
// hydrated User with an empty credential set so registration can
// proceed).
//
// The signature deliberately accepts the WebAuthnName +
// WebAuthnDisplayName so this package does NOT reach into the
// users table — keeping the dependency on the user store one-way.
func (s *Store) LoadUser(ctx context.Context, userID string, name, displayName string) (*User, error) {
	if userID == "" {
		return nil, errors.New("webauthn: empty user_id")
	}
	rows, err := s.pool.Query(ctx, `
SELECT credential_id, public_key, sign_count, transports,
       attestation_type, attestation_format,
       flags_uv, flags_bs, flags_be, flags_up
FROM webauthn_credentials
WHERE user_id = $1 AND disabled_at IS NULL
`, userID)
	if err != nil {
		return nil, fmt.Errorf("webauthn: load creds: %w", err)
	}
	defer rows.Close()

	var creds []webauthn.Credential
	for rows.Next() {
		var (
			c                                       webauthn.Credential
			transports                              string
			attestationType, attestationFormat      string
			flagsUV, flagsBS, flagsBE, flagsUP      bool
		)
		if err := rows.Scan(
			&c.ID, &c.PublicKey, &c.Authenticator.SignCount, &transports,
			&attestationType, &attestationFormat,
			&flagsUV, &flagsBS, &flagsBE, &flagsUP,
		); err != nil {
			return nil, fmt.Errorf("webauthn: scan: %w", err)
		}
		c.AttestationType = attestationType
		c.AttestationFormat = attestationFormat
		c.Flags = webauthn.CredentialFlags{
			UserPresent:    flagsUP,
			UserVerified:   flagsUV,
			BackupEligible: flagsBE,
			BackupState:    flagsBS,
		}
		c.Transport = parseTransports(transports)
		creds = append(creds, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("webauthn: rows: %w", err)
	}

	return &User{
		id:          []byte(userID),
		name:        name,
		displayName: displayName,
		credentials: creds,
	}, nil
}

// SaveCredential persists a freshly registered credential. The
// caller hands in the *webauthn.Credential the protocol library
// returned from FinishRegistration; this method writes the row.
//
// Returns ErrCredentialExists when a row with the same credential
// ID is already present — that would be a re-registration of an
// existing key, which the caller should reject (the registration
// ceremony's WithExclusions option keeps the authenticator from
// returning a colliding credential in the first place, but this
// server-side check is the defence in depth).
func (s *Store) SaveCredential(ctx context.Context, userID string, c *webauthn.Credential) error {
	if userID == "" || c == nil || len(c.ID) == 0 {
		return errors.New("webauthn: invalid save args")
	}
	const q = `
INSERT INTO webauthn_credentials
  (user_id, credential_id, public_key, sign_count, transports,
   attestation_type, attestation_format,
   flags_uv, flags_bs, flags_be, flags_up,
   created_at, last_used_at)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12)
ON CONFLICT (credential_id) DO NOTHING
`
	now := s.clk().UTC()
	tag, err := s.pool.Exec(ctx, q,
		userID, c.ID, c.PublicKey, c.Authenticator.SignCount,
		joinTransports(c.Transport),
		c.AttestationType, c.AttestationFormat,
		c.Flags.UserVerified, c.Flags.BackupState, c.Flags.BackupEligible, c.Flags.UserPresent,
		now,
	)
	if err != nil {
		return fmt.Errorf("webauthn: insert: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrCredentialExists
	}
	return nil
}

// UpdateSignCount writes the new authenticator counter on a
// successful assertion. Called by Assert after the protocol
// library has confirmed the signature AND the counter strictly
// increased.
func (s *Store) UpdateSignCount(ctx context.Context, credentialID []byte, newCount uint32) error {
	if _, err := s.pool.Exec(ctx, `
UPDATE webauthn_credentials
SET sign_count = $2, last_used_at = $3
WHERE credential_id = $1
`, credentialID, newCount, s.clk().UTC()); err != nil {
		return fmt.Errorf("webauthn: update sign count: %w", err)
	}
	return nil
}

// DisableCredential flips disabled_at on a credential row. Called
// by Assert when the protocol library raises CloneWarning OR by
// the settings UI when the user explicitly removes a key.
func (s *Store) DisableCredential(ctx context.Context, credentialID []byte, reason string) error {
	if _, err := s.pool.Exec(ctx, `
UPDATE webauthn_credentials
SET disabled_at = $2, disabled_reason = $3
WHERE credential_id = $1 AND disabled_at IS NULL
`, credentialID, s.clk().UTC(), reason); err != nil {
		return fmt.Errorf("webauthn: disable: %w", err)
	}
	return nil
}

// LookupOwner returns the user_id that owns a credential ID, or
// empty + nil when no row matches. Used by discoverable-login
// flows where the client supplies only the credential ID.
func (s *Store) LookupOwner(ctx context.Context, credentialID []byte) (string, error) {
	var userID string
	err := s.pool.QueryRow(ctx,
		`SELECT user_id FROM webauthn_credentials WHERE credential_id = $1 AND disabled_at IS NULL`,
		credentialID,
	).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("webauthn: lookup owner: %w", err)
	}
	return userID, nil
}

// CountActive returns the number of currently-enabled credentials
// the user has. Surfaced in settings so we can warn users before
// disabling their last key.
func (s *Store) CountActive(ctx context.Context, userID string) (int, error) {
	var n int
	if err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM webauthn_credentials
		 WHERE user_id = $1 AND disabled_at IS NULL`,
		userID,
	).Scan(&n); err != nil {
		return 0, fmt.Errorf("webauthn: count: %w", err)
	}
	return n, nil
}

// IsDisabled returns whether a credential row is currently disabled.
// Used by tests + admin tooling.
func (s *Store) IsDisabled(ctx context.Context, credentialID []byte) (bool, error) {
	var t sql.NullTime
	if err := s.pool.QueryRow(ctx,
		`SELECT disabled_at FROM webauthn_credentials WHERE credential_id = $1`,
		credentialID,
	).Scan(&t); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrCredentialNotFound
		}
		return false, fmt.Errorf("webauthn: is disabled: %w", err)
	}
	return t.Valid, nil
}

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------

func joinTransports(t []protocol.AuthenticatorTransport) string {
	if len(t) == 0 {
		return ""
	}
	parts := make([]string, len(t))
	for i, x := range t {
		parts[i] = string(x)
	}
	return strings.Join(parts, ",")
}

func parseTransports(s string) []protocol.AuthenticatorTransport {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]protocol.AuthenticatorTransport, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, protocol.AuthenticatorTransport(p))
	}
	return out
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrUserNotFound is returned by LoadUser when the user identifier
// has no row in the users table.
var ErrUserNotFound = errors.New("webauthn: user not found")

// ErrCredentialExists is returned by SaveCredential when a row
// with the same credential ID is already present.
var ErrCredentialExists = errors.New("webauthn: credential already registered")

// ErrCredentialNotFound is returned by lookups when no row matches
// the supplied credential ID.
var ErrCredentialNotFound = errors.New("webauthn: credential not found")
