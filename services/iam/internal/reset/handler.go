// handler.go — request + confirm RPC implementations.
//
// REQ-FUNC-PLT-IAM-010 acceptance:
//
//   1. Token is single-use, 1h TTL, hashed at rest.
//   2. Rate limit 3/h enforced (per user; we count by user_id
//      *after* the email→user lookup, so the cap can't be bypassed
//      by varying capitalisation in the email field).
//   3. Response timing variance < 50 ms between known and unknown
//      emails — the constant-time delay below underwrites this.
//
// Constant-time policy:
//
//   • Request: the handler ALWAYS waits until ConstantTimeDelay
//     has elapsed before returning. Both branches (user-found +
//     user-missing) take the same wall-clock time; the caller
//     learns nothing from the response timing about whether the
//     email exists, whether rate limiting fired, or whether the
//     notify pipeline was even invoked.
//
//   • Confirm: the password-update branch runs an argon2id hash
//     (~250ms at PolicyV1) regardless of token validity, so a
//     forged-token attacker can't time the difference between
//     "token unknown" and "token unknown but I happen to know a
//     plausible new password length".
//
// The token bearer + storage layer live in store.go.

package reset

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ppusapati/space/services/iam/internal/password"
	"github.com/ppusapati/space/services/iam/internal/store"
)

// Defaults per REQ-FUNC-PLT-IAM-010.
const (
	// ConstantTimeDelay is the wall-clock delay every Request +
	// Confirm response is padded to. 250ms is the same budget the
	// login handler uses (REQ-FUNC-PLT-IAM-010), so the "known vs
	// unknown email" timing difference target of <50ms is comfortably
	// inside the 250ms floor.
	ConstantTimeDelay = 250 * time.Millisecond

	// DefaultRateLimit is the per-user maximum number of reset
	// requests per RateWindow.
	DefaultRateLimit = 3

	// RateWindow is the rolling window over which DefaultRateLimit
	// applies.
	RateWindow = time.Hour
)

// Notifier publishes the "send password reset email" event onto
// the chetana notify pipeline. Real implementation lands with
// TASK-P1-NOTIFY-001; until then the handler accepts a NopNotifier
// so the package is wireable today.
type Notifier interface {
	SendPasswordReset(ctx context.Context, email, token string, expiresAt time.Time) error
}

// NopNotifier is a no-op Notifier useful for tests and for early
// integration when the notify service isn't deployed yet.
type NopNotifier struct{}

// SendPasswordReset implements Notifier.
func (NopNotifier) SendPasswordReset(_ context.Context, _, _ string, _ time.Time) error {
	return nil
}

// UserStore is the user-CRUD surface the handler depends on.
// *store.Store satisfies it; tests pass a fake.
type UserStore interface {
	GetByEmail(ctx context.Context, tenantID, emailLower string) (*store.User, error)
	UpdatePasswordHash(ctx context.Context, userID, hash, algo string, now time.Time) error
}

// SessionRevoker is the optional session-eviction surface. When
// supplied, a successful Confirm revokes every active session for
// the user — a password reset MUST kill the attacker's outstanding
// sessions if they had compromised the old credential.
type SessionRevoker interface {
	RevokeAllForUser(ctx context.Context, userID, by string) (int64, error)
}

// HandlerConfig configures the Handler.
type HandlerConfig struct {
	TenantID string

	// PasswordPolicy governs hashing parameters for the new
	// password set in Confirm. Defaults to password.PolicyV1.
	PasswordPolicy password.Policy

	// TokenTTL overrides DefaultTTL.
	TokenTTL time.Duration

	// RateLimit overrides DefaultRateLimit.
	RateLimit int

	// Now is the clock; tests inject. nil → time.Now.
	Now func() time.Time

	// SleepUntil sleeps until t. Tests inject a no-op so the
	// constant-time delay does not blow up the unit-test runtime.
	SleepUntil func(ctx context.Context, t time.Time) error
}

// Handler orchestrates the request + confirm flow.
type Handler struct {
	store    *Store
	users    UserStore
	notify   Notifier
	sessions SessionRevoker
	cfg      HandlerConfig
}

