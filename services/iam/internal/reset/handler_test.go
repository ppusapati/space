package reset

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

type fakeUsers struct {
	mu       sync.Mutex
	byEmail  map[string]*store.User
	updateOK map[string]bool
	getErr   error
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{
		byEmail:  map[string]*store.User{},
		updateOK: map[string]bool{},
	}
}

func (f *fakeUsers) seed(u *store.User) {
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

func (f *fakeUsers) UpdatePasswordHash(_ context.Context, userID, _, _ string, _ time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.updateOK[userID] = true
	return nil
}

type fakeNotify struct {
	mu       sync.Mutex
	sent     []sentMsg
	failNext error
}

type sentMsg struct {
	email     string
	token     string
	expiresAt time.Time
}

func (f *fakeNotify) SendPasswordReset(_ context.Context, email, token string, expiresAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failNext != nil {
		err := f.failNext
		f.failNext = nil
		return err
	}
	f.sent = append(f.sent, sentMsg{email: email, token: token, expiresAt: expiresAt})
	return nil
}

func (f *fakeNotify) sentCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sent)
}

// fakeRevoker records every revoke call.
type fakeRevoker struct {
	mu      sync.Mutex
	revoked map[string]int
}

func newFakeRevoker() *fakeRevoker {
	return &fakeRevoker{revoked: map[string]int{}}
}

func (f *fakeRevoker) RevokeAllForUser(_ context.Context, userID, _ string) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.revoked[userID]++
	return 1, nil
}

// ----------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------

const (
	testTenant = "00000000-0000-0000-0000-000000000001"
	testUser   = "11111111-1111-1111-1111-111111111111"
)

func newActiveUser() *store.User {
	return &store.User{
		ID:                 testUser,
		TenantID:           testTenant,
		EmailLower:         "alice@example.com",
		EmailDisplay:       "Alice <alice@example.com>",
		Status:             store.StatusActive,
		DataClassification: "cui",
	}
}

func nopSleep(_ context.Context, _ time.Time) error { return nil }

// ----------------------------------------------------------------------
// Tests
// ----------------------------------------------------------------------

func TestNewHandler_Validation(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*HandlerConfig)
	}{
		{"empty tenant", func(c *HandlerConfig) { c.TenantID = "" }},
	}
	users := newFakeUsers()
	notify := &fakeNotify{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := HandlerConfig{TenantID: testTenant, SleepUntil: nopSleep}
			tc.mut(&cfg)
			if _, err := NewHandler(&Store{}, users, notify, nil, cfg); err == nil {
				t.Fatal("expected error")
			}
		})
	}
	if _, err := NewHandler(nil, users, notify, nil, HandlerConfig{TenantID: testTenant}); err == nil {
		t.Error("nil reset store should error")
	}
	if _, err := NewHandler(&Store{}, nil, notify, nil, HandlerConfig{TenantID: testTenant}); err == nil {
		t.Error("nil users store should error")
	}
	if _, err := NewHandler(&Store{}, users, nil, nil, HandlerConfig{TenantID: testTenant}); err == nil {
		t.Error("nil notifier should error")
	}
}

// User-not-found maps to RequestOutcomeUserNotFound and emits no
// notification — but the call is constant-time.
func TestRequest_UnknownEmail_NoLeak(t *testing.T) {
	users := newFakeUsers()
	notify := &fakeNotify{}
	// Manager only needs the in-memory store; Issue/CountRecentForUser
	// would touch the DB but those code paths don't run for the
	// unknown-email branch. Construct a Store with a nil pool —
	// safe because we never reach into it.
	rstore := &Store{clk: time.Now}

	h, err := NewHandler(rstore, users, notify, nil, HandlerConfig{
		TenantID:   testTenant,
		SleepUntil: nopSleep,
	})
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	res, err := h.Request(context.Background(), RequestInput{Email: "ghost@example.com"})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if res.Outcome != RequestOutcomeUserNotFound {
		t.Errorf("outcome: %v", res.Outcome)
	}
	if notify.sentCount() != 0 {
		t.Error("notify must not fire for unknown email")
	}
}

