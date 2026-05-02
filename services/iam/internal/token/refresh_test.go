//go:build integration

package token

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// dsnEnv is the env var the integration suite uses to find a real Postgres.
const dsnEnv = "IAM_TEST_DATABASE_URL"

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv(dsnEnv)
	if dsn == "" {
		t.Skipf("%s not set — skipping integration test", dsnEnv)
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`TRUNCATE refresh_tokens RESTART IDENTITY CASCADE`)
		pool.Close()
	})
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE refresh_tokens RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

func TestRefreshStore_IssueAndRotate_HappyPath(t *testing.T) {
	pool := newTestPool(t)
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	store := NewRefreshStore(pool, clock)

	first, err := store.Issue(context.Background(), RefreshIssue{
		UserID:    "11111111-1111-1111-1111-111111111111",
		TenantID:  "22222222-2222-2222-2222-222222222222",
		SessionID: "sess-1",
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if first.Token == "" || first.ID == "" || first.FamilyID == "" {
		t.Fatalf("first issue empty fields: %+v", first)
	}
	if !first.ExpiresAt.Equal(now.Add(DefaultRefreshTokenTTL)) {
		t.Errorf("expires_at: %v want %v", first.ExpiresAt, now.Add(DefaultRefreshTokenTTL))
	}

	second, err := store.Rotate(context.Background(), first.Token)
	if err != nil {
		t.Fatalf("rotate: %v", err)
	}
	if second.Token == first.Token {
		t.Error("rotated token should differ from original")
	}
	if second.FamilyID != first.FamilyID {
		t.Errorf("family must persist across rotation: %q vs %q", second.FamilyID, first.FamilyID)
	}
}

func TestRefreshStore_Rotate_ReuseRevokesFamily(t *testing.T) {
	pool := newTestPool(t)
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	store := NewRefreshStore(pool, func() time.Time { return now })
	ctx := context.Background()

	first, err := store.Issue(ctx, RefreshIssue{
		UserID:    "11111111-1111-1111-1111-111111111111",
		TenantID:  "22222222-2222-2222-2222-222222222222",
		SessionID: "sess-2",
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	// Legitimate rotation: first → second.
	second, err := store.Rotate(ctx, first.Token)
	if err != nil {
		t.Fatalf("rotate 1: %v", err)
	}

	// Replay the original (already-consumed) token: must be flagged as
	// reuse and revoke the entire family.
	if _, err := store.Rotate(ctx, first.Token); !errors.Is(err, ErrReusedRefresh) {
		t.Fatalf("reuse: got %v want ErrReusedRefresh", err)
	}

	// After family revocation, even the legitimate "current" token is
	// dead — presenting `second` must return ErrInvalidRefresh because
	// the row was marked revoked by the family-wide UPDATE.
	if _, err := store.Rotate(ctx, second.Token); !errors.Is(err, ErrInvalidRefresh) {
		t.Fatalf("post-revoke rotate: got %v want ErrInvalidRefresh", err)
	}

	// Verify in DB: every row in the family has revoked = true.
	var revokedCount int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM refresh_tokens WHERE family_id=$1 AND revoked=true`,
		first.FamilyID,
	).Scan(&revokedCount); err != nil {
		t.Fatalf("count: %v", err)
	}
	if revokedCount < 2 {
		t.Errorf("expected >=2 revoked rows in family, got %d", revokedCount)
	}
}

func TestRefreshStore_Rotate_ExpiredToken(t *testing.T) {
	pool := newTestPool(t)
	now := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	clock := now
	store := NewRefreshStore(pool, func() time.Time { return clock })
	ctx := context.Background()

	first, err := store.Issue(ctx, RefreshIssue{
		UserID:    "11111111-1111-1111-1111-111111111111",
		TenantID:  "22222222-2222-2222-2222-222222222222",
		SessionID: "sess-3",
		TTL:       time.Minute,
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	// Jump past expiry.
	clock = first.ExpiresAt.Add(time.Second)
	if _, err := store.Rotate(ctx, first.Token); !errors.Is(err, ErrExpiredRefresh) {
		t.Errorf("expired: got %v want ErrExpiredRefresh", err)
	}
}

func TestRefreshStore_Rotate_MalformedBearer(t *testing.T) {
	pool := newTestPool(t)
	store := NewRefreshStore(pool, nil)

	cases := []string{
		"",
		"no-dot",
		".only-secret",
		"only-id.",
		"id.not!base64!",
	}
	for _, tc := range cases {
		t.Run(tc, func(t *testing.T) {
			if _, err := store.Rotate(context.Background(), tc); !errors.Is(err, ErrInvalidRefresh) {
				t.Errorf("got %v want ErrInvalidRefresh", err)
			}
		})
	}
}

func TestRefreshStore_RevokeFamily(t *testing.T) {
	pool := newTestPool(t)
	store := NewRefreshStore(pool, nil)
	ctx := context.Background()

	first, err := store.Issue(ctx, RefreshIssue{
		UserID:    "11111111-1111-1111-1111-111111111111",
		TenantID:  "22222222-2222-2222-2222-222222222222",
		SessionID: "sess-4",
	})
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if err := store.RevokeFamily(ctx, first.FamilyID); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if _, err := store.Rotate(ctx, first.Token); !errors.Is(err, ErrInvalidRefresh) {
		t.Errorf("post-revoke: got %v want ErrInvalidRefresh", err)
	}
}

func TestRefreshStore_Issue_Validation(t *testing.T) {
	pool := newTestPool(t)
	store := NewRefreshStore(pool, nil)
	ctx := context.Background()

	cases := []RefreshIssue{
		{TenantID: "t", SessionID: "s"},
		{UserID: "u", SessionID: "s"},
		{UserID: "u", TenantID: "t"},
	}
	for _, c := range cases {
		if _, err := store.Issue(ctx, c); err == nil {
			t.Errorf("want error for %+v", c)
		}
	}
}

func TestRefreshRecord_IsActiveAt(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		rec  RefreshRecord
		want bool
	}{
		{"fresh", RefreshRecord{ExpiresAt: now.Add(time.Hour)}, true},
		{"revoked", RefreshRecord{Revoked: true, ExpiresAt: now.Add(time.Hour)}, false},
		{"expired", RefreshRecord{ExpiresAt: now.Add(-time.Hour)}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rec.IsActiveAt(now); got != tc.want {
				t.Errorf("got %v want %v", got, tc.want)
			}
		})
	}
}

func TestEncodeDecodeRefreshBearer_Roundtrip(t *testing.T) {
	secret := []byte("0123456789abcdef0123456789abcdef")
	bearer := encodeRefreshBearer("rowid", secret)
	id, got, err := decodeRefreshBearer(bearer)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if id != "rowid" {
		t.Errorf("id: %q", id)
	}
	if string(got) != string(secret) {
		t.Errorf("secret roundtrip mismatch")
	}
}
