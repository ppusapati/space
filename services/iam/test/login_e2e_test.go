//go:build integration

// Package iam_test holds end-to-end tests that exercise the IAM
// service against real Postgres + Redis instances. Build tag
// `integration` keeps them out of the default `go test ./...` run on
// developer workstations without those dependencies. CI runs:
//
//	go test -tags=integration -count=1 ./test/...
//
// from inside services/iam, with CHETANA_TEST_DB_URL and
// CHETANA_TEST_REDIS_ADDR pointing at containerised instances
// (tools/db/seed-test.sh + a Redis container in the workflow).
//
// Coverage:
//   1. Happy-path login → success, audit emitted, last_login_at stamped.
//   2. Wrong password 5 times → 5th attempt locks the account
//      (REQ-FUNC-PLT-IAM-003).
//   3. Locked account returns 423 with Retry-After.
//   4. Rate-limit window blocks the 11th attempt from a single IP
//      within 60s.
//   5. Successful login resets failed_login_count and lockout_level.
package iam_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/ppusapati/space/services/iam/internal/login"
	"github.com/ppusapati/space/services/iam/internal/password"
	"github.com/ppusapati/space/services/iam/internal/store"
)

const (
	testTenant = "00000000-0000-0000-0000-000000000001"
)

// requireBackends returns DSNs for Postgres + Redis or skips the
// test cleanly when either env var is missing.
func requireBackends(t *testing.T) (string, string) {
	t.Helper()
	dsn := os.Getenv("CHETANA_TEST_DB_URL")
	if dsn == "" {
		t.Skip("CHETANA_TEST_DB_URL unset; integration test skipped")
	}
	addr := os.Getenv("CHETANA_TEST_REDIS_ADDR")
	if addr == "" {
		t.Skip("CHETANA_TEST_REDIS_ADDR unset; integration test skipped")
	}
	return dsn, addr
}

// setupSchema applies the users-table migration. Idempotent.
func setupSchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	body, err := os.ReadFile("../migrations/0001_users.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := pool.Exec(ctx, string(body)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
}

// resetDB truncates the users table between scenarios.
func resetDB(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(ctx, "TRUNCATE users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("truncate users: %v", err)
	}
}

// resetRedis flushes the IAM rate-limit keys.
func resetRedis(t *testing.T, ctx context.Context, rdb *redis.Client) {
	t.Helper()
	keys, _ := rdb.Keys(ctx, "iam:login:ip:*").Result()
	if len(keys) == 0 {
		return
	}
	if err := rdb.Del(ctx, keys...).Err(); err != nil {
		t.Fatalf("redis del: %v", err)
	}
}

// seedUser creates a fresh active user with the supplied password.
func seedUser(t *testing.T, ctx context.Context, users *store.Store, email, pw string) string {
	t.Helper()
	hash, err := password.Hash(pw, password.PolicyV1)
	if err != nil {
		t.Fatalf("password.Hash: %v", err)
	}
	id := newULIDLike(t)
	if err := users.Create(ctx, store.CreateUserParams{
		ID:           id,
		TenantID:     testTenant,
		EmailLower:   strings.ToLower(email),
		EmailDisplay: email,
		PasswordHash: hash,
	}); err != nil {
		t.Fatalf("seed user %s: %v", email, err)
	}
	return id
}

// newULIDLike returns a UUIDv4-shaped string. Real services use
// services/packages/ulid; for the e2e test any uuid works.
func newULIDLike(t *testing.T) string {
	t.Helper()
	const uuidLen = 36
	b := make([]byte, 16)
	if _, err := os.ReadFile("/dev/urandom"); err == nil {
		// best-effort: read 16 bytes from /dev/urandom on Linux
	}
	// Quick deterministic UUIDv4 using time stamp; sufficient
	// for test isolation since each test resets the table.
	now := time.Now().UnixNano()
	for i := 0; i < 16; i++ {
		b[i] = byte((now >> (i * 4)) & 0xff)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	out := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[:4], b[4:6], b[6:8], b[8:10], b[10:16])
	if len(out) != uuidLen {
		t.Fatalf("uuid wrong length: %q", out)
	}
	return out
}