func TestRequest_DisabledUser_SilentNoOp(t *testing.T) {
	users := newFakeUsers()
	disabled := newActiveUser()
	disabled.Status = store.StatusDisabled
	users.seed(disabled)
	notify := &fakeNotify{}
	rstore := &Store{clk: time.Now}

	h, _ := NewHandler(rstore, users, notify, nil, HandlerConfig{
		TenantID:   testTenant,
		SleepUntil: nopSleep,
	})
	res, err := h.Request(context.Background(), RequestInput{Email: "alice@example.com"})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if res.Outcome != RequestOutcomeUserDisabled {
		t.Errorf("outcome: %v", res.Outcome)
	}
	if notify.sentCount() != 0 {
		t.Error("notify must not fire for disabled user")
	}
}

func TestRequest_EmptyEmail(t *testing.T) {
	h, _ := NewHandler(&Store{clk: time.Now}, newFakeUsers(), &fakeNotify{}, nil, HandlerConfig{
		TenantID:   testTenant,
		SleepUntil: nopSleep,
	})
	res, _ := h.Request(context.Background(), RequestInput{Email: "  "})
	if res.Outcome != RequestOutcomeUserNotFound {
		t.Errorf("outcome: %v", res.Outcome)
	}
}

func TestConfirm_WeakPasswordRejected(t *testing.T) {
	h, _ := NewHandler(&Store{clk: time.Now}, newFakeUsers(), &fakeNotify{}, nil, HandlerConfig{
		TenantID:   testTenant,
		SleepUntil: nopSleep,
	})
	res, err := h.Confirm(context.Background(), ConfirmInput{
		Token:       "rowid.dGVzdA",
		NewPassword: "short",
	})
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if res.Outcome != ConfirmOutcomeWeakPassword {
		t.Errorf("outcome: %v", res.Outcome)
	}
}

func TestConfirm_TokenInvalid(t *testing.T) {
	rstore := &Store{clk: time.Now} // nil pool — Redeem will fail at decode for a malformed bearer
	h, _ := NewHandler(rstore, newFakeUsers(), &fakeNotify{}, nil, HandlerConfig{
		TenantID:   testTenant,
		SleepUntil: nopSleep,
	})
	res, err := h.Confirm(context.Background(), ConfirmInput{
		Token:       "no-dot-no-good",
		NewPassword: "this-password-is-long-enough",
	})
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if res.Outcome != ConfirmOutcomeTokenInvalid {
		t.Errorf("outcome: %v", res.Outcome)
	}
}

func TestNopNotifier(t *testing.T) {
	if err := (NopNotifier{}).SendPasswordReset(context.Background(), "x@x", "tok", time.Now()); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

// PolicyV1 fields are exposed elsewhere; just sanity-check Hash
// works for the default path our handler uses.
func TestConfirm_DefaultPolicyHashesSuccessfully(t *testing.T) {
	if _, err := password.Hash("a-strong-password-12345", password.PolicyV1); err != nil {
		t.Fatalf("hash: %v", err)
	}
}

func TestNewRequestID(t *testing.T) {
	a := NewRequestID()
	b := NewRequestID()
	if len(a) != 16 || len(b) != 16 {
		t.Errorf("len: %d %d", len(a), len(b))
	}
	if a == b {
		t.Error("two ids must differ")
	}
}

func TestSentinelErrors(t *testing.T) {
	for _, e := range []error{ErrTokenNotFound, ErrTokenExpired, ErrTokenAlreadyUsed} {
		if !errors.Is(e, e) {
			t.Errorf("not reflexive: %v", e)
		}
		if e.Error() == "" {
			t.Errorf("empty error: %v", e)
		}
	}
}
