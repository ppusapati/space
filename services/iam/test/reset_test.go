//go:build integration

// reset_test.go — TASK-P1-IAM-008 integration tests.
//
// Three acceptance criteria:
//
//   1. Token is single-use, 1h TTL, hashed at rest.
//   2. Rate limit 3/h enforced.
//   3. Response timing variance < 50ms between known and unknown
//      emails.

package iam_test

import (
	"context"
	"errors"
	"math"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/iam/internal/password"
	"github.com/ppusapati/space/services/iam/internal/reset"
	"github.com/ppusapati/space/services/iam/internal/store"
)

func newResetPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("IAM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("IAM_TEST_DATABASE_URL not set — skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`TRUNCATE password_resets, sessions, users RESTART IDENTITY CASCADE`)
		pool.Close()
	})
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE password_resets, sessions, users RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

const (
	resetTestTenant = "44444444-4444-4444-4444-444444444444"
)

// resetRig holds the wired-up handler + dependencies.
type resetRig struct {
	pool   *pgxpool.Pool
	users  *store.Store
	rstore *reset.Store
	notify *captureNotify
	h      *reset.Handler
	user   *store.User
}

type captureNotify struct {
	sent []captured
}

type captured struct {
	Email string
	Token string
}

func (c *captureNotify) SendPasswordReset(_ context.Context, email, token string, _ time.Time) error {
	c.sent = append(c.sent, captured{Email: email, Token: token})
	return nil
}