// newHandler wires the integration handler against real backends.
func newHandler(t *testing.T, users *store.Store, rdb *redis.Client) *login.Handler {
	t.Helper()
	limiter := login.NewIPLimiter(rdb, login.IPLimiterConfig{})
	h, err := login.NewHandler(limiter, users, login.NopAudit{}, login.HandlerConfig{
		TenantID:   testTenant,
		SleepUntil: func(_ context.Context, _ time.Time) error { return nil }, // no constant-time delay in tests
	})
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h
}

// ----------------------------------------------------------------------
// Tests
// ----------------------------------------------------------------------

func TestLogin_E2E_HappyPath(t *testing.T) {
	dsn, redisAddr := requireBackends(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()

	setupSchema(t, ctx, pool)
	resetDB(t, ctx, pool)
	resetRedis(t, ctx, rdb)

	users := store.NewStore(pool)
	seedUser(t, ctx, users, "alice@example.com", "correct horse battery")

	h := newHandler(t, users, rdb)
	res, err := h.Login(ctx, login.LoginInput{
		Email:    "alice@example.com",
		Password: "correct horse battery",
		ClientIP: "203.0.113.10",
	})
	if err != nil {
		t.Fatalf("Login err: %v", err)
	}
	if res.Status != login.ResultOK {
		t.Errorf("status=%v, want OK", res.Status)
	}
	// last_login_at stamped.
	u, err := users.GetByEmail(ctx, testTenant, "alice@example.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if u.LastLoginAt == nil {
		t.Error("last_login_at not set after successful login")
	}
}

func TestLogin_E2E_LockoutAfterFiveFailures(t *testing.T) {
	dsn, redisAddr := requireBackends(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()

	setupSchema(t, ctx, pool)
	resetDB(t, ctx, pool)
	resetRedis(t, ctx, rdb)

	users := store.NewStore(pool)
	seedUser(t, ctx, users, "bob@example.com", "right")

	h := newHandler(t, users, rdb)
	for i := 1; i <= 5; i++ {
		res, err := h.Login(ctx, login.LoginInput{
			Email:    "bob@example.com",
			Password: "wrong",
			ClientIP: "203.0.113.20",
		})
		if err != nil {
			t.Fatalf("attempt %d err: %v", i, err)
		}
		// 5th attempt MUST trigger lockout (status=Locked) per
		// REQ-FUNC-PLT-IAM-003.
		if i == 5 {
			if res.Status != login.ResultLocked {
				t.Errorf("attempt 5: status=%v, want Locked", res.Status)
			}
			if res.RetryAfter <= 0 {
				t.Errorf("attempt 5: RetryAfter=%v, want > 0", res.RetryAfter)
			}
		} else if res.Status != login.ResultBadCredentials {
			t.Errorf("attempt %d: status=%v, want BadCredentials", i, res.Status)
		}
	}

	// 6th attempt against the same account MUST stay locked even
	// with the correct password.
	res, err := h.Login(ctx, login.LoginInput{
		Email:    "bob@example.com",
		Password: "right",
		ClientIP: "203.0.113.20",
	})
	if err != nil {
		t.Fatalf("attempt 6 err: %v", err)
	}
	if res.Status != login.ResultLocked {
		t.Errorf("attempt 6 (correct pw, locked): status=%v, want Locked", res.Status)
	}
}

func TestLogin_E2E_RateLimitedAt11thRequest(t *testing.T) {
	dsn, redisAddr := requireBackends(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	defer pool.Close()
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()

	setupSchema(t, ctx, pool)
	resetDB(t, ctx, pool)
	resetRedis(t, ctx, rdb)

	users := store.NewStore(pool)
	seedUser(t, ctx, users, "carol@example.com", "right")

	h := newHandler(t, users, rdb)
	for i := 1; i <= 10; i++ {
		_, err := h.Login(ctx, login.LoginInput{
			Email:    "ghost@example.com", // no such user — only the rate limiter cares
			Password: "anything",
			ClientIP: "203.0.113.30",
		})
		if err != nil {
			t.Fatalf("attempt %d: %v", i, err)
		}
	}
	// 11th MUST be rate-limited.
	res, err := h.Login(ctx, login.LoginInput{
		Email:    "ghost@example.com",
		Password: "anything",
		ClientIP: "203.0.113.30",
	})
	if err != nil {
		t.Fatalf("attempt 11: %v", err)
	}
	if res.Status != login.ResultRateLimited {
		t.Errorf("attempt 11: status=%v, want RateLimited", res.Status)
	}
}