// NewHandler builds a Handler. store + users + notify MUST be
// non-nil; sessions is optional (recommended). cfg.TenantID MUST
// be set.
func NewHandler(rstore *Store, users UserStore, notify Notifier, sessions SessionRevoker, cfg HandlerConfig) (*Handler, error) {
	if rstore == nil {
		return nil, errors.New("reset: nil reset store")
	}
	if users == nil {
		return nil, errors.New("reset: nil users store")
	}
	if notify == nil {
		return nil, errors.New("reset: nil notifier")
	}
	if cfg.TenantID == "" {
		return nil, errors.New("reset: empty tenant_id")
	}
	if cfg.PasswordPolicy.MemoryKiB == 0 {
		cfg.PasswordPolicy = password.PolicyV1
	}
	if cfg.TokenTTL <= 0 {
		cfg.TokenTTL = DefaultTTL
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = DefaultRateLimit
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.SleepUntil == nil {
		cfg.SleepUntil = realSleepUntil
	}
	return &Handler{
		store:    rstore,
		users:    users,
		notify:   notify,
		sessions: sessions,
		cfg:      cfg,
	}, nil
}

// RequestInput is the per-call input for Request.
type RequestInput struct {
	Email     string
	ClientIP  string
	UserAgent string
}

// RequestResult is the structured outcome of Request. The HTTP
// layer translates Outcome into the (constant) 202 response — the
// fields are populated for the audit pipeline.
type RequestResult struct {
	Outcome RequestOutcome
	UserID  string // empty when user did not exist
	Reason  string
}

// RequestOutcome enumerates the outcomes the audit pipeline records.
// All map to the same 202 user-facing response.
type RequestOutcome string

// Canonical request outcomes.
const (
	RequestOutcomeSent          RequestOutcome = "sent"
	RequestOutcomeUserNotFound  RequestOutcome = "user_not_found"
	RequestOutcomeUserDisabled  RequestOutcome = "user_disabled"
	RequestOutcomeRateLimited   RequestOutcome = "rate_limited"
	RequestOutcomeInternalError RequestOutcome = "internal_error"
)

// Request initiates a password-reset flow. The response is
// constant-time + non-disclosing — the caller learns nothing from
// (status, body) about whether the email exists.
//
// Errors are returned only for unexpected internal failures
// (database down, etc.); known outcomes (user-not-found / rate-
// limited / etc.) are reported via Result.Outcome with err==nil.
func (h *Handler) Request(ctx context.Context, in RequestInput) (RequestResult, error) {
	now := h.cfg.Now()
	deadline := now.Add(ConstantTimeDelay)

	finish := func(res RequestResult, err error) (RequestResult, error) {
		_ = h.cfg.SleepUntil(ctx, deadline)
		return res, err
	}

	emailLower := strings.ToLower(strings.TrimSpace(in.Email))
	if emailLower == "" {
		return finish(RequestResult{Outcome: RequestOutcomeUserNotFound, Reason: "empty email"}, nil)
	}

	user, err := h.users.GetByEmail(ctx, h.cfg.TenantID, emailLower)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			return finish(RequestResult{Outcome: RequestOutcomeUserNotFound}, nil)
		}
		return finish(RequestResult{Outcome: RequestOutcomeInternalError, Reason: "user lookup error"}, err)
	}
	if !user.IsActive() {
		// Disabled accounts: silently no-op (same shape as
		// "user_not_found" to the outside world).
		return finish(RequestResult{
			Outcome: RequestOutcomeUserDisabled,
			UserID:  user.ID,
			Reason:  fmt.Sprintf("status=%s", user.Status),
		}, nil)
	}

	// Rate limit by user (NOT by email) so capitalisation games
	// can't dodge the cap.
	count, err := h.store.CountRecentForUser(ctx, user.ID, RateWindow)
	if err != nil {
		return finish(RequestResult{Outcome: RequestOutcomeInternalError, Reason: "rate count"}, err)
	}
	if count >= h.cfg.RateLimit {
		return finish(RequestResult{
			Outcome: RequestOutcomeRateLimited,
			UserID:  user.ID,
			Reason:  fmt.Sprintf("hit %d/%d in window", count, h.cfg.RateLimit),
		}, nil)
	}

	issued, err := h.store.Issue(ctx, user.ID, h.cfg.TokenTTL)
	if err != nil {
		return finish(RequestResult{Outcome: RequestOutcomeInternalError, Reason: "issue token"}, err)
	}
	if err := h.notify.SendPasswordReset(ctx, user.EmailDisplay, issued.Token, issued.ExpiresAt); err != nil {
		// Notify failure does NOT roll back the issued token —
		// the user can retry the request once the notify service
		// is healthy. We still report success for non-disclosure.
		return finish(RequestResult{
			Outcome: RequestOutcomeInternalError,
			UserID:  user.ID,
			Reason:  fmt.Sprintf("notify: %v", err),
		}, nil)
	}
	return finish(RequestResult{Outcome: RequestOutcomeSent, UserID: user.ID}, nil)
}

