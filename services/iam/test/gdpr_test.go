//go:build integration

// gdpr_test.go — TASK-P1-IAM-009 integration tests covering
// Article 15 (SAR), 16 (rectify), 17 (erasure), and 20
// (portability) end-to-end against real Postgres.

package iam_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/iam/internal/gdpr"
	"github.com/ppusapati/space/services/iam/internal/password"
	"github.com/ppusapati/space/services/iam/internal/store"
)

const gdprTestTenant = "55555555-5555-5555-5555-555555555555"

func newGDPRPool(t *testing.T) *pgxpool.Pool {
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
			`TRUNCATE password_resets, sessions, refresh_tokens, oauth2_auth_codes,
			            mfa_backup_codes, mfa_totp_secrets, webauthn_credentials,
			            users RESTART IDENTITY CASCADE`)
		pool.Close()
	})
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE password_resets, sessions, refresh_tokens, oauth2_auth_codes,
		            mfa_backup_codes, mfa_totp_secrets, webauthn_credentials,
		            users RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

func seedGDPRUser(t *testing.T, users *store.Store) *store.User {
	t.Helper()
	hash, err := password.Hash("orig-password-12345", password.PolicyV1)
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	const userID = "11111111-1111-1111-1111-111111111111"
	if err := users.Create(context.Background(), store.CreateUserParams{
		ID:                 userID,
		TenantID:           gdprTestTenant,
		EmailLower:         "subject@example.com",
		EmailDisplay:       "Subject <subject@example.com>",
		PasswordHash:       hash,
		DataClassification: "cui",
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	u, err := users.GetByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("get seeded: %v", err)
	}
	return u
}

// captureExporter records the EnqueueSAR call so the test can
// inspect the snapshot the SAR service would have shipped to the
// real export service.
type captureExporter struct {
	calls []gdpr.EnqueueSARInput
}

func (c *captureExporter) EnqueueSAR(_ context.Context, in gdpr.EnqueueSARInput) (gdpr.JobID, error) {
	c.calls = append(c.calls, in)
	return gdpr.JobID("job-" + in.UserID), nil
}

// Acceptance #1: SAR completes within minutes — synchronously
// returns a job id + the IAM-side snapshot.
func TestGDPR_SAR_RoundTrip(t *testing.T) {
	pool := newGDPRPool(t)
	users := store.NewStore(pool)
	u := seedGDPRUser(t, users)

	exporter := &captureExporter{}
	svc, err := gdpr.NewSARService(pool, exporter, nil, time.Now)
	if err != nil {
		t.Fatalf("sar service: %v", err)
	}

	res, err := svc.Request(context.Background(), gdpr.SARRequest{
		UserID: u.ID,
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if res.JobID == "" {
		t.Error("missing job id")
	}
	if res.Snapshot == nil || res.Snapshot.User.ID != u.ID {
		t.Fatalf("snapshot: %+v", res.Snapshot)
	}
	if res.Snapshot.User.EmailLower != "subject@example.com" {
		t.Errorf("snapshot email: %q", res.Snapshot.User.EmailLower)
	}
	if len(exporter.calls) != 1 {
		t.Fatalf("exporter calls: %d", len(exporter.calls))
	}
	// The captured snapshot must match what was returned to the caller.
	if exporter.calls[0].Snapshot != res.Snapshot {
		t.Error("captured snapshot pointer differs from returned snapshot")
	}
}

func TestGDPR_SAR_UnknownUser(t *testing.T) {
	pool := newGDPRPool(t)
	svc, _ := gdpr.NewSARService(pool, gdpr.NopExporter{}, nil, time.Now)
	if _, err := svc.Request(context.Background(), gdpr.SARRequest{
		UserID: "00000000-0000-0000-0000-000000000000",
	}); !errors.Is(err, gdpr.ErrUserNotFound) {
		t.Errorf("got %v want ErrUserNotFound", err)
	}
}

// Acceptance #2: erasure anonymises email_lower deterministically
// and preserves the audit chain (the user_id stays referencable).
func TestGDPR_Erase_AnonymisesAndPurgesOperationalState(t *testing.T) {
	pool := newGDPRPool(t)
	users := store.NewStore(pool)
	u := seedGDPRUser(t, users)
	ctx := context.Background()

	// Plant some operational state so the erasure has rows to
	// hard-delete.
	if _, err := pool.Exec(ctx, `
INSERT INTO sessions
  (id, user_id, tenant_id, absolute_expires_at, idle_expires_at)
VALUES ('sess-1', $1, $2, now() + INTERVAL '24 hours', now() + INTERVAL '1 hour')
`, u.ID, gdprTestTenant); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	svc, err := gdpr.NewEraseService(pool, time.Now)
	if err != nil {
		t.Fatalf("erase service: %v", err)
	}
	res, err := svc.Erase(ctx, gdpr.ErasureRequest{
		UserID: u.ID,
		Reason: "user_request",
	})
	if err != nil {
		t.Fatalf("erase: %v", err)
	}

	// 1. The anonymisation hash matches the deterministic helper.
	wantHash := gdpr.AnonymisedEmailFor(u.ID, gdprTestTenant)
	if res.AnonHashPrefix != wantHash {
		t.Errorf("anon hash: got %q want %q", res.AnonHashPrefix, wantHash)
	}

	// 2. Operational state hard-deleted.
	if res.HardDeleted.Sessions != 1 {
		t.Errorf("sessions hard-deleted: %d want 1", res.HardDeleted.Sessions)
	}

	// 3. The users row is anonymised in place — NOT deleted —
	// because the audit chain still references the user_id.
	var (
		emailLower, emailDisplay, status string
		gdprAt                           *time.Time
	)
	if err := pool.QueryRow(ctx,
		`SELECT email_lower, email_display, status, gdpr_anonymized_at FROM users WHERE id = $1`,
		u.ID,
	).Scan(&emailLower, &emailDisplay, &status, &gdprAt); err != nil {
		t.Fatalf("post-erasure user: %v", err)
	}
	if emailLower != wantHash {
		t.Errorf("email_lower: %q want %q", emailLower, wantHash)
	}
	if emailDisplay != "(erased)" {
		t.Errorf("email_display: %q", emailDisplay)
	}
	if status != "deleted" {
		t.Errorf("status: %q", status)
	}
	if gdprAt == nil {
		t.Error("gdpr_anonymized_at should be set")
	}

	// 4. Idempotency: re-erasing is a no-op (anonymisation
	// timestamp does NOT advance).
	res2, err := svc.Erase(ctx, gdpr.ErasureRequest{UserID: u.ID})
	if err != nil {
		t.Fatalf("second erase: %v", err)
	}
	if !res2.AnonymizedAt.Equal(res.AnonymizedAt) {
		t.Errorf("anonymized_at moved on re-erase: %v vs %v", res2.AnonymizedAt, res.AnonymizedAt)
	}
}

func TestGDPR_Erase_UnknownUser(t *testing.T) {
	pool := newGDPRPool(t)
	svc, _ := gdpr.NewEraseService(pool, time.Now)
	if _, err := svc.Erase(context.Background(), gdpr.ErasureRequest{
		UserID: "00000000-0000-0000-0000-000000000000",
	}); !errors.Is(err, gdpr.ErrUserNotFound) {
		t.Errorf("got %v want ErrUserNotFound", err)
	}
}

// Article 16: rectify happy path + error matrix.
func TestGDPR_RectifyEmail_HappyPath(t *testing.T) {
	pool := newGDPRPool(t)
	users := store.NewStore(pool)
	u := seedGDPRUser(t, users)

	svc, err := gdpr.NewRectifyService(pool, time.Now)
	if err != nil {
		t.Fatalf("rectify service: %v", err)
	}
	res, err := svc.RectifyEmail(context.Background(), gdpr.RectifyEmailRequest{
		UserID:   u.ID,
		NewEmail: "subject.fixed@example.com",
	})
	if err != nil {
		t.Fatalf("rectify: %v", err)
	}
	if res.NewEmail != "subject.fixed@example.com" {
		t.Errorf("new email: %q", res.NewEmail)
	}

	// Verify in the DB.
	got, _ := users.GetByEmail(context.Background(), gdprTestTenant, "subject.fixed@example.com")
	if got == nil || got.ID != u.ID {
		t.Errorf("post-rectify lookup: %+v", got)
	}
}

func TestGDPR_RectifyEmail_DuplicateRejected(t *testing.T) {
	pool := newGDPRPool(t)
	users := store.NewStore(pool)
	u := seedGDPRUser(t, users)

	// Plant a second user whose email we'll try to steal.
	otherHash, _ := password.Hash("other-password-12345", password.PolicyV1)
	if err := users.Create(context.Background(), store.CreateUserParams{
		ID:                 "22222222-2222-2222-2222-222222222222",
		TenantID:           gdprTestTenant,
		EmailLower:         "other@example.com",
		EmailDisplay:       "Other <other@example.com>",
		PasswordHash:       otherHash,
		DataClassification: "cui",
	}); err != nil {
		t.Fatalf("seed second: %v", err)
	}

	svc, _ := gdpr.NewRectifyService(pool, time.Now)
	_, err := svc.RectifyEmail(context.Background(), gdpr.RectifyEmailRequest{
		UserID:   u.ID,
		NewEmail: "other@example.com",
	})
	if !errors.Is(err, gdpr.ErrEmailInUse) {
		t.Errorf("got %v want ErrEmailInUse", err)
	}
}

func TestGDPR_RectifyEmail_AfterErasureRejected(t *testing.T) {
	pool := newGDPRPool(t)
	users := store.NewStore(pool)
	u := seedGDPRUser(t, users)
	ctx := context.Background()

	erase, _ := gdpr.NewEraseService(pool, time.Now)
	if _, err := erase.Erase(ctx, gdpr.ErasureRequest{UserID: u.ID}); err != nil {
		t.Fatalf("erase: %v", err)
	}

	rect, _ := gdpr.NewRectifyService(pool, time.Now)
	_, err := rect.RectifyEmail(ctx, gdpr.RectifyEmailRequest{
		UserID:   u.ID,
		NewEmail: "restored@example.com",
	})
	if !errors.Is(err, gdpr.ErrAlreadyErased) {
		t.Errorf("got %v want ErrAlreadyErased", err)
	}
}

func TestGDPR_RectifyEmail_InvalidShape(t *testing.T) {
	pool := newGDPRPool(t)
	users := store.NewStore(pool)
	u := seedGDPRUser(t, users)

	svc, _ := gdpr.NewRectifyService(pool, time.Now)
	cases := []string{"", "no-at-sign", "@example.com", "u@nodomain"}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			_, err := svc.RectifyEmail(context.Background(), gdpr.RectifyEmailRequest{
				UserID:   u.ID,
				NewEmail: c,
			})
			if !errors.Is(err, gdpr.ErrInvalidEmail) {
				t.Errorf("got %v want ErrInvalidEmail", err)
			}
		})
	}
}

