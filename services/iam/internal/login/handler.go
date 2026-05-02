// handler.go — login RPC implementation.
//
// REQ-FUNC-PLT-IAM-001 (auth + Argon2id), REQ-FUNC-PLT-IAM-003
// (rate limit + lockout), REQ-FUNC-PLT-IAM-010 (constant-time
// response).
//
// Decision flow per Login(req):
//
//	1. Validate the request (non-empty email + password + IP).
//	2. IPLimiter.Allow → 429 if denied.
//	3. Lookup user by (tenant_id, email_lower):
//	     • not found  → constant-time delay + 401
//	     • disabled   → 401 (treat-as-not-found to keep enumeration
//	                          resistant)
//	     • locked     → 423 with Retry-After
//	4. Argon2id Verify the password.
//	     • mismatch   → RecordFailedLogin + 401
//	     • match      → RecordSuccessfulLogin + emit login.attempted
//	                    (success=true) + return tokens (empty in
//	                    Phase-1-IAM-001; populated in IAM-002)
//
// JWT + refresh-token issuance — TASK-P1-IAM-002. When a TokenIssuer
// is wired into HandlerConfig, a successful login also mints an
// access token + a fresh refresh-token family; both are surfaced on
// the Result and the connect mapper copies them onto the proto
// response. Handlers without an issuer keep the IAM-001 behaviour
// (auth + audit only) so unit tests don't require a key store.

package login

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

// FailureThresholdPerAccount is the per-account budget of consecutive
// failed login attempts before lockout escalates. REQ-FUNC-PLT-IAM-003.
const FailureThresholdPerAccount = 5

// constantTimeDelay is the artificial latency the handler adds to
// every login response so failed (user-not-found / wrong-password)
// and successful paths are indistinguishable to a timing attacker.
// Per REQ-FUNC-PLT-IAM-010.
//
// 250ms keeps the rate well above acceptable login UX (see TASK-P1-NFR-001
// which gates ≤100ms p95 — the 100ms target excludes the deliberate
// delay; the bench scenario subtracts it before measuring).
const constantTimeDelay = 250 * time.Millisecond

// AuditEmitter publishes a login event to the audit pipeline. The
// IAM service writes via Kafka in production (TASK-P1-AUDIT-001
// supplies the wire); a no-op implementation is used in unit tests.
type AuditEmitter interface {
	Emit(ctx context.Context, event Event) error
}

// Limiter is the rate-limiter shape the handler depends on. The
// production *IPLimiter satisfies it; tests pass a fake.
type Limiter interface {
	Allow(ctx context.Context, ip string) (LimitResult, error)
}

// UserStore is the user-CRUD shape the handler depends on. The
// production *store.Store satisfies it; tests pass a fake.
type UserStore interface {
	GetByEmail(ctx context.Context, tenantID, emailLower string) (*store.User, error)
	RecordSuccessfulLogin(ctx context.Context, userID string, now time.Time) error
	RecordFailedLogin(ctx context.Context, userID string, threshold int, now time.Time) (*store.User, error)
}

// Event is the event payload emitted to the audit pipeline.
type Event struct {
	TenantID    string    `json:"tenant_id"`
	UserID      string    `json:"user_id"`     // empty when user did not exist
	EmailLower  string    `json:"email_lower"` // recorded for forensic analysis
	ClientIP    string    `json:"client_ip"`
	UserAgent   string    `json:"user_agent"`
	OccurredAt  time.Time `json:"occurred_at"`
	Outcome     Outcome   `json:"outcome"`
	Reason      string    `json:"reason,omitempty"`
}

// Outcome enumerates the login outcomes the audit pipeline records.
type Outcome string

// Canonical login outcomes.
const (
	OutcomeSuccess        Outcome = "success"
	OutcomeBadCredentials Outcome = "bad_credentials"
	OutcomeUserNotFound   Outcome = "user_not_found"
	OutcomeUserDisabled   Outcome = "user_disabled"
	OutcomeLocked         Outcome = "locked"
	OutcomeRateLimited    Outcome = "rate_limited"
	OutcomeError          Outcome = "internal_error"
)

// Result is the structured result of Handler.Login. Carries enough
// context for the RPC layer to translate into the appropriate HTTP
// status code (200/401/423/429/500). The Connect handler defined
// alongside this package takes care of the proto / status mapping.
type Result struct {
	Status              ResultStatus
	UserID              string
	SessionID           string
	RetryAfter          time.Duration // populated for RateLimited + Locked
	Reason              string
	AccessToken         string
	AccessTokenExpires  time.Time
	RefreshToken        string
	RefreshTokenExpires time.Time
}

