//go:build integration

// runner_integration_test.go asserts that the platform-wide migrations
// under services/packages/db/migrate/migrations apply cleanly to a real
// Postgres+TimescaleDB+PostGIS instance and that re-applying them is a
// no-op (TASK-P0-DB-001 acceptance criteria #1, #2, #3).
//
// Run with:    go test -tags=integration ./db/migrate/...
//
// Driver:
//   The test reads CHETANA_TEST_DB_URL — when unset the test is skipped.
//   tools/db/seed-test.sh sets up a local TimescaleDB container; the CI
//   workflow runs the test against a service container with the same
//   image. This avoids pulling testcontainers-go into the module graph.
package migrate_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"p9e.in/chetana/packages/db/migrate"
)

const envDSN = "CHETANA_TEST_DB_URL"

// requireDSN returns the DSN or skips the test if not configured.
func requireDSN(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv(envDSN)
	if dsn == "" {
		t.Skipf("integration test: set %s to run (see tools/db/seed-test.sh)", envDSN)
	}
	return dsn
}

func TestEnsureUp_FreshDatabase_AppliesAllMigrations(t *testing.T) {
	dsn := requireDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resetMigrations(t, ctx, dsn)

	if err := migrate.EnsureUp(ctx, migrate.Config{DSN: dsn}); err != nil {
		t.Fatalf("EnsureUp: %v", err)
	}

	conn := mustConnect(t, ctx, dsn)
	defer conn.Close(ctx)

	t.Run("extensions installed", func(t *testing.T) {
		for _, ext := range []string{"timescaledb", "postgis", "pg_trgm", "pgcrypto"} {
			var present bool
			if err := conn.QueryRow(ctx,
				`SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = $1)`, ext,
			).Scan(&present); err != nil {
				t.Fatalf("query pg_extension %s: %v", ext, err)
			}
			if !present {
				t.Errorf("extension %q missing — acceptance criterion #2", ext)
			}
		}
	})

	t.Run("migrations bookkeeping", func(t *testing.T) {
		want := map[string]bool{"0001_extensions": true, "0002_retention_policies": true}
		rows, err := conn.Query(ctx,
			`SELECT version FROM chetana_schema_migrations ORDER BY version`)
		if err != nil {
			t.Fatalf("query bookkeeping table: %v", err)
		}
		defer rows.Close()
		got := map[string]bool{}
		for rows.Next() {
			var v string
			if err := rows.Scan(&v); err != nil {
				t.Fatalf("scan: %v", err)
			}
			got[v] = true
		}
		for v := range want {
			if !got[v] {
				t.Errorf("expected applied migration %q not present", v)
			}
		}
	})
}

func TestEnsureUp_SecondRun_IsNoop(t *testing.T) {
	dsn := requireDSN(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Make sure the first run completed.
	if err := migrate.EnsureUp(ctx, migrate.Config{DSN: dsn}); err != nil {
		t.Fatalf("first EnsureUp: %v", err)
	}

	// Capture the applied_at + duration_ms columns to assert the second
	// run does NOT update them (i.e. it's a true no-op, not a re-apply).
	conn := mustConnect(t, ctx, dsn)
	defer conn.Close(ctx)

	type rev struct {
		appliedAt time.Time
		duration  int64
	}
	loadRevs := func() map[string]rev {
		rows, err := conn.Query(ctx,
			`SELECT version, applied_at, duration_ms FROM chetana_schema_migrations`)
		if err != nil {
			t.Fatalf("load revs: %v", err)
		}
		defer rows.Close()
		out := map[string]rev{}
		for rows.Next() {
			var v string
			var r rev
			if err := rows.Scan(&v, &r.appliedAt, &r.duration); err != nil {
				t.Fatalf("scan rev: %v", err)
			}
			out[v] = r
		}
		return out
	}

	before := loadRevs()
	if len(before) == 0 {
		t.Fatalf("no migrations recorded after first EnsureUp; cannot assert idempotency")
	}

	// Acceptance criterion #1: re-apply is a no-op.
	if err := migrate.EnsureUp(ctx, migrate.Config{DSN: dsn}); err != nil {
		t.Fatalf("second EnsureUp: %v", err)
	}

	after := loadRevs()
	for v, b := range before {
		a, ok := after[v]
		if !ok {
			t.Errorf("migration %q disappeared after second run", v)
			continue
		}
		if !a.appliedAt.Equal(b.appliedAt) {
			t.Errorf("migration %q re-applied (applied_at changed): before=%s after=%s",
				v, b.appliedAt, a.appliedAt)
		}
	}
}

// resetMigrations drops the bookkeeping table and any platform-managed
// extensions so the test starts from a known-clean baseline.
//
// We deliberately do NOT drop the database — the test runner controls
// that. We only undo what migrate.EnsureUp creates.
func resetMigrations(t *testing.T, ctx context.Context, dsn string) {
	t.Helper()
	conn := mustConnect(t, ctx, dsn)
	defer conn.Close(ctx)

	stmts := []string{
		`DROP TABLE IF EXISTS chetana_schema_migrations`,
		// Best-effort cleanup of extensions; ignore errors since some
		// extensions may not be present on a truly-fresh DB.
		`DROP EXTENSION IF EXISTS timescaledb CASCADE`,
		`DROP EXTENSION IF EXISTS postgis CASCADE`,
		`DROP EXTENSION IF EXISTS pg_trgm CASCADE`,
	}
	for _, s := range stmts {
		if _, err := conn.Exec(ctx, s); err != nil {
			t.Logf("reset (ignored): %s -> %v", s, err)
		}
	}
}

func mustConnect(t *testing.T, ctx context.Context, dsn string) *pgx.Conn {
	t.Helper()
	c, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("pgx.Connect: %v", err)
	}
	return c
}
