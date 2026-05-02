//go:build integration

// session_test.go — TASK-P1-IAM-007 integration tests.
//
// Exercises the session.Manager end-to-end against real Postgres,
// covering the three acceptance criteria:
//
//   1. 6th concurrent session evicts the oldest (concurrency cap).
//   2. Idle > 1h → Touch returns ErrSessionIdleTimeout.
//   3. Revoke immediately invalidates the next Touch.

package iam_test

import (
	"context"
	"errors"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/iam/internal/session"
)

func newSessionPool(t *testing.T) *pgxpool.Pool {
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
			`TRUNCATE sessions RESTART IDENTITY CASCADE`)
		pool.Close()
	})
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE sessions RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

const (
	sessionTestUserA   = "11111111-1111-1111-1111-111111111111"
	sessionTestUserB   = "22222222-2222-2222-2222-222222222222"
	sessionTestTenant  = "33333333-3333-3333-3333-333333333333"
)

// Acceptance #1: 6th concurrent session evicts the oldest.
func TestSession_ConcurrencyCap_EvictsOldest(t *testing.T) {
	pool := newSessionPool(t)
	mgr, err := session.NewManager(pool, session.Config{})
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	ctx := context.Background()

	// Open the first 5 sessions back-to-back. None should evict.
	var sessionIDs []string
	for i := 0; i < session.DefaultMaxConcurrent; i++ {
		out, err := mgr.Create(ctx, session.CreateInput{
			UserID:   sessionTestUserA,
			TenantID: sessionTestTenant,
			AMR:      []string{"pwd"},
		})
		if err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
		if len(out.EvictedSessionIDs) != 0 {
			t.Errorf("session %d should not evict, got %v", i, out.EvictedSessionIDs)
		}
		sessionIDs = append(sessionIDs, out.SessionID)
		// Tiny sleep so issued_at orders distinctly.
		time.Sleep(5 * time.Millisecond)
	}

	// Verify the cap is at exactly the limit.
	count, err := mgr.CountActiveForUser(ctx, sessionTestUserA)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != session.DefaultMaxConcurrent {
		t.Errorf("active count: got %d want %d", count, session.DefaultMaxConcurrent)
	}

	// 6th login: must evict the 1st (oldest by issued_at).
	out6, err := mgr.Create(ctx, session.CreateInput{
		UserID:   sessionTestUserA,
		TenantID: sessionTestTenant,
		AMR:      []string{"pwd"},
	})
	if err != nil {
		t.Fatalf("create 6: %v", err)
	}
	if len(out6.EvictedSessionIDs) != 1 || out6.EvictedSessionIDs[0] != sessionIDs[0] {
		t.Errorf("evicted: got %v want [%s]", out6.EvictedSessionIDs, sessionIDs[0])
	}

	// Active count remains at the cap.
	count, _ = mgr.CountActiveForUser(ctx, sessionTestUserA)
	if count != session.DefaultMaxConcurrent {
		t.Errorf("active count after eviction: got %d want %d", count, session.DefaultMaxConcurrent)
	}

	// The evicted session must now refuse Touch with the right
	// reason — revoked, not idle/absolute.
	_, err = mgr.Touch(ctx, sessionIDs[0])
	if !errors.Is(err, session.ErrSessionRevoked) {
		t.Errorf("evicted touch: got %v want ErrSessionRevoked", err)
	}
}

