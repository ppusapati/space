package login

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ppusapati/space/services/iam/internal/password"
	"github.com/ppusapati/space/services/iam/internal/store"
)

// ----------------------------------------------------------------------
// Fakes
// ----------------------------------------------------------------------

// fakeLimiter records every Allow call and returns a pre-configured
// LimitResult. Tests set Verdict per scenario.
type fakeLimiter struct {
	mu      sync.Mutex
	calls   int
	verdict LimitResult
	err     error
}

func (f *fakeLimiter) Allow(_ context.Context, _ string) (LimitResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.verdict, f.err
}

// fakeUsers is an in-memory UserStore good enough for the handler
// test. Production lives in services/iam/internal/store; the e2e
// test exercises that path against a real Postgres.
type fakeUsers struct {
	mu             sync.Mutex
	byID           map[string]*store.User
	byEmail        map[string]*store.User
	getErr         error
	successErr     error
	failedErr      error
	successCalls   int
	failedCalls    int
	postFailedRow  *store.User // returned from RecordFailedLogin
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{
		byID:    make(map[string]*store.User),
		byEmail: make(map[string]*store.User),
	}
}

func (f *fakeUsers) seed(u *store.User) {
	f.byID[u.ID] = u
	f.byEmail[u.EmailLower] = u
}

func (f *fakeUsers) GetByEmail(_ context.Context, _, emailLower string) (*store.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.getErr != nil {
		return nil, f.getErr
	}
	u, ok := f.byEmail[emailLower]
	if !ok {
		return nil, store.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeUsers) RecordSuccessfulLogin(_ context.Context, userID string, now time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.successCalls++
	if f.successErr != nil {
		return f.successErr
	}
	if u, ok := f.byID[userID]; ok {
		u.FailedLoginCount = 0
		u.LockedUntil = nil
		u.LockoutLevel = 0
		u.LastLoginAt = &now
	}
	return nil
}

func (f *fakeUsers) RecordFailedLogin(_ context.Context, userID string, _ int, _ time.Time) (*store.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failedCalls++
	if f.failedErr != nil {
		return nil, f.failedErr
	}
	if f.postFailedRow != nil {
		return f.postFailedRow, nil
	}
	if u, ok := f.byID[userID]; ok {
		u.FailedLoginCount++
		return u, nil
	}
	return nil, store.ErrUserNotFound
}

// recordingAudit captures every Emit call.
type recordingAudit struct {
	mu     sync.Mutex
	events []Event
	err    error
}

func (r *recordingAudit) Emit(_ context.Context, e Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.events = append(r.events, e)
	return r.err
}

// noSleep skips the constant-time delay so tests run fast. The
// real handler uses realSleepUntil.
func noSleep(_ context.Context, _ time.Time) error { return nil }

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------

const testTenantID = "00000000-0000-0000-0000-000000000001"

func mustHash(t *testing.T, pw string) string {
	t.Helper()
	h, err := password.Hash(pw, password.PolicyV1)
	if err != nil {
		t.Fatalf("password.Hash: %v", err)
	}
	return h
}

func newSeededHandler(t *testing.T, limiter Limiter, users UserStore, audit AuditEmitter, frozenNow time.Time) *Handler {
	t.Helper()
	h, err := NewHandler(limiter, users, audit, HandlerConfig{
		TenantID:   testTenantID,
		Now:        func() time.Time { return frozenNow },
		SleepUntil: noSleep,
	})
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h
}

// ----------------------------------------------------------------------
// Tests
// ----------------------------------------------------------------------

// TestNewHandler_RejectsNilCollaborators covers the constructor
// guards.
func TestNewHandler_RejectsNilCollaborators(t *testing.T) {
	cases := map[string]struct {
		limiter Limiter
		users   UserStore
		audit   AuditEmitter
		tenant  string
	}{
		"nil limiter": {nil, newFakeUsers(), &recordingAudit{}, testTenantID},
		"nil users":   {&fakeLimiter{verdict: LimitResult{Allowed: true}}, nil, &recordingAudit{}, testTenantID},
		"nil audit":   {&fakeLimiter{verdict: LimitResult{Allowed: true}}, newFakeUsers(), nil, testTenantID},
		"empty tenant": {&fakeLimiter{verdict: LimitResult{Allowed: true}}, newFakeUsers(), &recordingAudit{}, ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := NewHandler(c.limiter, c.users, c.audit, HandlerConfig{TenantID: c.tenant})
			if err == nil {
				t.Errorf("expected error for %s", name)
			}
		})
	}
}