// Article 20: portability snapshot serialises every IAM-owned row
// referencing the subject.
func TestGDPR_PortabilitySnapshot(t *testing.T) {
	pool := newGDPRPool(t)
	users := store.NewStore(pool)
	u := seedGDPRUser(t, users)
	ctx := context.Background()

	// Plant one session row so the snapshot has something to
	// serialise besides the user.
	if _, err := pool.Exec(ctx, `
INSERT INTO sessions
  (id, user_id, tenant_id, absolute_expires_at, idle_expires_at)
VALUES ('snap-sess', $1, $2, now() + INTERVAL '24 hours', now() + INTERVAL '1 hour')
`, u.ID, gdprTestTenant); err != nil {
		t.Fatalf("seed session: %v", err)
	}

	builder := gdpr.NewSnapshotBuilder(pool, time.Now)
	snap, err := builder.Build(ctx, u.ID)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if snap.User.ID != u.ID {
		t.Errorf("user id: %q", snap.User.ID)
	}
	if len(snap.Sessions) != 1 || snap.Sessions[0].SessionID != "snap-sess" {
		t.Errorf("sessions: %+v", snap.Sessions)
	}
}

// Re-erasure is idempotent and does not double-purge counts.
func TestGDPR_Erase_Idempotent_NoDoublePurge(t *testing.T) {
	pool := newGDPRPool(t)
	users := store.NewStore(pool)
	u := seedGDPRUser(t, users)
	ctx := context.Background()

	svc, _ := gdpr.NewEraseService(pool, time.Now)
	if _, err := svc.Erase(ctx, gdpr.ErasureRequest{UserID: u.ID}); err != nil {
		t.Fatalf("first erase: %v", err)
	}
	res2, err := svc.Erase(ctx, gdpr.ErasureRequest{UserID: u.ID})
	if err != nil {
		t.Fatalf("second erase: %v", err)
	}
	// Operational tables are already empty so subsequent counts
	// are all zero.
	if res2.HardDeleted.Sessions != 0 {
		t.Errorf("second erase deleted sessions again: %d", res2.HardDeleted.Sessions)
	}
}