func setupResetRig(t *testing.T) *resetRig {
	t.Helper()
	pool := newResetPool(t)

	// Seed a user with a known initial password.
	users := store.NewStore(pool)
	hash, err := password.Hash("original-password-12345", password.PolicyV1)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	if err := users.Create(context.Background(), store.CreateUserParams{
		ID:                 "11111111-1111-1111-1111-111111111111",
		TenantID:           resetTestTenant,
		EmailLower:         "alice@example.com",
		EmailDisplay:       "Alice <alice@example.com>",
		PasswordHash:       hash,
		DataClassification: "cui",
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	user, err := users.GetByEmail(context.Background(), resetTestTenant, "alice@example.com")
	if err != nil {
		t.Fatalf("get seeded: %v", err)
	}

	rstore := reset.NewStore(pool, time.Now)
	notify := &captureNotify{}

	// Use a zero constant-time delay for the lifecycle tests so
	// they run quickly. The dedicated timing test (below)
	// constructs its own handler with the real delay.
	h, err := reset.NewHandler(rstore, users, notify, nil, reset.HandlerConfig{
		TenantID: resetTestTenant,
		SleepUntil: func(_ context.Context, _ time.Time) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("handler: %v", err)
	}
	return &resetRig{
		pool:   pool,
		users:  users,
		rstore: rstore,
		notify: notify,
		h:      h,
		user:   user,
	}
}

// Acceptance #1: Token is single-use + 1h TTL + hashed at rest.
func TestReset_TokenLifecycle(t *testing.T) {
	rig := setupResetRig(t)
	ctx := context.Background()

	res, err := rig.h.Request(ctx, reset.RequestInput{Email: "alice@example.com"})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if res.Outcome != reset.RequestOutcomeSent {
		t.Fatalf("outcome: %v", res.Outcome)
	}
	if len(rig.notify.sent) != 1 {
		t.Fatalf("notify count: %d", len(rig.notify.sent))
	}
	token := rig.notify.sent[0].Token
	if token == "" {
		t.Fatal("token empty")
	}

	// Hashed at rest: the row's token_hash MUST NOT equal the
	// bearer secret (we can't recompute the hash without the
	// helper, but we can at least check the column doesn't carry
	// the bearer string).
	var hash, dotID string
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			dotID = token[:i]
			break
		}
	}
	if dotID == "" {
		t.Fatal("bearer must contain '.'")
	}
	if err := rig.pool.QueryRow(ctx,
		`SELECT token_hash FROM password_resets WHERE id = $1`,
		dotID,
	).Scan(&hash); err != nil {
		t.Fatalf("query: %v", err)
	}
	if hash == "" {
		t.Error("token_hash should be populated")
	}
	if hash == token {
		t.Error("token_hash must not equal the bearer string")
	}

	// Confirm with the new password.
	confirm, err := rig.h.Confirm(ctx, reset.ConfirmInput{
		Token:       token,
		NewPassword: "fresh-password-99",
	})
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if confirm.Outcome != reset.ConfirmOutcomeOK {
		t.Fatalf("confirm outcome: %v reason=%s", confirm.Outcome, confirm.Reason)
	}

	// The user's password hash must have been replaced.
	updated, _ := rig.users.GetByEmail(ctx, resetTestTenant, "alice@example.com")
	if updated.PasswordHash == rig.user.PasswordHash {
		t.Error("password hash unchanged after confirm")
	}

	// Re-presentation of the same token must be rejected as reused.
	confirm2, _ := rig.h.Confirm(ctx, reset.ConfirmInput{
		Token:       token,
		NewPassword: "fresh-password-99",
	})
	if confirm2.Outcome != reset.ConfirmOutcomeTokenReused {
		t.Errorf("reuse outcome: %v", confirm2.Outcome)
	}
}

// Acceptance #1 follow-up: expired token returns the expired
// outcome.
func TestReset_TokenExpiry(t *testing.T) {
	rig := setupResetRig(t)
	ctx := context.Background()

	// Build a clock-injected store + handler for this test so we
	// can advance past the TTL without sleeping.
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	currentTime := now
	rstore := reset.NewStore(rig.pool, func() time.Time { return currentTime })
	h, err := reset.NewHandler(rstore, rig.users, rig.notify, nil, reset.HandlerConfig{
		TenantID: resetTestTenant,
		Now:      func() time.Time { return currentTime },
		SleepUntil: func(_ context.Context, _ time.Time) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("handler: %v", err)
	}

	rig.notify.sent = nil
	res, err := h.Request(ctx, reset.RequestInput{Email: "alice@example.com"})
	if err != nil || res.Outcome != reset.RequestOutcomeSent {
		t.Fatalf("request: %v %v", res.Outcome, err)
	}
	token := rig.notify.sent[0].Token

	// Jump 1h+1s past issuance.
	currentTime = now.Add(reset.DefaultTTL + time.Second)

	confirm, err := h.Confirm(ctx, reset.ConfirmInput{
		Token:       token,
		NewPassword: "fresh-password-99",
	})
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if confirm.Outcome != reset.ConfirmOutcomeTokenExpired {
		t.Errorf("expired outcome: %v", confirm.Outcome)
	}
}

// Acceptance #2: rate limit 3/h enforced.
func TestReset_RateLimit_3PerHour(t *testing.T) {
	rig := setupResetRig(t)
	ctx := context.Background()

	for i := 0; i < reset.DefaultRateLimit; i++ {
		res, err := rig.h.Request(ctx, reset.RequestInput{Email: "alice@example.com"})
		if err != nil {
			t.Fatalf("req %d: %v", i, err)
		}
		if res.Outcome != reset.RequestOutcomeSent {
			t.Fatalf("req %d outcome: %v", i, res.Outcome)
		}
	}
	// 4th must be rate-limited.
	res, err := rig.h.Request(ctx, reset.RequestInput{Email: "alice@example.com"})
	if err != nil {
		t.Fatalf("4th: %v", err)
	}
	if res.Outcome != reset.RequestOutcomeRateLimited {
		t.Errorf("4th outcome: %v want %v", res.Outcome, reset.RequestOutcomeRateLimited)
	}
	// Notify should have fired exactly RateLimit times.
	if got := len(rig.notify.sent); got != reset.DefaultRateLimit {
		t.Errorf("notify sent: %d want %d", got, reset.DefaultRateLimit)
	}
}

// Acceptance #3: response timing variance < 50ms between known and
// unknown emails.
//
// The handler pads every response to ConstantTimeDelay (250ms),
// so the floor is 250ms regardless of branch. We assert the median
// of N samples for each branch is within 50ms of the other.
func TestReset_TimingVariance_KnownVsUnknownEmail(t *testing.T) {
	rig := setupResetRig(t)

	// Construct a handler with the REAL constant-time delay (the
	// rig's default uses a no-op for speed).
	h, err := reset.NewHandler(rig.rstore, rig.users, rig.notify, nil, reset.HandlerConfig{
		TenantID: resetTestTenant,
		// SleepUntil left as default → realSleepUntil.
	})
	if err != nil {
		t.Fatalf("handler: %v", err)
	}

	const samples = 5
	knownTimes := make([]time.Duration, samples)
	unknownTimes := make([]time.Duration, samples)

	// Reset the password_resets table so the rate-limit doesn't
	// fire mid-loop.
	if _, err := rig.pool.Exec(context.Background(), `TRUNCATE password_resets`); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	// Interleave so background system noise affects both branches
	// roughly equally.
	for i := 0; i < samples; i++ {
		// Known branch: clear table first so the per-user rate
		// limit doesn't trip after sample 3.
		if _, err := rig.pool.Exec(context.Background(), `TRUNCATE password_resets`); err != nil {
			t.Fatalf("truncate: %v", err)
		}

		start := time.Now()
		_, _ = h.Request(context.Background(), reset.RequestInput{Email: "alice@example.com"})
		knownTimes[i] = time.Since(start)

		start = time.Now()
		_, _ = h.Request(context.Background(), reset.RequestInput{Email: "ghost@example.com"})
		unknownTimes[i] = time.Since(start)
	}

	knownMedian := median(knownTimes)
	unknownMedian := median(unknownTimes)

	delta := time.Duration(math.Abs(float64(knownMedian - unknownMedian)))
	t.Logf("known median:   %s", knownMedian)
	t.Logf("unknown median: %s", unknownMedian)
	t.Logf("delta:          %s", delta)

	if delta > 50*time.Millisecond {
		t.Errorf("timing variance %s > 50ms (known=%s unknown=%s)",
			delta, knownMedian, unknownMedian)
	}
	// Both branches MUST sit on the constant-time floor.
	if knownMedian < reset.ConstantTimeDelay-20*time.Millisecond {
		t.Errorf("known median %s < ConstantTimeDelay floor %s",
			knownMedian, reset.ConstantTimeDelay)
	}
	if unknownMedian < reset.ConstantTimeDelay-20*time.Millisecond {
		t.Errorf("unknown median %s < ConstantTimeDelay floor %s",
			unknownMedian, reset.ConstantTimeDelay)
	}
}

func median(durs []time.Duration) time.Duration {
	tmp := make([]time.Duration, len(durs))
	copy(tmp, durs)
	for i := 1; i < len(tmp); i++ {
		for j := i; j > 0 && tmp[j-1] > tmp[j]; j-- {
			tmp[j-1], tmp[j] = tmp[j], tmp[j-1]
		}
	}
	return tmp[len(tmp)/2]
}

// Confirm-with-malformed-token returns the invalid outcome rather
// than leaking a stack trace.
func TestReset_Confirm_MalformedTokenInvalid(t *testing.T) {
	rig := setupResetRig(t)
	res, err := rig.h.Confirm(context.Background(), reset.ConfirmInput{
		Token:       "no-dot-no-good",
		NewPassword: "this-is-long-enough",
	})
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if res.Outcome != reset.ConfirmOutcomeTokenInvalid {
		t.Errorf("outcome: %v", res.Outcome)
	}
}

// A successful Confirm with a SessionRevoker wired in evicts the
// user's outstanding sessions.
func TestReset_Confirm_RevokesSessions(t *testing.T) {
	rig := setupResetRig(t)
	ctx := context.Background()

	// Wire a tiny in-memory revoker so we don't need to spin up
	// the real session.Manager here (its own test file owns that
	// path).
	revoker := &countingRevoker{}
	h, err := reset.NewHandler(rig.rstore, rig.users, rig.notify, revoker, reset.HandlerConfig{
		TenantID: resetTestTenant,
		SleepUntil: func(_ context.Context, _ time.Time) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("handler: %v", err)
	}

	res, _ := h.Request(ctx, reset.RequestInput{Email: "alice@example.com"})
	if res.Outcome != reset.RequestOutcomeSent {
		t.Fatalf("request outcome: %v", res.Outcome)
	}
	token := rig.notify.sent[len(rig.notify.sent)-1].Token

	confirm, _ := h.Confirm(ctx, reset.ConfirmInput{
		Token:       token,
		NewPassword: "fresh-password-99",
	})
	if confirm.Outcome != reset.ConfirmOutcomeOK {
		t.Fatalf("confirm outcome: %v", confirm.Outcome)
	}
	if revoker.count == 0 {
		t.Error("session revoker should have been invoked")
	}
}

type countingRevoker struct {
	count int
}

func (c *countingRevoker) RevokeAllForUser(_ context.Context, _, by string) (int64, error) {
	c.count++
	if by != "password_reset" {
		return 0, errors.New("unexpected reason")
	}
	return 0, nil
}