// Acceptance #1 follow-up: a different user's sessions don't
// affect the cap accounting.
func TestSession_ConcurrencyCap_PerUser(t *testing.T) {
	pool := newSessionPool(t)
	mgr, _ := session.NewManager(pool, session.Config{})
	ctx := context.Background()

	for i := 0; i < session.DefaultMaxConcurrent; i++ {
		if _, err := mgr.Create(ctx, session.CreateInput{UserID: sessionTestUserA, TenantID: sessionTestTenant}); err != nil {
			t.Fatalf("user A %d: %v", i, err)
		}
	}
	// User B's first session must NOT trip user A's cap.
	out, err := mgr.Create(ctx, session.CreateInput{UserID: sessionTestUserB, TenantID: sessionTestTenant})
	if err != nil {
		t.Fatalf("user B create: %v", err)
	}
	if len(out.EvictedSessionIDs) != 0 {
		t.Errorf("user B eviction leaked across users: %v", out.EvictedSessionIDs)
	}

	// User A still has the full cap.
	if n, _ := mgr.CountActiveForUser(ctx, sessionTestUserA); n != session.DefaultMaxConcurrent {
		t.Errorf("user A count: %d", n)
	}
	if n, _ := mgr.CountActiveForUser(ctx, sessionTestUserB); n != 1 {
		t.Errorf("user B count: %d", n)
	}
}

// Acceptance #2: idle > 1h → Touch returns ErrSessionIdleTimeout.
func TestSession_IdleTimeout(t *testing.T) {
	pool := newSessionPool(t)

	// Drive the manager with a controllable clock so the test
	// runs in milliseconds rather than waiting an hour.
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	currentTime := now
	mgr, err := session.NewManager(pool, session.Config{
		IdleTimeout:      time.Hour,
		AbsoluteLifetime: 24 * time.Hour,
		Now:              func() time.Time { return currentTime },
	})
	if err != nil {
		t.Fatalf("manager: %v", err)
	}
	ctx := context.Background()

	out, err := mgr.Create(ctx, session.CreateInput{UserID: sessionTestUserA, TenantID: sessionTestTenant})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Within idle window — touches succeed and bump last_seen_at.
	currentTime = now.Add(30 * time.Minute)
	if _, err := mgr.Touch(ctx, out.SessionID); err != nil {
		t.Fatalf("touch within idle: %v", err)
	}

	// Half an hour past the previous touch — STILL within idle
	// (the previous Touch reset the horizon to t+1h).
	currentTime = currentTime.Add(50 * time.Minute)
	if _, err := mgr.Touch(ctx, out.SessionID); err != nil {
		t.Fatalf("touch within rolling idle: %v", err)
	}

	// Sit idle for an hour and a second — must trip idle timeout.
	currentTime = currentTime.Add(time.Hour + time.Second)
	_, err = mgr.Touch(ctx, out.SessionID)
	if !errors.Is(err, session.ErrSessionIdleTimeout) {
		t.Errorf("idle: got %v want ErrSessionIdleTimeout", err)
	}
}

// Acceptance #2 follow-up: absolute lifetime caps a continuously-
// active session at AbsoluteLifetime, regardless of touches.
func TestSession_AbsoluteLifetime(t *testing.T) {
	pool := newSessionPool(t)
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	currentTime := now
	mgr, _ := session.NewManager(pool, session.Config{
		IdleTimeout:      time.Hour,
		AbsoluteLifetime: 24 * time.Hour,
		Now:              func() time.Time { return currentTime },
	})
	ctx := context.Background()

	out, err := mgr.Create(ctx, session.CreateInput{UserID: sessionTestUserA, TenantID: sessionTestTenant})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Tick forward with continuous touches every 30 minutes — idle
	// horizon is constantly pushed forward, BUT once we cross the
	// 24h absolute mark the next Touch must reject.
	for i := 0; i < 47; i++ {
		currentTime = currentTime.Add(30 * time.Minute)
		if _, err := mgr.Touch(ctx, out.SessionID); err != nil {
			t.Fatalf("touch at +%dh: %v", i*30/60, err)
		}
	}
	// We're now at +23.5h. Cross the line:
	currentTime = currentTime.Add(31 * time.Minute) // → +24h01m
	_, err = mgr.Touch(ctx, out.SessionID)
	if !errors.Is(err, session.ErrSessionAbsoluteExpired) {
		t.Errorf("absolute: got %v want ErrSessionAbsoluteExpired", err)
	}
}

