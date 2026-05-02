//go:build integration

// mfa_test.go — TASK-P1-IAM-003 integration tests.
//
// Exercises the MFA store layer end-to-end against a real Postgres,
// covering the three acceptance criteria in plan/todo.md:
//
//   1. Enroll → submit code completes within one HTTP round-trip
//      (here we simulate the round-trip as a SaveEnrollment +
//      MarkVerified pair gated by a Verify call).
//   2. Each backup code is single-use; reuse rejected.
//   3. Replay of the same TOTP code within the same time-step is
//      rejected via the in-memory replay cache.

package iam_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/iam/internal/mfa"
)

func newMFAPool(t *testing.T) *pgxpool.Pool {
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
			`TRUNCATE mfa_totp_secrets, mfa_backup_codes RESTART IDENTITY CASCADE`)
		pool.Close()
	})
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE mfa_totp_secrets, mfa_backup_codes RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

// Acceptance #1: enroll → submit code in one round-trip.
func TestMFA_EnrollmentLifecycle(t *testing.T) {
	pool := newMFAPool(t)
	ctx := context.Background()
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	store := mfa.NewStore(pool, func() time.Time { return now })

	const userID = "11111111-1111-1111-1111-111111111111"

	secret, err := mfa.GenerateSecret()
	if err != nil {
		t.Fatalf("secret: %v", err)
	}

	// Save the pending enrollment + return the QR URI to the user.
	if err := store.SaveEnrollment(ctx, userID, secret); err != nil {
		t.Fatalf("save enrollment: %v", err)
	}
	uri, err := mfa.EnrollmentURI("Chetana", "user@example.com", secret)
	if err != nil || uri == "" {
		t.Fatalf("uri: %v %q", err, uri)
	}

	// User submits a code; we verify, then mark the enrollment active.
	pending, err := store.LoadPending(ctx, userID)
	if err != nil || pending == nil {
		t.Fatalf("load pending: %v %v", err, pending)
	}
	code := mfa.Generate(pending.Secret, now)
	step, err := mfa.Verify(pending.Secret, code, now)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !store.ConsumeReplayWindow(userID, step, code) {
		t.Fatal("first presentation rejected by replay cache")
	}
	if err := store.MarkVerified(ctx, userID); err != nil {
		t.Fatalf("mark verified: %v", err)
	}

	// LoadActive should now return the row.
	active, err := store.LoadActive(ctx, userID)
	if err != nil || active == nil {
		t.Fatalf("load active: %v %v", err, active)
	}
	if !active.VerifiedAt.Valid {
		t.Error("verified_at not set")
	}
}

// Acceptance #2: each backup code single-use.
func TestMFA_BackupCodes_SingleUse(t *testing.T) {
	pool := newMFAPool(t)
	ctx := context.Background()
	store := mfa.NewStore(pool, time.Now)
	const userID = "11111111-1111-1111-1111-111111111111"

	codes, err := mfa.GenerateBackupCodes()
	if err != nil {
		t.Fatalf("gen: %v", err)
	}
	if err := store.SaveBackupCodes(ctx, userID, codes); err != nil {
		t.Fatalf("save: %v", err)
	}

	count, err := store.CountActiveBackupCodes(ctx, userID)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != mfa.BackupCodeCount {
		t.Fatalf("active count: got %d want %d", count, mfa.BackupCodeCount)
	}

	// First use of code 0 succeeds; replay returns ErrBackupCodeReused.
	target := codes[0].Plain
	if err := store.ConsumeBackupCode(ctx, userID, target); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if err := store.ConsumeBackupCode(ctx, userID, target); !errors.Is(err, mfa.ErrBackupCodeReused) {
		t.Errorf("reuse: got %v want ErrBackupCodeReused", err)
	}

	count, _ = store.CountActiveBackupCodes(ctx, userID)
	if count != mfa.BackupCodeCount-1 {
		t.Errorf("after consume: got %d want %d", count, mfa.BackupCodeCount-1)
	}

	// A code that was never minted is not found.
	if err := store.ConsumeBackupCode(ctx, userID, "ZZZZZZZZ"); !errors.Is(err, mfa.ErrBackupCodeNotFound) {
		t.Errorf("unknown: got %v want ErrBackupCodeNotFound", err)
	}
}

// Acceptance #2 follow-up: regenerating the book invalidates older codes.
func TestMFA_BackupCodes_RegenerationInvalidatesOldBook(t *testing.T) {
	pool := newMFAPool(t)
	ctx := context.Background()
	store := mfa.NewStore(pool, time.Now)
	const userID = "11111111-1111-1111-1111-111111111111"

	first, _ := mfa.GenerateBackupCodes()
	if err := store.SaveBackupCodes(ctx, userID, first); err != nil {
		t.Fatalf("save first: %v", err)
	}
	second, _ := mfa.GenerateBackupCodes()
	if err := store.SaveBackupCodes(ctx, userID, second); err != nil {
		t.Fatalf("save second: %v", err)
	}

	// A code from the first batch must no longer be accepted.
	if err := store.ConsumeBackupCode(ctx, userID, first[0].Plain); !errors.Is(err, mfa.ErrBackupCodeNotFound) {
		t.Errorf("old book: got %v want ErrBackupCodeNotFound", err)
	}
	// A code from the new batch works.
	if err := store.ConsumeBackupCode(ctx, userID, second[0].Plain); err != nil {
		t.Errorf("new book: %v", err)
	}
}

// Acceptance #3: TOTP replay within the same step rejected.
func TestMFA_TOTP_ReplayRejection(t *testing.T) {
	// Replay cache is process-local, no DB needed but the helper
	// uses a real pool for parity with the other tests.
	pool := newMFAPool(t)
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	store := mfa.NewStore(pool, func() time.Time { return now })

	secret, _ := mfa.GenerateSecret()
	const userID = "11111111-1111-1111-1111-111111111111"

	code := mfa.Generate(secret, now)
	step, err := mfa.Verify(secret, code, now)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !store.ConsumeReplayWindow(userID, step, code) {
		t.Fatal("first presentation must succeed")
	}
	if store.ConsumeReplayWindow(userID, step, code) {
		t.Fatal("replay must be rejected")
	}
}
