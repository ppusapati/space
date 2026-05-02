// Command audit is the chetana audit service.
//
// Phase 1: TASK-P1-AUDIT-001 ships the append-only hash-chain
// store + the `audit_writer`/`audit_reader` Postgres role split.
// TASK-P1-AUDIT-002 adds the search + signed-export + retention
// surface on top.
//
// REQ-FUNC-PLT-AUDIT-001..006 + design.md §4.2.

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"p9e.in/chetana/packages/observability/serverobs"
	"p9e.in/chetana/packages/region"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("audit failed", slog.Any("err", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg := loadConfig()
	logger.Info("audit starting",
		slog.String("http_addr", cfg.HTTPAddr),
		slog.String("metrics_addr", cfg.MetricsAddr),
		slog.String("version", cfg.Version),
		slog.String("git_sha", cfg.GitSHA),
	)

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 15*time.Second)
	pool, err := pgxpool.New(dbCtx, cfg.DatabaseDSN)
	dbCancel()
	if err != nil {
		return err
	}
	defer pool.Close()

	srv := serverobs.NewServer(
		serverobs.ServerConfig{Addr: cfg.HTTPAddr},
		serverobs.ObservabilityConfig{
			Build:       serverobs.BuildInfo{Version: cfg.Version, GitSHA: cfg.GitSHA},
			MetricsAddr: cfg.MetricsAddr,
			DepChecks: []serverobs.DepCheck{
				serverobs.PostgresCheck{Pool: pool},
			},
		},
	)

	// Connect RPC handler registration (Append + Verify + Search +
	// Export) lands once `audit.proto` regenerates (BSR auth
	// blocked by OQ-004). The internal/chain package is fully
	// functional today and is what those handlers will call.

	if err := pool.Ping(context.Background()); err != nil {
		logger.Warn("postgres ping failed at boot", slog.Any("err", err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info("audit stopped cleanly")
	return nil
}

type auditConfig struct {
	HTTPAddr    string
	MetricsAddr string
	DatabaseDSN string
	Version     string
	GitSHA      string
}

func loadConfig() auditConfig {
	return auditConfig{
		HTTPAddr:    getenvOr("HTTP_ADDR", ":8082"),
		MetricsAddr: getenvOr("METRICS_ADDR", ":9092"),
		DatabaseDSN: getenvOr("DATABASE_URL", region.PostgresDSN("audit")),
		Version:     getenvOr("CHETANA_VERSION", "v0.0.0-dev"),
		GitSHA:      getenvOr("CHETANA_GIT_SHA", "unknown"),
	}
}

func getenvOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