// TokenIssuer mints an access token + a fresh refresh-token family
// for a successful login. Implemented by services/iam/internal/token
// (Issuer + RefreshStore wrapper). Optional: if Handler is configured
// without one, the handler returns Result without token fields.
type TokenIssuer interface {
	IssueLoginTokens(ctx context.Context, in TokenIssueInput) (TokenIssueOutput, error)
}

// TokenIssueInput is the per-login payload handed to TokenIssuer.
type TokenIssueInput struct {
	UserID         string
	TenantID       string
	SessionID      string
	IsUSPerson     bool
	ClearanceLevel string
	Nationality    string
	Roles          []string
	Scopes         []string
	AMR            []string
}

// TokenIssueOutput carries the access + refresh credentials returned
// to the client.
type TokenIssueOutput struct {
	AccessToken         string
	AccessTokenExpires  time.Time
	RefreshToken        string
	RefreshTokenExpires time.Time
}

// ResultStatus enumerates the broad outcomes of a Login call.
type ResultStatus string

// Canonical result statuses.
const (
	ResultOK            ResultStatus = "ok"
	ResultBadCredentials ResultStatus = "bad_credentials" // 401
	ResultLocked         ResultStatus = "locked"          // 423
	ResultRateLimited    ResultStatus = "rate_limited"    // 429
	ResultDisabled       ResultStatus = "disabled"        // 401 (enum-resist)
	ResultInternalError  ResultStatus = "internal_error"  // 500
)

// Handler is the login flow orchestrator. Construct with NewHandler;
// the type holds references to its three collaborators (limiter,
// store, audit) and one configuration value (the active tenant ID).
type Handler struct {
	limiter Limiter
	users   UserStore
	audit   AuditEmitter
	cfg     HandlerConfig
}

// HandlerConfig configures the login handler. TenantID is the active
// tenant in the single-tenant runtime (TASK-P1-TENANT-001 keeps
// this constant in v1; multi-tenant lookup lands in v1.x via
// principal context).
type HandlerConfig struct {
	TenantID string
	// PasswordPolicy governs hashing parameters. Defaults to
	// password.PolicyV1.
	PasswordPolicy password.Policy
	// Now is the clock; tests inject. nil → time.Now.
	Now func() time.Time
	// SleepUntil sleeps until t. Tests inject a no-op so the
	// constant-time delay does not blow up the unit-test runtime.
	SleepUntil func(ctx context.Context, t time.Time) error
	// Tokens is the optional access/refresh token issuer. nil →
	// handler returns Result without token fields (used by unit
	// tests that don't exercise the token surface).
	Tokens TokenIssuer
	// AMR is the auth-methods-references list stamped onto issued
	// tokens. Defaults to {"pwd"} (single-factor password). Once
	// MFA lands the handler will append "mfa".
	AMR []string
}

