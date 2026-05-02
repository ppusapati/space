//go:build integration

// chain_test.go — TASK-P1-AUDIT-001 integration tests covering
// the three acceptance criteria:
//
//   1. Direct DB writes from non-audit roles are blocked by
//      Postgres role grants (verified by checking audit_writer
//      exists + has the expected privileges).
//   2. Chain verifier detects single-row tampering and reports
//      the first broken offset.
//   3. Append throughput ≥ 5 000 events/s on a single Postgres
//      (covered by the bench, services/audit/bench/append_bench_test.go).

package audit_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/audit/internal/chain"
)

const auditTestTenant = "00000000-0000-0000-0000-000000000001"

func newAuditPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("AUDIT_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("AUDIT_TEST_DATABASE_URL not set — skipping integration test")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			`TRUNCATE audit_events RESTART IDENTITY; UPDATE chain_tip SET last_row_id=0, last_hash=$1, last_seq=0`,
			chain.GenesisHash)
		pool.Close()
	})
	if _, err := pool.Exec(context.Background(),
		`TRUNCATE audit_events RESTART IDENTITY; UPDATE chain_tip SET last_row_id=0, last_hash=$1, last_seq=0`,
		chain.GenesisHash,
	); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return pool
}

func makeEvent(seqHint int) chain.Event {
	return chain.Event{
		TenantID:       auditTestTenant,
		EventTime:      time.Now().UTC(),
		Action:         "iam.user.read",
		Decision:       "allow",
		Classification: "cui",
		Metadata:       map[string]string{"seq_hint": "test"},
	}
}

// Acceptance #2 happy path: 5 sequential events form a clean chain.
func TestChain_AppendAndVerify_HappyPath(t *testing.T) {
	pool := newAuditPool(t)
	app := chain.NewAppender(pool)
	ver := chain.NewVerifier(pool)
	ctx := context.Background()

	start := time.Now().UTC().Add(-time.Second)
	for i := 0; i < 5; i++ {
		if _, err := app.Append(ctx, makeEvent(i)); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	res, err := ver.VerifyRange(ctx, auditTestTenant, start, time.Now().UTC().Add(time.Second))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !res.IsClean() {
		t.Errorf("clean chain reported broken: %+v", res)
	}
	if res.Verified != 5 {
		t.Errorf("verified count: got %d want 5", res.Verified)
	}
}

// Acceptance #2: tampering with one row's payload trips the
// verifier at exactly that chain_seq.
func TestChain_VerifyDetectsRowTampering(t *testing.T) {
	pool := newAuditPool(t)
	app := chain.NewAppender(pool)
	ver := chain.NewVerifier(pool)
	ctx := context.Background()

	for i := 0; i < 4; i++ {
		if _, err := app.Append(ctx, makeEvent(i)); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	// Tamper: change the action on chain_seq=2 (the 2nd row).
	if _, err := pool.Exec(ctx,
		`UPDATE audit_events SET action = 'iam.user.delete' WHERE tenant_id = $1 AND chain_seq = 2`,
		auditTestTenant,
	); err != nil {
		t.Fatalf("tamper: %v", err)
	}

	start := time.Now().UTC().Add(-time.Hour)
	end := time.Now().UTC().Add(time.Hour)
	res, err := ver.VerifyRange(ctx, auditTestTenant, start, end)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if res.IsClean() {
		t.Fatal("expected broken chain")
	}
	if res.Broken != 2 {
		t.Errorf("broken seq: got %d want 2", res.Broken)
	}
	if !strings.Contains(res.Reason, "chain_seq=2") {
		t.Errorf("reason missing seq: %q", res.Reason)
	}
}

// Tampering with prev_hash directly trips the prev_hash continuity
// check (a different code path from the row_hash recompute).
func TestChain_VerifyDetectsPrevHashTampering(t *testing.T) {
	pool := newAuditPool(t)
	app := chain.NewAppender(pool)
	ver := chain.NewVerifier(pool)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := app.Append(ctx, makeEvent(i)); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	// Replace chain_seq=2's prev_hash with a fake value.
	if _, err := pool.Exec(ctx,
		`UPDATE audit_events SET prev_hash = $2 WHERE tenant_id = $1 AND chain_seq = 2`,
		auditTestTenant, strings.Repeat("f", 64),
	); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	res, err := ver.VerifyRange(ctx, auditTestTenant,
		time.Now().UTC().Add(-time.Hour), time.Now().UTC().Add(time.Hour))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if res.IsClean() {
		t.Fatal("expected broken chain")
	}
	if res.Broken != 2 {
		t.Errorf("broken seq: got %d want 2", res.Broken)
	}
}

// VerifyRow on a clean row returns nil; a tampered row returns
// ErrChainBreak.
func TestChain_VerifyRow(t *testing.T) {
	pool := newAuditPool(t)
	app := chain.NewAppender(pool)
	ver := chain.NewVerifier(pool)
	ctx := context.Background()

	if _, err := app.Append(ctx, makeEvent(0)); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := ver.VerifyRow(ctx, auditTestTenant, 1); err != nil {
		t.Errorf("clean row: %v", err)
	}

	if _, err := pool.Exec(ctx,
		`UPDATE audit_events SET reason = 'tampered' WHERE tenant_id = $1 AND chain_seq = 1`,
		auditTestTenant,
	); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	if err := ver.VerifyRow(ctx, auditTestTenant, 1); err == nil {
		t.Error("expected ErrChainBreak after tamper")
	}
}

// Acceptance #1: the audit_writer role exists with the expected
// privileges. We don't actually CONNECT as that role here (the
// test connection runs as the migration owner); instead we
// inspect pg_roles + has_table_privilege.
func TestChain_AuditWriterRoleExistsWithGrants(t *testing.T) {
	pool := newAuditPool(t)
	ctx := context.Background()

	var exists bool
	if err := pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'audit_writer')`,
	).Scan(&exists); err != nil {
		t.Fatalf("role check: %v", err)
	}
	if !exists {
		t.Fatal("audit_writer role missing")
	}

	for _, priv := range []string{"INSERT", "SELECT", "UPDATE"} {
		var has bool
		if err := pool.QueryRow(ctx,
			`SELECT has_table_privilege('audit_writer', 'audit_events', $1)`, priv,
		).Scan(&has); err != nil {
			t.Fatalf("privilege check: %v", err)
		}
		if !has {
			t.Errorf("audit_writer missing %s on audit_events", priv)
		}
	}

	var readerExists bool
	if err := pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'audit_reader')`,
	).Scan(&readerExists); err != nil {
		t.Fatalf("reader check: %v", err)
	}
	if !readerExists {
		t.Fatal("audit_reader role missing")
	}
}

// Empty input → no rows → clean (no rows to verify).
func TestChain_VerifyEmptyRangeIsClean(t *testing.T) {
	pool := newAuditPool(t)
	ver := chain.NewVerifier(pool)
	res, err := ver.VerifyRange(context.Background(), auditTestTenant,
		time.Now().UTC().Add(-time.Hour), time.Now().UTC().Add(-time.Minute))
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !res.IsClean() {
		t.Errorf("empty range should be clean: %+v", res)
	}
	if res.Verified != 0 {
		t.Errorf("verified: %d", res.Verified)
	}
}
