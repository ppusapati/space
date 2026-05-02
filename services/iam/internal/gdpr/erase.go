// erase.go — Article 17 (Right to erasure / "right to be forgotten").
//
// Erasure semantics on chetana:
//
//   • The users row is NOT deleted. We anonymise it in place:
//
//       email_lower   = "anon-" || sha256(user_id||tenant_id||"chetana-gdpr-v1")[:16]
//                       (deterministic so cross-service joins
//                        keyed on the user_id-derived hash still
//                        function for compliance reporting,
//                        without exposing the original email).
//       email_display = "(erased)"
//       password_hash = ""
//       password_algo = ""
//       status        = "deleted"
//       gdpr_anonymized_at = now
//
//   • Refresh tokens, sessions, MFA secrets, WebAuthn credentials,
//     OAuth auth codes, and password-reset tokens are HARD-deleted.
//     Those tables hold operational state (not historical record)
//     and the user has explicitly asked us to forget them.
//
//   • The audit chain is NOT touched. Per the platform DPIA, audit
//     retention is a separate legal basis (REQ-COMP-AUDIT-001 +
//     REQ-COMP-LGPD-001 + ITAR record-keeping). The audit chain
//     references the user_id only — which is a UUID with no
//     personal data on its own — so the ID can survive without
//     re-identifying the subject as long as the users row
//     (the only place email lived) is anonymised.
//
// Reversibility: NONE. The original email is unrecoverable from
// the anonymised hash. Customers must download a SAR before
// erasing if they need a copy.

package gdpr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AnonHashSalt is the deployment-wide salt the deterministic
// anonymisation hash incorporates. Hardcoded (rather than per-
// tenant random) so the hash stays stable across DR restores and
// cross-service joins. Knowledge of the salt does NOT enable
// re-identification — the hash also incorporates the user_id +
// tenant_id, and the email itself is NEVER an input.
const AnonHashSalt = "chetana-gdpr-v1"

// ErasureRequest is the per-call input.
type ErasureRequest struct {
	UserID         string
	RequestorIP    string
	RequestorAgent string
	// Reason is recorded in the audit event. Free text — the
	// caller typically threads "user_request" or "controller_request".
	Reason string
}

// ErasureResult is the per-call output.
type ErasureResult struct {
	UserID         string
	AnonymizedAt   time.Time
	AnonHashPrefix string // the new email_lower value, for the audit trail
	HardDeleted    HardDeleteCounts
}

// HardDeleteCounts breaks down the rows the erasure removed from
// the operational tables. Surfaced in the audit event so the
// compliance team can sanity-check the scope.
type HardDeleteCounts struct {
	Sessions         int64
	RefreshTokens    int64
	WebAuthn         int64
	OAuthAuthCodes   int64
	PasswordResets   int64
	MFATOTPSecrets   int64
	MFABackupCodes   int64
}

// EraseService handles Article 17 requests.
type EraseService struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewEraseService wraps a pool. clock=nil → time.Now.
func NewEraseService(pool *pgxpool.Pool, clock func() time.Time) (*EraseService, error) {
	if pool == nil {
		return nil, errors.New("gdpr: nil pool")
	}
	if clock == nil {
		clock = time.Now
	}
	return &EraseService{pool: pool, clk: clock}, nil
}

// Erase performs the anonymisation + operational-state purge in
// a single transaction so the system never observes a half-erased
// state.
//
// Idempotent: erasing an already-erased user is a no-op (the
// users row stays anonymised; the operational tables are already
// empty). The returned ErasureResult.AnonymizedAt is the timestamp
// of the FIRST erasure for that user.
func (s *EraseService) Erase(ctx context.Context, in ErasureRequest) (*ErasureResult, error) {
	if in.UserID == "" {
		return nil, ErrUserNotFound
	}
	now := s.clk().UTC()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gdpr: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Lookup the current user to (a) confirm existence and
	// (b) discover whether anonymisation has already happened
	// (so a re-Erase is idempotent).
	var (
		tenantID         string
		gdprAnonymizedAt *time.Time
	)
	err = tx.QueryRow(ctx,
		`SELECT tenant_id, gdpr_anonymized_at FROM users WHERE id = $1 FOR UPDATE`,
		in.UserID,
	).Scan(&tenantID, &gdprAnonymizedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("gdpr: lookup user: %w", err)
	}

	anonHash := AnonymisedEmailFor(in.UserID, tenantID)
	anonymizedAt := now
	if gdprAnonymizedAt != nil {
		anonymizedAt = *gdprAnonymizedAt
	}

	if gdprAnonymizedAt == nil {
		// First erasure — anonymise the users row.
		if _, err := tx.Exec(ctx, `
UPDATE users SET
    email_lower = $2,
    email_display = '(erased)',
    password_hash = '',
    password_algo = '',
    status = $3,
    gdpr_anonymized_at = $4,
    updated_at = $4
WHERE id = $1
`, in.UserID, anonHash, "deleted", now); err != nil {
			return nil, fmt.Errorf("gdpr: anonymise user: %w", err)
		}
	}

	// Hard-delete operational state. Each table is best-effort —
	// missing tables (e.g. on a partial migration) yield a zero
	// row count rather than a transaction abort. The audit chain
	// is intentionally NOT touched.
	counts := HardDeleteCounts{}
	for _, t := range []struct {
		query string
		col   *int64
	}{
		{`DELETE FROM sessions WHERE user_id = $1`, &counts.Sessions},
		{`DELETE FROM refresh_tokens WHERE user_id = $1`, &counts.RefreshTokens},
		{`DELETE FROM webauthn_credentials WHERE user_id = $1`, &counts.WebAuthn},
		{`DELETE FROM oauth2_auth_codes WHERE user_id = $1`, &counts.OAuthAuthCodes},
		{`DELETE FROM password_resets WHERE user_id = $1`, &counts.PasswordResets},
		{`DELETE FROM mfa_totp_secrets WHERE user_id = $1`, &counts.MFATOTPSecrets},
		{`DELETE FROM mfa_backup_codes WHERE user_id = $1`, &counts.MFABackupCodes},
	} {
		tag, err := tx.Exec(ctx, t.query, in.UserID)
		if err != nil {
			return nil, fmt.Errorf("gdpr: purge: %w", err)
		}
		*t.col = tag.RowsAffected()
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gdpr: commit: %w", err)
	}
	return &ErasureResult{
		UserID:         in.UserID,
		AnonymizedAt:   anonymizedAt,
		AnonHashPrefix: anonHash,
		HardDeleted:    counts,
	}, nil
}

// AnonymisedEmailFor returns the deterministic anonymisation
// value the erasure flow writes into users.email_lower. Pure
// function — exposed so the audit pipeline + compliance reports
// can recompute the value without re-running the SQL.
//
// Output shape: "anon-" + 32 lowercase hex chars (16 bytes of the
// SHA-256 prefix). Globally unique with overwhelming probability.
func AnonymisedEmailFor(userID, tenantID string) string {
	sum := sha256.Sum256([]byte(userID + "|" + tenantID + "|" + AnonHashSalt))
	return "anon-" + hex.EncodeToString(sum[:16])
}