// TestLogin_HappyPath covers the success flow.
func TestLogin_HappyPath(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	users := newFakeUsers()
	users.seed(&store.User{
		ID:           "user-1",
		TenantID:     testTenantID,
		EmailLower:   "alice@example.com",
		EmailDisplay: "Alice@Example.com",
		PasswordHash: mustHash(t, "correct horse battery"),
		Status:       store.StatusActive,
	})
	limiter := &fakeLimiter{verdict: LimitResult{Allowed: true, HitsInWindow: 1, Limit: 10}}
	audit := &recordingAudit{}

	h := newSeededHandler(t, limiter, users, audit, now)

	res, err := h.Login(context.Background(), LoginInput{
		Email:    "Alice@Example.com",
		Password: "correct horse battery",
		ClientIP: "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("Login err: %v", err)
	}
	if res.Status != ResultOK {
		t.Errorf("Status=%v, want OK", res.Status)
	}
	if res.UserID != "user-1" {
		t.Errorf("UserID=%q, want user-1", res.UserID)
	}
	if res.SessionID == "" {
		t.Error("SessionID empty")
	}
	if users.successCalls != 1 {
		t.Errorf("RecordSuccessfulLogin calls=%d, want 1", users.successCalls)
	}
	// Audit MUST emit success.
	if len(audit.events) != 1 || audit.events[0].Outcome != OutcomeSuccess {
		t.Errorf("expected one success audit event; got %+v", audit.events)
	}
}

// TestLogin_WrongPassword exercises the bad-credentials path:
// returns 401, increments failed counter, audits bad_credentials.
func TestLogin_WrongPassword(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	users := newFakeUsers()
	users.seed(&store.User{
		ID:           "user-1",
		TenantID:     testTenantID,
		EmailLower:   "alice@example.com",
		PasswordHash: mustHash(t, "right"),
		Status:       store.StatusActive,
	})
	limiter := &fakeLimiter{verdict: LimitResult{Allowed: true}}
	audit := &recordingAudit{}

	h := newSeededHandler(t, limiter, users, audit, now)

	res, err := h.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "wrong",
		ClientIP: "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("Login err: %v", err)
	}
	if res.Status != ResultBadCredentials {
		t.Errorf("Status=%v, want BadCredentials", res.Status)
	}
	if users.failedCalls != 1 {
		t.Errorf("RecordFailedLogin calls=%d, want 1", users.failedCalls)
	}
	if len(audit.events) != 1 || audit.events[0].Outcome != OutcomeBadCredentials {
		t.Errorf("expected bad_credentials audit; got %+v", audit.events)
	}
}

// TestLogin_UserNotFoundReturnsBadCredentials — same response as
// wrong password (constant-time + enumeration-resistant).
func TestLogin_UserNotFoundReturnsBadCredentials(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	limiter := &fakeLimiter{verdict: LimitResult{Allowed: true}}
	audit := &recordingAudit{}

	h := newSeededHandler(t, limiter, newFakeUsers(), audit, now)

	res, err := h.Login(context.Background(), LoginInput{
		Email:    "ghost@example.com",
		Password: "anything",
		ClientIP: "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("Login err: %v", err)
	}
	if res.Status != ResultBadCredentials {
		t.Errorf("Status=%v, want BadCredentials", res.Status)
	}
	if len(audit.events) != 1 || audit.events[0].Outcome != OutcomeUserNotFound {
		t.Errorf("expected user_not_found audit; got %+v", audit.events)
	}
}

// TestLogin_DisabledUserReturnsBadCredentials — disabled accounts
// MUST be indistinguishable from non-existent ones.
func TestLogin_DisabledUserReturnsBadCredentials(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	users := newFakeUsers()
	users.seed(&store.User{
		ID:           "user-1",
		TenantID:     testTenantID,
		EmailLower:   "alice@example.com",
		PasswordHash: mustHash(t, "right"),
		Status:       store.StatusDisabled,
	})
	limiter := &fakeLimiter{verdict: LimitResult{Allowed: true}}
	audit := &recordingAudit{}
	h := newSeededHandler(t, limiter, users, audit, now)

	res, err := h.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "right",
		ClientIP: "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("Login err: %v", err)
	}
	if res.Status != ResultBadCredentials {
		t.Errorf("Status=%v, want BadCredentials (enum-resist)", res.Status)
	}
	if len(audit.events) != 1 || audit.events[0].Outcome != OutcomeUserDisabled {
		t.Errorf("expected user_disabled audit; got %+v", audit.events)
	}
}