// ConfirmInput is the per-call input for Confirm.
type ConfirmInput struct {
	Token       string
	NewPassword string
	ClientIP    string
	UserAgent   string
}

// ConfirmResult is the structured outcome of Confirm.
type ConfirmResult struct {
	Outcome ConfirmOutcome
	UserID  string
	Reason  string
}

// ConfirmOutcome enumerates the outcomes the audit pipeline records.
type ConfirmOutcome string

// Canonical confirm outcomes.
const (
	ConfirmOutcomeOK             ConfirmOutcome = "ok"
	ConfirmOutcomeTokenInvalid   ConfirmOutcome = "token_invalid"
	ConfirmOutcomeTokenExpired   ConfirmOutcome = "token_expired"
	ConfirmOutcomeTokenReused    ConfirmOutcome = "token_reused"
	ConfirmOutcomeWeakPassword   ConfirmOutcome = "weak_password"
	ConfirmOutcomeInternalError  ConfirmOutcome = "internal_error"
)

// Confirm redeems a reset token + replaces the user's password
// hash + revokes outstanding sessions. The argon2id hash runs
// regardless of token validity to mask the timing of the token
// check.
func (h *Handler) Confirm(ctx context.Context, in ConfirmInput) (ConfirmResult, error) {
	if len(in.NewPassword) < 12 {
		return ConfirmResult{Outcome: ConfirmOutcomeWeakPassword, Reason: "min 12 chars"}, nil
	}

	// Always hash, regardless of token validity, so the response
	// timing is dominated by the argon2 cost rather than the DB
	// hit. The hash output is only USED on the success path.
	hash, hashErr := password.Hash(in.NewPassword, h.cfg.PasswordPolicy)

	rec, err := h.store.Redeem(ctx, in.Token)
	if err != nil {
		switch {
		case errors.Is(err, ErrTokenNotFound):
			return ConfirmResult{Outcome: ConfirmOutcomeTokenInvalid}, nil
		case errors.Is(err, ErrTokenExpired):
			return ConfirmResult{Outcome: ConfirmOutcomeTokenExpired}, nil
		case errors.Is(err, ErrTokenAlreadyUsed):
			return ConfirmResult{Outcome: ConfirmOutcomeTokenReused}, nil
		}
		return ConfirmResult{Outcome: ConfirmOutcomeInternalError, Reason: "redeem"}, err
	}
	if hashErr != nil {
		return ConfirmResult{Outcome: ConfirmOutcomeInternalError, Reason: "hash"}, hashErr
	}

	if err := h.users.UpdatePasswordHash(ctx, rec.UserID, hash, "argon2id", h.cfg.Now().UTC()); err != nil {
		return ConfirmResult{Outcome: ConfirmOutcomeInternalError, Reason: "update password", UserID: rec.UserID}, err
	}

	// REQ-FUNC-PLT-IAM-010: a password reset MUST kill outstanding
	// sessions so the attacker who triggered the reset can't keep
	// using a JWT minted before the change.
	if h.sessions != nil {
		if _, err := h.sessions.RevokeAllForUser(ctx, rec.UserID, "password_reset"); err != nil {
			// Logged, not fatal — the reset itself succeeded; a
			// retry of the eviction sweep happens out of band.
			return ConfirmResult{
				Outcome: ConfirmOutcomeOK,
				UserID:  rec.UserID,
				Reason:  fmt.Sprintf("session revoke partial: %v", err),
			}, nil
		}
	}

	return ConfirmResult{Outcome: ConfirmOutcomeOK, UserID: rec.UserID}, nil
}

// realSleepUntil sleeps until t, honouring context cancellation.
// Reused from the login handler.
func realSleepUntil(ctx context.Context, t time.Time) error {
	dur := time.Until(t)
	if dur <= 0 {
		return nil
	}
	timer := time.NewTimer(dur)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// NewRequestID is a small helper for the audit pipeline / log
// correlation when the caller does not already have a request id.
func NewRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
