// Package migrate runs the cluster-wide platform migrations defined under
// infra/atlas/migrations/ at service boot.
//
// TASK-P0-DB-001 — design.md §5.1.
//
// Usage from a service entrypoint:
//
//	cfg := migrate.Config{DSN: cfg.Postgres.DSN, Logger: slog.Default()}
//	if err := migrate.EnsureUp(ctx, cfg); err != nil {
//	    return fmt.Errorf("platform migrations: %w", err)
//	}
//
// Behavior:
//   - Reads the embedded SQL files in lexicographic order
//     (0001_extensions.sql, 0002_retention_policies.sql, …).
//   - Records applied versions in chetana_schema_migrations to make
//     re-application a no-op.
//   - Files containing the "atlas:txmode none" directive are executed
//     with autocommit (CREATE EXTENSION timescaledb requires this).
//   - Other files run inside a single transaction per file.
//
// This runner is intentionally minimal — Atlas itself (the binary) is
// the source of truth in CI for migration linting (`atlas migrate lint`)
// and checksum verification (`atlas migrate hash`). At runtime services
// apply migrations themselves so we don't need to ship the Atlas binary
// in service container images.
package migrate

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MigrationsFS is the embedded migration directory. Tests may override it
// via OptionsWithFS.
//
// We embed the platform-wide migrations here so any service that imports
// this package picks up the same set without filesystem dependencies.
//
//go:embed migrations/*.sql
var defaultMigrationsFS embed.FS

// migrationsSubdir is the directory inside the embed FS holding *.sql.
const migrationsSubdir = "migrations"

// schemaMigrationsTable is the bookkeeping table the runner uses to track
// applied versions. We use a name distinct from atlas_schema_revisions so
// the runtime runner and the Atlas CLI do not interfere with each other —
// the runner is the runtime source of truth; the CLI is the dev/CI tool.
const schemaMigrationsTable = "chetana_schema_migrations"

// txModeNoneDirective marks a migration that must run with autocommit
// instead of inside a single transaction. Required by CREATE EXTENSION
// timescaledb on some Postgres builds.
const txModeNoneDirective = "atlas:txmode none"

// Config configures EnsureUp.
type Config struct {
	// DSN is the Postgres connection string. Required.
	DSN string

	// Logger is used for migration progress. Defaults to slog.Default().
	Logger *slog.Logger

	// Timeout caps the total migration duration. Defaults to 5 minutes.
	Timeout time.Duration

	// FS overrides the embedded migrations FS. Tests use this; production
	// callers leave it nil to use the embedded set.
	FS fs.FS
}

// EnsureUp opens a short-lived pool, applies any pending migrations from
// the embedded migrations directory (or cfg.FS when set), then closes the
// pool. It is safe to call concurrently from multiple service replicas —
// Postgres advisory locks serialise the apply.
func EnsureUp(ctx context.Context, cfg Config) error {
	if cfg.DSN == "" {
		return errors.New("migrate.EnsureUp: empty DSN")
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DSN)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer pool.Close()

	migrations, err := readMigrations(cfg.FS)
	if err != nil {
		return err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire conn: %w", err)
	}
	defer conn.Release()

	// Advisory lock — namespaced to a fixed app-defined ID (CRC32 of
	// "chetana_migrate"). Multiple replicas attempting EnsureUp in
	// parallel will serialise here.
	const lockID int64 = 0x6368746E // ascii "chtn"
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", lockID); err != nil {
		return fmt.Errorf("acquire advisory lock: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.Background(), "SELECT pg_advisory_unlock($1)", lockID)
	}()

	if err := ensureMigrationsTable(ctx, conn.Conn()); err != nil {
		return err
	}

	applied, err := loadApplied(ctx, conn.Conn())
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if existing, ok := applied[m.Version]; ok {
			if existing.Checksum != m.Checksum {
				return fmt.Errorf(
					"migration %s checksum drift: applied=%s embedded=%s",
					m.Version, existing.Checksum, m.Checksum,
				)
			}
			cfg.Logger.Debug("migration already applied", "version", m.Version)
			continue
		}

		cfg.Logger.Info("applying migration", "version", m.Version, "tx_mode", m.TxModeLabel())
		start := time.Now()
		if err := applyMigration(ctx, conn.Conn(), m); err != nil {
			return fmt.Errorf("apply %s: %w", m.Version, err)
		}
		if err := recordApplied(ctx, conn.Conn(), m, time.Since(start)); err != nil {
			return fmt.Errorf("record %s: %w", m.Version, err)
		}
		cfg.Logger.Info("migration applied", "version", m.Version, "duration", time.Since(start))
	}

	return nil
}