// TestLogin_LockedAccountReturnsLocked covers the lockout window.
func TestLogin_LockedAccountReturnsLocked(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	until := now.Add(15 * time.Minute)
	users := newFakeUsers()
	users.seed(&store.User{
		ID:           "user-1",
		TenantID:     testTenantID,
		EmailLower:   "alice@example.com",
		PasswordHash: mustHash(t, "right"),
		Status:       store.StatusActive,
		LockedUntil:  &until,
		LockoutLevel: 1,
	})
	limiter := &fakeLimiter{verdict: LimitResult{Allowed: true}}
	audit := &recordingAudit{}
	h := newSeededHandler(t, limiter, users, audit, now)

	res, err := h.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "right", // doesn't matter — locked path short-circuits
		ClientIP: "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("Login err: %v", err)
	}
	if res.Status != ResultLocked {
		t.Errorf("Status=%v, want Locked", res.Status)
	}
	if res.RetryAfter != 15*time.Minute {
		t.Errorf("RetryAfter=%v, want 15m", res.RetryAfter)
	}
	if len(audit.events) != 1 || audit.events[0].Outcome != OutcomeLocked {
		t.Errorf("expected locked audit; got %+v", audit.events)
	}
}

// TestLogin_FailedAttemptThatTriggersLockoutReturnsLocked —
// when RecordFailedLogin reports a row that is now locked, the
// handler MUST surface 423 not 401.
func TestLogin_FailedAttemptThatTriggersLockoutReturnsLocked(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	postLockUntil := now.Add(15 * time.Minute)

	users := newFakeUsers()
	users.seed(&store.User{
		ID:           "user-1",
		TenantID:     testTenantID,
		EmailLower:   "alice@example.com",
		PasswordHash: mustHash(t, "right"),
		Status:       store.StatusActive,
	})
	users.postFailedRow = &store.User{
		ID:           "user-1",
		TenantID:     testTenantID,
		EmailLower:   "alice@example.com",
		PasswordHash: mustHash(t, "right"),
		Status:       store.StatusActive,
		LockedUntil:  &postLockUntil,
		LockoutLevel: 1,
	}

	limiter := &fakeLimiter{verdict: LimitResult{Allowed: true}}
	audit := &recordingAudit{}
	h := newSeededHandler(t, limiter, users, audit, now)

	res, err := h.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "wrong",
		ClientIP: "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("Login err: %v", err)
	}
	if res.Status != ResultLocked {
		t.Errorf("Status=%v, want Locked", res.Status)
	}
	if res.RetryAfter <= 0 {
		t.Errorf("RetryAfter should be > 0; got %v", res.RetryAfter)
	}
	if len(audit.events) != 1 || audit.events[0].Outcome != OutcomeLocked {
		t.Errorf("expected locked audit; got %+v", audit.events)
	}
}

