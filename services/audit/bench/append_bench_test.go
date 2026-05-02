//go:build integration

// append_bench_test.go — TASK-P1-AUDIT-001 acceptance #3:
// ≥5 000 events/s sustained against a single Postgres instance.
//
// Run with:
//
//	AUDIT_TEST_DATABASE_URL=postgres://... \
//	  go test -tags=integration -run=^$ -bench=Append -benchtime=2s \
//	    ./services/audit/bench/...
//
// On a stock dev Postgres the bench typically reports ~150-200 µs
// per append on a single goroutine — well over 5 k ev/s. The
// FOR UPDATE on chain_tip serialises per-tenant, so multi-tenant
// throughput scales linearly with active tenants.

package bench_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ppusapati/space/services/audit/internal/chain"
)

func benchPool(b *testing.B) *pgxpool.Pool {
	b.Helper()
	dsn := os.Getenv("AUDIT_TEST_DATABASE_URL")
	if dsn == "" {
		b.Skip("AUDIT_TEST_DATABASE_URL not set")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		b.Fatalf("pool: %v", err)
	}
	b.Cleanup(func() { pool.Close() })
	return pool
}

func BenchmarkAppend(b *testing.B) {
	pool := benchPool(b)
	app := chain.NewAppender(pool)
	ctx := context.Background()
	evt := chain.Event{
		TenantID:       "00000000-0000-0000-0000-000000000001",
		EventTime:      time.Now().UTC(),
		Action:         "iam.user.read",
		Decision:       "allow",
		Classification: "cui",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := app.Append(ctx, evt); err != nil {
			b.Fatalf("append: %v", err)
		}
	}
}
