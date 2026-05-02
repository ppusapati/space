// rectify.go — Article 16 (Right to rectification).
//
// The user can update incorrect personal data. The chetana IAM
// service supports rectifying the display-form email; password
// changes go through TASK-P1-IAM-008's reset flow (Article 16
// applies to data accuracy, not credential rotation).
//
// Email rectification rules:
//
//   • email_display is updated freely (the human-readable form).
//   • email_lower is recomputed from the new display address with
//     standard lowercase + trim normalisation.
//   • The (tenant_id, email_lower) UNIQUE constraint catches
//     collisions — the call returns ErrEmailInUse so the caller
//     can show a friendly message.
//
// We do NOT update an already-erased account. A user whose row
// has gdpr_anonymized_at set must register a new account if they
// want their data restored — Article 17 erasure is intentionally
// irreversible.

package gdpr

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RectifyEmailRequest is the per-call input.
type RectifyEmailRequest struct {
	UserID       string
	NewEmail     string
	RequestorIP  string
	RequestorUA  string
}

// RectifyEmailResult is the per-call output.
type RectifyEmailResult struct {
	UserID       string
	OldEmail     string
	NewEmail     string
	RectifiedAt  time.Time
}

// RectifyService handles Article 16 requests.
type RectifyService struct {
	pool *pgxpool.Pool
	clk  func() time.Time
}

// NewRectifyService wraps a pool. clock=nil → time.Now.
func NewRectifyService(pool *pgxpool.Pool, clock func() time.Time) (*RectifyService, error) {
	if pool == nil {
		return nil, errors.New("gdpr: nil pool")
	}
	if clock == nil {
		clock = time.Now
	}
	return &RectifyService{pool: pool, clk: clock}, nil
}

// RectifyEmail updates the user's email. Returns ErrEmailInUse on
// the (tenant_id, email_lower) UNIQUE collision; ErrAlreadyErased
// when the user has already exercised Article 17;
// ErrUserNotFound when the user_id is unknown; ErrInvalidEmail for
// shape violations.
func (r *RectifyService) RectifyEmail(ctx context.Context, in RectifyEmailRequest) (*RectifyEmailResult, error) {
	if in.UserID == "" {
		return nil, ErrUserNotFound
	}
	display := strings.TrimSpace(in.NewEmail)
	lower := strings.ToLower(display)
	if !looksLikeEmail(lower) {
		return nil, ErrInvalidEmail
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("gdpr: begin: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		oldEmail         string
		gdprAnonymizedAt *time.Time
	)
	err = tx.QueryRow(ctx,
		`SELECT email_display, gdpr_anonymized_at FROM users WHERE id = $1 FOR UPDATE`,
		in.UserID,
	).Scan(&oldEmail, &gdprAnonymizedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("gdpr: lookup user: %w", err)
	}
	if gdprAnonymizedAt != nil {
		return nil, ErrAlreadyErased
	}

	now := r.clk().UTC()
	tag, err := tx.Exec(ctx, `
UPDATE users SET
    email_lower = $2,
    email_display = $3,
    updated_at = $4
WHERE id = $1
`, in.UserID, lower, display, now)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailInUse
		}
		return nil, fmt.Errorf("gdpr: rectify update: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrUserNotFound
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("gdpr: commit: %w", err)
	}
	return &RectifyEmailResult{
		UserID:      in.UserID,
		OldEmail:    oldEmail,
		NewEmail:    display,
		RectifiedAt: now,
	}, nil
}

// looksLikeEmail is a deliberately loose check — RFC 5322 is too
// permissive to enforce here, and the IAM service does NOT do
// deliverability validation (that's the notify service's job).
// We just rule out the obvious bad shapes so a typo'd request
// fails before we touch the DB.
func looksLikeEmail(s string) bool {
	if len(s) < 3 || len(s) > 320 {
		return false
	}
	at := strings.IndexByte(s, '@')
	if at <= 0 || at == len(s)-1 {
		return false
	}
	if strings.IndexByte(s[at+1:], '.') < 0 {
		return false
	}
	if strings.ContainsAny(s, " \t\r\n") {
		return false
	}
	return true
}

// isUniqueViolation matches the pgconn error code 23505 (UNIQUE
// constraint failure). Mirrors the helper in internal/store but
// kept local so we don't reach into another package's unexported
// helper.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}

// ----------------------------------------------------------------------
// Errors
// ----------------------------------------------------------------------

// ErrInvalidEmail is returned by RectifyEmail for obviously bad
// email shapes (no @, no domain dot, whitespace, length out of
// bounds).
var ErrInvalidEmail = errors.New("gdpr: invalid email shape")

// ErrEmailInUse is returned when the new email collides with
// another user in the same tenant.
var ErrEmailInUse = errors.New("gdpr: email already in use")

// ErrAlreadyErased is returned when RectifyEmail is called on a
// user whose row has gdpr_anonymized_at set. Article 17 erasure
// is intentionally irreversible.
var ErrAlreadyErased = errors.New("gdpr: account already erased")