// NewHandler builds a Handler. limiter / users / audit MUST be
// non-nil; cfg.TenantID MUST be set.
func NewHandler(limiter Limiter, users UserStore, audit AuditEmitter, cfg HandlerConfig) (*Handler, error) {
	if limiter == nil {
		return nil, errors.New("login: nil limiter")
	}
	if users == nil {
		return nil, errors.New("login: nil users store")
	}
	if audit == nil {
		return nil, errors.New("login: nil audit emitter")
	}
	if cfg.TenantID == "" {
		return nil, errors.New("login: empty tenant_id")
	}
	if cfg.PasswordPolicy.MemoryKiB == 0 {
		cfg.PasswordPolicy = password.PolicyV1
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.SleepUntil == nil {
		cfg.SleepUntil = realSleepUntil
	}
	if len(cfg.AMR) == 0 {
		cfg.AMR = []string{"pwd"}
	}
	return &Handler{
		limiter: limiter,
		users:   users,
		audit:   audit,
		cfg:     cfg,
	}, nil
}

// LoginInput is the handler's input — a thin wrapper around the
// proto request that does NOT take a dependency on the generated
// proto types (so this package builds without buf-generate).
type LoginInput struct {
	Email     string
	Password  string
	ClientIP  string
	UserAgent string
}

// Login is the main entry point. Returns Result + nil for normal
// outcomes; returns Result + non-nil err only for unexpected errors
// (Redis down, Postgres down, etc.) — those should propagate to the
// RPC layer as Internal Server Error.
func (h *Handler) Login(ctx context.Context, in LoginInput) (Result, error) {
	now := h.cfg.Now()
	deadline := now.Add(constantTimeDelay)

	// Always ensure we wait until the constant-time deadline before
	// returning, regardless of which branch we hit. We capture the
	// deferred sleep here so even error returns honour it.
	finish := func(res Result, err error) (Result, error) {
		_ = h.cfg.SleepUntil(ctx, deadline)
		return res, err
	}

	if strings.TrimSpace(in.Email) == "" || in.Password == "" {
		h.emitOrLog(ctx, Event{
			TenantID:   h.cfg.TenantID,
			ClientIP:   in.ClientIP,
			UserAgent:  in.UserAgent,
			OccurredAt: now,
			Outcome:    OutcomeBadCredentials,
			Reason:     "empty email or password",
		})
		return finish(Result{Status: ResultBadCredentials, Reason: "missing credentials"}, nil)
	}

	// 1. Per-IP sliding window.
	limit, err := h.limiter.Allow(ctx, in.ClientIP)
	if err != nil {
		h.emitOrLog(ctx, Event{
			TenantID:   h.cfg.TenantID,
			EmailLower: strings.ToLower(in.Email),
			ClientIP:   in.ClientIP,
			UserAgent:  in.UserAgent,
			OccurredAt: now,
			Outcome:    OutcomeError,
			Reason:     "rate-limit backend error",
		})
		return finish(Result{Status: ResultInternalError, Reason: "rate limiter error"}, err)
	}
	if !limit.Allowed {
		h.emitOrLog(ctx, Event{
			TenantID:   h.cfg.TenantID,
			EmailLower: strings.ToLower(in.Email),
			ClientIP:   in.ClientIP,
			UserAgent:  in.UserAgent,
			OccurredAt: now,
			Outcome:    OutcomeRateLimited,
			Reason:     fmt.Sprintf("ip %s hit %d/%d in window", in.ClientIP, limit.HitsInWindow, limit.Limit),
		})
		return finish(Result{Status: ResultRateLimited, RetryAfter: limit.RetryAfter}, nil)
	}

	emailLower := strings.ToLower(strings.TrimSpace(in.Email))

	// 2. User lookup.
	user, err := h.users.GetByEmail(ctx, h.cfg.TenantID, emailLower)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			// Constant-time response — never reveal which side of
			// the credential is wrong. Audit records the user-not-
			// found outcome so internal forensics can still see it.
			h.emitOrLog(ctx, Event{
				TenantID:   h.cfg.TenantID,
				EmailLower: emailLower,
				ClientIP:   in.ClientIP,
				UserAgent:  in.UserAgent,
				OccurredAt: now,
				Outcome:    OutcomeUserNotFound,
			})
			return finish(Result{Status: ResultBadCredentials, Reason: "invalid credentials"}, nil)
		}
		h.emitOrLog(ctx, Event{
			TenantID:   h.cfg.TenantID,
			EmailLower: emailLower,
			ClientIP:   in.ClientIP,
			UserAgent:  in.UserAgent,
			OccurredAt: now,
			Outcome:    OutcomeError,
			Reason:     "user lookup failed",
		})
		return finish(Result{Status: ResultInternalError, Reason: "user lookup error"}, err)
	}

	// 3. Lifecycle gates.
	if !user.IsActive() {
		h.emitOrLog(ctx, Event{
			TenantID:   h.cfg.TenantID,
			UserID:     user.ID,
			EmailLower: emailLower,
			ClientIP:   in.ClientIP,
			UserAgent:  in.UserAgent,
			OccurredAt: now,
			Outcome:    OutcomeUserDisabled,
			Reason:     fmt.Sprintf("status=%s", user.Status),
		})
		// Treat as bad-credentials to avoid leaking whether the
		// account exists in a disabled state.
		return finish(Result{Status: ResultBadCredentials, Reason: "invalid credentials"}, nil)
	}
	if user.IsLockedAt(now) {
		retry := user.LockoutRemaining(now)
		h.emitOrLog(ctx, Event{
			TenantID:   h.cfg.TenantID,
			UserID:     user.ID,
			EmailLower: emailLower,
			ClientIP:   in.ClientIP,
			UserAgent:  in.UserAgent,
			OccurredAt: now,
			Outcome:    OutcomeLocked,
			Reason:     fmt.Sprintf("level=%d remaining=%s", user.LockoutLevel, retry),
		})
		return finish(Result{Status: ResultLocked, UserID: user.ID, RetryAfter: retry}, nil)
	}

	// 4. Verify password.
	ok, err := password.Verify(in.Password, user.PasswordHash)
	if err != nil {
		// A weak / malformed stored hash. Audit + 500 — the
		// operator must investigate; do NOT silently treat as a
		// password mismatch (REQ-CONST-013 — never paper over).
		h.emitOrLog(ctx, Event{
			TenantID:   h.cfg.TenantID,
			UserID:     user.ID,
			EmailLower: emailLower,
			ClientIP:   in.ClientIP,
			UserAgent:  in.UserAgent,
			OccurredAt: now,
			Outcome:    OutcomeError,
			Reason:     fmt.Sprintf("password verify error: %v", err),
		})
		return finish(Result{Status: ResultInternalError, Reason: "password backend error"}, err)
	}
	if !ok {
		// Mismatch — increment failed-login counter; may escalate
		// lockout via store.RecordFailedLogin.
		updated, err := h.users.RecordFailedLogin(ctx, user.ID, FailureThresholdPerAccount, now)
		if err != nil {
			return finish(Result{Status: ResultInternalError, Reason: "failed-login record"}, err)
		}
		// If THIS attempt pushed us into lockout, prefer the locked
		// status code so the client + Retry-After header match
		// reality.
		if updated != nil && updated.IsLockedAt(now) {
			retry := updated.LockoutRemaining(now)
			h.emitOrLog(ctx, Event{
				TenantID:   h.cfg.TenantID,
				UserID:     user.ID,
				EmailLower: emailLower,
				ClientIP:   in.ClientIP,
				UserAgent:  in.UserAgent,
				OccurredAt: now,
				Outcome:    OutcomeLocked,
				Reason:     fmt.Sprintf("escalated to level %d", updated.LockoutLevel),
			})
			return finish(Result{Status: ResultLocked, UserID: user.ID, RetryAfter: retry}, nil)
		}
		h.emitOrLog(ctx, Event{
			TenantID:   h.cfg.TenantID,
			UserID:     user.ID,
			EmailLower: emailLower,
			ClientIP:   in.ClientIP,
			UserAgent:  in.UserAgent,
			OccurredAt: now,
			Outcome:    OutcomeBadCredentials,
		})
		return finish(Result{Status: ResultBadCredentials, Reason: "invalid credentials"}, nil)
	}

	// 5. Success.
	if err := h.users.RecordSuccessfulLogin(ctx, user.ID, now); err != nil {
		return finish(Result{Status: ResultInternalError, Reason: "success record"}, err)
	}

	sessionID, err := newSessionID()
	if err != nil {
		return finish(Result{Status: ResultInternalError, Reason: "session id"}, err)
	}

	res := Result{
		Status:    ResultOK,
		UserID:    user.ID,
		SessionID: sessionID,
	}

	if h.cfg.Tokens != nil {
		out, err := h.cfg.Tokens.IssueLoginTokens(ctx, TokenIssueInput{
			UserID:    user.ID,
			TenantID:  h.cfg.TenantID,
			SessionID: sessionID,
			AMR:       h.cfg.AMR,
			// Phase 1: clearance / nationality / role projection lands
			// once the user-attributes table (TASK-P1-IAM-USER-ATTRS)
			// ships. Until then we issue tokens with the conservative
			// default ("internal" clearance, no role grants).
			ClearanceLevel: "internal",
		})
		if err != nil {
			return finish(Result{Status: ResultInternalError, Reason: "token mint"}, err)
		}
		res.AccessToken = out.AccessToken
		res.AccessTokenExpires = out.AccessTokenExpires
		res.RefreshToken = out.RefreshToken
		res.RefreshTokenExpires = out.RefreshTokenExpires
	}

	h.emitOrLog(ctx, Event{
		TenantID:   h.cfg.TenantID,
		UserID:     user.ID,
		EmailLower: emailLower,
		ClientIP:   in.ClientIP,
		UserAgent:  in.UserAgent,
		OccurredAt: now,
		Outcome:    OutcomeSuccess,
	})

	return finish(res, nil)
}

// emitOrLog publishes the event to the audit pipeline. We do NOT
// fail the login if the audit emit fails — the audit pipeline is a
// best-effort secondary writer; the primary flow MUST keep working
// when Kafka is down. Errors are logged via the supplied audit
// emitter's own logger (caller-side).
func (h *Handler) emitOrLog(ctx context.Context, e Event) {
	_ = h.audit.Emit(ctx, e)
}

// newSessionID returns a 128-bit random hex token. Used as the
// session_id claim in the JWT once TASK-P1-IAM-002 lands; for now
// it's just a stable identifier returned in the LoginResponse.
func newSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("login: read random: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// realSleepUntil sleeps until t, honouring context cancellation. The
// public test surface uses a no-op variant.
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

// NopAudit is a no-op AuditEmitter useful for tests. Real services
// supply a Kafka-backed implementation wired in TASK-P1-AUDIT-001.
type NopAudit struct{}

// Emit implements AuditEmitter.
func (NopAudit) Emit(_ context.Context, _ Event) error { return nil }