// Acceptance #3: Revoke immediately invalidates the next Touch.
func TestSession_Revoke_ImmediatelyInvalidates(t *testing.T) {
	pool := newSessionPool(t)
	mgr, _ := session.NewManager(pool, session.Config{})
	ctx := context.Background()

	out, err := mgr.Create(ctx, session.CreateInput{UserID: sessionTestUserA, TenantID: sessionTestTenant})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Touch works pre-revoke.
	if _, err := mgr.Touch(ctx, out.SessionID); err != nil {
		t.Fatalf("pre-revoke touch: %v", err)
	}

	// Revoke.
	if err := mgr.Revoke(ctx, out.SessionID, "user_logout"); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	// Touch immediately fails with the right typed error +
	// canonical reason.
	_, err = mgr.Touch(ctx, out.SessionID)
	if !errors.Is(err, session.ErrSessionRevoked) {
		t.Fatalf("touch post-revoke: got %v want ErrSessionRevoked", err)
	}
	if got := session.Reason(err); got != "session_revoked" {
		t.Errorf("Reason: %q", got)
	}

	// Idempotent: re-revoking is a no-op.
	if err := mgr.Revoke(ctx, out.SessionID, "second_call"); err != nil {
		t.Errorf("second revoke: %v", err)
	}
}

// Acceptance #3 follow-up: RevokeAllForUser kills every active
// session for the user but leaves other users untouched.
func TestSession_RevokeAllForUser(t *testing.T) {
	pool := newSessionPool(t)
	mgr, _ := session.NewManager(pool, session.Config{})
	ctx := context.Background()

	// 3 sessions for A, 2 for B.
	var aIDs, bIDs []string
	for i := 0; i < 3; i++ {
		out, err := mgr.Create(ctx, session.CreateInput{UserID: sessionTestUserA, TenantID: sessionTestTenant})
		if err != nil {
			t.Fatalf("create A: %v", err)
		}
		aIDs = append(aIDs, out.SessionID)
	}
	for i := 0; i < 2; i++ {
		out, err := mgr.Create(ctx, session.CreateInput{UserID: sessionTestUserB, TenantID: sessionTestTenant})
		if err != nil {
			t.Fatalf("create B: %v", err)
		}
		bIDs = append(bIDs, out.SessionID)
	}

	n, err := mgr.RevokeAllForUser(ctx, sessionTestUserA, "admin_revoke")
	if err != nil {
		t.Fatalf("revoke all: %v", err)
	}
	if n != 3 {
		t.Errorf("revoked count: got %d want 3", n)
	}

	// Every A session is now revoked.
	sort.Strings(aIDs)
	for _, sid := range aIDs {
		if _, err := mgr.Touch(ctx, sid); !errors.Is(err, session.ErrSessionRevoked) {
			t.Errorf("A session %s touch: got %v want ErrSessionRevoked", sid, err)
		}
	}
	// Every B session is still alive.
	for _, sid := range bIDs {
		if _, err := mgr.Touch(ctx, sid); err != nil {
			t.Errorf("B session %s touch should still work: %v", sid, err)
		}
	}

	// Counts.
	if c, _ := mgr.CountActiveForUser(ctx, sessionTestUserA); c != 0 {
		t.Errorf("A count: %d want 0", c)
	}
	if c, _ := mgr.CountActiveForUser(ctx, sessionTestUserB); c != 2 {
		t.Errorf("B count: %d want 2", c)
	}
}

// Touch on a missing session returns ErrSessionNotFound.
func TestSession_Touch_NotFound(t *testing.T) {
	pool := newSessionPool(t)
	mgr, _ := session.NewManager(pool, session.Config{})
	if _, err := mgr.Touch(context.Background(), "00000000000000000000000000000000"); !errors.Is(err, session.ErrSessionNotFound) {
		t.Errorf("got %v want ErrSessionNotFound", err)
	}
}