// TestLogin_RateLimitedReturns429 covers the per-IP gate.
func TestLogin_RateLimitedReturns429(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	limiter := &fakeLimiter{verdict: LimitResult{
		Allowed:      false,
		RetryAfter:   30 * time.Second,
		HitsInWindow: 11,
		Limit:        10,
	}}
	audit := &recordingAudit{}
	h := newSeededHandler(t, limiter, newFakeUsers(), audit, now)

	res, err := h.Login(context.Background(), LoginInput{
		Email:    "alice@example.com",
		Password: "anything",
		ClientIP: "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("Login err: %v", err)
	}
	if res.Status != ResultRateLimited {
		t.Errorf("Status=%v, want RateLimited", res.Status)
	}
	if res.RetryAfter != 30*time.Second {
		t.Errorf("RetryAfter=%v, want 30s", res.RetryAfter)
	}
	if len(audit.events) != 1 || audit.events[0].Outcome != OutcomeRateLimited {
		t.Errorf("expected rate_limited audit; got %+v", audit.events)
	}
}

// TestLogin_EmptyCredentialsRejected covers the input validation guard.
func TestLogin_EmptyCredentialsRejected(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	cases := map[string]LoginInput{
		"empty email":    {Email: "", Password: "x", ClientIP: "203.0.113.10"},
		"empty password": {Email: "x@y", Password: "", ClientIP: "203.0.113.10"},
		"both empty":     {Email: "", Password: "", ClientIP: "203.0.113.10"},
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			h := newSeededHandler(t,
				&fakeLimiter{verdict: LimitResult{Allowed: true}},
				newFakeUsers(),
				&recordingAudit{},
				now)
			res, err := h.Login(context.Background(), in)
			if err != nil {
				t.Fatalf("Login err: %v", err)
			}
			if res.Status != ResultBadCredentials {
				t.Errorf("Status=%v, want BadCredentials", res.Status)
			}
		})
	}
}

// TestLogin_RateLimiterBackendErrorReturns500 verifies that a
// limiter error propagates as ResultInternalError + audit
// internal_error.
func TestLogin_RateLimiterBackendErrorReturns500(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	limiter := &fakeLimiter{err: errors.New("redis down")}
	audit := &recordingAudit{}
	h := newSeededHandler(t, limiter, newFakeUsers(), audit, now)

	res, err := h.Login(context.Background(), LoginInput{
		Email: "x@y", Password: "p", ClientIP: "1.1.1.1",
	})
	if err == nil {
		t.Error("expected propagated error")
	}
	if res.Status != ResultInternalError {
		t.Errorf("Status=%v, want InternalError", res.Status)
	}
	if len(audit.events) != 1 || audit.events[0].Outcome != OutcomeError {
		t.Errorf("expected internal_error audit; got %+v", audit.events)
	}
}

// TestLogin_AuditFailureDoesNotBreakLoginPath — audit emit returning
// an error MUST NOT break login. Confirms the design choice to log
// audit failures rather than fail the request.
func TestLogin_AuditFailureDoesNotBreakLoginPath(t *testing.T) {
	now := time.Date(2026, 5, 2, 12, 0, 0, 0, time.UTC)
	users := newFakeUsers()
	users.seed(&store.User{
		ID:           "user-1",
		TenantID:     testTenantID,
		EmailLower:   "alice@example.com",
		PasswordHash: mustHash(t, "right"),
		Status:       store.StatusActive,
	})
	limiter := &fakeLimiter{verdict: LimitResult{Allowed: true}}
	audit := &recordingAudit{err: errors.New("kafka down")}
	h := newSeededHandler(t, limiter, users, audit, now)

	res, err := h.Login(context.Background(), LoginInput{
		Email: "alice@example.com", Password: "right", ClientIP: "1.1.1.1",
	})
	if err != nil {
		t.Errorf("Login should not return audit error; got %v", err)
	}
	if res.Status != ResultOK {
		t.Errorf("Status=%v, want OK", res.Status)
	}
}

// TestNewSessionID_StableShape covers the helper used to mint the
// session ID in success responses.
func TestNewSessionID_StableShape(t *testing.T) {
	a, err := newSessionID()
	if err != nil {
		t.Fatalf("newSessionID: %v", err)
	}
	if len(a) != 32 {
		t.Errorf("session id length=%d, want 32 hex chars", len(a))
	}
	b, _ := newSessionID()
	if a == b {
		t.Error("two newSessionID calls returned identical values — RNG broken?")
	}
}