// migration represents one *.sql file from the migrations directory.
type migration struct {
	Version  string // filename without extension, e.g. "0001_extensions"
	SQL      string
	Checksum string // sha256 hex of SQL bytes
	TxNone   bool   // true when the file declares atlas:txmode none
}

func (m migration) TxModeLabel() string {
	if m.TxNone {
		return "autocommit"
	}
	return "transactional"
}

// readMigrations enumerates the embedded *.sql files in lexicographic
// order. cfgFS may be nil (use the package-level embedded FS).
func readMigrations(cfgFS fs.FS) ([]migration, error) {
	var src fs.FS = defaultMigrationsFS
	subdir := migrationsSubdir
	if cfgFS != nil {
		src = cfgFS
		subdir = "."
	}

	entries, err := fs.ReadDir(src, subdir)
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var out []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		// Skip Atlas housekeeping files (atlas.sum etc.).
		if strings.HasPrefix(e.Name(), "atlas.") {
			continue
		}
		path := e.Name()
		if subdir != "." {
			path = subdir + "/" + e.Name()
		}
		raw, err := fs.ReadFile(src, path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		sum := sha256.Sum256(raw)
		out = append(out, migration{
			Version:  strings.TrimSuffix(e.Name(), ".sql"),
			SQL:      string(raw),
			Checksum: hex.EncodeToString(sum[:]),
			TxNone:   strings.Contains(string(raw), txModeNoneDirective),
		})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Version < out[j].Version })
	return out, nil
}

// ensureMigrationsTable creates chetana_schema_migrations if absent.
func ensureMigrationsTable(ctx context.Context, conn *pgx.Conn) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS ` + schemaMigrationsTable + ` (
    version       text PRIMARY KEY,
    checksum      text NOT NULL,
    applied_at    timestamptz NOT NULL DEFAULT now(),
    duration_ms   bigint NOT NULL
)`
	_, err := conn.Exec(ctx, ddl)
	if err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}
	return nil
}

// appliedRow represents one row in chetana_schema_migrations.
type appliedRow struct {
	Checksum string
}

// loadApplied returns the set of applied migrations keyed by version.
func loadApplied(ctx context.Context, conn *pgx.Conn) (map[string]appliedRow, error) {
	rows, err := conn.Query(ctx, "SELECT version, checksum FROM "+schemaMigrationsTable)
	if err != nil {
		return nil, fmt.Errorf("query applied: %w", err)
	}
	defer rows.Close()

	out := map[string]appliedRow{}
	for rows.Next() {
		var v, c string
		if err := rows.Scan(&v, &c); err != nil {
			return nil, fmt.Errorf("scan applied: %w", err)
		}
		out[v] = appliedRow{Checksum: c}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied: %w", err)
	}
	return out, nil
}

// applyMigration runs the SQL either in a single transaction (default) or
// with autocommit (when atlas:txmode none directive is present).
func applyMigration(ctx context.Context, conn *pgx.Conn, m migration) error {
	if m.TxNone {
		// Autocommit. Run statement-by-statement to avoid implicit txn.
		// We do a coarse split on ';' at end-of-line — this is fine for
		// our handwritten CREATE EXTENSION migration; complex DO blocks
		// stay within transactional migrations.
		stmts := splitSimpleStatements(m.SQL)
		for _, s := range stmts {
			if strings.TrimSpace(s) == "" {
				continue
			}
			if _, err := conn.Exec(ctx, s); err != nil {
				return fmt.Errorf("exec %q: %w", firstLine(s), err)
			}
		}
		return nil
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, m.SQL); err != nil {
		return fmt.Errorf("exec migration: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// recordApplied writes the bookkeeping row.
func recordApplied(ctx context.Context, conn *pgx.Conn, m migration, dur time.Duration) error {
	_, err := conn.Exec(ctx,
		`INSERT INTO `+schemaMigrationsTable+` (version, checksum, duration_ms)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (version) DO UPDATE
		 SET checksum = EXCLUDED.checksum, duration_ms = EXCLUDED.duration_ms`,
		m.Version, m.Checksum, dur.Milliseconds())
	return err
}

// splitSimpleStatements splits on top-level semicolons followed by a
// newline. This is sufficient for autocommit migrations that contain only
// straight DDL (CREATE EXTENSION …); it is NOT safe for files containing
// DO blocks, dollar-quoted strings, or embedded ;. Such files MUST run
// transactionally (without atlas:txmode none).
func splitSimpleStatements(sql string) []string {
	var out []string
	var b strings.Builder
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		// Skip pure comment lines / Atlas directives.
		if strings.HasPrefix(trimmed, "--") {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
		if strings.HasSuffix(trimmed, ";") {
			out = append(out, b.String())
			b.Reset()
		}
	}
	if rem := strings.TrimSpace(b.String()); rem != "" {
		out = append(out, b.String())
	}
	return out
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}
